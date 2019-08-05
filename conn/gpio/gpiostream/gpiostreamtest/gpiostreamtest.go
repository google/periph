// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpiostreamtest enables testing device driver using gpiostream.PinIn
// or PinOut.
package gpiostreamtest

import (
	"fmt"
	"io"
	"reflect"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/pin"
)

// InOp represents an expected replay StreamIn operation in PinIn.
type InOp struct {
	gpio.Pull
	gpiostream.BitStream
}

// PinIn implements gpiostream.PinIn that accepts BitStream only.
//
// Embed in a struct with gpiotest.Pin for more functionality.
type PinIn struct {
	// These should be immutable.
	N         string
	DontPanic bool

	// Grab the Mutex before accessing the following members.
	sync.Mutex
	Ops   []InOp
	Count int
}

// Close verifies that all the expected Ops have been consumed.
func (p *PinIn) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: expected playback to be empty: I/O count %d; expected %d", p.Count, len(p.Ops))
	}
	return nil
}

// String implements conn.Resource.
func (p *PinIn) String() string {
	return p.N
}

// Name implements pin.Pin.
func (p *PinIn) Name() string {
	return p.N
}

// Number implements pin.Pin.
func (p *PinIn) Number() int {
	return -1
}

// Function implements pin.Pin.
func (p *PinIn) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *PinIn) Func() pin.Func {
	return gpio.IN
}

// SupportedFuncs implements pin.PinFunc.
func (p *PinIn) SupportedFuncs() []pin.Func {
	return []pin.Func{gpio.IN}
}

// SetFunc implements pin.PinFunc.
func (p *PinIn) SetFunc(f pin.Func) error {
	if f == gpio.IN {
		return nil
	}
	return errorf(p.DontPanic, "gpiostreamtest: not supported")
}

// Halt implements conn.Resource.
func (p *PinIn) Halt() error {
	return nil
}

// StreamIn implements gpiostream.PinIn.
func (p *PinIn) StreamIn(pull gpio.Pull, b gpiostream.Stream) error {
	s, ok := b.(*gpiostream.BitStream)
	if !ok {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn(%t)", b)
	}
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) <= p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() (count #%d) expecting %#v", p.Count, b)
	}
	if s.Freq != p.Ops[p.Count].Freq {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() Freq (count #%d) expected %s, got %s", p.Count, p.Ops[p.Count].Freq, s.Freq)
	}
	if len(s.Bits) != len(p.Ops[p.Count].Bits) {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() len(Bits) (count #%d) expected %d, got %d", p.Count, len(p.Ops[p.Count].Bits), len(s.Bits))
	}
	if s.LSBF != p.Ops[p.Count].LSBF {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() LSBF (count #%d) expected %t, got %t", p.Count, p.Ops[p.Count].LSBF, s.LSBF)
	}
	if pull != p.Ops[p.Count].Pull {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() pull (count #%d) expected %s, got %s", p.Count, p.Ops[p.Count].Pull, pull)
	}
	copy(s.Bits, p.Ops[p.Count].Bits)
	p.Count++
	return nil
}

// PinOutPlayback implements gpiostream.PinOut.
//
// Embed in a struct with gpiotest.Pin for more functionality.
type PinOutPlayback struct {
	// These should be immutable.
	N         string
	DontPanic bool

	// Grab the Mutex before accessing the following members.
	sync.Mutex
	Ops   []gpiostream.Stream
	Count int
}

// Close verifies that all the expected Ops have been consumed.
func (p *PinOutPlayback) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: expected playback to be empty: I/O count %d; expected %d", p.Count, len(p.Ops))
	}
	return nil
}

// String implements conn.Resource.
func (p *PinOutPlayback) String() string {
	return p.N
}

// Name implements pin.Pin.
func (p *PinOutPlayback) Name() string {
	return p.N
}

// Number implements pin.Pin.
func (p *PinOutPlayback) Number() int {
	return -1
}

// Function implements pin.Pin.
func (p *PinOutPlayback) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *PinOutPlayback) Func() pin.Func {
	return gpio.OUT
}

