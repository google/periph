// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiostream

import (
	"testing"
	"time"
)

func TestBitStreamLSB(t *testing.T) {
	s := BitStreamLSB{Res: time.Second, Bits: make(BitsLSB, 100)}
	if r := s.Resolution(); r != time.Second {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 100*8*time.Second {
		t.Fatal(d)
	}
	s = BitStreamLSB{Res: time.Second}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
}

func TestBitStreamMSB(t *testing.T) {
	s := BitStreamMSB{Res: time.Second, Bits: make(BitsMSB, 100)}
	if r := s.Resolution(); r != time.Second {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 100*8*time.Second {
		t.Fatal(d)
	}
	s = BitStreamMSB{Res: time.Second}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
}

func TestBitStream(t *testing.T) {
	s := BitStream{Res: time.Second, Bits: make(Bits, 100)}
	if r := s.Resolution(); r != time.Second {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 100*time.Second {
		t.Fatal(d)
	}
	s = BitStream{Res: time.Second}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
}

func TestEdgeStream(t *testing.T) {
	s := EdgeStream{Res: time.Second, Edges: []time.Duration{time.Second, time.Millisecond}}
	if r := s.Resolution(); r != time.Second {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 1001*time.Millisecond {
		t.Fatal(d)
	}
	s = EdgeStream{Res: time.Second}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
	s = EdgeStream{Edges: []time.Duration{time.Second, time.Millisecond}}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
}

func TestProgram(t *testing.T) {
	s := Program{
		Parts: []Stream{
			&EdgeStream{Res: time.Second, Edges: []time.Duration{time.Second, time.Millisecond}},
			&BitStreamLSB{Res: time.Second, Bits: make(BitsLSB, 100)},
		},
		Loops: 2,
	}
	if r := s.Resolution(); r != 500*time.Millisecond {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 2*(100*8*time.Second+1001*time.Millisecond) {
		t.Fatal(d)
	}
	s = Program{Loops: 0}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
	s = Program{Parts: []Stream{&Program{}}, Loops: -1}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
}

func TestProgram_Nyquist(t *testing.T) {
	s := Program{
		Parts: []Stream{
			&BitStreamLSB{Res: time.Second + 2*time.Millisecond, Bits: make(BitsLSB, 1)},
			&BitStreamLSB{Res: time.Second, Bits: make(BitsLSB, 1)},
			&BitStreamLSB{Res: 5 * time.Second, Bits: make(BitsLSB, 1)},
		},
		Loops: 1,
	}
	// TODO(maruel): This will cause small aliasing on the first BitStream.
	if r := s.Resolution(); r != 500*time.Millisecond {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 7*8*time.Second+2*8*time.Millisecond {
		t.Fatal(d)
	}
}
