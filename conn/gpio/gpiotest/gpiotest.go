// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpiotest is meant to be used to test drivers using fake Pins.
package gpiotest

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/periph/conn/gpio"
)

// Pin implements gpio.Pin.
//
// Modify its members to simulate hardware events.
type Pin struct {
	N   string // Should be immutable
	Num int    // Should be immutable
	Fn  string // Should be immutable

	sync.Mutex            // Grab the Mutex before modifying the members to keep it concurrent safe
	L          gpio.Level // Used for both input and output
	P          gpio.Pull
	EdgesChan  chan gpio.Level // Use it to fake edges
}

func (p *Pin) String() string {
	return fmt.Sprintf("%s(%d)", p.N, p.Num)
}

// Name returns the name of the pin.
func (p *Pin) Name() string {
	return p.N
}

// Number returns the pin number.
func (p *Pin) Number() int {
	return p.Num
}

// Function return the value of the Fn field of the pin.
func (p *Pin) Function() string {
	return p.Fn
}

// In is concurrent safe.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	p.Lock()
	defer p.Unlock()
	p.P = pull
	if pull == gpio.Down {
		p.L = gpio.Low
	} else if pull == gpio.Up {
		p.L = gpio.High
	}
	if edge != gpio.None && p.EdgesChan == nil {
		return errors.New("gpiotest: please set p.EdgesChan first")
	}
	for {
		select {
		case <-p.EdgesChan:
		default:
			return nil
		}
	}
}

// Read is concurrent safe.
func (p *Pin) Read() gpio.Level {
	p.Lock()
	defer p.Unlock()
	return p.L
}

// WaitForEdge implements gpio.PinIn.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	if timeout == -1 {
		p.Out(<-p.EdgesChan)
		return true
	}
	select {
	case <-time.After(timeout):
		return false
	case l := <-p.EdgesChan:
		p.Out(l)
		return true
	}
}

// Pull implements gpio.PinIn.
func (p *Pin) Pull() gpio.Pull {
	return p.P
}

// Out is concurrent safe.
func (p *Pin) Out(l gpio.Level) error {
	p.Lock()
	defer p.Unlock()
	p.L = l
	return nil
}

// PWM implements gpio.PinOut.
func (p *Pin) PWM(duty int) error {
	return errors.New("gpiotest: pwm is not implemented")
}

var _ gpio.PinIO = &Pin{}
