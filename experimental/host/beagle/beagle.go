// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package beagle

import (
	"errors"
	"os"

	"periph.io/x/periph"
)

// Present returns true if there is evidence that we have a
// Beagleboard present.
func Present() bool {
	if isArm {
		_, err := os.Stat("/sys/firmware/devicetree/base/bone_capemgr")

		return err == nil
	}
	return false
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "beagle"
}

func (d *driver) Prerequisites() []string {
	return []string{"sysfs"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("BeagleBoard board not detected")
	}

	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&driver{})
	}
}

var _ periph.Driver = &driver{}
