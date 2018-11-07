// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package analog defines analog pins, both digital to analog converter (DAC)
// and analog to digital converter (ADC).
package analog

import (
	"errors"

	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
)

// Reading is the result of PinADC.Read().
type Reading struct {
	// V is the interpreted electrical level.
	V physic.ElectricPotential
	// Raw is the raw measurement.
	Raw int32
}

// PinADC is an analog-to-digital-conversion input.
type PinADC interface {
	pin.Pin
	// Range returns the maximum supported range [min, max] of the values.
	Range() (Reading, Reading)
	// Read returns the current pin level.
	Read() (Reading, error)
}

// PinDAC is an digital-to-analog-conversion output.
type PinDAC interface {
	pin.Pin
	// Range returns the maximum supported range [min, max] of the values.
	//
	// It is possible for a DAC that the Reading.V value is not set.
	Range() (Reading, Reading)
	// Out sets an analog output value.
	Out(v int32) error
}

// INVALID implements both PinADC and PinDAC and fails on all access.
var INVALID invalidPin

//

// errInvalidPin is returned when trying to use INVALID.
var errInvalidPin = errors.New("invalid pin")

// invalidPin implements PinIO for compatibility but fails on all access.
type invalidPin struct {
}

func (invalidPin) Number() int {
	return -1
}

func (invalidPin) Name() string {
	return "INVALID"
}

func (invalidPin) String() string {
	return "INVALID"
}

func (invalidPin) Function() string {
	return ""
}

func (invalidPin) Halt() error {
	return errInvalidPin
}

func (invalidPin) Range() (Reading, Reading) {
	return Reading{}, Reading{}
}

func (invalidPin) Read() (Reading, error) {
	return Reading{}, errInvalidPin
}

func (invalidPin) Out(v int32) error {
	return errInvalidPin
}

var _ PinADC = &INVALID
var _ PinDAC = &INVALID
