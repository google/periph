// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package uart defines the UART protocol.
package uart

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/google/periph/conn/gpio"
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
	fmt.Stringer
	io.Writer
	Tx(w, r []byte) error
	// Speed changes the bus speed.
	Speed(baud int64) error
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

// All returns all the UART buses available on this host.
func All() map[string]Opener {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]Opener, len(byName))
	for k, v := range byName {
		out[k] = v
	}
	return out
}

// New returns an open handle to the UART bus.
//
// Specify busNumber -1 to get the first available bus. This is the
// recommended value.
func New(busNumber int) (ConnCloser, error) {
	opener, err := find(busNumber)
	if err != nil {
		return nil, err
	}
	return opener()
}

// Opener opens an UART bus.
type Opener func() (ConnCloser, error)

// Register registers an UART bus.
func Register(name string, busNumber int, opener Opener) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return fmt.Errorf("uart: registering the same port %s twice", name)
	}
	if busNumber != -1 {
		if _, ok := byNumber[busNumber]; ok {
			return fmt.Errorf("uart: registering the same port #%d twice", busNumber)
		}
	}

	byName[name] = opener
	if busNumber != -1 {
		byNumber[busNumber] = opener
	}
	return nil
}

// Unregister removes a previously registered UART bus.
//
// This can happen when an UART bus is exposed via an USB device and the device
// is unplugged.
func Unregister(name string, busNumber int) error {
	mu.Lock()
	defer mu.Unlock()
	_, ok := byName[name]
	if !ok {
		return fmt.Errorf("uart: unknown port name %s", name)
	}
	if _, ok := byNumber[busNumber]; !ok {
		return fmt.Errorf("uart: unknown port number %d", busNumber)
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
		return nil, errors.New("uart: no port found; did you forget to call Init()?")
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
		return nil, fmt.Errorf("uart: no port %d found", busNumber)
	}
	return bus, nil
}

var (
	mu       sync.Mutex
	byName   = map[string]Opener{}
	byNumber = map[int]Opener{}
)
