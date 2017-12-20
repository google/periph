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

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiostream"
)

// InOpLSBF represents an expected replay StreamIn operation in PinInLSBF.
type InOpLSBF struct {
	gpio.Pull
	gpiostream.BitStreamLSBF
}

// InOpMSBF represents an expected replay StreamIn operation in PinInMSBF.
type InOpMSBF struct {
	gpio.Pull
	gpiostream.BitStreamMSBF
}

// PinInLSBF implements gpiostream.PinIn that accepts BitStreamLSBF only.
//
// Embed in a struct with gpiotest.Pin for more functionality.
type PinInLSBF struct {
	sync.Mutex
	N         string
	DontPanic bool
	Ops       []InOpLSBF
	Count     int
}

// Close verifies that all the expected Ops have been consumed.
func (p *PinInLSBF) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: expected playback to be empty: I/O count %d; expected %d", p.Count, len(p.Ops))
	}
	return nil
}

func (p *PinInLSBF) String() string {
	p.Lock()
	defer p.Unlock()
	return p.N
}

// StreamIn implements gpiostream.PinIn.
func (p *PinInLSBF) StreamIn(pull gpio.Pull, b gpiostream.Stream) error {
	s, ok := b.(*gpiostream.BitStreamLSBF)
	if !ok {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn(%t)", b)
	}
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) <= p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() (count #%d) expecting %#v", p.Count, b)
	}
	if s.Res != p.Ops[p.Count].Res {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() Res (count #%d) expected %s, got %s", p.Count, p.Ops[p.Count].Res, s.Res)
	}
	if len(s.Bits) != len(p.Ops[p.Count].Bits) {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() len(Bits) (count #%d) expected %d, got %d", p.Count, len(p.Ops[p.Count].Bits), len(s.Bits))
	}
	if pull != p.Ops[p.Count].Pull {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() pull (count #%d) expected %s, got %s", p.Count, p.Ops[p.Count].Pull, pull)
	}
	copy(s.Bits, p.Ops[p.Count].Bits)
	p.Count++
	return nil
}

// PinInMSBF implements gpiostream.PinIn that accepts BitStreamMSBF only.
//
// Embed in a struct with gpiotest.Pin for more functionality.
type PinInMSBF struct {
	sync.Mutex
	N         string
	DontPanic bool
	Ops       []InOpMSBF
	Count     int
}

// Close verifies that all the expected Ops have been consumed.
func (p *PinInMSBF) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: expected playback to be empty: I/O count %d; expected %d", p.Count, len(p.Ops))
	}
	return nil
}

func (p *PinInMSBF) String() string {
	p.Lock()
	defer p.Unlock()
	return p.N
}

// StreamIn implements gpiostream.PinIn.
func (p *PinInMSBF) StreamIn(pull gpio.Pull, b gpiostream.Stream) error {
	s, ok := b.(*gpiostream.BitStreamMSBF)
	if !ok {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn(%t)", b)
	}
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) <= p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() (count #%d) expecting %#v", p.Count, b)
	}
	if s.Res != p.Ops[p.Count].Res {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() Res (count #%d) expected %s, got %s", p.Count, p.Ops[p.Count].Res, s.Res)
	}
	if len(s.Bits) != len(p.Ops[p.Count].Bits) {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamIn() len(Bits) (count #%d) expected %d, got %d", p.Count, len(p.Ops[p.Count].Bits), len(s.Bits))
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
	sync.Mutex
	N         string
	DontPanic bool
	Ops       []gpiostream.Stream
	Count     int
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

func (p *PinOutPlayback) String() string {
	return p.N
}

// StreamOut implements gpiostream.PinOut.
func (p *PinOutPlayback) StreamOut(s gpiostream.Stream) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) <= p.Count {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamOut() (count #%d) expecting %#v", p.Count, s)
	}
	if !reflect.DeepEqual(s, p.Ops[p.Count]) {
		return errorf(p.DontPanic, "gpiostreamtest: unexpected StreamOut() content (count #%d) expected %#v, got %#v", p.Count, p.Ops[p.Count], s)
	}
	p.Count++
	return nil
}

// PinOutRecord implements gpiostream.PinOut that records operations.
//
// Embed in a struct with gpiotest.Pin for more functionality.
type PinOutRecord struct {
	sync.Mutex
	N         string
	DontPanic bool
	Ops       []gpiostream.Stream
}

func (p *PinOutRecord) String() string {
	return p.N
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
	case *gpiostream.BitStreamLSBF:
		o := &gpiostream.BitStreamLSBF{Bits: make(gpiostream.BitsLSBF, len(t.Bits)), Res: t.Res}
		copy(o.Bits, t.Bits)
		return o, nil
	case *gpiostream.BitStreamMSBF:
		o := &gpiostream.BitStreamMSBF{Bits: make(gpiostream.BitsMSBF, len(t.Bits)), Res: t.Res}
		copy(o.Bits, t.Bits)
		return o, nil
	case *gpiostream.EdgeStream:
		o := &gpiostream.EdgeStream{Edges: make([]uint16, len(t.Edges)), Res: t.Res}
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

var _ io.Closer = &PinInLSBF{}
var _ io.Closer = &PinInMSBF{}
var _ io.Closer = &PinOutPlayback{}
var _ fmt.Stringer = &PinInLSBF{}
var _ fmt.Stringer = &PinInMSBF{}
var _ fmt.Stringer = &PinOutPlayback{}
var _ fmt.Stringer = &PinOutRecord{}
var _ gpiostream.PinIn = &PinInLSBF{}
var _ gpiostream.PinIn = &PinInMSBF{}
var _ gpiostream.PinOut = &PinOutPlayback{}
var _ gpiostream.PinOut = &PinOutRecord{}
