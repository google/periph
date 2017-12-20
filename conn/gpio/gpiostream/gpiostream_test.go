// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiostream

import (
	"testing"
	"time"
)

func TestBitStreamLSBF(t *testing.T) {
	s := BitStreamLSBF{Res: time.Second, Bits: make(BitsLSBF, 100)}
	if r := s.Resolution(); r != time.Second {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 100*8*time.Second {
		t.Fatal(d)
	}
	s = BitStreamLSBF{Res: time.Second}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
}

func TestBitStreamMSBF(t *testing.T) {
	s := BitStreamMSBF{Res: time.Second, Bits: make(BitsMSBF, 100)}
	if r := s.Resolution(); r != time.Second {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 100*8*time.Second {
		t.Fatal(d)
	}
	s = BitStreamMSBF{Res: time.Second}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
}

func TestEdgeStream(t *testing.T) {
	s := EdgeStream{Res: time.Millisecond, Edges: []uint16{1000, 1}}
	if r := s.Resolution(); r != time.Millisecond {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 1001*time.Millisecond {
		t.Fatal(d)
	}
	s = EdgeStream{Res: time.Second}
	if r := s.Resolution(); r != 0 {
		t.Fatal(r)
	}
	s = EdgeStream{Edges: []uint16{1000, 1}}
	if d := s.Duration(); d != 0 {
		t.Fatal(d)
	}
}

func TestProgram(t *testing.T) {
	s := Program{
		Parts: []Stream{
			&EdgeStream{Res: time.Millisecond, Edges: []uint16{1000, 1}},
			&BitStreamLSBF{Res: time.Second, Bits: make(BitsLSBF, 100)},
		},
		Loops: 2,
	}
	if r := s.Resolution(); r != time.Millisecond {
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
			&BitStreamLSBF{Res: time.Second + 2*time.Millisecond, Bits: make(BitsLSBF, 1)},
			&BitStreamLSBF{Res: time.Second, Bits: make(BitsLSBF, 1)},
			&BitStreamLSBF{Res: 5 * time.Second, Bits: make(BitsLSBF, 1)},
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
