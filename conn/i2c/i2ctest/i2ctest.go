// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package i2ctest is meant to be used to test drivers over a fake I²C bus.
package i2ctest

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/i2c"
)

// IO registers the I/O that happened on either a real or fake I²C bus.
type IO struct {
	Addr  uint16
	Write []byte
	Read  []byte
}

// Record implements i2c.Conn that records everything written to it.
//
// This can then be used to feed to Playback to do "replay" based unit tests.
type Record struct {
	sync.Mutex
	Conn i2c.Conn // Conn can be nil if only writes are being recorded.
	Ops  []IO
}

func (r *Record) String() string {
	return "record"
}

// Tx implements i2c.Conn.
func (r *Record) Tx(addr uint16, w, read []byte) error {
	r.Lock()
	defer r.Unlock()
	if r.Conn == nil {
		if len(read) != 0 {
			return errors.New("read unsupported when no bus is connected")
		}
	} else {
		if err := r.Conn.Tx(addr, w, read); err != nil {
			return err
		}
	}
	io := IO{Addr: addr, Write: make([]byte, len(w))}
	if len(read) != 0 {
		io.Read = make([]byte, len(read))
	}
	copy(io.Write, w)
	copy(io.Read, read)
	r.Ops = append(r.Ops, io)
	return nil
}

// Speed implements i2c.Conn.
func (r *Record) Speed(hz int64) error {
	if r.Conn != nil {
		return r.Conn.Speed(hz)
	}
	return nil
}

// SCL implements i2c.Pins.
func (r *Record) SCL() gpio.PinIO {
	if p, ok := r.Conn.(i2c.Pins); ok {
		return p.SCL()
	}
	return gpio.INVALID
}

// SDA implements i2c.Pins.
func (r *Record) SDA() gpio.PinIO {
	if p, ok := r.Conn.(i2c.Pins); ok {
		return p.SDA()
	}
	return gpio.INVALID
}

// Playback implements i2c.Conn and plays back a recorded I/O flow.
//
// While "replay" type of unit tests are of limited value, they still present
// an easy way to do basic code coverage.
type Playback struct {
	sync.Mutex
	Ops []IO
}

func (p *Playback) String() string {
	return "playback"
}

// Close implements i2c.ConnCloser.
func (p *Playback) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != 0 {
		return fmt.Errorf("expected playback to be empty:\n%#v", p.Ops)
	}
	return nil
}

// Tx implements i2c.Conn.
func (p *Playback) Tx(addr uint16, w, r []byte) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) == 0 {
		// log.Fatal() ?
		return errors.New("unexpected Tx()")
	}
	if addr != p.Ops[0].Addr {
		return fmt.Errorf("unexpected addr %d != %d", addr, p.Ops[0].Addr)
	}
	if !bytes.Equal(p.Ops[0].Write, w) {
		return fmt.Errorf("unexpected write %#v != %#v", w, p.Ops[0].Write)
	}
	if len(p.Ops[0].Read) != len(r) {
		return fmt.Errorf("unexpected read buffer length %d != %d", len(r), len(p.Ops[0].Read))
	}
	copy(r, p.Ops[0].Read)
	p.Ops = p.Ops[1:]
	return nil
}

// Speed implements i2c.Conn.
func (p *Playback) Speed(hz int64) error {
	return nil
}

var _ i2c.Conn = &Record{}
var _ i2c.Pins = &Record{}
var _ i2c.Conn = &Playback{}
