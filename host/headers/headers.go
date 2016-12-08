// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package headers

import (
	"fmt"
	"sync"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/pins"
)

// All contains all the on-board headers on a micro computer.
//
// The map key is the header name, e.g. "P1" or "EULER" and the value is a
// slice of slice of pins. For a 2x20 header, it's going to be a slice of
// [20][2]pins.Pin.
func All() map[string][][]pins.Pin {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string][][]pins.Pin, len(allHeaders))
	for k, v := range allHeaders {
		outV := make([][]pins.Pin, len(v))
		for i, w := range v {
			outW := make([]pins.Pin, len(w))
			copy(outW, w)
			outV[i] = outW
		}
		out[k] = outV
	}
	return out
}

// Position returns the position on a pin if found.
//
// The header and the pin number. Pin numbers are 1-based.
//
// Returns "", 0 if not connected.
func Position(p pins.Pin) (string, int) {
	mu.Lock()
	defer mu.Unlock()
	pos, _ := byPin[realPin(p).Name()]
	return pos.name, pos.number
}

// IsConnected returns true if the pin is on a header.
func IsConnected(p pins.Pin) bool {
	_, i := Position(p)
	return i != 0
}

// Register registers a physical header.
func Register(name string, allPins [][]pins.Pin) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := allHeaders[name]; ok {
		return fmt.Errorf("headers: header %q was already registered", name)
	}
	for i, line := range allPins {
		for j, pin := range line {
			if pin == nil || len(pin.Name()) == 0 {
				return fmt.Errorf("headers: invalid pin on header %s[%d][%d]", name, i+1, j+1)
			}
		}
	}
	allHeaders[name] = allPins
	number := 1
	for _, line := range allPins {
		for _, p := range line {
			byPin[realPin(p).Name()] = position{name, number}
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
	mu         sync.Mutex
	allHeaders = map[string][][]pins.Pin{} // every known headers as per internal lookup table
	byPin      = map[string]position{}     // GPIO pin name to position
)

// realPin returns the real pin from an alias.
func realPin(p pins.Pin) pins.Pin {
	for {
		if r, ok := p.(gpio.RealPin); ok {
			p = r.Real()
		} else {
			return p
		}
	}
}
