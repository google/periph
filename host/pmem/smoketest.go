// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"bytes"
	"math/rand"
)

// TestCopy is used by CPU drivers to verify that the DMA engine works
// correctly.
//
// It is not meant to be used by end users.
//
// TestCopy allocates two buffer via `alloc`, once as the source and one as the
// destination. It fills the source with random data and the destination with
// 0x11.
//
// `copyMem` is expected to copy the memory from pSrc to pDst, with an offset
// of `hole` and size `size-2*hole`.
//
// The function `copyMem` being tested is only given the buffer physical
// addresses and must copy the data without other help. It is expected to
//
// This confirm misaligned DMA copying works.
// leverage the host's DMA engine.
func TestCopy(size, holeSize int, alloc func(size int) (Mem, error), copyMem func(pDst, pSrc uint64) error) error {
	pSrc, err2 := alloc(size)
	if err2 != nil {
		return err2
	}
	defer pSrc.Close()
	pDst, err2 := alloc(size)
	if err2 != nil {
		return err2
	}
	defer pDst.Close()
	dst := pDst.Bytes()
	for i := range dst {
		dst[i] = 0x11
	}
	src := make([]byte, size)
	for i := range src {
		src[i] = byte(rand.Int31())
	}
	copy(pSrc.Bytes(), src[:])

	// Run the driver supplied memory copying code.
	if err := copyMem(pDst.PhysAddr(), pSrc.PhysAddr()); err != nil {
		return err
	}

	// Verifications.
	for i := 0; i < holeSize; i++ {
		if dst[i] != 0x11 {
			return wrapf("DMA corrupted the buffer header: %x", dst[:holeSize])
		}
		if dst[size-1-i] != 0x11 {
			return wrapf("DMA corrupted the buffer footer: %x", dst[size-1-holeSize:])
		}
	}

	// Headers and footers were not corupted in the destination. Verify the inner
	// view that should match.
	x := src[:size-2*holeSize]
	y := dst[holeSize : size-holeSize]
	if !bytes.Equal(x, y) {
		offset := 0
		for len(x) != 0 && x[0] == y[0] {
			x = x[1:]
			y = y[1:]
			offset++
		}
		for len(x) != 0 && x[len(x)-1] == y[len(y)-1] {
			x = x[:len(x)-1]
			y = y[:len(y)-1]
		}
		if len(x) > 32 {
			x = x[:32]
		}
		if len(y) > 32 {
			y = y[:32]
		}
		return wrapf("DMA corrupted the buffer at offset %d:\n%x\n%x", offset, x, y)
	}
	return nil
}
