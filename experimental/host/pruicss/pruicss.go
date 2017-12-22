// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pruicss

import (
	"errors"

	"periph.io/x/periph"
)

// Present returns true if an Texas Instrument PRU-ICSS processor is detected.
//
// TODO(maruel): Implement.
func Present() bool {
	if isArm {
		return false
	}
	return false
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "pruicss"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("real time PRU-ICSS CPU not detected")
	}
	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&driver{})
	}
}
