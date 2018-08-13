// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package uart defines the UART protocol.
//
// As described in https://periph.io/x/periph/conn#hdr-Concepts, periph.io uses
// the concepts of Bus, Port and Conn.
//
// In the package uart, 'Bus' is not exposed, as the protocol is primarily
// point-to-point.
//
// Use Port.Connect() converts the uninitialized Port into a Conn.
//
// TODO(maruel): The Port -> Conn dance is unusual for UART users and feels
// unnatural.
//
// TODO(maruel): UART users talks in term of bauds, not hertz.
//
// TODO(maruel): The LimitSpeed() function feels weird, as generally it's not
// the device driver that gets to decide the speed (?) Well it 'depends'.
//
// There's a great implementation at https://github.com/albenik/go-serial but
// it uses cgo a lot. Maybe making an adaptor and moving this into extra is the
// best choice here?
package uart

import (
	"fmt"
	"io"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
)

// Flow determines the data flow to use, if any.
type Flow uint32

const (
	// NoFlow specifies that no flow control is used.
	NoFlow Flow = 0x10000
	// XOnXOff specifies XOn/XOff flow control, also called Software flow control.
	//
	// See https://en.wikipedia.org/wiki/Software_flow_control for more
	// information.
	XOnXOff Flow = 0x20000
	// RTSCTS specifies RTS/CTS flow contro. This uses RTS and CTS lines for flow
	// control, also called Hardware flow control. This enables more reliable
	// communication. The lines are driven Low when they are ready to receive
	// more data.
	RTSCTS Flow = 0x40000

	mask Flow = 0xFFFF0000
)

// MakeXOnXOffFlow returns an initialized Flow to enable software based flow
// control.
func MakeXOnXOffFlow(xon, xoff byte) Flow {
	return XOnXOff | Flow(xon)<<8 | Flow(xoff)
}

func (f Flow) String() string {
	switch f {
	case NoFlow:
		return "None"
	case RTSCTS:
		return "RTS/CTS"
	default:
		if f&mask == XOnXOff {
			return fmt.Sprintf("XOn(%c)/XOff(%c)", byte(f>>8), byte(f))
		}
		return fmt.Sprintf("Flow(%x)", uint32(f))
	}
}

// Parity determines the parity bit when transmitting, if any.
type Parity byte

const (
	// NoParity means no parity bit.
	NoParity Parity = 'N'
	// Odd means 1 when sum is odd.
	Odd Parity = 'O'
	// Even means 1 when sum is even.
	Even Parity = 'E'
	// Mark means always 1.
	Mark Parity = 'M'
	// Space means always 0.
	Space Parity = 'S'
)

// Stop determines what stop bit to use.
type Stop int8

const (
	// One is 1 stop bit.
	One Stop = 1
	// OneHalf is 1.5 stop bits.
	OneHalf Stop = 15
	// Two is 2 stop bits.
	Two Stop = 2
)

// Port is the interface to be provided to device drivers.
//
// The device driver, that is the driver for the peripheral connected over
// this port, calls Connect() to retrieve a configured connection as Conn.
type Port interface {
	String() string
	// Connect sets the communication parameters of the connection for use by a
	// device.
	//
	// The device driver must call this function exactly once.
	//
	// f must specify the maximum rated speed by the device's spec. For example
	// if a device is known to not work at over 115200 bauds, it should specify
	// 115200Hz.
	//
	// The lowest speed between the port speed and the device speed is selected.
	//
	// There's rarely a reason to use anything else than One stop bit and 8 bits
	// per character.
	Connect(f physic.Frequency, stopBit Stop, parity Parity, flow Flow, bits int) (conn.Conn, error)
}

// PortCloser is a UART port that can be closed.
//
// This interface is meant to be handled by the application.
type PortCloser interface {
	io.Closer
	Port
	// LimitSpeed sets the maximum port speed.
	//
	// It lets an application use a device at a lower speed than the maximum
	// speed as rated by the device driver. This is useful for example when the
	// wires are long or the connection is of poor quality, and you want to try
	// to run at lower speed like 19200 bauds.
	//
	// This function can be called multiple times and resets the previous value.
	// 0 is not a valid value for f. The lowest speed between the port speed and
	// the device speed is selected.
	LimitSpeed(f physic.Frequency) error
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
	// RTS returns the request to send pin, if present.
	RTS() gpio.PinOut
	// CTS returns the clear to send pin, if present.
	CTS() gpio.PinIn
}
