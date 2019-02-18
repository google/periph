// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package videocore interacts with the VideoCore GPU found on bcm283x.
//
// This package shouldn't be used directly, it is used by bcm283x's DMA
// implementation.
//
// Datasheet
//
// While not an actual datasheet, this is the closest to actual formal
// documentation:
// https://github.com/raspberrypi/firmware/wiki/Mailbox-property-interface
package videocore

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"periph.io/x/periph/host/fs"
	"periph.io/x/periph/host/pmem"
)

// Mem represents contiguous physically locked memory that was allocated by
// VideoCore.
//
// The memory is mapped in user space.
type Mem struct {
	*pmem.View
	handle uint32
}

// Close unmaps the physical memory allocation.
//
// It is important to call this function otherwise the memory is locked until
// the host reboots.
func (m *Mem) Close() error {
	if err := m.View.Close(); err != nil {
		return wrapf("failed to close physical view: %v", err)
	}
	if _, err := mailboxTx32(mbUnlockMemory, m.handle); err != nil {
		return wrapf("failed to unlock memory: %v", err)
	}
	if _, err := mailboxTx32(mbReleaseMemory, m.handle); err != nil {
		return wrapf("failed to release memory: %v", err)
	}
	return nil
}

// Alloc allocates a continuous chunk of physical memory for use with DMA
// controller.
//
// Size must be rounded to 4Kb.
func Alloc(size int) (*Mem, error) {
	if size <= 0 {
		return nil, wrapf("memory size must be > 0; got %d", size)
	}
	if size&0xFFF != 0 {
		return nil, wrapf("memory size must be rounded to 4096 pages; got %d", size)
	}
	if err := openMailbox(); err != nil {
		return nil, wrapf("failed to open the mailbox to the GPU: %v", err)
	}
	// Size, Alignment, Flags; returns an opaque handle to be used to release the
	// memory.
	handle, err := mailboxTx32(mbAllocateMemory, uint32(size), 4096, flagDirect)
	if err != nil {
		return nil, wrapf("failed request to allocate memory: %v", err)
	}
	if handle == 0 {
		return nil, wrapf("failed to allocate %d bytes", size)
	}
	// Lock the memory to retrieve a physical memory address.
	p, err := mailboxTx32(mbLockMemory, handle)
	if err != nil {
		return nil, wrapf("failed request to lock memory: %v", err)
	}
	if p == 0 {
		return nil, wrapf("failed to lock memory")
	}
	b, err := pmem.Map(uint64(p&^0xC0000000), size)
	if err != nil {
		return nil, wrapf("failed to memory map phyisical pages: %v", err)
	}
	return &Mem{View: b, handle: handle}, nil
}

//

var (
	mu         sync.Mutex
	mailbox    messager
	mailboxErr error

	mbIoctl = fs.IOWR('d', 0, uint(unsafe.Sizeof(new(byte))))
)

const (
	// All of these return anything but zero (‽)
	mbFirmwareVersion = 0x1     // 0, 4
	mbBoardModel      = 0x10001 // 0, 4
	mbBoardRevision   = 0x10002 // 0, 4
	mbBoardMAC        = 0x10003 // 0, 6
	mbBoardSerial     = 0x10004 // 0, 8
	mbARMMemory       = 0x10005 // 0, 8
	mbVCMemory        = 0x10006 // 0, 8
	mbClocks          = 0x10007 // 0, variable
	// These work:
	mbAllocateMemory = 0x3000C    // 12, 4
	mbLockMemory     = 0x3000D    // 4, 4
	mbUnlockMemory   = 0x3000E    // 4, 4
	mbReleaseMemory  = 0x3000F    // 4, 4
	mbReply          = 0x80000000 // High bit means a reply

	flagDiscardable     = 1 << 0                    // Can be resized to 0 at any time. Use for cached data.
	flagNormal          = 0 << 2                    // Normal allocating alias. Don't use from ARM.
	flagDirect          = 1 << 2                    // 0xCxxxxxxx Uncached
	flagCoherent        = 2 << 2                    // 0x8xxxxxxx Non-allocating in L2 but coherent
	flagL1Nonallocating = flagDirect | flagCoherent // Allocating in L2
	flagZero            = 1 << 4                    // Initialise buffer to all zeros
	flagNoInit          = 1 << 5                    // Don't initialise (default is initialise to all ones
	flagHintPermalock   = 1 << 6                    // Likely to be locked for long periods of time
)

type messager interface {
	sendMessage(b []uint32) error
}

type messageBox struct {
	f *fs.File
}

func (m *messageBox) sendMessage(b []uint32) error {
	return m.f.Ioctl(mbIoctl, uintptr(unsafe.Pointer(&b[0])))
}

func openMailbox() error {
	mu.Lock()
	defer mu.Unlock()
	if mailbox != nil || mailboxErr != nil {
		return mailboxErr
	}
	f, err := fs.Open("/dev/vcio", os.O_RDWR|os.O_SYNC)
	mailboxErr = err
	if err == nil {
		mailbox = &messageBox{f}
		mailboxErr = smokeTest()
	}
	return mailboxErr
}

// genPacket creates a message to be sent to the GPU via the "mailbox".
//
// The message must be 16-byte aligned because only the upper 28 bits are
// passed; the lower bits are used to select the channel.
func genPacket(cmd uint32, replyLen uint32, args ...uint32) []uint32 {
	p := make([]uint32, 48)
	offset := uintptr(unsafe.Pointer(&p[0])) & 15
	b := p[16-offset : 32+16-offset]
	max := uint32(len(args) * 4)
	if replyLen > max {
		max = replyLen
	}
	max = ((max + 3) / 4) * 4
	// size + zero + cmd + in + out + <max> + zero
	b[0] = uint32(6*4) + max     // message total length in bytes, including trailing zero
	b[2] = cmd                   //
	b[3] = uint32(len(args)) * 4 // inputs length in bytes
	b[4] = replyLen              // outputs length in bytes
	copy(b[5:], args)
	return b[:6+max/4]
}

func sendPacket(b []uint32) error {
	if err := mailbox.sendMessage(b); err != nil {
		return fmt.Errorf("failed to send IOCTL: %v", err)
	}
	if b[1] != mbReply {
		// 0x80000001 means partial response.
		return fmt.Errorf("got unexpected reply bit 0x%08x", b[1])
	}
	return nil
}

func mailboxTx32(cmd uint32, args ...uint32) (uint32, error) {
	b := genPacket(cmd, 4, args...)
	if err := sendPacket(b); err != nil {
		return 0, err
	}
	if b[4] != mbReply|4 {
		return 0, fmt.Errorf("got unexpected reply size 0x%08x", b[4])
	}
	return b[5], nil
}

/*
// mailboxTx is the generic version of mailboxTx32. It is not currently needed.
func mailboxTx(cmd uint32, reply []byte, args ...uint32) error {
	b := genPacket(cmd, uint32(len(reply)), args...)
	if err := sendPacket(b); err != nil {
		return err
	}
	rep := b[4]
	if rep&mbReply == 0 {
		return fmt.Errorf("got unexpected reply size 0x%08x", b[4])
	}
	rep &^= mbReply
	if rep == 0 || rep > uint32(len(reply)) {
		return fmt.Errorf("got unexpected reply size 0x%08x", b[4])
	}
	return nil
}
*/

func smokeTest() error {
	// It returns 0 on a RPi3 but don't assert this in case the VC firmware gets
	// updated.
	_, err := mailboxTx32(mbFirmwareVersion)
	return err
}

func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("videocore: "+format, a...)
}

var _ pmem.Mem = &Mem{}
