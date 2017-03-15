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
	"unsafe"
)

// Slice can be transparently viewed as []byte, []uint32 or a struct.
type Slice []byte

// Uint32 returns a view of the byte slice as a []uint32.
func (s *Slice) Uint32() []uint32 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(s))
	header.Len /= 4
	header.Cap /= 4
	return *(*[]uint32)(unsafe.Pointer(&header))
}

// Bytes implements Mem.
func (s *Slice) Bytes() []byte {
	return *s
}

// Struct implements Mem.
func (s *Slice) Struct(pp reflect.Value) error {
	size, err := isPP(pp)
	if err != nil {
		return err
	}
	p := pp.Elem()
	t := p.Type().Elem()
	if size > len(*s) {
		return fmt.Errorf("pmem: can't map struct %s (size %d) on [%d]byte", t, size, len(*s))
	}
	// Use casting black magic to read the internal slice headers.
	dest := unsafe.Pointer(((*reflect.SliceHeader)(unsafe.Pointer(s))).Data)
	// Use reflection black magic to write to the original pointer.
	p.Set(reflect.NewAt(t, dest))
	return nil
}

// View represents a view of physical memory memory mapped into user space.
//
// It is usually used to map CPU registers into user space, usually I/O
// registers and the likes.
//
// It is not required to call Close(), the kernel will clean up on process
// shutdown.
type View struct {
	Slice
	orig []uint8 // Reference rounded to the lowest 4Kb page containing Slice.
	phys uint64  // physical address of the base of Slice.
}

// Close unmaps the memory from the user address space.
//
// This is done naturally by the OS on process teardown (when the process
// exits) so this is not a hard requirement to call this function.
func (v *View) Close() error {
	return munmap(v.orig)
}

// PhysAddr implements Mem.
func (v *View) PhysAddr() uint64 {
	return v.phys
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

// MapStruct is a shorthand to call Map(base, sizeof(v)) then Struct(v).
func MapStruct(base uint64, v reflect.Value) error {
	size, err := isPP(v)
	if err != nil {
		return err
	}
	m, err := Map(base, size)
	if err != nil {
		return err
	}
	return m.Struct(v)
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
			if i, err := mmap(f.Fd(), 0, pageSize); err == nil {
				gpioMemView = &View{Slice: i, orig: i, phys: 0}
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
	i, err := mmap(f.Fd(), int64(base&^0xFFF), (size+offset+0xFFF)&^0xFFF)
	if err != nil {
		return nil, fmt.Errorf("pmem: mapping at 0x%x failed: %v", base, err)
	}
	return &View{Slice: i[offset : offset+size], orig: i, phys: base + uint64(offset)}, nil
}

func openDevMemLinux() (*os.File, error) {
	mu.Lock()
	defer mu.Unlock()
	if devMem == nil && devMemErr == nil {
		devMem, devMemErr = os.OpenFile("/dev/mem", os.O_RDWR|os.O_SYNC, 0)
	}
	return devMem, devMemErr
}

// isPP makes sure it is a pointer to a nil-pointer to something. It does
// sanity checks to reduce likelihood of a panic().
func isPP(pp reflect.Value) (int, error) {
	if k := pp.Kind(); k != reflect.Ptr {
		return 0, fmt.Errorf("pmem: require Ptr, got %s", k)
	}
	if pp.IsNil() {
		return 0, errors.New("pmem: require Ptr to be valid")
	}
	p := pp.Elem()
	if k := p.Kind(); k != reflect.Ptr {
		return 0, fmt.Errorf("pmem: require Ptr to Ptr, got %s", k)
	}
	if !p.IsNil() {
		return 0, errors.New("pmem: require Ptr to Ptr to be nil")
	}
	// p.Elem() can't be used since it's a nil pointer. Use the type instead.
	t := p.Type().Elem()
	if k := t.Kind(); k != reflect.Struct && k != reflect.Array {
		return 0, fmt.Errorf("pmem: require Ptr to Ptr to a struct or an array, got Ptr to Ptr to %d", k)
	}
	return int(t.Size()), nil
}
