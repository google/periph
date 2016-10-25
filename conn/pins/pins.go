// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package pins declare well known pins.
//
// Pins is about physical pins, not about their logical function.
//
// While not a protocol strictly speaking, these are "well known constants".
package pins

import "fmt"

// These are well known pins.
var (
	INVALID Pin = &BasicPin{N: "INVALID"} // Either floating or invalid pin
	GROUND  Pin = &BasicPin{N: "GROUND"}  // Ground
	V1_8    Pin = &BasicPin{N: "V1_8"}    // 1.8 volt
	V3_3    Pin = &BasicPin{N: "V3_3"}    // 3.3 volt
	V5      Pin = &BasicPin{N: "V5"}      // 5 vol
)

// Pin is the minimal common interface shared between gpio.PinIO and
// analog.PinIO.
type Pin interface {
	// String() typically returns the pin name and number, ex: "PD6(45)"
	fmt.Stringer
	// Name returns the name of the pin.
	Name() string
	// Number returns the logical pin number or a negative number if the pin is
	// not a GPIO, e.g. GROUND, V3_3, etc.
	Number() int
	// Function returns a user readable string representation of what the pin is
	// configured to do. Common case is In and Out but it can be bus specific pin
	// name.
	Function() string
}

//

// BasicPin implements Pin as a non-functional pin.
type BasicPin struct {
	N string
}

// String returns the pin name.
func (b *BasicPin) String() string {
	return b.N
}

// Name returns the pin name.
func (b *BasicPin) Name() string {
	return b.N
}

// Number returns -1 as pin number.
func (b *BasicPin) Number() int {
	return -1
}

// Function returns "" as pin function.
func (b *BasicPin) Function() string {
	return ""
}
