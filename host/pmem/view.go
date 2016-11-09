// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"syscall"
	"unsafe"
)

// View represents a view of physical memory memory mapped into user space.
//
// It is usually used to map CPU registers into user space, usually I/O
// registers and the likes.
//
// It is not required to call Close(), the kernel will clean up on process
// shutdown.
type View struct {
	base   []uint8 // base is always rounded to 4096 page.
	offset int     // offset in the 4096 page.
}

// Close unmaps the memory from the user address space.
//
// This is done naturally by the OS on process teardown (when the process
// exits) so this is not a hard requirement to call this function.
func (v *View) Close() error {
	return syscall.Munmap(v.base)
}

// Uint8 returns the memory map as a slice of uint8.
func (v *View) Uint8() []uint8 {
	return v.base[v.offset:]
}

// Uint32 returns the memory map as a slice of uint32.
func (v *View) Uint32() []uint32 {
	return unsafeRemap(v.base[v.offset:])
}

// Struct initializes a pointer to a struct to point to the memory mapped
// region.
//
// i must be a pointer to a pointer to a struct and the pointer to struct must
// be nil. panics otherwise.
func (v *View) Struct(i unsafe.Pointer) {
	// This looks dangerous but it works. If someone knows how to reformat the
	// following code to not trigger a go vet warning, please send a PR.
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&v.base))
	dest := (*int)(unsafe.Pointer(header.Data + uintptr(v.offset)))
	p := (**int)(i)
	if *p != nil {
		panic(*p)
	}
	*p = dest
}

// MapGPIO returns a CPU specific memory mapping of the CPU I/O registers using
// /dev/gpiomem.
//
// At the moment, /dev/gpiomem is only supported on Raspbian Jessie via a
// specific kernel driver.
func MapGPIO() (*View, error) {
	if isLinux {
		return mapGPIOLinux()
	}
	return nil, errors.New("pmem: /dev/gpiomem is not support on this platform")
}

// Map returns a memory mapped view of arbitrary physical memory range using OS
// provided functionality.
//
// Maps size of memory, rounded on a 4kb window.
//
// This function is dangerous and should be used wisely. It normally requires
// super privileges (root). On Linux, it leverages /dev/mem.
func Map(base uint64, size int) (*View, error) {
	if isLinux {
		return mapLinux(base, size)
	}
	return nil, errors.New("pmem: /dev/mem is not supported on this platform")
}

//

// Keep a cache of open file handles instead of opening and closing repeatedly.
var (
	mu          sync.Mutex
	gpioMemErr  error
	gpioMemView *View
	devMem      *os.File
	devMemErr   error
)

// mapGPIOLinux is purely Raspbian specific.
func mapGPIOLinux() (*View, error) {
	mu.Lock()
	defer mu.Unlock()
	if gpioMemView == nil && gpioMemErr == nil {
		if f, err := os.OpenFile("/dev/gpiomem", os.O_RDWR|os.O_SYNC, 0); err == nil {
			defer f.Close()
			if i, err := syscall.Mmap(int(f.Fd()), 0, 4096, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED); err == nil {
				gpioMemView = &View{base: i}
			} else {
				gpioMemErr = err
			}
		} else {
			gpioMemErr = err
		}
	}
	return gpioMemView, gpioMemErr
}

// mapLinux leverages /dev/mem to map a view of physical memory.
func mapLinux(base uint64, size int) (*View, error) {
	f, err := openDevMemLinux()
	if err != nil {
		return nil, err
	}
	// Align base and size at 4Kb.
	offset := int(base & 0xFFF)
	i, err := syscall.Mmap(
		int(f.Fd()),
		int64(base&^0xFFF),
		(size+offset+0xFFF)&^0xFFF,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("physview: mapping at 0x%x failed: %v", base, err)
	}
	return &View{i, int(offset)}, nil
}

func openDevMemLinux() (*os.File, error) {
	mu.Lock()
	defer mu.Unlock()
	if devMem == nil && devMemErr == nil {
		devMem, devMemErr = os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0)
	}
	return devMem, devMemErr
}

func unsafeRemap(i []byte) []uint32 {
	// I/O needs to happen as 32 bits operation, so remap accordingly.
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&i))
	header.Len /= 4
	header.Cap /= 4
	return *(*[]uint32)(unsafe.Pointer(&header))
}
