// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains the Allwinner driver struct and the top-level initialization code.

package allwinner

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"unsafe"

	"github.com/google/periph"
	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/host/pmem"
)

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "allwinner"
}

func (d *driver) Prerequisites() []string {
	return nil
}

// Init does nothing if an allwinner processor is not detected. If one is
// detected, it memory maps gpio CPU registers and then sets up the pin mapping
// for the exact processor model detected.
func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("Allwinner CPU not detected")
	}
	m, err := pmem.Map(getBaseAddress(), 4096)
	if err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}
	m.Struct(unsafe.Pointer(&gpioMemory))

	switch {
	case IsA64():
		mapA64Pins()
	case IsR8():
		mapR8Pins()
	default:
		return false, errors.New("Unknown Allwinner CPU model")
	}

	return true, initPins()
}

func init() {
	if isArm {
		periph.MustRegister(&driver{})
	}
}

// getBaseAddress queries the virtual file system to retrieve the base address
// of the GPIO registers for GPIO pins in groups PB to PH.
//
// Defaults to 0x01C20800 as per datasheet if it could not query the file system.
func getBaseAddress() uint64 {
	base := uint64(0x01C20800)
	link, err := os.Readlink("/sys/bus/platform/drivers/sun50i-pinctrl/driver")
	if err != nil {
		return base
	}
	parts := strings.SplitN(path.Base(link), ".", 2)
	if len(parts) != 2 {
		return base
	}
	base2, err := strconv.ParseUint(parts[0], 16, 64)
	if err != nil {
		return base
	}
	return base2
}

// Ensure that the various structs implement the interfaces they're supposed to.

var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ gpio.PinIO = &Pin{}
