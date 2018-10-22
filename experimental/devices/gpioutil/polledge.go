// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
)

// pollEdge is a gpio.PinIO where edge detection is done manually.
type pollEdge struct {
	// Immutable.
	gpio.PinIO
	// period is the delay between each poll.
	period time.Duration
	die    chan struct{}

	// Mutable.
	// edge is the current edge detection.
	edge gpio.Edge
}

// PollEdge returns a gpio.PinIO which implements edge detection via polling.
//
// Example of GPIOs without edge detection are GPIOs accessible over an I²C
// chip or over USB.
//
// freq must be above 0. A reasonable value is 20Hz reading. High rate
// essentially means a busy loop.
func PollEdge(p gpio.PinIO, freq physic.Frequency) gpio.PinIO {
	return &pollEdge{PinIO: p, period: freq.Duration(), die: make(chan struct{}, 1)}
}

// In implements gpio.PinIO.
func (p *pollEdge) In(pull gpio.Pull, edge gpio.Edge) error {
	p.edge = gpio.NoEdge
	err := p.PinIO.In(pull, gpio.NoEdge)
	if err == nil {
		p.edge = edge
	}
	return err
}

// WaitForEdge implements gpio.PinIO.
func (p *pollEdge) WaitForEdge(timeout time.Duration) bool {
	select {
	case <-p.die:
	default:
	}
	defer func() {
		select {
		case <-p.die:
		default:
		}
	}()
	curr := p.PinIO.Read()
	// -1 means to wait indefinitely.
	if timeout >= 0 {
		defer time.AfterFunc(timeout, func() {
			p.die <- struct{}{}
		}).Stop()
	}
	// Sadly it's not possible to stop then restart a ticker, so we can't cache
	// it in the object.
	t := time.NewTicker(p.period)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			n := p.PinIO.Read()
			if n != curr {
				switch p.edge {
				case gpio.RisingEdge:
					if n == gpio.High {
						return true
					}
					curr = n
				case gpio.FallingEdge:
					if n == gpio.Low {
						return true
					}
					curr = n
				case gpio.BothEdges:
					return true
				}
			}
		case <-p.die:
			return false
		}
	}
}

// Halt implements gpio.PinIO.
//
// It unblocks any WaitForEdge loop.
func (p *pollEdge) Halt() error {
	select {
	// If a WaitForEdge was pending, it will be unblocked.
	case p.die <- struct{}{}:
	default:
	}
	return nil
}

// Real implements gpio.RealPin.
func (p *pollEdge) Real() gpio.PinIO {
	if r, ok := p.PinIO.(gpio.RealPin); ok {
		return r.Real()
	}
	return p.PinIO
}

var _ gpio.PinIO = &pollEdge{}
