// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package beagle

import (
	"errors"

	"periph.io/x/periph"
)

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "beagle"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("BeagleBoard/BeagleBone board not detected")
	}
	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&driver{})
	}
}

var _ periph.Driver = &driver{}
