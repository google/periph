// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spi

import (
	"fmt"
	"testing"
)

func ExamplePins() {
	//b, err := spireg.Open("")
	//defer b.Close()
	var b Conn

	// Prints out the gpio pin used.
	if p, ok := b.(Pins); ok {
		fmt.Printf("  CLK : %s", p.CLK())
		fmt.Printf("  MOSI: %s", p.MOSI())
		fmt.Printf("  MISO: %s", p.MISO())
		fmt.Printf("  CS  : %s", p.CS())
	}
}

//

func TestMode_String(t *testing.T) {
	if s := Mode(^int(0)).String(); s != "Mode3|HalfDuplex|NoCS|LSBFirst|0xffffffffffffffe0" {
		t.Fatal(s)
	}
	if s := Mode0.String(); s != "Mode0" {
		t.Fatal(s)
	}
	if s := Mode1.String(); s != "Mode1" {
		t.Fatal(s)
	}
	if s := Mode2.String(); s != "Mode2" {
		t.Fatal(s)
	}
}
