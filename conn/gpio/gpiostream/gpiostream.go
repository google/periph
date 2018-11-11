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
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
)

// Stream is the interface to define a generic stream.
type Stream interface {
	// Frequency is the minimum data rate at which the binary stream is usable.
	//
	// For example, a bit stream may have a 10kHz data rate.
	Frequency() physic.Frequency
	// Duration of the binary stream. For infinitely looping streams, it is the
	// duration of the non-looping part.
	Duration() time.Duration
}

// BitStream is a stream of bits to be written or read.
type BitStream struct {
	// Bits is a densely packed bitstream.
	//
	// The stream is required to be a multiple of 8 samples.
	Bits []byte
	// Freq is the rate at each the bit (not byte) stream should be processed.
	Freq physic.Frequency
	// LSBF when true means than Bits is in LSB-first. When false, the data is
	// MSB-first.
	//
	// With MSBF, the first bit processed is the most significant one (0x80). For
	// example, I²C, I2S PCM and SPI use MSB-first at the word level. This
	// requires to pack words correctly.
	//
	// With LSBF, the first bit processed is the least significant one (0x01).
	// For example, Ethernet uses LSB-first at the byte level and MSB-first at
	// the word level.
	LSBF bool
}

// Frequency implements Stream.
func (b *BitStream) Frequency() physic.Frequency {
	return b.Freq
}

// Duration implements Stream.
func (b *BitStream) Duration() time.Duration {
	if b.Freq == 0 {
		return 0
	}
	return b.Freq.Period() * time.Duration(len(b.Bits)*8)
}

// GoString implements fmt.GoStringer.
func (b *BitStream) GoString() string {
	return fmt.Sprintf("&gpiostream.BitStream{Bits: %x, Freq:%s, LSBF:%t}", b.Bits, b.Freq, b.LSBF)
}

// EdgeStream is a stream of edges to be written.
//
// This struct is more efficient than BitStream for short repetitive pulses,
// like controlling a servo. A PWM can be created by specifying a slice of
// twice the same resolution and make it looping via a Program.
type EdgeStream struct {
	// Edges is the list of Level change. It is assumed that the signal starts
	// with gpio.High. Use a duration of 0 for Edges[0] to start with a Low
	// instead of the default High.
	//
	// The value is a multiple of Res. Use a 0 value to 'extend' a continuous
	// signal that lasts more than "2^16-1*Res" duration by skipping a pulse.
	Edges []uint16
	// Res is the minimum resolution at which the edges should be
	// rasterized.
	//
	// The lower the value, the more memory shall be used when rasterized.
	Freq physic.Frequency
}

// Frequency implements Stream.
func (e *EdgeStream) Frequency() physic.Frequency {
	return e.Freq
}

// Duration implements Stream.
func (e *EdgeStream) Duration() time.Duration {
	if e.Freq == 0 {
		return 0
	}
	t := 0
	for _, edge := range e.Edges {
		t += int(edge)
	}
	return e.Freq.Period() * time.Duration(t)
}

// Program is a loop of streams.
//
// This is itself a stream, it can be used to reduce memory usage when repeated
// patterns are used.
type Program struct {
	Parts []Stream // Each part must be a BitStream, EdgeStream or Program
	Loops int      // Set to -1 to create an infinite loop
}

// Frequency implements Stream.
func (p *Program) Frequency() physic.Frequency {
	if p.Loops == 0 {
		return 0
	}
	var buf [16]physic.Frequency
	freqs := buf[:0]
	for _, part := range p.Parts {
		if f := part.Frequency(); f != 0 {
			freqs = insertFreq(freqs, f)
		}
	}
	if len(freqs) == 0 {
		return 0
	}
	f := freqs[0]
	for i := 1; i < len(freqs); i++ {
		if r := freqs[i]; r*2 < f {
			break
		}
		// Take in account Nyquist rate. https://wikipedia.org/wiki/Nyquist_rate
		f *= 2
	}
	return f
}

// Duration implements Stream.
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
	pin.Pin
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
	pin.Pin
	StreamOut(s Stream) error
}

//

// insertFreq inserts in reverse order, highest frequency first.
func insertFreq(l []physic.Frequency, f physic.Frequency) []physic.Frequency {
	i := search(len(l), func(i int) bool { return l[i] < f })
	l = append(l, 0)
	copy(l[i+1:], l[i:])
	l[i] = f
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

var _ Stream = &BitStream{}
var _ Stream = &EdgeStream{}
var _ Stream = &Program{}
