// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package uart

import "fmt"

func ExamplePins() {
	//b, err := uartreg.Open("")
	//defer b.Close()
	var b Conn

	// Prints out the gpio pin used.
	if p, ok := b.(Pins); ok {
		fmt.Printf("  RX : %s", p.RX())
		fmt.Printf("  TX : %s", p.TX())
		fmt.Printf("  RTS: %s", p.RTS())
		fmt.Printf("  CTS: %s", p.CTS())
	}
}
