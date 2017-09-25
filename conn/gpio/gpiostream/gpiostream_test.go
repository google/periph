// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiostream

import (
	"fmt"
	"log"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

func ExamplePinIn() {
	// Read one second of sample at 1ms resolution and print the values read.
	p := gpioreg.ByName("GPIO3")
	r, ok := p.(PinIn)
	if !ok {
		log.Fatalf("pin streaming is not supported on pin %s", p)
	}
	b := BitStream{Res: time.Millisecond, Bits: make(Bits, 1000/8)}
	if err := r.StreamIn(gpio.PullNoChange, &b); err != nil {
		log.Fatal(err)
	}
	for i, bit := range b.Bits {
		for j := 0; j < 8; j++ {
			fmt.Printf("%4s, ", gpio.Level(bit&(1<<uint(j)) != 0))
		}
		if i&1 == 1 {
			fmt.Printf("\n")
		}
	}
}

func ExamplePinOut() {
	// Generates a 25% duty cycle PWM at 1kHz for 5 seconds with a precision of
	// 1µs.
	p := gpioreg.ByName("GPIO3")
	r, ok := p.(PinOut)
	if !ok {
		log.Fatalf("pin streaming is not supported on pin %s", p)
	}
	b := Program{
		Parts: []Stream{
			&EdgeStream{
				Res:   time.Microsecond,
				Edges: []time.Duration{250 * time.Microsecond, 750 * time.Microsecond},
			},
		},
		Loops: 5000,
	}
	if err := r.StreamOut(&b); err != nil {
		log.Fatal(err)
	}
}

//

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
			&BitStream{Res: time.Second, Bits: make(Bits, 100)},
		},
		Loops: 2,
	}
	if r := s.Resolution(); r != 500*time.Millisecond {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 2*(100*time.Second+1001*time.Millisecond) {
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
			&BitStream{Res: time.Second + 2*time.Millisecond, Bits: make(Bits, 1)},
			&BitStream{Res: time.Second, Bits: make(Bits, 1)},
			&BitStream{Res: 5 * time.Second, Bits: make(Bits, 1)},
		},
		Loops: 1,
	}
	// TODO(maruel): This will cause small aliasing on the first BitStream.
	if r := s.Resolution(); r != 500*time.Millisecond {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 7*time.Second+2*time.Millisecond {
		t.Fatal(d)
	}
}
