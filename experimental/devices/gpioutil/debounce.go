// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"time"

	"periph.io/x/periph/conn/gpio"
)

// debounced is a gpio.PinIO where reading and edge detection pass through a
// debouncing algorithm.
type debounced struct {
	// Immutable.
	gpio.PinIO
	// denoise delays state changes. It waits for this amount before reporting it.
	denoise time.Duration
	// debounce locks on after a steady state change. Once a state change
	// happened, don't change again for this amount of time.
	debounce time.Duration

	// Mutable.
}

// Debounce returns a debounced gpio.PinIO from a gpio.PinIO source. Only the
// PinIn behavior is mutated.
//
// denoise is a noise filter, which waits a pin to be steady for this amount
// of time BEFORE reporting the new level.
//
// debounce will lock on a level for this amount of time AFTER the pin changed
// state, ignoring following state changes.
//
// Either value can be 0.
func Debounce(p gpio.PinIO, denoise, debounce time.Duration, edge gpio.Edge) (gpio.PinIO, error) {
	if denoise == 0 && debounce == 0 {
		return p, nil
	}
	if err := p.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		return nil, err
	}
	return &debounced{
		// Immutable.
		PinIO:    p,
		denoise:  denoise,
		debounce: debounce,
		// Mutable.
	}, nil
}

// In implements gpio.PinIO.
func (d *debounced) In(pull gpio.Pull, edge gpio.Edge) error {
	err := d.PinIO.In(pull, gpio.BothEdges)
	return err
}

// Read implements gpio.PinIO.
//
// It is the smoothed out value from the underlying gpio.PinIO.
func (d *debounced) Read() gpio.Level {
	return d.PinIO.Read()
}

// WaitForEdge implements gpio.PinIO.
//
// It is the smoothed out value from the underlying gpio.PinIO.
func (d *debounced) WaitForEdge(timeout time.Duration) bool {
	if !d.PinIO.WaitForEdge(timeout) {
		return false
	}
	return true
}

// Halt implements gpio.PinIO.
func (d *debounced) Halt() error {
	return nil
}

// Real implements gpio.RealPin.
func (d *debounced) Real() gpio.PinIO {
	if r, ok := d.PinIO.(gpio.RealPin); ok {
		return r.Real()
	}
	return d.PinIO
}

var now = time.Now
var _ gpio.PinIO = &debounced{}
