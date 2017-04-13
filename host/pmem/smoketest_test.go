// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"errors"
	"testing"
	"unsafe"
)

func TestSmokeTest_fail(t *testing.T) {
	count := 0
	alloc := func(size int) (Mem, error) {
		if count == 0 {
			return nil, errors.New("oops")
		}
		count--
		return allocRam(size)
	}

	if CopyTest(1024, 1, alloc, copyOk) == nil {
		t.Fatal("first alloc failed")
	}
	count = 1
	if CopyTest(1024, 1, alloc, copyOk) == nil {
		t.Fatal("second alloc failed")
	}

	copyFail := func(d, s uint64) error {
		return errors.New("oops")
	}
	if CopyTest(1024, 1, allocRam, copyFail) == nil {
		t.Fatal("copy failed")
	}

	copyNop := func(d, s uint64) error {
		return nil
	}
	if CopyTest(1024, 1, allocRam, copyNop) == nil {
		t.Fatal("no copy")
	}

	copyPartial := func(d, s uint64) error {
		return copyRam(d, s, 1024, 2)
	}
	if CopyTest(1024, 1, allocRam, copyPartial) == nil {
		t.Fatal("copy corrupted")
	}

	copyHdr := func(d, s uint64) error {
		toSlice(d)[0] = 0
		return nil
	}
	if CopyTest(1024, 1, allocRam, copyHdr) == nil {
		t.Fatal("header corrupted")
	}

	copyFtr := func(d, s uint64) error {
		toSlice(d)[1023] = 0
		return copyRam(d, s, 1024, 1)
	}
	if CopyTest(1024, 1, allocRam, copyFtr) == nil {
		t.Fatal("footer corrupted")
	}

	copyOffset := func(d, s uint64) error {
		copyRam(d, s, 1024, 1)
		toSlice(d)[3] = 0
		return nil
	}
	if CopyTest(1024, 1, allocRam, copyOffset) == nil {
		t.Fatal("copy corrupted")
	}
}

func TestSmokeTest(t *testing.T) {
	// Successfully copy the memory.
	if err := CopyTest(1024, 1, allocRam, copyOk); err != nil {
		t.Fatal(err)
	}
}

// allocRam allocates memory and fake it is physical memory.
func allocRam(size int) (Mem, error) {
	p := make([]byte, size)
	return &MemAlloc{
		View: View{
			Slice: Slice(p),
			orig:  p,
			phys:  uint64(uintptr(unsafe.Pointer(&p))),
		},
	}, nil
}

func copyOk(d, s uint64) error {
	return copyRam(d, s, 1024, 1)
}

// copyRam copies the memory.
func copyRam(pDst, pSrc uint64, size, hole int) error {
	dst := toSlice(pDst)
	src := toSlice(pSrc)
	copy(dst[hole:size-hole], src)
	return nil
}

func toSlice(p uint64) []byte {
	return *(*[]byte)(unsafe.Pointer(uintptr(p)))
}
