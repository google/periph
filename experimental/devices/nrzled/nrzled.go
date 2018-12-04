// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package nrzled

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/display"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

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

// NewStream opens a handle to a compatible LED strip.
func NewStream(p gpiostream.PinOut, opts *Opts) (*Dev, error) {
	// Allow a wider range in case there's new devices with higher supported
	// frequency.
	if opts.Freq < 10*physic.KiloHertz || opts.Freq > 100*physic.MegaHertz {
		return nil, errors.New("nrzled: specify valid frequency")
	}
	if opts.Channels != 3 && opts.Channels != 4 {
		return nil, errors.New("nrzled: specify valid number of channels (3 or 4)")
	}
	// 3 symbol bytes per byte, 3/4 bytes per pixel.
	streamLen := 3 * (opts.Channels * opts.NumPixels)
	// 3 bytes for latch. TODO: duration.
	bufSize := streamLen + 3
	buf := make([]byte, bufSize)
	return &Dev{
		name:      "nrzled{" + p.Name() + "}",
		p:         p,
		numPixels: opts.NumPixels,
		channels:  opts.Channels,
		b:         gpiostream.BitStream{Freq: opts.Freq, Bits: buf, LSBF: false},
		rawBuf:    buf[:streamLen],
		rect:      image.Rect(0, 0, opts.NumPixels, 1),
	}, nil
}

// NewSPI returns a strip that communicates over SPI to NRZ encoded LEDs.
//
// Due to the tight timing demands of these LEDs, the SPI port speed must be a
// reliable 2.4~2.5MHz; this is 3x 800kHz.
//
// The driver's SPI buffer must be at least 12*num_pixels+3 bytes long.
func NewSPI(p spi.Port, opts *Opts) (*Dev, error) {
	const spiFreq = 2500 * physic.KiloHertz
	if opts.Freq != spiFreq {
		return nil, errors.New("nrzled: expected Freq " + spiFreq.String())
	}
	if opts.Channels != 3 && opts.Channels != 4 {
		return nil, errors.New("nrzled: specify valid number of channels (3 or 4)")
	}
	// 4 symbol bytes per byte, 3/4 bytes per pixel.
	streamLen := 4 * (opts.Channels * opts.NumPixels)
	// 3 bytes for latch. 24*400ns = 9600ns. In practice this could be skipped,
	// as the overhead for SPI Tx() tear down and the next one is likely at least
	// 10µs.
	bufSize := streamLen + 3
	if l, ok := p.(conn.Limits); ok {
		if s := l.MaxTxSize(); s < bufSize {
			return nil, errors.New("spi port buffer is too short for the specified number of pixels")
		}
	}
	c, err := p.Connect(spiFreq, spi.Mode3|spi.NoCS, 8)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, bufSize)
	return &Dev{
		name:      "nrzled{" + c.String() + "}",
		s:         c,
		numPixels: opts.NumPixels,
		channels:  opts.Channels,
		b:         gpiostream.BitStream{Freq: opts.Freq, Bits: buf, LSBF: false},
		rawBuf:    buf[:streamLen],
		rect:      image.Rect(0, 0, opts.NumPixels, 1),
	}, nil
}

// Dev is a handle to the LED strip.
type Dev struct {
	// Immutable.
	name      string
	s         spi.Conn
	p         gpiostream.PinOut
	numPixels int
	channels  int             // Number of channels per pixel
	rect      image.Rectangle // Device bounds

	// Mutable.
	b      gpiostream.BitStream // NRZ encoded bits; cached to reduce heap fragmentation
	rawBuf []byte               // NRZ encoded bits; excluding the padding
	pixels []byte               // Double buffer of RGB/RGBW pixels; enables partial Draw()
}

func (d *Dev) String() string {
	return d.name
}

