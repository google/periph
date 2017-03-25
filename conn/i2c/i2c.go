// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package i2c defines an I²C bus.
//
// It includes an adapter to directly address an I²C device on a I²C bus
// without having to continuously specify the address when doing I/O. This
// enables the support of conn.Conn.
package i2c

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
)

// Bus defines the interface a concrete I²C driver must implement.
//
// This interface is consummed by a device driver for a device sitting on a bus.
//
// This interface doesn't implement conn.Conn since a device address must be
// specified. Use i2cdev.Dev as an adapter to get a conn.Conn compatible
// object.
type Bus interface {
	fmt.Stringer
	Tx(addr uint16, w, r []byte) error
	// Speed changes the bus speed, if supported.
	Speed(hz int64) error
}

// BusCloser is an I²C bus that can be closed.
//
// This interface is meant to be handled by the application.
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
	return fmt.Sprintf("%s(%d)", d.Bus, d.Addr)
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

// ReadReg writes the register number to the I²C bus, then reads data.
//

// All returns all the I²C buses available on this host.
func All() map[string]Opener {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]Opener, len(byName))
	for k, v := range byName {
		out[k] = v
	}
	return out
}

// New returns an open handle to the first available I²C bus.
//
// Specify busNumber -1 to get the first available bus. This is the recommended
// value.
func New(busNumber int) (BusCloser, error) {
	opener, err := find(busNumber)
	if err != nil {
		return nil, err
	}
	return opener()
}

// Opener opens an I²C bus.
type Opener func() (BusCloser, error)

// Register registers an I²C bus.
//
// Registering the same bus name twice is an error.
func Register(name string, busNumber int, opener Opener) error {
	if opener == nil {
		return errors.New("i2c: nil opener")
	}
	if len(name) == 0 {
		return errors.New("i2c: empty name")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return fmt.Errorf("i2c: registering the same bus %s twice", name)
	}
	if busNumber != -1 {
		if _, ok := byNumber[busNumber]; ok {
			return fmt.Errorf("i2c: registering the same bus %d twice", busNumber)
		}
	}

	byName[name] = opener
	if busNumber != -1 {
		byNumber[busNumber] = opener
	}
	return nil
}

// Unregister removes a previously registered I²C bus.
//
// This can happen when an I²C bus is exposed via an USB device and the device
// is unplugged.
func Unregister(name string, busNumber int) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; !ok {
		return fmt.Errorf("i2c: unknown bus name %q", name)
	}
	if _, ok := byNumber[busNumber]; !ok {
		return fmt.Errorf("i2c: unknown bus number %d", busNumber)
	}

	delete(byName, name)
	delete(byNumber, busNumber)
	return nil
}

//

func find(busNumber int) (Opener, error) {
	mu.Lock()
	defer mu.Unlock()
	if len(byNumber) == 0 {
		return nil, errors.New("i2c: no bus found; did you forget to call Init()?")
	}
	if busNumber == -1 {
		busNumber = int((^uint(0)) >> 1)
		for n := range byNumber {
			if busNumber > n {
				busNumber = n
			}
		}
	}
	bus, ok := byNumber[busNumber]
	if !ok {
		return nil, fmt.Errorf("i2c: no bus %d", busNumber)
	}
	return bus, nil
}

var (
	mu       sync.Mutex
	byName   = map[string]Opener{}
	byNumber = map[int]Opener{}
)

var _ conn.Conn = &Dev{}
