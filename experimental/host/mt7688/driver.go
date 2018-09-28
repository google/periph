// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mt7688

import (
	"errors"

	"periph.io/x/periph/host/sysfs"
)

// driverGPIO implements periph.Driver.
type driverGPIO struct {
	// gpioMemory is the memory map of the CPU GPIO registers.
	gpioMemory *gpioMap
}

func (d driverGPIO) String() string {
	return "mt7688-gpio"
}

func (d driverGPIO) Prerequisites() []string {
	return nil
}

func (d driverGPIO) After() []string {
	return []string{"sysfs-gpio"}
}

func (d *driverGPIO) Init() (bool, error) {
	if !Present() {
		return false, errors.New("mt7688 board not detected")
	}

	for _, p := range cpuPins {
		// Initialize sysfs access right away.
		p.sysfsPin = sysfs.Pins[p.number]
	}

	return true, nil
}

var drvGPIO driverGPIO