// Halt turns the lights off.
//
// It doesn't affect the back buffer.
func (d *Dev) Halt() error {
	if d.s == nil {
		// zero := nrzMSB3[0]
		const a = 0x92
		const b = 0x49
		const c = 0x24
		for i := 0; i < len(d.rawBuf); i += 3 {
			d.rawBuf[i+0] = a
			d.rawBuf[i+1] = b
			d.rawBuf[i+2] = c
		}
		if err := d.p.StreamOut(&d.b); err != nil {
			return fmt.Errorf("nrzled: %v", err)
		}
		return nil
	}

	// Zap out the buffer. 0x88 is '0'.
	for i := range d.rawBuf {
		d.rawBuf[i] = 0x88
	}
	if err := d.s.Tx(d.b.Bits, nil); err != nil {
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
	if srcR.Empty() {
		return nil
	}
	if d.s != nil {
		d.rasterSPIImg(d.rawBuf, r, src, srcR)
		return d.s.Tx(d.b.Bits, nil)
	}
	if d.pixels == nil {
		// Allocate d.pixels on first Draw() call, in case the user only wants to
		// use .Write().
		d.pixels = make([]byte, d.numPixels*d.channels)
	}
	if img, ok := src.(*image.NRGBA); ok {
		// Fast path for image.NRGBA.
		base := srcR.Min.Y * img.Stride
		rasterBits(d.b.Bits, img.Pix[base+4*srcR.Min.X:base+4*srcR.Max.X], d.channels, 4)
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
	if d.s == nil {
		rasterBits(d.b.Bits, pixels, d.channels, d.channels)
		if err := d.p.StreamOut(&d.b); err != nil {
			return 0, fmt.Errorf("nrzled: %v", err)
		}
		return len(pixels), nil
	}
	d.rasterSPI(d.rawBuf, pixels, false)
	return len(pixels), d.s.Tx(d.b.Bits, nil)
}

// Bits

// rasterBits converts a RGB/RGBW input stream into a MSB binary output stream
// as it must be sent over the GPIO pin.
//
// `in` is RGB 24 bits or RGBW 32 bits. Each bit is encoded over 3 bits so the
// length of `out` must be 3x as large as `in`.
//
// Encoded output format is GRB as 72 bits (24 * 3) or 96 bits (32 * 3).
func rasterBits(out, in []byte, outChannels, inChannels int) {
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

// SPI

// rasterSPI serializes a buffer of RGB bytes to the WS2812b SPI format.
//
// It is expected to be given the part where pixels are, not the header nor
// footer.
//
// dst is in WS2812b SPI 32 bits word format. src is in RGB 24 bits, or 32 bits
// word format when srcHasAlpha is true. The src alpha channel is ignored in
// this case.
//
// src cannot be longer in pixel count than dst.
func (d *Dev) rasterSPI(dst []byte, src []byte, srcHasAlpha bool) {
	pBytes := 3
	if srcHasAlpha {
		pBytes = 4
	}
	length := len(src) / pBytes
	stride := 4 //number of spi-bytes in color-byte
	for i := 0; i < length; i++ {
		sOff := pBytes * i
		dOff := 3 * stride * i //3 channels * stride
		r, g, b := src[sOff], src[sOff+1], src[sOff+2]
		//grb color order, msb first
		copy(dst[dOff+stride*0:dOff+stride*1], nrzMSB4[r][:])
		copy(dst[dOff+stride*1:dOff+stride*2], nrzMSB4[g][:])
		copy(dst[dOff+stride*2:dOff+stride*3], nrzMSB4[b][:])
	}
}

// rasterSPIImg is the generic version of raster that converts an image instead
// of raw RGB values.
//
// It has 'fast paths' for image.RGBA and image.NRGBA that extract and convert
// the RGB values directly.  For other image types, it converts to image.RGBA
// and then does the same.  In all cases, alpha values are ignored.
//
// rect specifies where into the output buffer to draw.
//
// srcR specifies what portion of the source image to use.
func (d *Dev) rasterSPIImg(dst []byte, rect image.Rectangle, src image.Image, srcR image.Rectangle) {
	// Render directly into the buffer for maximum performance and to keep
	// untouched sections intact.
	switch im := src.(type) {
	case *image.RGBA:
		start := im.PixOffset(srcR.Min.X, srcR.Min.Y)
		// srcR.Min.Y since the output display has only a single column
		end := im.PixOffset(srcR.Max.X, srcR.Min.Y)
		// Offset into the output buffer using rect
		d.rasterSPI(dst[4*rect.Min.X:], im.Pix[start:end], true)
	case *image.NRGBA:
		// Ignores alpha
		start := im.PixOffset(srcR.Min.X, srcR.Min.Y)
		// srcR.Min.Y since the output display has only a single column
		end := im.PixOffset(srcR.Max.X, srcR.Min.Y)
		// Offset into the output buffer using rect
		d.rasterSPI(dst[4*rect.Min.X:], im.Pix[start:end], true)
	default:
		// Slow path.  Convert to RGBA
		b := im.Bounds()
		m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
		start := m.PixOffset(srcR.Min.X, srcR.Min.Y)
		// srcR.Min.Y since the output display has only a single column
		end := m.PixOffset(srcR.Max.X, srcR.Min.Y)
		// Offset into the output buffer using rect
		d.rasterSPI(dst[4*rect.Min.X:], m.Pix[start:end], true)
	}
}

var _ display.Drawer = &Dev{}