// SupportedFuncs implements pin.PinFunc.
func (p *PinOutPlayback) SupportedFuncs() []pin.Func {
	return []pin.Func{gpio.OUT}
}

// SetFunc implements pin.PinFunc.
func (p *PinOutPlayback) SetFunc(f pin.Func) error {
	if f == gpio.OUT {
		return nil
	}
	return errorf(p.DontPanic, "gpiostreamtest: not supported")
}

// Halt implements conn.Resource.
func (p *PinOutPlayback) Halt() error {
	return nil
}

// StreamOut implements gpiostream.PinOut.
func (p *PinOutPlayback) StreamOut(s gpiostream.Stream) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) <= p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamOut() (count #%d) expecting %#v", p.Count, s)
	}
	if !reflect.DeepEqual(s, p.Ops[p.Count]) {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamOut() content (count #%d)\nexpected: %#v\ngot:      %#v", p.Count, p.Ops[p.Count], s)
	}
	p.Count++
	return nil
}

// PinOutRecord implements gpiostream.PinOut that records operations.
//
// Embed in a struct with gpiotest.Pin for more functionality.
type PinOutRecord struct {
	// These should be immutable.
	N         string
	DontPanic bool

	// Grab the Mutex before accessing the following members.
	sync.Mutex
	Ops []gpiostream.Stream
}

// String implements conn.Resource.
func (p *PinOutRecord) String() string {
	return p.N
}

// Name implements pin.Pin.
func (p *PinOutRecord) Name() string {
	return p.N
}

// Number implements pin.Pin.
func (p *PinOutRecord) Number() int {
	return -1
}

// Function implements pin.Pin.
func (p *PinOutRecord) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *PinOutRecord) Func() pin.Func {
	return gpio.OUT
}

// SupportedFuncs implements pin.PinFunc.
func (p *PinOutRecord) SupportedFuncs() []pin.Func {
	return []pin.Func{gpio.OUT}
}

// SetFunc implements pin.PinFunc.
func (p *PinOutRecord) SetFunc(f pin.Func) error {
	if f == gpio.OUT {
		return nil
	}
	return errorf(p.DontPanic, "gpiostreamtest: not supported")
}

// Halt implements conn.Resource.
func (p *PinOutRecord) Halt() error {
	return nil
}

// StreamOut implements gpiostream.PinOut.
func (p *PinOutRecord) StreamOut(s gpiostream.Stream) error {
	p.Lock()
	defer p.Unlock()
	d, err := deepCopy(s)
	if err != nil {
		return errorf(p.DontPanic, "gpiostreamtest: %s", err)
	}
	p.Ops = append(p.Ops, d)
	return nil
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

func deepCopy(s gpiostream.Stream) (gpiostream.Stream, error) {
	switch t := s.(type) {
	case *gpiostream.BitStream:
		o := &gpiostream.BitStream{Bits: make([]byte, len(t.Bits)), Freq: t.Freq, LSBF: t.LSBF}
		copy(o.Bits, t.Bits)
		return o, nil
	case *gpiostream.EdgeStream:
		o := &gpiostream.EdgeStream{Edges: make([]uint16, len(t.Edges)), Freq: t.Freq}
		copy(o.Edges, t.Edges)
		return o, nil
	case *gpiostream.Program:
		o := &gpiostream.Program{Loops: t.Loops}
		for _, p := range t.Parts {
			x, err := deepCopy(p)
			if err != nil {
				return nil, err
			}
			o.Parts = append(o.Parts, x)
		}
		return o, nil
	default:
		return nil, fmt.Errorf("invalid type %T", s)
	}
}

var _ io.Closer = &PinIn{}
var _ io.Closer = &PinOutPlayback{}
var _ conn.Resource = &PinIn{}
var _ conn.Resource = &PinOutPlayback{}
var _ conn.Resource = &PinOutRecord{}
var _ gpiostream.PinIn = &PinIn{}
var _ gpiostream.PinOut = &PinOutPlayback{}
var _ gpiostream.PinOut = &PinOutRecord{}
