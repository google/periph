// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package nrzled

import (
	"errors"
	"fmt"
	"image"
	"image/color"

	"periph.io/x/periph/conn/display"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
)

// Strip is deprecated and will soon be removed.
type Strip interface {
	display.Drawer
	Write(pixels []byte) (int, error)
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
		rect: image.Rect(0, 0, opts.NumPixels, 1),
	}, nil
}

// Dev is a handle to the LED strip.
type Dev struct {
	p         gpiostream.PinOut
	numPixels int
	channels  int                  // Number of channels per pixel
	b         gpiostream.BitStream // NRZ encoded bits; cached to reduce heap fragmentation
	buf       []byte               // Double buffer of RGB/RGBW pixels; enables partial Draw()
	rect      image.Rectangle      // Device bounds
}

func (d *Dev) String() string {
	return fmt.Sprintf("nrzled{%s}", d.p)
}

// Halt turns the lights off.
//
// It doesn't affect the back buffer.
func (d *Dev) Halt() error {
	zero := nrzMSB3[0]
	for i := 0; i < d.channels*d.numPixels; i++ {
		d.b.Bits[3*i+0] = zero[0]
		d.b.Bits[3*i+1] = zero[1]
		d.b.Bits[3*i+2] = zero[2]
	}
	if err := d.p.StreamOut(&d.b); err != nil {
		return fmt.Errorf("nrzled: %v", err)
	}
	return nil
}

// ColorModel implements display.Drawer.
//
// It is color.NRGBAModel.
func (d *Dev) ColorModel() color.Model {
	return color.NRGBAModel
}

// Bounds implements display.Drawer. Min is guaranteed to be {0, 0}.
func (d *Dev) Bounds() image.Rectangle {
	return d.rect
}

// Draw implements display.Drawer.
//
// Using something else than image.NRGBA is 10x slower and is not recommended.
// When using image.NRGBA, the alpha channel is ignored in RGB mode and used as
// White channel in RGBW mode.
//
// A back buffer is kept so that partial updates are supported, albeit the full
// LED strip is updated synchronously.
func (d *Dev) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	if r = r.Intersect(d.rect); r.Empty() {
		return nil
	}
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
				putNRZMSB3(d.b.Bits[3*(j+0):], c.G)
				putNRZMSB3(d.b.Bits[3*(j+1):], c.R)
				putNRZMSB3(d.b.Bits[3*(j+2):], c.B)
			}
		} else {
			for i := 0; i < m; i++ {
				c := color.NRGBAModel.Convert(src.At(srcR.Min.X+i, srcR.Min.Y)).(color.NRGBA)
				j := 4 * i
				putNRZMSB3(d.b.Bits[3*(j+0):], c.G)
				putNRZMSB3(d.b.Bits[3*(j+1):], c.R)
				putNRZMSB3(d.b.Bits[3*(j+2):], c.B)
				putNRZMSB3(d.b.Bits[3*(j+3):], c.A)
			}
		}
	}
	return d.p.StreamOut(&d.b)
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
			putNRZMSB3(out[3*(k+0):], in[j+1])
			putNRZMSB3(out[3*(k+1):], in[j+0])
			putNRZMSB3(out[3*(k+2):], in[j+2])
		}
	} else {
		for i := 0; i < pixels; i++ {
			j := i * inChannels
			k := 4 * i
			putNRZMSB3(out[3*(k+0):], in[j+1])
			putNRZMSB3(out[3*(k+1):], in[j+0])
			putNRZMSB3(out[3*(k+2):], in[j+2])
			putNRZMSB3(out[3*(k+3):], in[j+3])
		}
	}
}

// putNRZMSB3 writes the byte v as an MSB-first NRZ encoded triplet byte into
// out.
func putNRZMSB3(out []byte, v byte) {
	copy(out, nrzMSB3[v][:])
}

var _ display.Drawer = &Dev{}
