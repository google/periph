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

func ExampleBitsLSB() {
	// Format is LSB; least significant bit first.
	stream := Bits{0x80, 0x01, 0xAA, 0x55}
	for _, l := range stream {
		fmt.Printf("0x%02X: ", l)
		for j := 0; j < 8; j++ {
			mask := byte(1) << uint(j)
			fmt.Printf("%4s,", gpio.Level(l&mask != 0))
			if j != 7 {
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}
	// Output:
	// 0x80:  Low,  Low,  Low,  Low,  Low,  Low,  Low, High,
	// 0x01: High,  Low,  Low,  Low,  Low,  Low,  Low,  Low,
	// 0xAA:  Low, High,  Low, High,  Low, High,  Low, High,
	// 0x55: High,  Low, High,  Low, High,  Low, High,  Low,
}

func ExampleBitsMSB() {
	// Format is MSB; most significant bit first.
	stream := Bits{0x80, 0x01, 0xAA, 0x55}
	for _, l := range stream {
		fmt.Printf("0x%02X: ", l)
		for j := 7; j >= 0; j-- {
			mask := byte(1) << uint(j)
			fmt.Printf("%4s,", gpio.Level(l&mask != 0))
			if j != 0 {
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}
	// Output:
	// 0x80: High,  Low,  Low,  Low,  Low,  Low,  Low,  Low,
	// 0x01:  Low,  Low,  Low,  Low,  Low,  Low,  Low, High,
	// 0xAA: High,  Low, High,  Low, High,  Low, High,  Low,
	// 0x55:  Low, High,  Low, High,  Low, High,  Low, High,
}

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
	for i, l := range b.Bits {
		// Bits format is in MSB; the most significant bit is streamed first.
		for j := 7; j >= 0; j-- {
			mask := byte(1) << uint(j)
			fmt.Printf("%4s, ", gpio.Level(l&mask != 0))
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

func TestBitStreamLSB(t *testing.T) {
	s := BitStreamLSB{Res: time.Second, Bits: make(BitsLSB, 100)}
	if r := s.Resolution(); r != time.Second {
		t.Fatal(r)
	}
	if d := s.Duration(); d != 100*time.Second {
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
	if d := s.Duration(); d != 100*time.Second {
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
	if d := s.Duration(); d != 7*time.Second+2*time.Millisecond {
		t.Fatal(d)
	}
}
