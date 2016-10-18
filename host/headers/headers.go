// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package headers

import (
	"fmt"
	"sync"

	"github.com/google/pio/conn/pins"
)

// All contains all the on-board headers on a micro computer.
//
// The map key is the header name, e.g. "P1" or "EULER" and the value is a
// slice of slice of pins. For a 2x20 header, it's going to be a slice of
// [20][2]pins.Pin.
func All() map[string][][]pins.Pin {
	lock.Lock()
	defer lock.Unlock()
	// TODO(maruel): Return a copy?
	return allHeaders
}

// Position returns the position on a pin if found.
//
// The header and the pin number. Pin numbers are 1-based.
func Position(p pins.Pin) (string, int) {
	lock.Lock()
	defer lock.Unlock()
	pos := byPin[p.Name()]
	return pos.name, pos.number
}

// IsConnected returns true if the pin is on a header.
func IsConnected(p pins.Pin) bool {
	lock.Lock()
	defer lock.Unlock()
	return connected[p.Name()]
}

// Register registers a physical header.
func Register(name string, pins [][]pins.Pin) error {
	lock.Lock()
	defer lock.Unlock()
	// TODO(maruel): Copy the slices?
	if _, ok := allHeaders[name]; ok {
		return fmt.Errorf("header %q was already registered", name)
	}
	for i, line := range pins {
		for j, pin := range line {
			if pin == nil || len(pin.Name()) == 0 {
				return fmt.Errorf("missing pin on header %s[%d][%d]\n", name, i+1, j+1)
			}
		}
	}

	allHeaders[name] = pins
	number := 1
	for _, line := range pins {
		for _, pin := range line {
			n := pin.Name()
			byPin[n] = position{name, number}
			connected[n] = true
			number++
		}
	}
	return nil
}

//

type position struct {
	name   string // Header name
	number int    // Pin number
}

var (
	lock       sync.Mutex
	allHeaders = map[string][][]pins.Pin{} // every known headers as per internal lookup table
	byPin      = map[string]position{}     // GPIO pin name to position
	connected  = map[string]bool{}         // GPIO pin name to position
)
