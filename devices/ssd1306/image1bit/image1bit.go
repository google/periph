// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package image1bit implements black and white (1 bit per pixel) 2D graphics.
//
// It is compatible with package image/draw.
//
// VerticalLSB is the only bit packing implemented as it is used by the
// ssd1306. Others would be VerticalMSB, HorizontalLSB and HorizontalMSB.
package image1bit

import (
	"image"
	"image/color"
	"image/draw"
)

// Bit implements a 1 bit color.
type Bit bool

// RGBA returns either all white or all black.
//
// Technically the monochrome display could be colored but this information is
// unavailable here. To use a colored display, use the 1 bit image as a mask
// for a color.
func (b Bit) RGBA() (uint32, uint32, uint32, uint32) {
	if b {
		return 65535, 65535, 65535, 65535
	}
	return 0, 0, 0, 65535
}

func (b Bit) String() string {
	if b {
		return "On"
	}
	return "Off"
}

// Possible bitness.
const (
	On  Bit = true
	Off Bit = false
)

// BitModel is the color Model for 1 bit color.
var BitModel = color.ModelFunc(convert)

// VerticalLSB is a 1 bit (black and white) image.
//
// Each byte is 8 vertical pixels. Each stride is an horizontal band of 8
// pixels high with LSB first. So the first byte represent the following
// pixels, with lowest bit being the top left pixel.
//
//   0 x x x x x x x
//   1 x x x x x x x
//   2 x x x x x x x
//   3 x x x x x x x
//   4 x x x x x x x
//   5 x x x x x x x
//   6 x x x x x x x
//   7 x x x x x x x
//
// It is designed specifically to work with SSD1306 OLED display controler.
type VerticalLSB struct {
	// Pix holds the image's pixels, as vertically LSB-first packed bitmap. It
	// can be passed directly to ssd1306.Dev.Write()
	Pix []byte
	// Stride is the Pix stride (in bytes) between vertically adjacent 8 pixels
	// horizontal bands.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewVerticalLSB returns an initialized VerticalLSB instance.
func NewVerticalLSB(r image.Rectangle) *VerticalLSB {
	w := r.Dx()
	// Round down.
	minY := r.Min.Y &^ 7
	// Round up.
	maxY := (r.Max.Y + 7) & ^7
	bands := (maxY - minY) / 8
	return &VerticalLSB{Pix: make([]byte, w*bands), Stride: w, Rect: r}
}

// ColorModel implements image.Image.
func (i *VerticalLSB) ColorModel() color.Model {
	return BitModel
}

// Bounds implements image.Image.
func (i *VerticalLSB) Bounds() image.Rectangle {
	return i.Rect
}

// At implements image.Image.
func (i *VerticalLSB) At(x, y int) color.Color {
	return i.BitAt(x, y)
}

// BitAt is the optimized version of At().
func (i *VerticalLSB) BitAt(x, y int) Bit {
	if !(image.Point{x, y}.In(i.Rect)) {
		return Off
	}
	offset, mask := i.PixOffset(x, y)
	return Bit(i.Pix[offset]&mask != 0)
}

// Opaque scans the entire image and reports whether it is fully opaque.
func (i *VerticalLSB) Opaque() bool {
	return true
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y) and the corresponding mask.
func (i *VerticalLSB) PixOffset(x, y int) (int, byte) {
	// Adjust band.
	minY := i.Rect.Min.Y &^ 7
	pY := (y - minY)
	offset := pY/8*i.Stride + (x - i.Rect.Min.X)
	bit := uint(pY & 7)
	return offset, 1 << bit
}

// Set implements draw.Image
func (i *VerticalLSB) Set(x, y int, c color.Color) {
	i.SetBit(x, y, convertBit(c))
}

// SetBit is the optimized version of Set().
func (i *VerticalLSB) SetBit(x, y int, b Bit) {
	if !(image.Point{x, y}.In(i.Rect)) {
		return
	}
	offset, mask := i.PixOffset(x, y)
	if b {
		i.Pix[offset] |= mask
	} else {
		i.Pix[offset] &^= mask
	}
}

/*
// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (i *VerticalLSB) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(i.Rect)
	// If r1 and r2 are Rectangles, r1.Intersect(r2) is not guaranteed to be
	// inside either r1 or r2 if the intersection is empty. Without explicitly
	// checking for this, the Pix[i:] expression below can panic.
	if r.Empty() {
		return &VerticalLSB{}
	}
	offset, mask := i.PixOffset(r.Min.X, r.Min.Y)
	// TODO(maruel): Adjust with mask.
	return &VerticalLSB{
		Pix:    i.Pix[offset:],
		Stride: i.Stride,
		Rect:   r,
	}
}
*/

//

var _ draw.Image = &VerticalLSB{}

// Anything not transparent and not pure black is white.
func convert(c color.Color) color.Color {
	return convertBit(c)
}

// Anything not transparent and not pure black is white.
func convertBit(c color.Color) Bit {
	switch t := c.(type) {
	case Bit:
		return t
	default:
		r, g, b, _ := c.RGBA()
		return Bit((r | g | b) >= 0x8000)
	}
}
