// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package onewire defines a Dallas Semiconductor (now Maxim) 1-wire bus
//
// It includes an adapter to directly address a 1-wire device on a 1-wire bus
// without having to continuously specify the address when doing I/O. This
// enables the support of conn.Conn.
package onewire

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/google/periph/conn/gpio"
)

// Conn defines the function a concrete driver for a 1-wire bus must implement.
//
// This interface doesn't implement conn.Conn since a device address must be
// specified. Use onewire.Dev as an adapter to get a conn.Conn compatible
// object.
type Conn interface {
	// Tx performs a "match ROM" command on the bus, which selects at most
	// one device and then transmits and receives the specified bytes.
	Tx(addr uint64, w, r []byte) error

	// TxPup performs a "match ROM" command on the bus, which selects at most
	// one device, then transmits and receives the specified bytes, and ends
	// by turning-on a strong pull-up on the bus to power devices (such as
	// temperature sensors performing a conversion).
	TxPup(addr uint64, w, r []byte) error

	// All performs a "skip ROM" command on the bus, which selects all devices,
	// and then transmits the specified bytes.
	All(w []byte) error

	// AllPup performs a "skip ROM" command on the bus, which selects all devices,
	// then transmits the specified bytes, and ends by turning-on a strong pull-up
	// on the bus to power devices (such as temperature sensors performing a
	// conversion).
	AllPup(w []byte) error

	// Search performs a "search" cycle on the 1-wire bus and returns the
	// addresses of all devices on the bus if alarmOnly is false and of all
	// devices in alarm state if alarmOnly is true.
	Search(alarmOnly bool) ([]uint64, error)
}

// ConnCloser is a 1-wire bus that can be closed.
//
// This interface is meant to be handled by the application.
type ConnCloser interface {
	io.Closer
	Conn
}

// Pins defines the pins that a 1-wire bus interconnect is using on the host.
//
// It is expected that a implementer of Conn also implement Pins but this is
// not a requirement.
type Pins interface {
	// Q returns the 1-wire Q (data) pin
	Q() gpio.PinIO
}

// Dev is a device on a 1-wire bus.
//
// It implements conn.Conn.
//
// It saves from repeatedly specifying the device address and implements
// utility functions.
type Dev struct {
	Conn Conn
	Addr uint64
}

func (d *Dev) String() string {
	return fmt.Sprintf("%s(%d)", d.Conn, d.Addr)
}

// Tx does a transaction by adding the device's address to each command.
//
// It's a wrapper for Dev.Conn.Tx().
func (d *Dev) Tx(w, r []byte) error {
	return d.Conn.Tx(d.Addr, w, r)
}

// TxPup does a transaction ending in a string pull-up by adding the
// device's address to each command.
//
// It's a wrapper for Dev.Conn.TxPup().
func (d *Dev) TxPup(w, r []byte) error {
	return d.Conn.TxPup(d.Addr, w, r)
}

// Write writes to the 1-wire bus without reading, implementing io.Writer.
//
// It's a wrapper for Tx()
func (d *Dev) Write(b []byte) (int, error) {
	if err := d.Tx(b, nil); err != nil {
		return 0, err
	}
	return len(b), nil
}

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
func New(busNumber int) (ConnCloser, error) {
	opener, err := find(busNumber)
	if err != nil {
		return nil, err
	}
	return opener()
}

// Opener opens a 1-wire bus.
type Opener func() (ConnCloser, error)

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

	if first == nil {
		first = opener
	}
	byName[name] = opener
	if busNumber != -1 {
		byNumber[busNumber] = opener
	}
	return nil
}

// Unregister removes a previously registered 1-wire bus.
//
// This can happen, for example, when a 1-wire bus is exposed via an USB device
// and the device is unplugged.
func Unregister(name string, busNumber int) error {
	mu.Lock()
	defer mu.Unlock()

	_, ok := byName[name]
	if !ok {
		return errors.New("unknown name")
	}
	if _, ok := byNumber[busNumber]; !ok {
		return errors.New("unknown number")
	}

	delete(byName, name)
	delete(byNumber, busNumber)
	first = nil
	/* TODO(maruel): Figure out a way.
	if first == bus {
		first = nil
		last := ""
		for name, b := range byName {
			if last == "" || last > name {
				last = name
				first = b
			}
		}
	}
	*/
	return nil
}

//

func find(busNumber int) (Opener, error) {
	mu.Lock()
	defer mu.Unlock()
	if busNumber == -1 {
		if first == nil {
			return nil, errors.New("no 1-wire bus found")
		}
		return first, nil
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
	first    Opener
)
