// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package onewiretest is meant to be used to test drivers over a fake 1-wire bus.
package onewiretest

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/experimental/conn/onewire"
)

// IO registers the I/O that happened on either a real or fake 1-wire bus.
type IO struct {
	Write []byte
	Read  []byte
	Pull  onewire.Pullup
}

// Record implements onewire.Bus that records everything written to it.
//
// This can then be used to feed to Playback to do "replay" based unit tests.
type Record struct {
	sync.Mutex
	Bus onewire.Bus // Bus can be nil if only writes are being recorded.
	Ops []IO
}

func (r *Record) String() string {
	return "record"
}

// Tx implements onewire.Bus.
func (r *Record) Tx(w, read []byte, pull onewire.Pullup) error {
	r.Lock()
	defer r.Unlock()
	if r.Bus == nil {
		if len(read) != 0 {
			return errors.New("read unsupported when no bus is connected")
		}
	} else {
		if err := r.Bus.Tx(w, read, pull); err != nil {
			return err
		}
	}
	io := IO{Write: make([]byte, len(w)), Pull: pull}
	if len(read) != 0 {
		io.Read = make([]byte, len(read))
	}
	copy(io.Write, w)
	copy(io.Read, read)
	r.Ops = append(r.Ops, io)
	return nil
}

// Q implements onewire.Pins.
func (r *Record) Q() gpio.PinIO {
	if p, ok := r.Bus.(onewire.Pins); ok {
		return p.Q()
	}
	return gpio.INVALID
}

// Search implements onewire.Bus
func (r *Record) Search(alarmOnly bool) ([]onewire.Address, error) {
	return nil, nil
}

// Playback implements onewire.Bus and plays back a recorded I/O flow.
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

// Close implements onewire.BusCloser.
func (p *Playback) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != 0 {
		return fmt.Errorf("expected playback to be empty:\n%#v", p.Ops)
	}
	return nil
}

// Tx implements onewire.Bus.
func (p *Playback) Tx(w, r []byte, pull onewire.Pullup) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) == 0 {
		// log.Fatal() ?
		return errors.New("unexpected Tx()")
	}
	if !bytes.Equal(p.Ops[0].Write, w) {
		return fmt.Errorf("unexpected write %#v != %#v", w, p.Ops[0].Write)
	}
	if len(p.Ops[0].Read) != len(r) {
		return fmt.Errorf("unexpected read buffer length %d != %d", len(r), len(p.Ops[0].Read))
	}
	if pull != p.Ops[0].Pull {
		return fmt.Errorf("unexpected pullup %d != %d", pull, p.Ops[0].Pull)
	}
	copy(r, p.Ops[0].Read)
	p.Ops = p.Ops[1:]
	return nil
}

// Search implements onewire.Bus
func (p *Playback) Search(alarmOnly bool) ([]onewire.Address, error) {
	return nil, nil
}

var _ onewire.Bus = &Record{}
var _ onewire.Pins = &Record{}
var _ onewire.Bus = &Playback{}
