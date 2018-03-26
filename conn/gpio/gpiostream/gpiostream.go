// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpiostream defines digital streams.
//
// Warning
//
// This package is still in flux as development is on-going.
package gpiostream

import (
	"time"

	"periph.io/x/periph/conn/gpio"
)

// Stream is the interface to define a generic stream.
type Stream interface {
	// Resolution is the minimum resolution of the binary stream at which it is
	// usable.
	Resolution() time.Duration
	// Duration of the binary stream. For infinitely looping streams, it is the
	// duration of the non-looping part.
	Duration() time.Duration
}

// BitsLSBF is a densely packed LSB-first bitstream.
//
// The format is LSB-first, the first bit processed is the least significant
// one (0x01).
//
// For example, Ethernet uses LSB-first at the byte level and MSB-first at the
// word level.
//
// The stream is required to be a multiple of 8 samples.
type BitsLSBF []byte

// BitsMSBF is a densely packed MSB-first bitstream.
//
// The format is MSB-first, the first bit processed is the most significant one
// (0x80).
//
// For example, I²C, I2S PCM and SPI use MSB-first at the word level. This
// requires to pack words correctly.
//
// The stream is required to be a multiple of 8 samples.
type BitsMSBF []byte

// BitStreamLSBF is a stream of BitsLSBF to be written or read.
type BitStreamLSBF struct {
	Bits BitsLSBF
	// The duration each bit represents.
	Res time.Duration
}

// Resolution implement Stream.
func (b *BitStreamLSBF) Resolution() time.Duration {
	if len(b.Bits) == 0 {
		return 0
	}
	return b.Res
}

// Duration implement Stream.
func (b *BitStreamLSBF) Duration() time.Duration {
	return b.Res * time.Duration(len(b.Bits)*8)
}

// BitStreamMSBF is a stream of Bits.MSB to be written or read.
type BitStreamMSBF struct {
	Bits BitsMSBF
	// The duration each bit represents.
	Res time.Duration
}

// Resolution implement Stream.
func (b *BitStreamMSBF) Resolution() time.Duration {
	if len(b.Bits) == 0 {
		return 0
	}
	return b.Res
}

// Duration implement Stream.
func (b *BitStreamMSBF) Duration() time.Duration {
	return b.Res * time.Duration(len(b.Bits)*8)
}

//

// EdgeStream is a stream of edges to be written.
//
// This struct is more efficient than BitStreamxSB for repetitive pulses, like
// controlling a servo. A PWM can be created by specifying a slice of twice the
// same resolution and make it looping via a Program.
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
			rates = insertTime(rates, r)
		}
	}
	if len(rates) == 0 {
		return 0
	}
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

//

// PinIn allows to read a bit stream from a pin.
//
// Caveat
//
// This interface doesn't enable sampling multiple pins in a
// synchronized way or reading in a continuous uninterrupted way. As such, it
// should be considered experimental.
type PinIn interface {
	// StreamIn reads for the pin at the specified resolution to fill the
	// provided buffer.
	//
	// May only support a subset of the structs implementing Stream.
	StreamIn(p gpio.Pull, b Stream) error
}

// PinOut allows to stream to a pin.
//
// The Stream may be a Program, a BitStream or an EdgeStream. If it is a
// Program that is an infinite loop, a separate goroutine can be used to cancel
// the program. In this case StreamOut() returns without an error.
//
// Caveat
//
// This interface doesn't enable streaming to multiple pins in a
// synchronized way or reading in a continuous uninterrupted way. As such, it
// should be considered experimental.
type PinOut interface {
	StreamOut(s Stream) error
}

//

func insertTime(l []time.Duration, t time.Duration) []time.Duration {
	i := search(len(l), func(i int) bool { return l[i] > t })
	l = append(l, 0)
	copy(l[i+1:], l[i:])
	l[i] = t
	return l
}

// search implements the same algorithm as sort.Search().
//
// It was extracted to to not depend on sort, which depends on reflect.
func search(n int, f func(int) bool) int {
	lo := 0
	for hi := n; lo < hi; {
		if i := int(uint(lo+hi) >> 1); !f(i) {
			lo = i + 1
		} else {
			hi = i
		}
	}
	return lo
}

var _ Stream = &BitStreamLSBF{}
var _ Stream = &BitStreamMSBF{}
var _ Stream = &EdgeStream{}
var _ Stream = &Program{}
