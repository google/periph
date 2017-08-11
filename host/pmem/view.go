// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"unsafe"

	"periph.io/x/periph/host/fs"
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

// AsPOD implements Mem.
func (s *Slice) AsPOD(pp interface{}) error {
	if pp == nil {
		return wrapf("require Ptr, got nil")
	}
	vpp := reflect.ValueOf(pp)
	if elemSize, err := isPS(len(*s), vpp); err == nil {
		p := vpp.Elem()
		t := p.Type().Elem()
		if elemSize > len(*s) {
			return wrapf("can't map slice of struct %s (size %d) on [%d]byte", t, elemSize, len(*s))
		}
		nbElems := len(*s) / elemSize
		// Use casting black magic to set the internal slice headers.
		hdr := (*reflect.SliceHeader)(unsafe.Pointer(p.UnsafeAddr()))
		hdr.Data = ((*reflect.SliceHeader)(unsafe.Pointer(s))).Data
		hdr.Len = nbElems
		hdr.Cap = nbElems
		return nil
	}

	size, err := isPP(vpp)
	if err != nil {
		return err
	}
	p := vpp.Elem()
	t := p.Type().Elem()
	if size > len(*s) {
		return wrapf("can't map struct %s (size %d) on [%d]byte", t, size, len(*s))
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
	return nil, wrapf("/dev/gpiomem is not supported on this platform")
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
	return nil, wrapf("physical memory mapping is not supported on this platform")
}

// MapAsPOD is a leaky shorthand of calling Map(base, sizeof(v)) then AsPOD(v).
//
// There is no way to reclaim the memory map.
//
// A slice cannot be used, as it does not have inherent size. Use an aray
// instead.
func MapAsPOD(base uint64, i interface{}) error {
	// Automatically determine the necessary size. Because of this, slice of
	// unspecified length cannot be used here.
	if i == nil {
		return wrapf("require Ptr, got nil")
	}
	v := reflect.ValueOf(i)
	size, err := isPP(v)
	if err != nil {
		return err
	}
	m, err := Map(base, size)
	if err != nil {
		return err
	}
	return m.AsPOD(i)
}

//

// Keep a cache of open file handles instead of opening and closing repeatedly.
var (
	mu          sync.Mutex
	gpioMemErr  error
	gpioMemView *View
	devMem      fileIO
	devMemErr   error
	openFile    = openFileOrig
)

type fileIO interface {
	io.Closer
	io.Seeker
	io.Reader
	Fd() uintptr
}

func openFileOrig(path string, flag int) (fileIO, error) {
	f, err := fs.Open(path, flag)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// mapGPIOLinux is purely Raspbian specific.
func mapGPIOLinux() (*View, error) {
	mu.Lock()
	defer mu.Unlock()
	if gpioMemView == nil && gpioMemErr == nil {
		if f, err := openFile("/dev/gpiomem", os.O_RDWR|os.O_SYNC); err == nil {
			defer f.Close()
			if i, err := mmap(f.Fd(), 0, pageSize); err == nil {
				gpioMemView = &View{Slice: i, orig: i, phys: 0}
			} else {
				gpioMemErr = wrapf("failed to memory map in user space GPIO memory: %v", err)
			}
		} else {
			gpioMemErr = wrapf("failed to open GPIO memory: %v", err)
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
		return nil, wrapf("mapping at 0x%x failed: %v", base, err)
	}
	return &View{Slice: i[offset : offset+size], orig: i, phys: base + uint64(offset)}, nil
}

func openDevMemLinux() (fileIO, error) {
	mu.Lock()
	defer mu.Unlock()
	if devMem == nil && devMemErr == nil {
		if devMem, devMemErr = openFile("/dev/mem", os.O_RDWR|os.O_SYNC); devMemErr != nil {
			devMemErr = wrapf("failed to open physical memory: %v", devMemErr)
		}
	}
	return devMem, devMemErr
}

func isAcceptableInner(t reflect.Type) error {
	switch k := t.Kind(); k {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	case reflect.Array:
		return isAcceptableInner(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if err := isAcceptableInner(t.Field(i).Type); err != nil {
				return err
			}
		}
		return nil
	default:
		return wrapf("require Ptr to Ptr to a POD type, got Ptr to Ptr to %s", k)
	}
}

// isPP makes sure it is a pointer to a nil-pointer to something. It does
// sanity checks to reduce likelihood of a panic().
func isPP(pp reflect.Value) (int, error) {
	if k := pp.Kind(); k != reflect.Ptr {
		return 0, wrapf("require Ptr, got %s of %s", k, pp.Type().Name())
	}
	p := pp.Elem()
	if k := p.Kind(); k != reflect.Ptr {
		return 0, wrapf("require Ptr to Ptr, got %s", k)
	}
	if !p.IsNil() {
		return 0, wrapf("require Ptr to Ptr to be nil")
	}
	// p.Elem() can't be used since it's a nil pointer. Use the type instead.
	t := p.Type().Elem()
	if err := isAcceptableInner(t); err != nil {
		return 0, err
	}
	return int(t.Size()), nil
}

// isPS makes sure it is a pointer to a nil-slice of something. It does
// sanity checks to reduce likelihood of a panic().
func isPS(bufSize int, ps reflect.Value) (int, error) {
	if k := ps.Kind(); k != reflect.Ptr {
		return 0, wrapf("require Ptr, got %s of %s", k, ps.Type().Name())
	}
	s := ps.Elem()
	if k := s.Kind(); k != reflect.Slice {
		return 0, wrapf("require Ptr to Slice, got %s", k)
	}
	if !s.IsNil() {
		return 0, wrapf("require Ptr to Slice to be nil")
	}
	// s.Elem() can't be used since it's a nil slice. Use the type instead.
	t := s.Type().Elem()
	if err := isAcceptableInner(t); err != nil {
		return 0, err
	}
	return int(t.Size()), nil
}

func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("pmem: "+format, a...)
}
