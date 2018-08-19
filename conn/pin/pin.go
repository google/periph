// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package pin declare well known pins.
//
// pin is about physical pins, not about their logical function.
//
// While not a protocol strictly speaking, these are "well known constants".
package pin

import (
	"errors"

	"periph.io/x/periph/conn"
)

// These are well known pins.
var (
	INVALID  *BasicPin // Either floating or invalid pin
	GROUND   *BasicPin // Ground
	V1_8     *BasicPin // 1.8V (filtered)
	V2_8     *BasicPin // 2.8V (filtered)
	V3_3     *BasicPin // 3.3V (filtered)
	V5       *BasicPin // 5V (filtered)
	DC_IN    *BasicPin // DC IN; this is normally the 5V input
	BAT_PLUS *BasicPin // LiPo Battery + connector
)

// Pin is the minimal common interface shared between gpio.PinIO and
// analog.PinIO.
type Pin interface {
	conn.Resource
	// Name returns the name of the pin.
	Name() string
	// Number returns the logical pin number or a negative number if the pin is
	// not a GPIO, e.g. GROUND, V3_3, etc.
	Number() int
	// Function returns a user readable string representation of what the pin is
	// configured to do. Common case is In and Out but it can be bus specific pin
	// name.
	//
	// Deprecated: Use PinFunc.Func. Will be removed in v4.
	Function() string
}

// PinFunc is a supplementary interface that enables specifically querying for
// the pin function.
//
// TODO(maruel): It will be merged into interface Pin for v4.
type PinFunc interface {
	// Func returns the pin's current function.
	//
	// The returned value may be specialized or generalized, depending on the
	// actual port. For example it will likely be generalized for ports served
	// over USB (like a FT232H with D0 set as SPI_MOSI) but specialized for
	// ports on the base board (like a RPi3 with GPIO10 set as SPI0_MOSI).
	Func() Func
	// SupportedFuncs returns the possible functions this pin support.
	//
	// Do not mutate the returned slice.
	SupportedFuncs() []Func
	// SetFunc sets the pin function.
	//
	// Example use is to reallocate a RPi3's GPIO14 active function between
	// UART0_TX and UART1_TX.
	SetFunc(f Func) error
}

//

// BasicPin implements Pin as a static pin.
//
// It doesn't have a usable functionality.
type BasicPin struct {
	N string
}

// String implements conn.Resource.
func (b *BasicPin) String() string {
	return b.N
}

// Halt implements conn.Resource.
func (b *BasicPin) Halt() error {
	return nil
}

// Name implements Pin.
func (b *BasicPin) Name() string {
	return b.N
}

// Number implements Pin.
//
// Returns -1 as pin number.
func (b *BasicPin) Number() int {
	return -1
}

// Function implements Pin.
//
// Returns "" as pin function.
func (b *BasicPin) Function() string {
	return ""
}

// Func implements PinFunc.
//
// Returns FuncNone as pin function.
func (b *BasicPin) Func() Func {
	return FuncNone
}

// SupportedFuncs implements PinFunc.
//
// Returns nil.
func (b *BasicPin) SupportedFuncs() []Func {
	return nil
}

// SetFunc implements PinFunc.
func (b *BasicPin) SetFunc(f Func) error {
	return errors.New("pin: can't change static pin function")
}

func init() {
	INVALID = &BasicPin{N: "INVALID"}
	GROUND = &BasicPin{N: "GROUND"}
	V1_8 = &BasicPin{N: "1.8V"}
	V2_8 = &BasicPin{N: "2.8V"}
	V3_3 = &BasicPin{N: "3.3V"}
	V5 = &BasicPin{N: "5V"}
	DC_IN = &BasicPin{N: "DC_IN"}
	BAT_PLUS = &BasicPin{N: "BAT+"}
}

var _ Pin = INVALID
var _ PinFunc = INVALID
