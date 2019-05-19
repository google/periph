// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"encoding/binary"
	"errors"
	"fmt"

	"periph.io/x/periph/conn/gpio/gpiostream"
)

// uint32ToBitLSBF packs a bit offset found on slice `d` (that is actually
// uint32) back into a densely packed Bits stream.
func uint32ToBitLSBF(w []byte, d []uint8, bit uint8, skip int) {
	// Little endian.
	x := bit / 8
	d = d[x:]
	bit -= 8 * x
	mask := uint8(1) << bit
	for i := range w {
		w[i] = ((d[0]&mask)>>bit<<0 |
			(d[skip*1]&mask)>>bit<<1 |
			(d[skip*2]&mask)>>bit<<2 |
			(d[skip*3]&mask)>>bit<<3 |
			(d[skip*4]&mask)>>bit<<4 |
			(d[skip*5]&mask)>>bit<<5 |
			(d[skip*6]&mask)>>bit<<6 |
			(d[skip*7]&mask)>>bit<<7)
		d = d[skip*8:]
	}
}

func getBit(b byte, index int, msb bool) byte {
	var shift uint
	if msb {
		shift = uint(7 - index)
	} else {
		shift = uint(index)
	}
	return (b >> shift) & 1
}

func raster32Bits(s gpiostream.Stream, skip int, clear, set []uint32, mask uint32) error {
	var msb bool
	var bits []byte
	switch b := s.(type) {
	case *gpiostream.BitStream:
		msb = !b.LSBF
		bits = b.Bits
	default:
		return fmt.Errorf("unsupported type %T", b)
	}
	m := len(clear) / 8
	if n := len(bits); n < m {
		m = n
	}
	index := 0
	for i := 0; i < m; i++ {
		for j := 0; j < 8; j++ {
			if getBit(bits[i], j, msb) != 0 {
				for k := 0; k < skip; k++ {
					set[index] |= mask
					index++
				}
			} else {
				for k := 0; k < skip; k++ {
					clear[index] |= mask
					index++
				}
			}
		}
	}
	return nil
}

// raster32 rasters the stream into a uint32 stream with the specified masks to
// put in the correctly slice when the bit is set and when it is clear.
//
// `s` must be one of the types in this package.
func raster32(s gpiostream.Stream, skip int, clear, set []uint32, mask uint32) error {
	if mask == 0 {
		return errors.New("bcm283x: mask is 0")
	}
	if len(clear) == 0 {
		return errors.New("bcm283x: clear buffer is empty")
	}
	if len(set) == 0 {
		return errors.New("bcm283x: set buffer is empty")
	}
	if len(clear) != len(set) {
		return errors.New("bcm283x: clear and set buffers have different length")
	}
	switch x := s.(type) {
	case *gpiostream.BitStream:
		// TODO
		return raster32Bits(x, skip, clear, set, mask)
	case *gpiostream.EdgeStream:
		return errors.New("bcm283x: EdgeStream is not supported yet")
	case *gpiostream.Program:
		return errors.New("bcm283x: Program is not supported yet")
	default:
		return errors.New("bcm283x: unknown stream type")
	}
}

// PCM/PWM DMA buf is encoded as little-endian and MSB first.
func copyStreamToDMABuf(w gpiostream.Stream, dst []uint32) error {
	switch v := w.(type) {
	case *gpiostream.BitStream:
		if v.LSBF {
			return errors.New("TODO(simokawa): handle BitStream.LSBF")
		}
		// This is big-endian and MSB first.
		i := 0
		for ; i < len(v.Bits)/4; i++ {
			dst[i] = binary.BigEndian.Uint32(v.Bits[i*4:])
		}
		last := uint32(0)
		if mod := len(v.Bits) % 4; mod > 0 {
			for j := 0; j < mod; j++ {
				last |= (uint32(v.Bits[i*4+j])) << uint32(8*(3-j))
			}
			dst[i] = last
		}
		return nil
	case *gpiostream.EdgeStream:
		return errors.New("TODO(simokawa): handle EdgeStream")
	default:
		return errors.New("unsupported Stream type")
	}
}
