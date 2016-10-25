// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package image1bit implements black and white (1 bit per pixel) 2D graphics
// in the memory format of the ssd1306 controller.
//
// It is compatible with package image/draw.
package image1bit

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
)

// Bit implements a 1 bit color.
type Bit bool

// RGBA returns either all white or all black and transparent.
func (b Bit) RGBA() (uint32, uint32, uint32, uint32) {
	if b {
		return 65535, 65535, 65535, 65535
	}
	return 0, 0, 0, 0
}

func (b Bit) String() string {
	if b {
		return "On"
	}
	return "Off"
}

// Possible bitness.
const (
	On  = Bit(true)
	Off = Bit(false)
)

// Image is a 1 bit (black and white) image.
//
// The packing used is unusual, each byte is 8 vertical pixels, with each byte
// stride being an horizontal band of 8 pixels high.
//
// It is designed specifically to work with SSD1306 OLED display controler.
type Image struct {
	W   int
	H   int
	Buf []byte // Can be passed directly to ssd1306.(*Dev).Write()
}

// New returns an initialized Image instance.
func New(r image.Rectangle) (*Image, error) {
	h := r.Dy()
	w := r.Dx()
	if h&7 != 0 {
		return nil, errors.New("height must be multiple of 8")
	}
	return &Image{w, h, make([]byte, w*h/8)}, nil
}

// SetAll sets all pixels to On.
func (i *Image) SetAll() {
	for j := range i.Buf {
		i.Buf[j] = 0xFF
	}
}

// Clear sets all pixels to Off.
func (i *Image) Clear() {
	for j := range i.Buf {
		i.Buf[j] = 0
	}
}

// Inverse changes all On pixels to Off and Off pixels to On.
func (i *Image) Inverse() {
	for j := range i.Buf {
		i.Buf[j] ^= 0xFF
	}
}

// ColorModel implements image.Image.
func (i *Image) ColorModel() color.Model {
	return color.ModelFunc(convert)
}

// Bounds implements image.Image.
func (i *Image) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: i.W, Y: i.H}}
}

// At implements image.Image.
func (i *Image) At(x, y int) color.Color {
	return i.AtBit(x, y)
}

// AtBit is the optimized version of At().
func (i *Image) AtBit(x, y int) Bit {
	offset := x + y/8*i.W
	mask := byte(1 << byte(y&7))
	return Bit(i.Buf[offset]&mask != 0)
}

// Set implements draw.Image
func (i *Image) Set(x, y int, c color.Color) {
	i.SetBit(x, y, convertBit(c))
}

// SetBit is the optimized version of Set().
func (i *Image) SetBit(x, y int, b Bit) {
	if x >= 0 && x < i.W {
		if y >= 0 && y < i.H {
			offset := x + y/8*i.W
			mask := byte(1 << byte(y&7))
			if b {
				i.Buf[offset] |= mask
			} else {
				i.Buf[offset] &^= mask
			}
		}
	}
}

//

var _ draw.Image = &Image{}

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
