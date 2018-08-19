// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiostream

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/physic"
)

func TestBitStream(t *testing.T) {
	var b [16]byte
	for i := range b {
		b[i] = byte(i)
	}
	s := BitStream{Freq: physic.Hertz, Bits: b[:], LSBF: true}
	if f := s.Frequency(); f != physic.Hertz {
		t.Fatal(f)
	}
	if d := s.Duration(); d != 16*8*time.Second {
		t.Fatal(d)
	}
	if g := s.GoString(); g != "&gpiostream.BitStream{Bits: 000102030405060708090a0b0c0d0e0f, Freq:1Hz, LSBF:true}" {
		t.Fatal(g)
	}
}

func TestBitStream_Empty(t *testing.T) {
	var b [16]byte
	s := BitStream{Bits: b[:]}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
}

func TestEdgeStream(t *testing.T) {
	s := EdgeStream{Freq: physic.KiloHertz, Edges: []uint16{1000, 1}}
	if f := s.Frequency(); f != physic.KiloHertz {
		t.Fatal(f)
	}
	if d := s.Duration(); d != 1001*time.Millisecond {
		t.Fatal(d)
	}
	s = EdgeStream{Edges: []uint16{1000, 1}}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
}

func TestProgram(t *testing.T) {
	s := Program{
		Parts: []Stream{
			&EdgeStream{Freq: physic.KiloHertz, Edges: []uint16{1000, 1}},
			&BitStream{Freq: physic.Hertz, Bits: make([]byte, 100)},
		},
		Loops: 2,
	}
	if f := s.Frequency(); f != physic.KiloHertz {
		t.Fatal(f)
	}
	if d := s.Duration(); d != 2*(100*8*time.Second+1001*time.Millisecond) {
		t.Fatal(d)
	}
	s = Program{Loops: 0}
	if f := s.Frequency(); f != 0 {
		t.Fatal(f)
	}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
	s = Program{Parts: []Stream{&Program{}}, Loops: -1}
	if f := s.Frequency(); f != 0 {
		t.Fatal(f)
	}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
}

func TestProgram_Nyquist(t *testing.T) {
	s := Program{
		Parts: []Stream{
			&BitStream{Freq: 998 * physic.MilliHertz, Bits: make([]byte, 1)},
			&BitStream{Freq: physic.Hertz, Bits: make([]byte, 1)},
			&BitStream{Freq: 200 * physic.MilliHertz, Bits: make([]byte, 1)},
		},
		Loops: 1,
	}
	// TODO(maruel): This will cause small aliasing on the first BitStream.
	if f := s.Frequency(); f != 2*physic.Hertz {
		t.Fatal(f)
	}

	if d := s.Duration(); d != 56016032064*time.Nanosecond {
		t.Fatal(d)
	}
}
