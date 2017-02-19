// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spi defines the SPI protocol.
package spi

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"periph.io/x/periph/conn/gpio"
)

// Mode determines how communication is done. The bits can be OR'ed to change
// the polarity and phase used for communication.
//
// CPOL means the clock polarity. Idle is High when set.
//
// CPHA is the clock phase, sample on trailing edge when set.
type Mode int

// Valid SPI clock and phase.
const (
	Mode0 Mode = 0x0 // CPOL=0, CPHA=0
	Mode1 Mode = 0x1 // CPOL=0, CPHA=1
	Mode2 Mode = 0x2 // CPOL=1, CPHA=0
	Mode3 Mode = 0x3 // CPOL=1, CPHA=1
)

// Conn defines the interface a concrete SPI driver must implement.
//
// It implements conn.Conn.
type Conn interface {
	fmt.Stringer
	io.Writer
	Tx(w, r []byte) error
	// Speed changes the bus speed.
	Speed(hz int64) error
	// Configure changes the communication parameters of the bus.
	Configure(mode Mode, bits int) error
}

// ConnCloser is a SPI bus that can be closed.
//
// This interface is meant to be handled by the application.
type ConnCloser interface {
	io.Closer
	Conn
}

// Pins defines the pins that a SPI bus interconnect is using on the host.
//
// It is expected that a implementer of Conn also implement Pins but this is
// not a requirement.
type Pins interface {
	// CLK returns the SCK (clock) pin.
	CLK() gpio.PinOut
	// MOSI returns the SDO (master out, slave in) pin.
	MOSI() gpio.PinOut
	// MISO returns the SDI (master in, slave out) pin.
	MISO() gpio.PinIn
	// CS returns the CSN (chip select) pin.
	CS() gpio.PinOut
}

// All returns all the SPI buses available on this host.
func All() map[string]Opener {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]Opener, len(byName))
	for k, v := range byName {
		out[k] = v
	}
	return out
}

// New returns an open handle to the bus and CS line.
//
// Specify busNumber -1 to get the first available bus and its first CS line.
// This is the recommended value.
func New(busNumber, cs int) (ConnCloser, error) {
	opener, err := find(busNumber, cs)
	if err != nil {
		return nil, err
	}
	return opener()
}

// Opener opens an SPI bus.
type Opener func() (ConnCloser, error)

// Register registers a SPI bus.
//
// Registering the same bus name twice is an error.
func Register(name string, busNumber, cs int, opener Opener) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return fmt.Errorf("spi: registering the same bus %s twice", name)
	}
	if busNumber != -1 {
		if _, ok := byNumber[busNumber]; !ok {
			byNumber[busNumber] = map[int]Opener{}
		}
		if _, ok := byNumber[busNumber][cs]; ok {
			return fmt.Errorf("spi: registering the same bus %d.%d twice", busNumber, cs)
		}
	}

	byName[name] = opener
	if busNumber != -1 {
		byNumber[busNumber][cs] = opener
	}
	return nil
}

// Unregister removes a previously registered SPI bus.
//
// This can happen when a SPI bus is exposed via an USB device and the device is
// unplugged.
func Unregister(name string, busNumber int) error {
	mu.Lock()
	defer mu.Unlock()
	_, ok := byName[name]
	if !ok {
		return fmt.Errorf("spi: unknown bus name %q", name)
	}
	if _, ok := byNumber[busNumber]; !ok {
		return fmt.Errorf("spi: unknown bus number %d", busNumber)
	}

	delete(byName, name)
	delete(byNumber, busNumber)
	return nil
}

//

func find(busNumber, cs int) (Opener, error) {
	mu.Lock()
	defer mu.Unlock()
	if len(byNumber) == 0 {
		return nil, errors.New("no SPI bus found; did you forget to call Init()?")
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
		return nil, fmt.Errorf("spi: no bus %d found", busNumber)
	}
	opener, ok := bus[cs]
	if !ok {
		return nil, fmt.Errorf("spi: no bus %d.%d found", busNumber, cs)
	}
	return opener, nil
}

var (
	mu       sync.Mutex
	byName   = map[string]Opener{}
	byNumber = map[int]map[int]Opener{}
)
