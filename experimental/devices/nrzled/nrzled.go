// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package nrzled

import (
	"errors"
	"fmt"
	"image"
	"image/color"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices"
)

// NRZ converts a byte into the MSB-first Non-Return-to-Zero encoded 24 bits.
//
// The upper 8 bits are zeros and shall be ignored.
//
// The Non-return-to-zero protocol is a self-clocking signal that enables
// one-way communication without the need of a dedicated clock signal, unlike
// SPI driven LEDs like the apa102.
//
// See https://en.wikipedia.org/wiki/Non-return-to-zero for more technical
// details.
func NRZ(b byte) uint32 {
	// The stream is 1x01x01x01x01x01x01x01x0 with the x bits being the bits from
	// `b` in reverse order.
	out := uint32(0x924924)
	out |= uint32(b&0x80) << (3*7 + 1 - 7)
	out |= uint32(b&0x40) << (3*6 + 1 - 6)
	out |= uint32(b&0x20) << (3*5 + 1 - 5)
	out |= uint32(b&0x10) << (3*4 + 1 - 4)
	out |= uint32(b&0x08) << (3*3 + 1 - 3)
	out |= uint32(b&0x04) << (3*2 + 1 - 2)
	out |= uint32(b&0x02) << (3*1 + 1 - 1)
	out |= uint32(b&0x01) << (3*0 + 1 - 0)
	return out
}

// DefaultOpts is the recommended default options.
var DefaultOpts = Opts{
	NumPixels: 150,                    // 150 LEDs is a common strip length.
	Channels:  3,                      // RGB.
	Freq:      800 * physic.KiloHertz, // Fast LEDs, most common.
}

// Opts defines the options for the device.
type Opts struct {
	// NumPixels is the number of pixels to control. If too short, the following
	// pixels will be corrupted. If too long, the pixels will be drawn
	// unnecessarily but not visible issue will occur.
	NumPixels int
	// Channels is 1 for single color LEDs, 3 for RGB LEDs and 4 for RGBW (white)
	// LEDs.
	Channels int
	// Freq is the frequency to use to drive the LEDs. It should be either 800kHz
	// for fast ICs and 400kHz for the slow ones.
	Freq physic.Frequency
}

// New opens a handle to a compatible LED strip.
func New(p gpiostream.PinOut, opts *Opts) (*Dev, error) {
	// Allow a wider range in case there's new devices with higher supported
	// frequency.
	if opts.Freq < 10*physic.KiloHertz || opts.Freq > 100*physic.MegaHertz {
		return nil, errors.New("nrzled: specify valid frequency")
	}
	if opts.Channels != 3 && opts.Channels != 4 {
		return nil, errors.New("nrzled: specify valid number of channels (3 or 4)")
	}
	return &Dev{
		p:         p,
		numPixels: opts.NumPixels,
		channels:  opts.Channels,
		b: gpiostream.BitStream{
			Freq: opts.Freq,
			// Each bit is encoded on 3 bits.
			Bits: make([]byte, opts.NumPixels*3*opts.Channels),
			LSBF: false,
		},
	}, nil
}

// Dev is a handle to the LED strip.
type Dev struct {
	p         gpiostream.PinOut
	numPixels int
	channels  int                  // Number of channels per pixel
	b         gpiostream.BitStream // NRZ encoded bits; cached to reduce heap fragmentation
	buf       []byte               // Double buffer of RGB/RGBW pixels; enables partial Draw()
}

func (d *Dev) String() string {
	return fmt.Sprintf("nrzled{%s}", d.p)
}

// Halt turns the lights off.
//
// It doesn't affect the back buffer.
func (d *Dev) Halt() error {
	zero := NRZ(0)
	a := byte(zero >> 16)
	b := byte(zero >> 8)
	c := byte(zero)
	for i := 0; i < d.channels*d.numPixels; i++ {
		d.b.Bits[3*i+0] = a
		d.b.Bits[3*i+1] = b
		d.b.Bits[3*i+2] = c
	}
	if err := d.p.StreamOut(&d.b); err != nil {
		return fmt.Errorf("nrzled: %v", err)
	}
	return nil
}

// ColorModel implements devices.Display.
//
// It is color.NRGBAModel.
func (d *Dev) ColorModel() color.Model {
	return color.NRGBAModel
}

// Bounds implements devices.Display. Min is guaranteed to be {0, 0}.
func (d *Dev) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: d.numPixels, Y: 1}}
}

