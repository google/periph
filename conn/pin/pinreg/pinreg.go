// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pinreg

import (
	"errors"
	"strconv"
	"sync"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin"
)

// All contains all the on-board headers on a micro computer.
//
// The map key is the header name, e.g. "P1" or "EULER" and the value is a
// slice of slice of pin.Pin. For a 2x20 header, it's going to be a slice of
// [20][2]pin.Pin.
func All() map[string][][]pin.Pin {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string][][]pin.Pin, len(allHeaders))
	for k, v := range allHeaders {
		outV := make([][]pin.Pin, len(v))
		for i, w := range v {
			outW := make([]pin.Pin, len(w))
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
func Position(p pin.Pin) (string, int) {
	mu.Lock()
	defer mu.Unlock()
	pos := byPin[realPin(p).Name()]
	return pos.name, pos.number
}

// IsConnected returns true if the pin is on a header.
func IsConnected(p pin.Pin) bool {
	_, i := Position(p)
	return i != 0
}

// Register registers a physical header.
//
// It automatically registers all gpio pins to gpioreg.
func Register(name string, allPins [][]pin.Pin) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := allHeaders[name]; ok {
		return errors.New("pinreg: header " + strconv.Quote(name) + " was already registered")
	}
	for i, line := range allPins {
		for j, pin := range line {
			if pin == nil || len(pin.Name()) == 0 {
				return errors.New("pinreg: invalid pin on header " + name + "[" + strconv.Itoa(i+1) + "][" + strconv.Itoa(j+1) + "]")
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

	count := 0
	for _, row := range allPins {
		for _, p := range row {
			count++
			if _, ok := p.(gpio.PinIO); ok {
				if err := gpioreg.RegisterAlias(name+"_"+strconv.Itoa(count), p.Name()); err != nil {
					// Unregister as much as possible.
					_ = unregister(name)
					return errors.New("pinreg: " + err.Error())
				}
			}
		}
	}

	return nil
}

// Unregister removes a previously registered header.
//
// This can happen when an USB device, which exposed an header, is unplugged.
// This is also useful for unit testing.
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()
	return unregister(name)
}

//

type position struct {
	name   string // Header name
	number int    // Pin number
}

var (
	mu         sync.Mutex
	allHeaders = map[string][][]pin.Pin{} // every known headers as per internal lookup table
	byPin      = map[string]position{}    // GPIO pin name to position
)

func unregister(name string) error {
	if hdr, ok := allHeaders[name]; ok {
		var err error
		delete(allHeaders, name)
		count := 0
		for _, row := range hdr {
			for _, p := range row {
				count++
				if _, ok := p.(gpio.PinIO); ok {
					if err1 := gpioreg.Unregister(name + "_" + strconv.Itoa(count)); err1 != nil && err == nil {
						// Continue unregistering as much as possible.
						err = errors.New("pinreg: " + err1.Error())
					}
				}
			}
		}
		return err
	}
	return errors.New("pinreg: can't unregister unknown header name " + strconv.Quote(name))
}

// realPin returns the real pin from an alias.
func realPin(p pin.Pin) pin.Pin {
	if r, ok := p.(gpio.RealPin); ok {
		p = r.Real()
	}
	return p
}
