// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package uart defines the UART protocol.
package uart

import (
	"io"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
)

// Parity determines the parity bit when transmitting, if any.
type Parity byte

const (
	// None means no parity bit
	None Parity = 'N'
	// Odd means 1 when sum is odd
	Odd Parity = 'O'
	// Even means 1 when sum is even
	Even Parity = 'E'
	// Mark means always 1
	Mark Parity = 'M'
	// Space means always 0
	Space Parity = 'S'
)

// Stop determines what stop bit to use.
type Stop int8

const (
	// One is 1 stop bit
	One Stop = 0
	// OneHalf is 1.5 stop bits
	OneHalf Stop = 1
	// Two is 2 stop bits
	Two Stop = 2
)

// Conn defines the interface a concrete UART driver must implement.
//
// It implements conn.Conn.
type Conn interface {
	conn.Conn
	// Speed changes the bus speed.
	Speed(baud int) error
	// Configure changes the communication parameters of the bus.
	//
	// There's rarely a reason to use anything else than One stop bit and 8 bits
	// per character.
	Configure(stopBit Stop, parity Parity, bits int) error
}

// ConnCloser is a connection that can be closed.
type ConnCloser interface {
	io.Closer
	Conn
}

// Pins defines the pins that an UART bus interconnect is using on the host.
//
// It is expected that a implementer of Conn also implement Pins but this is
// not a requirement.
type Pins interface {
	// RX returns the receive pin.
	RX() gpio.PinIn
	// TX returns the transmit pin.
	TX() gpio.PinOut
	// RTS returns the request to send pin.
	RTS() gpio.PinIO
	// CTS returns the clear to send pin.
	CTS() gpio.PinIO
}
