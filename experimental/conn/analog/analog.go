// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package analog defines analog pins, both DAC and ADC.
package analog

import (
	"errors"

	"github.com/google/periph/conn/pins"
)

// ADC is an analog-to-digital-conversion input.
type ADC interface {
	pins.Pin
	// Range returns the maximum supported range [min, max] of the values.
	Range() (int32, int32)
	// Read returns the current pin level.
	Read() int32
}

// DAC is an digital-to-analog-conversion output.
type DAC interface {
	pins.Pin
	// Range returns the maximum supported range [min, max] of the values.
	Range() (int32, int32)
	// Out sets an analog output value.
	DAC(v int32)
}

// INVALID implements both ADC and DAC and fails on all access.
var INVALID invalidPin

//

// errInvalidPin is returned when trying to use INVALID.
var errInvalidPin = errors.New("invalid pin")

// invalidPin implements PinIO for compability but fails on all access.
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

func (invalidPin) Range() (int32, int32) {
	return 0, 0
}

func (invalidPin) Read() int32 {
	return 0
}

func (invalidPin) DAC(v int32) {
}
