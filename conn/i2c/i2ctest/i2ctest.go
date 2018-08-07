// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package i2ctest is meant to be used to test drivers over a fake I²C bus.
package i2ctest

import (
	"bytes"
	"sync"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

// IO registers the I/O that happened on either a real or fake I²C bus.
type IO struct {
	Addr uint16
	W    []byte
	R    []byte
}

// Record implements i2c.Bus that records everything written to it.
//
// This can then be used to feed to Playback to do "replay" based unit tests.
//
// Record doesn't implement i2c.BusCloser on purpose.
type Record struct {
	sync.Mutex
	Bus i2c.Bus // Bus can be nil if only writes are being recorded.
	Ops []IO
}

func (r *Record) String() string {
	return "record"
}

// Tx implements i2c.Bus
func (r *Record) Tx(addr uint16, w, read []byte) error {
	io := IO{Addr: addr}
	if len(w) != 0 {
		io.W = make([]byte, len(w))
		copy(io.W, w)
	}
	r.Lock()
	defer r.Unlock()
	if r.Bus == nil {
		if len(read) != 0 {
			return conntest.Errorf("i2ctest: read unsupported when no bus is connected")
		}
	} else {
		if err := r.Bus.Tx(addr, w, read); err != nil {
			return err
		}
	}
	if len(read) != 0 {
		io.R = make([]byte, len(read))
		copy(io.R, read)
	}
	r.Ops = append(r.Ops, io)
	return nil
}

// SetSpeed implements i2c.Bus.
func (r *Record) SetSpeed(f physic.Frequency) error {
	if r.Bus != nil {
		return r.Bus.SetSpeed(f)
	}
	return nil
}

// SCL implements i2c.Pins.
func (r *Record) SCL() gpio.PinIO {
	if p, ok := r.Bus.(i2c.Pins); ok {
		return p.SCL()
	}
	return gpio.INVALID
}

// SDA implements i2c.Pins.
func (r *Record) SDA() gpio.PinIO {
	if p, ok := r.Bus.(i2c.Pins); ok {
		return p.SDA()
	}
	return gpio.INVALID
}

// Playback implements i2c.Bus and plays back a recorded I/O flow.
//
// While "replay" type of unit tests are of limited value, they still present
// an easy way to do basic code coverage.
//
// Set DontPanic to true to return an error instead of panicking, which is the
// default.
type Playback struct {
	sync.Mutex
	Ops       []IO
	Count     int
	DontPanic bool
	SDAPin    gpio.PinIO
	SCLPin    gpio.PinIO
}

func (p *Playback) String() string {
	return "playback"
}

// Close implements i2c.BusCloser.
//
// Close() verifies that all the expected Ops have been consumed.
func (p *Playback) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != p.Count {
		return errorf(p.DontPanic, "i2ctest: expected playback to be empty: I/O count %d; expected %d", p.Count, len(p.Ops))
	}
	return nil
}

// Tx implements i2c.Bus.
func (p *Playback) Tx(addr uint16, w, r []byte) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) <= p.Count {
		return errorf(p.DontPanic, "i2ctest: unexpected Tx() (count #%d) expecting i2ctest.IO{Addr:%d, W:%#v, R:%#v}", p.Count, addr, w, r)
	}
	if addr != p.Ops[p.Count].Addr {
		return errorf(p.DontPanic, "i2ctest: unexpected addr (count #%d) %d != %d", p.Count, addr, p.Ops[p.Count].Addr)
	}
	if !bytes.Equal(p.Ops[p.Count].W, w) {
		return errorf(p.DontPanic, "i2ctest: unexpected write (count #%d) %#v != %#v", p.Count, w, p.Ops[p.Count].W)
	}
	if len(p.Ops[p.Count].R) != len(r) {
		return errorf(p.DontPanic, "i2ctest: unexpected read buffer length (count #%d) %d != %d", p.Count, len(r), len(p.Ops[p.Count].R))
	}
	copy(r, p.Ops[p.Count].R)
	p.Count++
	return nil
}

// SetSpeed implements i2c.Bus.
func (p *Playback) SetSpeed(f physic.Frequency) error {
	return nil
}

// SCL implements i2c.Pins.
func (p *Playback) SCL() gpio.PinIO {
	return p.SCLPin
}

// SDA implements i2c.Pins.
func (p *Playback) SDA() gpio.PinIO {
	return p.SDAPin
}

//

// errorf is the internal implementation that optionally panic.
//
// If dontPanic is false, it panics instead.
func errorf(dontPanic bool, format string, a ...interface{}) error {
	err := conntest.Errorf(format, a...)
	if !dontPanic {
		panic(err)
	}
	return err
}

var _ i2c.Bus = &Record{}
var _ i2c.Pins = &Record{}
var _ i2c.Bus = &Playback{}
var _ i2c.Pins = &Playback{}
