// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package am335x

import (
	"errors"
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/host/distro"
)

// Present returns true if a TM AM335x processor is detected.
func Present() bool {
	if isArm {
		return strings.HasPrefix(distro.DTModel(), "TI AM335x")
	}
	return false
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "am335x"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("am335x CPU not detected")
	}
	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&drv)
	}
}

var drv driver
