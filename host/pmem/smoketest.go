// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
)

// CopyTest is used by CPU drivers to verify that the DMA engine works
// correctly.
//
// It allocates a buffer of `size` via the `alloc` function provided. It fills
// the source buffer with random data and copies it to the destination with
// `holeSize` bytes as an untouched header and footer. This is done to confirm
// misaligned copying works.
//
// The function `f` being tested is only given the buffer physical addresses and
// must copy the data without other help.
func CopyTest(size, holeSize int, alloc func(size int) (Mem, error), f func(pDst, pSrc uint64) error) error {
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
	if err := f(pSrc.PhysAddr(), pDst.PhysAddr()); err != nil {
		return err
	}

	// Verifications.
	for i := 0; i < holeSize; i++ {
		if dst[i] != 0x11 {
			return fmt.Errorf("DMA corrupted the buffer header: %s", hex.EncodeToString(dst[:holeSize]))
		}
		if dst[size-1-i] != 0x11 {
			return fmt.Errorf("DMA corrupted the buffer footer: %s", hex.EncodeToString(dst[size-1-holeSize:]))
		}
	}
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
		return fmt.Errorf("DMA corrupted the buffer at offset %d:\n%s\n%s", offset, hex.EncodeToString(x), hex.EncodeToString(y))
	}
	return nil
}
