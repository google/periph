// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"errors"
	"reflect"
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
		return allocRAM(size)
	}

	if TestCopy(1024, 1, alloc, copyOk) == nil {
		t.Fatal("first alloc failed")
	}
	count = 1
	if TestCopy(1024, 1, alloc, copyOk) == nil {
		t.Fatal("second alloc failed")
	}

	copyFail := func(d, s uint64) error {
		return errors.New("oops")
	}
	if TestCopy(1024, 1, allocRAM, copyFail) == nil {
		t.Fatal("copy failed")
	}

	copyNop := func(d, s uint64) error {
		return nil
	}
	if TestCopy(1024, 1, allocRAM, copyNop) == nil {
		t.Fatal("no copy")
	}

	copyPartial := func(d, s uint64) error {
		return copyRAM(d, s, 1024, 2)
	}
	if TestCopy(1024, 1, allocRAM, copyPartial) == nil {
		t.Fatal("copy corrupted")
	}

	copyHdr := func(d, s uint64) error {
		toSlice(d, int(s))[0] = 0
		return nil
	}
	if TestCopy(1024, 1, allocRAM, copyHdr) == nil {
		t.Fatal("header corrupted")
	}

	copyFtr := func(d, s uint64) error {
		toSlice(d, int(s))[1023] = 0
		return copyRAM(d, s, 1024, 1)
	}
	if TestCopy(1024, 1, allocRAM, copyFtr) == nil {
		t.Fatal("footer corrupted")
	}

	copyOffset := func(d, s uint64) error {
		if err := copyRAM(d, s, 1024, 1); err != nil {
			return err
		}
		toSlice(d, int(s))[3] = 0
		return nil
	}
	if TestCopy(1024, 1, allocRAM, copyOffset) == nil {
		t.Fatal("copy corrupted")
	}
}

func TestSmokeTest(t *testing.T) {
	// Successfully copy the memory.
	if err := TestCopy(1024, 1, allocRAM, copyOk); err != nil {
		t.Fatal(err)
	}
}

// allocRAM allocates memory and fake it is physical memory.
func allocRAM(size int) (Mem, error) {
	p := make([]byte, size)
	return &MemAlloc{
		View: View{
			Slice: Slice(p),
			orig:  p,
			phys:  uint64(uintptr(unsafe.Pointer(&p[0]))),
		},
	}, nil
}

func copyOk(d, s uint64) error {
	return copyRAM(d, s, 1024, 1)
}

// copyRAM copies the memory.
func copyRAM(pDst, pSrc uint64, size, hole int) error {
	dst := toSlice(pDst, size)
	src := toSlice(pSrc, size)
	copy(dst[hole:size-hole], src)
	return nil
}

func toSlice(p uint64, size int) []byte {
	h := &reflect.SliceHeader{
		Data: uintptr(p),
		Len:  size,
		Cap:  size,
	}
	return *(*[]byte)(unsafe.Pointer(h))
}