// Draw implements devices.Display.
//
// Using something else than image.NRGBA is 10x slower and is not recommended.
// When using image.NRGBA, the alpha channel is ignored in RGB mode and used as
// White channel in RGBW mode.
//
// A back buffer is kept so that partial updates are supported, albeit the full
// LED strip is updated synchronously.
func (d *Dev) Draw(r image.Rectangle, src image.Image, sp image.Point) {
	r = r.Intersect(d.Bounds())
	srcR := src.Bounds()
	srcR.Min = srcR.Min.Add(sp)
	if dX := r.Dx(); dX < srcR.Dx() {
		srcR.Max.X = srcR.Min.X + dX
	}
	if dY := r.Dy(); dY < srcR.Dy() {
		srcR.Max.Y = srcR.Min.Y + dY
	}
	if d.buf == nil {
		// Allocate d.buf on first Draw() call, in case the user only wants to use
		// .Write().
		d.buf = make([]byte, d.numPixels*d.channels)
	}
	if img, ok := src.(*image.NRGBA); ok {
		// Fast path for image.NRGBA.
		base := srcR.Min.Y * img.Stride
		raster(d.b.Bits, img.Pix[base+4*srcR.Min.X:base+4*srcR.Max.X], d.channels, 4)
	} else {
		// Generic version.
		m := srcR.Max.X - srcR.Min.X
		if d.channels == 3 {
			for i := 0; i < m; i++ {
				c := color.NRGBAModel.Convert(src.At(srcR.Min.X+i, srcR.Min.Y)).(color.NRGBA)
				j := 3 * i
				put(d.b.Bits[3*(j+0):], c.G)
				put(d.b.Bits[3*(j+1):], c.R)
				put(d.b.Bits[3*(j+2):], c.B)
			}
		} else {
			for i := 0; i < m; i++ {
				c := color.NRGBAModel.Convert(src.At(srcR.Min.X+i, srcR.Min.Y)).(color.NRGBA)
				j := 4 * i
				put(d.b.Bits[3*(j+0):], c.G)
				put(d.b.Bits[3*(j+1):], c.R)
				put(d.b.Bits[3*(j+2):], c.B)
				put(d.b.Bits[3*(j+3):], c.A)
			}
		}
	}
	_ = d.p.StreamOut(&d.b)
}

// Write accepts a stream of raw RGB/RGBW pixels and sends it as NRZ encoded
// stream.
//
// This bypasses the back buffer.
func (d *Dev) Write(pixels []byte) (int, error) {
	if len(pixels)%d.channels != 0 || len(pixels) > d.numPixels*d.channels {
		return 0, errors.New("nrzled: invalid RGB stream length")
	}
	raster(d.b.Bits, pixels, d.channels, d.channels)
	if err := d.p.StreamOut(&d.b); err != nil {
		return 0, fmt.Errorf("nrzled: %v", err)
	}
	return len(pixels), nil
}

//

// raster converts a RGB/RGBW input stream into a MSB binary output stream as it
// must be sent over the GPIO pin.
//
// `in` is RGB 24 bits or RGBW 32 bits. Each bit is encoded over 3 bits so the
// length of `out` must be 3x as large as `in`.
//
// Encoded output format is GRB as 72 bits (24 * 3) or 96 bits (32 * 3).
func raster(out, in []byte, outChannels, inChannels int) {
	pixels := len(in) / inChannels
	if outChannels == 3 {
		for i := 0; i < pixels; i++ {
			j := i * inChannels
			k := 3 * i
			put(out[3*(k+0):], in[j+1])
			put(out[3*(k+1):], in[j+0])
			put(out[3*(k+2):], in[j+2])
		}
	} else {
		for i := 0; i < pixels; i++ {
			j := i * inChannels
			k := 4 * i
			put(out[3*(k+0):], in[j+1])
			put(out[3*(k+1):], in[j+0])
			put(out[3*(k+2):], in[j+2])
			put(out[3*(k+3):], in[j+3])
		}
	}
}

// put writes the byte v as an MSB-first NRZ encoded triplet byte into out.
func put(out []byte, v byte) {
	w := NRZ(v)
	out[0] = byte(w >> 16)
	out[1] = byte(w >> 8)
	out[2] = byte(w)
}

var _ conn.Resource = &Dev{}
var _ devices.Display = &Dev{}
var _ fmt.Stringer = &Dev{}
