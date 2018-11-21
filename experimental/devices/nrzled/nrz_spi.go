// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package nrzled

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/periph/conn/display"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

// SPIDev represents a strip of WS2812b LEDs as a strip connected over a SPI port.
// It accepts a stream of raw RGB pixels and converts it to a bit pattern consistent
// with the WS812b protocol.
// Includes intensity and temperature correction.
type SPIDev struct {
	s         spi.Conn        //
	numPixels int             //
	rawBuf    []byte          // Raw buffer sent over SPI. Cached to reduce heap fragmentation.
	pixels    []byte          // Double buffer of pixels, to enable partial painting via Draw(). Effectively points inside rawBuf.
	rect      image.Rectangle // Device bounds
	pixelMap  map[image.Point]int
}

// NewSPI returns a strip that communicates over SPI to NRZ encoded LEDs.
//
// Due to the tight timing demands of these LEDs,
// the SPI port speed must be a reliable 2.5MHz
//
// Note that your SPI buffer should be at least 12*num_pixels+2 bytes long
func NewSPI(p spi.Port, o *Opts) (*SPIDev, error) {
	c, err := p.Connect(2500*physic.KiloHertz, spi.Mode3|spi.NoCS, 8)
	if err != nil {
		return nil, err
	}
	rawBCt := 4 * (3 * o.NumPixels) //3 bytes per pixel, 4 symbol bytes per byte
	buf := make([]byte, rawBCt+3)   //3 bytes for latch. 24*400ns = 9600ns.
	tail := buf[rawBCt:]
	for i := range tail {
		tail[i] = 0x00
	}

	return &SPIDev{
		s:         c,
		numPixels: o.NumPixels,
		rawBuf:    buf,
		pixels:    buf[:rawBCt],
		rect:      image.Rect(0, 0, o.NumPixels, 1),
	}, nil
}

func (d *SPIDev) String() string {
	return fmt.Sprintf("nrzled: {%d, %s}", d.numPixels, d.s)
}

// ColorModel implements display.Drawer. There's no surprise, it is
// color.NRGBAModel.
func (d *SPIDev) ColorModel() color.Model {
	return color.NRGBAModel
}

// Bounds implements display.Drawer. Min is guaranteed to be {0, 0}.
func (d *SPIDev) Bounds() image.Rectangle {
	return d.rect
}

// Draw implements display.Drawer.
//
// Using something else than image.NRGBA is 10x slower. When using image.NRGBA,
// the alpha channel is ignored.
func (d *SPIDev) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
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
	d.rasterImg(d.pixels, r, src, srcR)
	return d.s.Tx(d.rawBuf, nil)
}

// Write accepts a stream of raw RGB pixels and sends it as WS2812b encoded
// stream.
func (d *SPIDev) Write(pixels []byte) (int, error) {
	if len(pixels)%3 != 0 || len(pixels)/3 > d.numPixels {
		return 0, errors.New("ws2812b: invalid RGB stream length")
	}
	// Do not touch footer.
	d.raster(d.pixels, pixels, false)
	err := d.s.Tx(d.rawBuf, nil)
	return len(pixels), err
}

// Halt turns off all the lights.
func (d *SPIDev) Halt() error {
	// Zap out the buffer.
	for i := range d.pixels {
		d.pixels[i] = 0x88
	}
	return d.s.Tx(d.rawBuf, nil)
}

// raster serializes a buffer of RGB bytes to the WS2812b SPI format.
//
// It is expected to be given the part where pixels are, not the header nor
// footer.
//
// dst is in WS2812b SPI 32 bits word format. src is in RGB 24 bits, or 32 bits
// word format when srcHasAlpha is true. The src alpha channel is ignored in
// this case.
//
// src cannot be longer in pixel count than dst.
func (d *SPIDev) raster(dst []byte, src []byte, srcHasAlpha bool) {
	pBytes := 3
	if srcHasAlpha {
		pBytes = 4
	}
	length := len(src) / pBytes
	if l := len(dst) / 4; l < length {
		length = l
	}
	if length == 0 {
		// Save ourself some unneeded processing.
		return
	}
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

// rasterImg is the generic version of raster that converts an image instead of raw RGB values.
//
// It has 'fast paths' for image.RGBA and image.NRGBA that extract and convert the RGB values
// directly.  For other image types, it converts to image.RGBA and then does the same.  In all
// cases, alpha values are ignored.
//
// rect specifies where into the output buffer to draw.
//
// srcR specifies what portion of the source image to use.
func (d *SPIDev) rasterImg(dst []byte, rect image.Rectangle, src image.Image, srcR image.Rectangle) {
	// Render directly into the buffer for maximum performance and to keep
	// untouched sections intact.
	switch im := src.(type) {
	case *image.RGBA:
		start := im.PixOffset(srcR.Min.X, srcR.Min.Y)
		// srcR.Min.Y since the output display has only a single column
		end := im.PixOffset(srcR.Max.X, srcR.Min.Y)
		// Offset into the output buffer using rect
		d.raster(dst[4*rect.Min.X:], im.Pix[start:end], true)
	case *image.NRGBA:
		// Ignores alpha
		start := im.PixOffset(srcR.Min.X, srcR.Min.Y)
		// srcR.Min.Y since the output display has only a single column
		end := im.PixOffset(srcR.Max.X, srcR.Min.Y)
		// Offset into the output buffer using rect
		d.raster(dst[4*rect.Min.X:], im.Pix[start:end], true)
	default:
		// Slow path.  Convert to RGBA
		b := im.Bounds()
		m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
		start := m.PixOffset(srcR.Min.X, srcR.Min.Y)
		// srcR.Min.Y since the output display has only a single column
		end := m.PixOffset(srcR.Max.X, srcR.Min.Y)
		// Offset into the output buffer using rect
		d.raster(dst[4*rect.Min.X:], m.Pix[start:end], true)
	}
}

var _ display.Drawer = &Dev{}
