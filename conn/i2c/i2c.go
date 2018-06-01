// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package i2c defines interface to an I²C bus and an I²C device.
//
// As described in https://periph.io/x/periph/conn#hdr-Concepts, periph.io uses
// the concepts of Bus, Port and Conn.
//
// In the package i2c, 'Port' is not exposed, since once you know the I²C
// device address, there's no unconfigured Port to configure.
//
// Instead, the package includes the adapter 'Dev' to directly convert an I²C
// bus 'i2c.Bus' into a connection 'conn.Conn' by only specifying the device
// I²C address.
package i2c

import (
	"io"
	"strconv"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
)

// Bus defines the interface a concrete I²C driver must implement.
//
// This interface is consummed by a device driver for a device sitting on a bus.
//
// This interface doesn't implement conn.Conn since a device address must be
// specified. Use i2cdev.Dev as an adapter to get a conn.Conn compatible
// object.
type Bus interface {
	String() string
	// Tx does a transaction at the specified device address.
	//
	// Write is done first, then read. One of 'w' or 'r' can be omitted for a
	// unidirectional operation.
	Tx(addr uint16, w, r []byte) error
	// SetSpeed changes the bus speed, if supported.
	//
	// On linux due to the way the I²C sysfs driver is exposed in userland,
	// calling this function will likely affect *all* I²C buses on the host.
	SetSpeed(f physic.Frequency) error
}

// BusCloser is an I²C bus that can be closed.
//
// This interface is meant to be handled by the application and not the device
// driver. A device driver doesn't "own" a bus, hence it must operate on a Bus,
// not a BusCloser.
type BusCloser interface {
	io.Closer
	Bus
}

// Pins defines the pins that an I²C bus interconnect is using on the host.
//
// It is expected that a implementer of Bus also implement Pins but this is not
// a requirement.
type Pins interface {
	// SCL returns the CLK (clock) pin.
	SCL() gpio.PinIO
	// SDA returns the DATA pin.
	SDA() gpio.PinIO
}

// Dev is a device on a I²C bus.
//
// It implements conn.Conn.
//
// It saves from repeatedly specifying the device address.
type Dev struct {
	Bus  Bus
	Addr uint16
}

func (d *Dev) String() string {
	s := "<nil>"
	if d.Bus != nil {
		s = d.Bus.String()
	}
	return s + "(" + strconv.Itoa(int(d.Addr)) + ")"
}

// Tx does a transaction by adding the device's address to each command.
//
// It's a wrapper for Bus.Tx().
func (d *Dev) Tx(w, r []byte) error {
	return d.Bus.Tx(d.Addr, w, r)
}

// Write writes to the I²C bus without reading, implementing io.Writer.
//
// It's a wrapper for Tx()
func (d *Dev) Write(b []byte) (int, error) {
	if err := d.Tx(b, nil); err != nil {
		return 0, err
	}
	return len(b), nil
}

// Duplex always return conn.Half for I²C.
func (d *Dev) Duplex() conn.Duplex {
	return conn.Half
}

//

var _ conn.Conn = &Dev{}
