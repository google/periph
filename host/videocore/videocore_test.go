// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package videocore

import (
	"errors"
	"log"
	"testing"

	"periph.io/x/periph/host/pmem"
)

func ExampleAlloc() {
	// Allocates physical memory on a Broadcom CPU by leveraging the GPU.
	// This memory can be leveraged to do DMA operations.
	m, err := Alloc(64536)
	if err != nil {
		log.Fatal(err)
	}
	// Use m
	m.Close()
}

//

func TestClose(t *testing.T) {
	defer reset()
	mailbox = &dummy{}
	m := Mem{View: &pmem.View{}}
	if m.Close() == nil {
		t.Fatal("can't close uninitalized pmem.View")
	}
}

func TestAlloc_fail(t *testing.T) {
	defer reset()
	if m, err := Alloc(0); m != nil || err == nil {
		t.Fatal("can't alloc 0 bytes")
	}
	if m, err := Alloc(1); m != nil || err == nil {
		t.Fatal("can't alloc non 4096 bytes increments")
	}
	mailboxErr = errors.New("error")
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("mailboxErr is not nil")
	}
	mailboxErr = nil
	mailbox = &dummy{}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("can't map arbitrary physical pages")
	}
	mailbox = &playback{}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("mailbox failed")
	}
	mailbox = &playback{
		reply: []uint32{0},
	}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("mailbox failed")
	}
	mailbox = &playback{
		reply: []uint32{failReply},
	}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("mailbox failed")
	}
	mailbox = &playback{
		reply: []uint32{failReplyLen},
	}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("mailbox failed")
	}
	mailbox = &playback{
		reply: []uint32{1, 0},
	}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("mailbox failed")
	}
	mailbox = &playback{
		reply: []uint32{1, failReply},
	}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("mailbox failed")
	}
}

func TestOpenMailbox(t *testing.T) {
	defer reset()
	// It is expected to fail unless running as root.
	if err := openMailbox(); err != nil {
		if mailboxErr != err {
			t.Fatal("error is different")
		}
		if mailbox != nil {
			t.Fatal("mailbox should be nil")
		}
	} else {
		if mailboxErr != nil {
			t.Fatal("should have error'ed")
		}
		if mailbox == nil {
			t.Fatal("mailbox should not be nil")
		}
	}
	// No-op in any case.
	openMailbox()
}

func TestSmokeTest(t *testing.T) {
	defer reset()
	mailbox = &dummy{}
	if err := smokeTest(); err != nil {
		t.Fatal(err)
	}
}

func TestGenPacket(t *testing.T) {
	defer reset()
	actual := genPacket(10, 12, 1, 2, 3)
	expected := []uint32{0x24, 0x0, 0xa, 0xc, 0xc, 0x1, 0x2, 0x3, 0x0}
	if !uint32Equals(actual, expected) {
		t.Fatal(actual)
	}
	actual = genPacket(10, 12, 1, 2)
	expected = []uint32{0x24, 0x0, 0xa, 0x8, 0xc, 0x1, 0x2, 0x0, 0x0}
	if !uint32Equals(actual, expected) {
		t.Fatal(actual)
	}
}

func TestIOCTL(t *testing.T) {
	// Sends an invalid ioctl to stdin. Not a big deal.
	m := messageBox{}
	if m.sendMessage([]uint32{0}) == nil {
		t.Fatal("can't send arbitrary ioctl to stdin")
	}
}

//

type dummy struct{}

func (d *dummy) sendMessage(b []uint32) error {
	b[1] = mbReply
	b[4] = mbReply | 4
	return nil
}

type playback struct {
	reply []uint32
	count int
}

const failReply uint32 = 0xFFFFFFFE
const failReplyLen uint32 = 0xFFFFFFFF

func (p *playback) sendMessage(b []uint32) error {
	if len(p.reply) <= p.count {
		return errors.New("exceeded count")
	}
	b[5] = p.reply[p.count]
	if b[5] != failReply {
		b[1] = mbReply
	}
	if b[5] != failReplyLen {
		b[4] = mbReply | 4
	}
	p.count++
	return nil
}

func uint32Equals(a []uint32, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	if mailbox != nil {
		if m, ok := mailbox.(*messageBox); ok {
			m.f.Close()
		}
	}
	mailbox = nil
	mailboxErr = nil
}
