// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spi defines the SPI protocol.
//
// As described in https://periph.io/x/periph/conn#hdr-Concepts, periph.io uses
// the concepts of Bus, Port and Conn.
//
// In the package spi, 'Bus' is not exposed, as it would be SPI bus number
// without a CS line, for example on linux asking for "/dev/spi0" without the
// ".0" suffix.
//
// The OS doesn't allow that so it is counter productive to express this at the
// API layer, so 'Port' is exposed directly instead.
//
// Use Port.Connect() converts the uninitialized Port into a Conn.
package spi

import (
	"io"
	"strconv"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
)

// Mode determines how communication is done.
//
// The bits can be OR'ed to change the parameters used for
// communication.
//
type Mode int

// Mode determines the SPI communication parameters.
//
// CPOL means the clock polarity. Idle is High when set.
//
// CPHA is the clock phase, sample on trailing edge when set.
const (
	Mode0 Mode = 0x0 // CPOL=0, CPHA=0
	Mode1 Mode = 0x1 // CPOL=0, CPHA=1
	Mode2 Mode = 0x2 // CPOL=1, CPHA=0
	Mode3 Mode = 0x3 // CPOL=1, CPHA=1

	// HalfDuplex specifies that MOSI and MISO use the same wire, and that only
	// one duplex is used at a time.
	HalfDuplex Mode = 0x4
	// NoCS request the driver to not use the CS line.
	NoCS Mode = 0x8
	// LSBFirst requests the words to be encoded in little endian instead of the
	// default big endian.
	LSBFirst = 0x10
)

func (m Mode) String() string {
	s := ""
	switch m & Mode3 {
	case Mode0:
		s = "Mode0"
	case Mode1:
		s = "Mode1"
	case Mode2:
		s = "Mode2"
	case Mode3:
		s = "Mode3"
	}
	m &^= Mode3
	if m&HalfDuplex != 0 {
		s += "|HalfDuplex"
	}
	m &^= HalfDuplex
	if m&NoCS != 0 {
		s += "|NoCS"
	}
	m &^= NoCS
	if m&LSBFirst != 0 {
		s += "|LSBFirst"
	}
	m &^= LSBFirst
	if m != 0 {
		s += "|0x"
		s += strconv.FormatUint(uint64(m), 16)
	}
	return s
}

// Packet represents one packet when sending multiple packets as a transaction.
type Packet struct {
	// W and R are the output and input data. When HalfDuplex is specified to
	// Connect, only one of the two can be set.
	W, R []byte
	// BitsPerWord overrides the default bits per word value set in Connect.
	BitsPerWord uint8
	// KeepCS tells the driver to keep CS asserted after this packet is
	// completed. This can be leveraged to create long transaction as multiple
	// packets like to use 9 bits commands then 8 bits data.
	//
	// Casual observation on a Rasberry Pi 3 is that two packets with
	// KeepCS:false, there is a few µs with CS asserted after the clock stops,
	// then 11.2µs with CS not asserted, then CS is asserted for (roughly) one
	// clock cycle before the clock starts again for the next packet. This seems
	// to be independent of the port clock speed but this wasn't fully verified.
	//
	// It cannot be expected that the driver will correctly keep CS asserted even
	// if KeepCS:true on the last packet.
	//
	// KeepCS is ignored when NoCS was specified to Connect.
	KeepCS bool
}

// Conn defines the interface a concrete SPI driver must implement.
//
// It is expected to implement fmt.Stringer and optionally io.Writer and
// io.Reader.
type Conn interface {
	conn.Conn
	// TxPackets does multiple operations over the SPI connection.
	//
	// The maximum number of bytes can be limited depending on the driver. Query
	// conn.Limits.MaxTxSize() can be used to determine the limit.
	//
	// If the last packet has KeepCS:true, the behavior is undefined. The CS line
	// will likely not stay asserted. This is a driver limitation.
	TxPackets(p []Packet) error
}

// Port is the interface to be provided to device drivers.
//
// The device driver, that is the driver for the peripheral connected over
// this port, calls Connect() to retrieve a configured connection as Conn.
type Port interface {
	// Connect sets the communication parameters of the connection for use by a
	// device.
	//
	// The device driver must call this function exactly once.
	//
	// maxHz must specify the maximum rated speed by the device's spec. The lowest
	// speed between the port speed and the device speed is selected. Use 0 for
	// maxHz if there is no known maximum value for this device.
	//
	// mode specifies the clock and signal polarities, if the port is using half
	// duplex (shared MISO and MOSI) or if CS is not needed.
	//
	// bits is the number of bits per word. Generally you should use 8.
	Connect(maxHz int64, mode Mode, bits int) (Conn, error)
}

// PortCloser is a SPI port that can be closed.
//
// This interface is meant to be handled by the application.
type PortCloser interface {
	io.Closer
	Port
	// LimitSpeed sets the maximum port speed.
	//
	// It lets an application use a device at a lower speed than the maximum
	// speed as rated by the device driver. This is useful for example when the
	// wires are long or the connection is of poor quality.
	//
	// This function can be called multiple times and resets the previous value.
	// 0 is not a value value for maxHz. The lowest speed between the port speed
	// and the device speed is selected.
	LimitSpeed(maxHz int64) error
}

// Pins defines the pins that a SPI port interconnect is using on the host.
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
