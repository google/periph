// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package green implements headers for the BeagleBone Green and BeagleBone
// Green Wireless micro-computers.
//
// Reference
//
// https://beagleboard.org/green
//
// https://beagleboard.org/green-wireless
//
// Datasheet
//
// http://wiki.seeedstudio.com/BeagleBone_Green/
package green

import (
	"errors"
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/distro"
	"periph.io/x/periph/host/sysfs"
)

// Headers found on BeagleBone Green.
var (
	// I2C Groove port.
	I2C_SCL gpio.PinIO = gpio.INVALID // GPIO13, I2C2_SCL, UART1_RTS, SPI1_CS1
	I2C_SDA gpio.PinIO = gpio.INVALID // GPIO12, I2C2_SDA, UART1_CTS, SPI1_CS0

	// UART Groove port connected to UART2.
	UART_TX gpio.PinIO = gpio.INVALID // GPIO3, EHRPWM0B, I2C2_SCL, UART2_TX, SPI0_MISO
	UART_RX gpio.PinIO = gpio.INVALID // GPIO2, EHRPWM0A, I2C2_SDA, UART2_RX, SPI0_CLK
)

// Present returns true if the host is a BeagleBone Green or BeagleBone Green
// Wireless.
func Present() bool {
	if isArm {
		return strings.HasPrefix(distro.DTModel(), "TI AM335x BeagleBone Green")
	}
	return false
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "beaglebone-green"
}

func (d *driver) Prerequisites() []string {
	return []string{"am335x"}
}

func (d *driver) After() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("BeagleBone Green board not detected")
	}

	I2C_SDA = sysfs.Pins[12]
	I2C_SCL = sysfs.Pins[13]
	hdr := [][]pin.Pin{{pin.GROUND}, {pin.V3_3}, {I2C_SDA}, {I2C_SCL}}
	if err := pinreg.Register("I2C", hdr); err != nil {
		return true, err
	}

	UART_TX = sysfs.Pins[3]
	UART_RX = sysfs.Pins[2]
	hdr = [][]pin.Pin{{pin.GROUND}, {pin.V3_3}, {UART_TX}, {UART_RX}}
	if err := pinreg.Register("UART", hdr); err != nil {
		return true, err
	}

	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&drv)
	}
}

var drv driver
