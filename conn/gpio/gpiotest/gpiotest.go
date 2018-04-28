// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpiotest is meant to be used to test drivers using fake Pins.
package gpiotest

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"periph.io/x/periph/conn/gpio"
)

// Pin implements gpio.PinIO.
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

// Halt implements conn.Resource.
//
// It has no effect.
func (p *Pin) Halt() error {
	return nil
}

// In is concurrent safe.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	p.Lock()
	defer p.Unlock()
	p.P = pull
	if pull == gpio.PullDown {
		p.L = gpio.Low
	} else if pull == gpio.PullUp {
		p.L = gpio.High
	}
	if edge != gpio.NoEdge && p.EdgesChan == nil {
		return errors.New("gpiotest: please set p.EdgesChan first")
	}
	// Flush any buffered edges.
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
		_ = p.Out(<-p.EdgesChan)
		return true
	}
	select {
	case <-time.After(timeout):
		return false
	case l := <-p.EdgesChan:
		_ = p.Out(l)
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

// PinPWM implements gpio.PinPWM.
type PinPWM struct {
	Pin

	D gpio.Duty     // duty
	P time.Duration // period
}

func (p *PinPWM) PWM(duty gpio.Duty, period time.Duration) error {
	p.Lock()
	defer p.Unlock()
	p.D = duty
	p.P = period
	return nil
}

// LogPinIO logs when its state changes.
type LogPinIO struct {
	gpio.PinIO
}

// Real implements gpio.RealPin.
func (p *LogPinIO) Real() gpio.PinIO {
	return p.PinIO
}

// In implements gpio.PinIO.
func (p *LogPinIO) In(pull gpio.Pull, edge gpio.Edge) error {
	log.Printf("%s.In(%s, %s)", p, pull, edge)
	return p.PinIO.In(pull, edge)
}

// Out mplements gpio.PinIO.
func (p *LogPinIO) Out(l gpio.Level) error {
	log.Printf("%s.Out(%s)", p, l)
	return p.PinIO.Out(l)
}

// Read implements gpio.PinIO.
func (p *LogPinIO) Read() gpio.Level {
	l := p.PinIO.Read()
	log.Printf("%s.Read() %s", p, l)
	return l
}

// Pull implements gpio.PinIO.
func (p *LogPinIO) Pull() gpio.Pull {
	log.Printf("%s.Read()", p)
	return p.PinIO.Pull()
}

// WaitForEdge implements gpio.PinIO.
func (p *LogPinIO) WaitForEdge(timeout time.Duration) bool {
	s := time.Now()
	r := p.PinIO.WaitForEdge(timeout)
	log.Printf("%s.WaitForEdge(%s) -> %t after %s", p, timeout, r, time.Since(s))
	return r
}

var _ gpio.PinIO = &Pin{}
var _ gpio.PinPWM = &PinPWM{}
