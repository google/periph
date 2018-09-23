// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mt7688

import (
	"errors"
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/host/distro"
)

// Present returns true if a mt7688 processor is detected.
func Present() bool {
	if isMIPS {
		hardware, ok := distro.CPUInfo()["system type"]
		return ok && strings.HasPrefix(hardware, "MediaTek MT7688")
		return true
	}
	return false
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "mt7688"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("mt7688 board not detected")
	}

	return true, nil
}

func init() {
	// Since isMIPS is a compile time constant, the compile can strip the
	// unnecessary code and unused private symbols.
	if isMIPS {
		periph.MustRegister(&drv)
	}
}

var drv driver
