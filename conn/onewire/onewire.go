// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package onewire defines a Dallas Semiconductor / Maxim Integrated 1-wire bus.
//
// As described in https://periph.io/x/periph/conn#hdr-Concepts, periph.io uses
// the concepts of Bus, Port and Conn.
//
// In the package onewire, 'Port' is not exposed, since once you know the 1-wire
// device address, there's no unconfigured Port to configure.
//
// Instead, the package includes the adapter 'Dev' to directly convert an 1-wire
// bus 'onewire.Bus' into a connection 'conn.Conn' by only specifying the device
// 1-wire address.
//
// References
//
// Overview: https://www.maximintegrated.com/en/app-notes/index.mvp/id/1796
//
// App notes: https://www.maximintegrated.com/en/design/techdocs/app-notes/index.mvp/id/1/c/1-Wire%26reg%3B%20Devices
package onewire

import (
	"strconv"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
)

// Bus defines the function a concrete driver for a 1-wire bus must implement.
//
// This interface doesn't implement conn.Conn since a device address must be
// specified for each transaction. Use onewire.Dev as an adapter to get a
// conn.Conn compatible object.
type Bus interface {
	String() string
	// Tx performs a bus transaction, sending and receiving bytes, and
	// ending by pulling the bus high either weakly or strongly depending
	// on the value of power.
	//
	// A strong pull-up is typically required to power temperature conversion or
	// EEPROM writes.
	Tx(w, r []byte, power Pullup) error

	// Search performs a "search" cycle on the 1-wire bus and returns the
	// addresses of all devices on the bus if alarmOnly is false and of all
	// devices in alarm state if alarmOnly is true.
	//
	// If an error occurs during the search the already-discovered devices are
	// returned with the error.
	//
	// Bus.Search may be implemented using onewire.Search if the bus implements
	// the BusSearcher interface or it may have a custom implementation, for
	// example a Linux sysfs implementation should return the list of devices
	// already discovered by the driver.
	Search(alarmOnly bool) ([]Address, error)
}

// Address represents a 1-wire device address in little-endian format.
//
// This means that the family code ends up in the lower byte, the CRC in the
// top byte, and the variable address part in the middle 6 bytes. E.g. a
// DS18B20 device, which has a family code of 0x28, might have address
// 0x7a00000131825228.
type Address uint64

// Pullup encodes the type of pull-up used at the end of a bus transaction.
type Pullup bool

const (
	// WeakPullup ends the transaction with weak pull-up
	WeakPullup Pullup = false
	// StrongPullup end the transaction with strong pull-up to power devices
	StrongPullup Pullup = true
)

func (p Pullup) String() string {
	if p {
		return "Strong"
	}
	return "Weak"
}

// BusCloser is a 1-wire bus that can be closed.
//
// It is expected that an implementer of Bus also implement BusCloser, but
// this is not required.
type BusCloser interface {
	Close() error
	Bus
}

// Pins defines the pins that a 1-wire bus interconnect is using on the host.
//
// It is expected that an implementer of Bus also implement Pins but this is
// not a requirement.
type Pins interface {
	// Q returns the 1-wire Q (data) pin.
	Q() gpio.PinIO
}

// NoDevicesError is an interface that should be implemented by errors that
// indicate that no devices responded with a presence pulse after a reset.
type NoDevicesError interface {
	NoDevices() bool // true if no presence pulse from any device has been detected
}

// noDevicesError implements error and NoDevicesError.
type noDevicesError string

func (e noDevicesError) Error() string   { return string(e) }
func (e noDevicesError) NoDevices() bool { return true }

// ShortedBusError is an interface that should be implemented by errors that
// indicate that the bus is electrically shorted (Q connected to GND).
//
// Errors that implement ShortedBusError should also implement BusError.
type ShortedBusError interface {
	IsShorted() bool // true if the bus is electrically shorted
}

// shortedBusError implements error and ShortedBusError.
type shortedBusError string

func (e shortedBusError) Error() string   { return string(e) }
func (e shortedBusError) IsShorted() bool { return true }
func (e shortedBusError) BusError() bool  { return true }

// BusError is an interface that should be implemented by errors that
// indicate that an error occurred on the bus, for example a CRC error
// or a non-responding device. These errors often indicate an electrical
// problem with the bus and may be worth retrying.
//
// BusError also helps to differentiate 1-wire errors from errors accessing
// the 1-wire bus interface chip or circuitry, which may be located on
// an I²C bus or gpio pin.
type BusError interface {
	BusError() bool // true if a bus error was detected
}

// busError implements error and BusError.
type busError string

func (e busError) Error() string  { return string(e) }
func (e busError) BusError() bool { return true }

// Dev is a device on a 1-wire bus.
//
// It implements conn.Conn.
//
// Compared to Bus it saves from repeatedly specifying the device address and
// implements utility functions.
type Dev struct {
	Bus  Bus     // the bus to which the device is connected
	Addr Address // address of the device on the bus
}

// String prints the bus name followed by the device address in parenthesis.
func (d *Dev) String() string {
	s := "<nil>"
	if d.Bus != nil {
		s = d.Bus.String()
	}
	a := strconv.FormatUint(uint64(d.Addr), 16)
	for len(a) < 16 {
		// O(n²) but since digits is expected to run for a few loops, it doesn't
		// matter.
		a = "0" + a
	}
	return s + "(0x" + a + ")"
}

// Tx performs a "match ROM" command on the bus to select the device
// and then transmits and receives the specified bytes. It ends by
// leaving a weak pull-up on the bus.
//
// It's a wrapper for Dev.Bus.Tx().
func (d *Dev) Tx(w, r []byte) error {
	// Issue ROM match command to select the device followed by the
	// bytes being written.
	ww := make([]byte, 9, len(w)+9)
	ww[0] = 0x55 // Match ROM
	putUint64(ww[1:], d.Addr)
	ww = append(ww, w...)
	return d.Bus.Tx(ww, r, WeakPullup)
}

// Duplex always return conn.Half for 1-wire.
func (d *Dev) Duplex() conn.Duplex {
	return conn.Half
}

// TxPower performs a "match ROM" command on the bus to select the device
// and then transmits and receives the specified bytes. It ends by
// leaving a strong pull-up on the bus suitable to power devices through
// an EEPROM write or a temperature conversion.
//
// It's a wrapper for Dev.Bus.Tx().
func (d *Dev) TxPower(w, r []byte) error {
	// Issue ROM match command to select the device followed by the
	// bytes being written.
	ww := make([]byte, 9, len(w)+9)
	ww[0] = 0x55 // Match ROM
	putUint64(ww[1:], d.Addr)
	ww = append(ww, w...)
	return d.Bus.Tx(ww, r, StrongPullup)
}

//

// putUint64 is littleEndian.PutUint64().
//
// It was extracted to to not depend on encoding/binary, which depends on
// reflect.
func putUint64(b []byte, v Address) {
	_ = b[7]
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

// Ensure that the appropriate interfaces are implemented.
var _ conn.Conn = &Dev{}
var _ NoDevicesError = noDevicesError("")
var _ ShortedBusError = shortedBusError("")
var _ BusError = busError("")
