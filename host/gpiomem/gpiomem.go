// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiomem

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"syscall"
	"unsafe"
)

// Mem represents memory mapped CPU registers (usually I/O).
type Mem struct {
	base   []uint8
	offset int
}

// Uint32 returns the memory map as a slice of uint32.
func (m *Mem) Uint32() []uint32 {
	return unsafeRemap(m.base[m.offset:])
}

// Struct initializes a point to a struct to point to the memory mapped region.
//
// i must be a pointer to a pointer to a struct and the pointer to struct must
// be nil. panics otherwise.
func (m *Mem) Struct(i unsafe.Pointer) {
	// This looks dangerous but it works. If someone knows how to reformat the
	// following code to not trigger a go vet warning, please send a PR.
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&m.base))
	dest := (*int)(unsafe.Pointer(header.Data + uintptr(m.offset)))
	v := (**int)(i)
	if *v != nil {
		panic(*v)
	}
	*v = dest
}

// Close unmaps the I/O registers.
func (m *Mem) Close() error {
	return syscall.Munmap(m.base)
}

// OpenGPIO returns a CPU specific memory mapping of the CPU I/O registers using
// /dev/gpiomem.
//
// /dev/gpiomem is only supported on Raspbian Jessie via a specific kernel
// driver.
func OpenGPIO() (*Mem, error) {
	if isLinux {
		return openGPIOLinux()
	}
	return nil, errors.New("/dev/gpiomem is not support on this platform")
}

// OpenMem returns a memory mapped view of arbitrary kernel memory range using
// /dev/mem.
//
// Maps 4kb of memory, rounded on a 4kb window.
//
// This function is dangerous and should be used wisely.
func OpenMem(base uint64) (*Mem, error) {
	if isLinux {
		return openMemLinux(base)
	}
	return nil, errors.New("/dev/mem is not support on this platform")
}

//

func openGPIOLinux() (*Mem, error) {
	f, err := os.OpenFile("/dev/gpiomem", os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// TODO(maruel): Map PWM, CLK, PADS, TIMER for more functionality.
	i, err := syscall.Mmap(int(f.Fd()), 0, 4096, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return &Mem{i, 0}, nil
}

func openMemLinux(base uint64) (*Mem, error) {
	f, err := os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Align at 4Kb then offset the returned uint32 array.
	i, err := syscall.Mmap(int(f.Fd()), int64(base&^0xFFF), 4096, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("gpiomem: mapping at 0x%x failed: %v", base, err)
	}
	return &Mem{i, int(base & 0xFFF)}, nil
}

func unsafeRemap(i []byte) []uint32 {
	// I/O needs to happen as 32 bits operation, so remap accordingly.
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&i))
	header.Len /= 4
	header.Cap /= 4
	return *(*[]uint32)(unsafe.Pointer(&header))
}
