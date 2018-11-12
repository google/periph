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

// Sample is one analog sample.
//
// Raw must be set, but V may or may not be set, depending if the device knows
// the electrical tension this measurement represents.
type Sample struct {
	// V is the interpreted electrical tension.
	V physic.ElectricPotential
	// Raw is the raw measurement.
	Raw int32
}

// PinADC is an analog-to-digital-conversion input.
type PinADC interface {
	pin.Pin
	// Range returns the maximum supported range [min, max] of the values.
	//
	// It is possible for a DAC that the Sample.V value is not set.
	Range() (Sample, Sample)
	// Read returns the current pin level.
	Read() (Sample, error)
}

// PinDAC is an digital-to-analog-conversion output.
type PinDAC interface {
	pin.Pin
	// Range returns the maximum supported range [min, max] of the values.
	//
	// It is possible for a DAC that the Sample.V value is not set.
	Range() (Sample, Sample)
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

func (invalidPin) Range() (Sample, Sample) {
	return Sample{}, Sample{}
}

func (invalidPin) Read() (Sample, error) {
	return Sample{}, errInvalidPin
}

func (invalidPin) Out(v int32) error {
	return errInvalidPin
}

var _ PinADC = &INVALID
var _ PinDAC = &INVALID
