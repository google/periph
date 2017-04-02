// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spi defines the SPI protocol.
package spi

import (
	"io"

	"periph.io/x/periph/conn"
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
// It is expected to implement fmt.Stringer and optionally io.Writer and
// io.Reader.
type Conn interface {
	conn.Conn
	// DevParams sets the communication parameters of the connection for use by a
	// device.
	//
	// The device driver calls this function exactly once. It must specify the
	// maximum rated speed by the device's spec. The lowest speed between the bus
	// speed and the device speed is selected. Use 0 for maxHz if there is no
	// known maximum value for this device.
	DevParams(maxHz int64, mode Mode, bits int) error
}

// ConnCloser is a SPI bus that can be closed.
//
// This interface is meant to be handled by the application.
type ConnCloser interface {
	io.Closer
	Conn
	// Speed sets the maximum bus speed.
	//
	// It lets an application use a device at a lower speed than the maximum
	// speed as rated by the device driver. This is useful for example when the
	// wires are long or the connection is of poor quality.
	//
	// This function can be called multiple times and resets the previous value.
	// 0 is not a value value for maxHz. The lowest speed between the bus speed
	// and the device speed is selected.
	Speed(maxHz int64) error
}

// Pins defines the pins that a SPI bus interconnect is using on the host.
//
// It is expected that a implementer of ConnCloser or Conn also implement Pins
// but this is not a requirement.
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
