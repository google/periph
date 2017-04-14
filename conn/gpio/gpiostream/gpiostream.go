// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpiostream defines digital streams.
package gpiostream

import (
	"sort"
	"time"
)

// Stream is the interface to define a generic stream
type Stream interface {
	// Resolution is the minimum resolution of the binary stream at which it is
	// usable.
	Resolution() time.Duration
	// Duration of the binary stream. For infinitely looping streams, it is the
	// duration of the non-looping part.
	Duration() time.Duration
}

// Bits is a densely packed bitstream. The format is LSB, bit 0 is sent first,
// up to bit 7.
//
// The stream is required to be a multiple of 8 samples.
type Bits []byte

// BitStream is a stream of bits to be written or read.
//
// This struct is useful for dense binary data, like controlling ws2812b LED
// strip or using the GPIO pin as an digital oscilloscope.
type BitStream struct {
	Bits Bits
	// The duration each bit represents.
	Res time.Duration
}

// Resolution implement Stream.
func (b *BitStream) Resolution() time.Duration {
	if len(b.Bits) == 0 {
		return 0
	}
	return b.Res
}

// Duration implement Stream.
func (b *BitStream) Duration() time.Duration {
	return b.Res * time.Duration(len(b.Bits))
}

// EdgeStream is a stream of edges to be written.
//
// This struct is more efficient than BitStream for repetitive pulses, like
// controlling a servo. A PWM can be created by specifying a slice of twice the
// same resolution and make it looping.
type EdgeStream struct {
	// Edges is the list of Level change. It is assumed that the signal starts
	// with gpio.High. Use a duration of 0 to start with a Low.
	Edges []time.Duration
	// Res is the minimum resolution at which the edges should be
	// rasterized.
	//
	// The lower the value, the more memory shall be used when rasterized.
	Res time.Duration
}

// Resolution implement Stream.
func (e *EdgeStream) Resolution() time.Duration {
	if len(e.Edges) == 0 {
		return 0
	}
	return e.Res
}

// Duration implement Stream.
func (e *EdgeStream) Duration() time.Duration {
	if e.Res == 0 {
		return 0
	}
	var t time.Duration
	for _, edge := range e.Edges {
		t += edge
	}
	return t
}

// Program is a loop of streams.
//
// This is itself a stream, it can be used to reduce memory usage when repeated
// patterns are used.
type Program struct {
	Parts []Stream // Each part must be a BitStream, EdgeStream or Program
	Loops int      // Set to -1 to create an infinite loop
}

// Resolution implement Stream.
func (p *Program) Resolution() time.Duration {
	if p.Loops == 0 {
		return 0
	}
	var rates []time.Duration
	for _, part := range p.Parts {
		if r := part.Resolution(); r != 0 {
			rates = append(rates, r)
		}
	}
	if len(rates) == 0 {
		return 0
	}
	sort.Slice(rates, func(i, j int) bool { return rates[i] < rates[j] })
	res := rates[0]
	for i := 1; i < len(rates); i++ {
		r := rates[i]
		if r > 2*res {
			break
		}
		// Take in account Nyquist rate. https://wikipedia.org/wiki/Nyquist_rate
		res /= 2
	}
	return res
}

// Duration implement Stream.
func (p *Program) Duration() time.Duration {
	if p.Loops == 0 {
		return 0
	}
	var d time.Duration
	for _, s := range p.Parts {
		d += s.Duration()
	}
	if p.Loops > 1 {
		d *= time.Duration(p.Loops)
	}
	return d
}

var _ Stream = &BitStream{}
var _ Stream = &EdgeStream{}
var _ Stream = &Program{}
