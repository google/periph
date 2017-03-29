// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spi

import "fmt"

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
