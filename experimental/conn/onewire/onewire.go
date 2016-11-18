// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package onewire defines a Dallas Semiconductor / Maxim Integrated 1-wire bus.
//
// It includes an adapter to directly address a 1-wire device on a 1-wire bus
// without having to continuously specify the address when doing I/O. This
// enables the support of conn.Conn.
//
// References
//
// Overview: https://www.maximintegrated.com/en/app-notes/index.mvp/id/1796
//
// App notes: https://www.maximintegrated.com/en/design/techdocs/app-notes/index.mvp/id/1/c/1-Wire%26reg%3B%20Devices#c1
package onewire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/google/periph/conn"
	"github.com/google/periph/conn/gpio"
)

// Bus defines the function a concrete driver for a 1-wire bus must implement.
//
// This interface doesn't implement conn.Conn since a device address must be
// specified for each transaction. Use onewire.Dev as an adapter to get a
// conn.Conn compatible object.
type Bus interface {
	fmt.Stringer

	// Tx performs a bus transaction, sending and receiving bytes, and
	// ending by pulling the bus high either weakly or strongly depending
	// on the value of power. A strong pull-up is typically required to
	// power temperature conversion or EEPROM writes.
	Tx(w, r []byte, power Pullup) error

	// Search performs a "search" cycle on the 1-wire bus and returns the
	// addresses of all devices on the bus if alarmOnly is false and of all
	// devices in alarm state if alarmOnly is true.
	//
	// If an error occurs during the search the already-discovered devices are
	// returned with the error.
	//
	// Bus.Search may be implemented using onewire.Search if the bus implements
	// the BusSearcher interface or it may have a custom implementation.
	Search(alarmOnly bool) ([]Address, error)
}

// Address represents a 1-wire device address in little-endian format. This means
// that the family code ends up in the lower byte, the CRC in the top byte,
// and the variable address part in the middle 6 bytes. E.g. a DS18B20 device,
// which has a family code of 0x28, might have address 0x7a00000131825228.
type Address uint64

// Pullup encodes the type of pull-up used at the end of a bus transaction.
type Pullup int

const (
	WeakPullup   Pullup = iota // end transaction with weak pull-up
	StrongPullup               // end transaction with strong pull-up to power devices
)

// BusCloser is a 1-wire bus that can be closed.
//
// It is expected that an implementer of Bus also implement BusCloser, but
// this is not required.
type BusCloser interface {
	io.Closer
	Bus
}

// Pins defines the pins that a 1-wire bus interconnect is using on the host.
//
// It is expected that an implementer of Bus also implement Pins but this is
// not a requirement.
type Pins interface {
	// Q returns the 1-wire Q (data) pin
	Q() gpio.PinIO
}

//===== Errors

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
type BusError interface {
	BusError() bool // true if a bus error was detected
}

// busError implements error and BusError.
type busError string

func (e busError) Error() string  { return string(e) }
func (e busError) BusError() bool { return true }

//===== A device on a 1-wire bus

// Dev is a device on a 1-wire bus.
//
// It implements conn.Conn.
//
// Compared to Bus it saves from repeatedly specifying the device address and
// implements utility functions.
type Dev struct {
	Bus  Bus     // the bus to which the device is connected
	Addr Address // the address of the device on the bus
}

// String prints the bus name followed by the device address in parenthesis.
func (d *Dev) String() string {
	return fmt.Sprintf("%s(%d)", d.Bus, d.Addr)
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
	binary.LittleEndian.PutUint64(ww[1:], uint64(d.Addr))
	ww = append(ww, w...)
	return d.Bus.Tx(w, r, WeakPullup)
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
	binary.LittleEndian.PutUint64(ww[1:], uint64(d.Addr))
	ww = append(ww, w...)
	return d.Bus.Tx(w, r, StrongPullup)
}

//===== Bus registry

// All returns all the 1-wire buses available on this host.
func All() map[string]Opener {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]Opener, len(byName))
	for k, v := range byName {
		out[k] = v
	}
	return out
}

// New returns an open handle to a 1-wire bus.
//
// Specify busNumber -1 to get the first available bus. This is the recommended value.
func New(busNumber int) (BusCloser, error) {
	opener, err := find(busNumber)
	if err != nil {
		return nil, err
	}
	return opener()
}

// Opener opens a 1-wire bus.
type Opener func() (BusCloser, error)

// Register registers a 1-wire bus.
//
// Registering the same bus name twice is an error.
func Register(name string, busNumber int, opener Opener) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := byName[name]; ok {
		return fmt.Errorf("registering the same 1-wire bus %s twice", name)
	}
	if busNumber != -1 {
		if _, ok := byNumber[busNumber]; ok {
			return fmt.Errorf("registering the same 1-wire bus %d twice", busNumber)
		}
	}

	if first == -1 {
		first = busNumber
	}
	byName[name] = opener
	if busNumber != -1 {
		byNumber[busNumber] = opener
	}
	return nil
}

// Unregister removes a previously registered 1-wire bus.
//
// This may be necessary, for example, when a 1-wire bus is exposed via an USB device
// and the device is unplugged.
func Unregister(name string, busNumber int) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := byName[name]; !ok {
		return errors.New("unknown name")
	}
	if _, ok := byNumber[busNumber]; !ok {
		return errors.New("unknown number")
	}

	// If the first bus is being deleted, pick another one.
	if busNumber == first {
		first = -1
		for f := range byNumber {
			if f != busNumber {
				first = f
				break
			}
		}
	}

	delete(byName, name)
	delete(byNumber, busNumber)
	return nil
}

//

// find interates through registered buses and returns the one with the desired number.
func find(busNumber int) (Opener, error) {
	mu.Lock()
	defer mu.Unlock()
	if busNumber == -1 {
		if first == -1 {
			return nil, errors.New("no 1-wire bus found")
		}
		return byNumber[first], nil
	}
	bus, ok := byNumber[busNumber]
	if !ok {
		return nil, fmt.Errorf("no 1-wire bus %d", busNumber)
	}
	return bus, nil
}

var (
	mu       sync.Mutex
	byName   = map[string]Opener{}
	byNumber = map[int]Opener{}
	first    = -1
)

// Ensure that the appropriate interfaces are implemented
var _ conn.Conn = &Dev{}
