// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiostream_test

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/host"
)

func ExampleBitsLSBF() {
	// Format is LSB-first; least significant bit first.
	stream := gpiostream.BitsLSBF{0x80, 0x01, 0xAA, 0x55}
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

func ExampleBitsMSBF() {
	// Format is MSB-first; most significant bit first.
	stream := gpiostream.BitsMSBF{0x80, 0x01, 0xAA, 0x55}
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
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Read one second of sample at 1ms resolution and print the values read.
	p := gpioreg.ByName("GPIO3")
	r, ok := p.(gpiostream.PinIn)
	if !ok {
		log.Fatalf("pin streaming is not supported on pin %s", p)
	}
	b := gpiostream.BitStreamMSBF{Res: time.Millisecond, Bits: make(gpiostream.BitsMSBF, 1000/8)}
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
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Generates a 25% duty cycle PWM at 1kHz for 5 seconds with a precision of
	// 1µs.
	p := gpioreg.ByName("GPIO3")
	r, ok := p.(gpiostream.PinOut)
	if !ok {
		log.Fatalf("pin streaming is not supported on pin %s", p)
	}
	b := gpiostream.Program{
		Parts: []gpiostream.Stream{
			&gpiostream.EdgeStream{
				Res:   time.Microsecond,
				Edges: []uint16{250, 750},
			},
		},
		Loops: 5000,
	}
	if err := r.StreamOut(&b); err != nil {
		log.Fatal(err)
	}
}
