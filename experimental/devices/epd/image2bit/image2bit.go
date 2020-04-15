// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package image2bit implements two bit gray scale (white, light gray,
// dark gray, black) 2D graphics.
//
// It is compatible with package image/draw.
//
// The bit packing format is the same as used by waveshare e-Paper
// displays such as the 4.2 inch display.
package image2bit

import (
	"image"
	"image/color"
	"image/draw"
)

// Gray implements a 2 bit color.
type Gray byte

// RGBA returns either black, dark gray, light gray or white.
func (b Gray) RGBA() (uint32, uint32, uint32, uint32) {
	switch b {
	case 0:
		return 0, 0, 0, 65535
	case 1:
		return 0x5555, 0x5555, 0x5555, 0xffff
	case 2:
		return 0xaaaa, 0xaaaa, 0xaaaa, 0xffff
	default:
		return 0xffff, 0xffff, 0xffff, 0xffff
	}
}

func (b Gray) String() string {
	switch b {
	case 0:
		return "black"
	case 1:
		return "dark gray"
	case 2:
		return "light gray"
	default:
		return "white"
	}
}

// All possible colors
const (
	White     Gray = 3
	LightGray Gray = 2
	DarkGray  Gray = 1
	Black     Gray = 0
)

// GrayModel is the color Model for 2 bit gray scale.
var GrayModel = color.ModelFunc(convert)

// BitPlane is a 2 bit gray scale image. To match the wire format
// for waveshare e-Paper the two bits per pixel is stored across two bitmaps.
// PixMSB contains the most significant bit, PixLSB contains the least significant bit.
//
//	        White LightGray DarkGray Black
//	PixMSB   1     1         0        0
//	PixLSB   1     0         1        0
//
// The following example shows the stored data for an 8 pixel wide image, 1 pixel high:
// PixMSB []byte{0b10100000}
// PixLSB []byte{0b10000000}
//
// It has a black background, the first pixel is white, and the third pixel LightGray.
type BitPlane struct {
	// PixMSB holds the image's most significant bit as a horizontally packed bitmap.
	PixMSB []byte
	// PixLSB holds the image's least significant bit as a horizontally packed bitmap.
	PixLSB []byte
	// Rect is the image's bounds.
	Rect image.Rectangle

	// Stride is the number of pixels on each horizontal line, including padding
	Stride int
}

// NewBitPlane returns an initialized BitPlane instance, all black.
func NewBitPlane(r image.Rectangle) *BitPlane {
	// stride is width rounded up to the next byte
	stride := ((r.Dx() + 7) &^ 7)

	size := (r.Dy() * stride) / 8
	return &BitPlane{PixMSB: make([]byte, size), PixLSB: make([]byte, size), Rect: r, Stride: stride}
}

// ColorModel implements image.Image.
func (i *BitPlane) ColorModel() color.Model {
	return GrayModel
}

// Bounds implements image.Image.
func (i *BitPlane) Bounds() image.Rectangle {
	return i.Rect
}

// At implements image.Image.
func (i *BitPlane) At(x, y int) color.Color {
	return i.GrayAt(x, y)
}

// GrayAt is the optimized version of At().
func (i *BitPlane) GrayAt(x, y int) Gray {
	if !(image.Point{x, y}.In(i.Rect)) {
		return Black
	}

	byteIndex, bitIndex, _ := i.getOffset(x, y)

	return Gray(((i.PixMSB[byteIndex]>>bitIndex)&0b1)<<1 | i.PixLSB[byteIndex]>>bitIndex&0b1)
}

// Opaque scans the entire image and reports whether it is fully opaque.
func (i *BitPlane) Opaque() bool {
	return true
}

// Set implements draw.Image
func (i *BitPlane) Set(x, y int, c color.Color) {
	i.SetGray(x, y, convertGray(c))
}

// SetGray is the optimized version of Set().
func (i *BitPlane) SetGray(x, y int, b Gray) {
	if !(image.Point{x, y}.In(i.Rect)) {
		return
	}

	byteIndex, bitIndex, mask := i.getOffset(x, y)

	i.PixMSB[byteIndex] = byte((i.PixMSB[byteIndex] & mask) | ((byte(b&0b10) >> 1) << bitIndex))
	i.PixLSB[byteIndex] = byte((i.PixLSB[byteIndex] & mask) | (byte(b&0b01) << bitIndex))
}

func (i *BitPlane) getOffset(x, y int) (byteIndex, bitIndex int, mask byte) {
	bitIndex = (y*i.Stride + x)
	byteIndex = bitIndex / 8
	bitIndex = 7 - (bitIndex % 8)
	mask = byte(0xff ^ (0x01 << bitIndex))
	return
}

// convert color to gray as color.Color
func convert(c color.Color) color.Color {
	return convertGray(c)
}

// convert color to gray
func convertGray(c color.Color) Gray {
	switch t := c.(type) {
	case Gray:
		return t
	default:
		r, g, b, _ := c.RGBA()
		// TODO something fancy, how to weight R/G/B
		return Gray((r | g | b) >> 14) // Use two most significant bits.
	}
}

// verify that we satisfy the draw.Image interface
var _ draw.Image = &BitPlane{}
