// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"sync"
	"unsafe"
)

const pageSize = 4096

// Mem represents a section of memory that is usable by the DMA controller.
//
// Since this is physically allocated memory, that could potentially have been
// allocated in spite of OS consent, for example by asking the GPU directly, it
// is important to call Close() before process exit.
type Mem interface {
	io.Closer
	// Bytes returns the user space memory mapped buffer address as a slice of
	// bytes.
	//
	// It is the raw view of the memory from this process.
	Bytes() []byte
	// AsPOD initializes a pointer to a POD (plain old data) to point to the
	// memory mapped region.
	//
	// pp must be a pointer to:
	//
	// - pointer to a base size type (uint8, int64, float32, etc)
	// - struct
	// - array of the above
	// - slice of the above
	//
	// and the value must be nil. Returns an error otherwise.
	//
	// If a pointer to a slice is passed in, it is initialized to the length and
	// capacity set to the maximum number of elements this slice can represent.
	//
	// The pointer initialized points to the same address as Bytes().
	AsPOD(pp interface{}) error
	// PhysAddr is the physical address. It can be either 32 bits or 64 bits,
	// depending on the bitness of the OS kernel, not on the user mode build,
	// e.g. you could have compiled on a 32 bits Go toolchain but running on a
	// 64 bits kernel.
	PhysAddr() uint64
}

// MemAlloc represents contiguous physically locked memory that was allocated.
//
// The memory is mapped in user space.
//
// MemAlloc implements Mem.
type MemAlloc struct {
	View
}

// Close unmaps the physical memory allocation.
func (m *MemAlloc) Close() error {
	if err := munlock(m.orig); err != nil {
		return err
	}
	return munmap(m.orig)
}

// Alloc allocates a continuous chunk of physical memory.
//
// Size must be rounded to 4Kb. Allocations of 4Kb will normally succeed.
// Allocations larger than 64Kb will likely fail due to kernel memory
// fragmentation; rebooting the host or reducing the number of running programs
// may help.
//
// The allocated memory is uncached.
func Alloc(size int) (*MemAlloc, error) {
	if size == 0 || size&(pageSize-1) != 0 {
		return nil, wrapf("allocated memory must be rounded to %d bytes", pageSize)
	}
	if isLinux && !isWSL() {
		return allocLinux(size)
	}
	return nil, wrapf("memory allocation is not supported on this platform")
}

//

var (
	wslOnce    sync.Once
	isWSLValue bool
)

// uallocMemLocked allocates user space memory and requests the OS to have the
// chunk to be locked into physical memory.
func uallocMemLocked(size int) ([]byte, error) {
	// It is important to write to the memory so it is forced to be present.
	b, err := uallocMem(size)
	if err == nil {
		for i := range b {
			b[i] = 0
		}
		if err := mlock(b); err != nil {
			// Ignore the unmap error.
			_ = munmap(b)
			return nil, wrapf("locking %d bytes failed: %v", size, err)
		}
	}
	return b, err
}

// allocLinux allocates physical memory and returns a user view to it.
func allocLinux(size int) (*MemAlloc, error) {
	// TODO(maruel): Implement the "shotgun approach". Allocate a ton of 4Kb
	// pages and lock them. Then look at their physical pages and only keep the
	// one useful. Then create a linear mapping in memory to simplify the user
	// mode with a single linear user space virtual address but keep the
	// individual page alive with their initial allocation. When done release
	// each individual page.
	if size > pageSize {
		return nil, wrapf("large allocation is not yet implemented")
	}
	// First allocate a chunk of user space memory.
	b, err := uallocMemLocked(size)
	if err != nil {
		return nil, err
	}
	pages := make([]uint64, (size+pageSize-1)/pageSize)
	// Figure out the physical memory addresses.
	for i := range pages {
		pages[i], err = virtToPhys(toRaw(b[pageSize*i:]))
		if err != nil {
			return nil, err
		}
		if pages[i] == 0 {
			return nil, wrapf("failed to read page %d", i)
		}
	}
	for i := 1; i < len(pages); i++ {
		// Fail if the memory is not contiguous.
		if pages[i] != pages[i-1]+pageSize {
			return nil, wrapf("failed to allocate %d bytes of continugous physical memory; page %d =0x%x; page %d=0x%x", size, i, pages[i], i-1, pages[i-1])
		}
	}

	return &MemAlloc{View{Slice: b, phys: pages[0], orig: b}}, nil
}

// virtToPhys returns the physical memory address backing a virtual
// memory address.
func virtToPhys(virt uintptr) (uint64, error) {
	physPage, err := ReadPageMap(virt)
	if err != nil {
		return 0, err
	}
	if physPage&(1<<63) == 0 {
		// If high bit is not set, the page doesn't exist.
		return 0, wrapf("0x%08x has no physical address", virt)
	}
	// Strip flags. See linux documentation on kernel.org for more details.
	physPage &^= 0x1FF << 55
	return physPage * pageSize, nil
}

func toRaw(b []byte) uintptr {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&b))
	return header.Data
}

// isWSL returns true if running under Windows Subsystem for Linux.
func isWSL() bool {
	wslOnce.Do(func() {
		if c, err := ioutil.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
			isWSLValue = bytes.Contains(c, []byte("Microsoft"))
		}
	})
	return isWSLValue
}

var _ Mem = &MemAlloc{}
