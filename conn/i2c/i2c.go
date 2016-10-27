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

	"github.com/google/periph/conn/gpio"
)

// Conn defines the function a concrete I²C driver must implement.
//
// This interface is consummed by a device driver for a device sitting on a bus.
//
// This interface doesn't implement conn.Conn since a device address must be
// specified. Use i2cdev.Dev as an adapter to get a conn.Conn compatible
// object.
type Conn interface {
	Tx(addr uint16, w, r []byte) error
	// Speed changes the bus speed, if supported.
	Speed(hz int64) error
}

// ConnCloser is an I²C bus that can be closed.
//
// This interface is meant to be handled by the application.
type ConnCloser interface {
	io.Closer
	Conn
}

// Pins defines the pins that an I²C bus interconnect is using on the host.
//
// It is expected that a implementer of Conn also implement Pins but this is
// not a requirement.
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
// It saves from repeatedly specifying the device address and implements
// utility functions.
type Dev struct {
	Conn Conn
	Addr uint16
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

// Write writes to the I²C bus without reading, implementing io.Writer.
//
// It's a wrapper for Tx()
func (d *Dev) Write(b []byte) (int, error) {
	if err := d.Tx(b, nil); err != nil {
		return 0, err
	}
	return len(b), nil
}

// ReadReg writes the register number to the I²C bus, then reads data.
//
// It's a wrapper for Tx()
func (d *Dev) ReadReg(reg byte, b []byte) error {
	return d.Tx([]byte{reg}, b)
}

// Shortcuts

// ReadRegUint8 reads a 8 bit register.
func (d *Dev) ReadRegUint8(reg byte) (uint8, error) {
	var v [1]byte
	err := d.ReadReg(reg, v[:])
	return v[0], err
}

// ReadRegUint16 reads a 16 bit register as big endian.
func (d *Dev) ReadRegUint16(reg byte) (uint16, error) {
	var v [2]byte
	err := d.ReadReg(reg, v[:])
	return uint16(v[0])<<8 | uint16(v[1]), err
}

// WriteRegUint8 writes a 8 bit register.
func (d *Dev) WriteRegUint8(reg byte, v uint8) error {
	_, err := d.Write([]byte{reg, v})
	return err
}

// WriteRegUint16 writes a 16 bit register.
func (d *Dev) WriteRegUint16(reg byte, v uint16) error {
	_, err := d.Write([]byte{reg, byte(v >> 8), byte(v)})
	return err
}

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
func New(busNumber int) (ConnCloser, error) {
	opener, err := find(busNumber)
	if err != nil {
		return nil, err
	}
	return opener()
}

// Opener opens an I²C bus.
type Opener func() (ConnCloser, error)

// Register registers an I²C bus.
//
// Registering the same bus name twice is an error.
func Register(name string, busNumber int, opener Opener) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return fmt.Errorf("registering the same I²C bus %s twice", name)
	}
	if busNumber != -1 {
		if _, ok := byNumber[busNumber]; ok {
			return fmt.Errorf("registering the same I²C bus %d twice", busNumber)
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

// Unregister removes a previously registered I²C bus.
//
// This can happen when an I²C bus is exposed via an USB device and the device
// is unplugged.
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
	if len(byNumber) == 0 {
		return nil, errors.New("no I²C bus found; did you forget to call Init()?")
	}
	if busNumber == -1 {
		if first == nil {
			return nil, errors.New("no I²C bus found")
		}
		return first, nil
	}
	bus, ok := byNumber[busNumber]
	if !ok {
		return nil, fmt.Errorf("no I²C bus %d", busNumber)
	}
	return bus, nil
}

var (
	mu       sync.Mutex
	byName   = map[string]Opener{}
	byNumber = map[int]Opener{}
	first    Opener
)
