// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package image14bit implements 14-bit per pixel images.
//
// It is compatible with the image/draw package.
package image14bit

import (
	"image"
	"image/color"
	"image/draw"
)

// Gray14 represents an image of 14-bit values.
type Gray14 struct {
	// Pix holds the image's pixels. Each uint16 element represents one 14-bit
	// pixel.
	Pix []uint16
	// Stride is the Pix stride (in pixels) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewGray14 returns an initialized Gray14 instance.
func NewGray14(r image.Rectangle) *Gray14 {
	w, h := r.Dx(), r.Dy()
	pix := make([]uint16, w*h)
	return &Gray14{Pix: pix, Stride: w, Rect: r}
}

// ColorModel implements image.Image.
func (i *Gray14) ColorModel() color.Model {
	return Intensity14Model
}

// Bounds implements image.Image.
func (i *Gray14) Bounds() image.Rectangle {
	return i.Rect
}

// Opaque returns whether the image is fully opaque.
func (i *Gray14) Opaque() bool {
	return true
}

// At implements image.Image.
func (i *Gray14) At(x, y int) color.Color {
	return i.Intensity14At(x, y)
}

// Intensity14At returns the Intensity14 value at a point.
func (i *Gray14) Intensity14At(x, y int) Intensity14 {
	if !(image.Point{x, y}.In(i.Rect)) {
		return Intensity14(0)
	}
	offset := i.PixOffset(x, y)
	return Intensity14(i.Pix[offset])
}

// PixOffset returns the index of the element of Pix that
// corresponds to the pixel at (x, y).
func (i *Gray14) PixOffset(x, y int) int {
	return (y-i.Rect.Min.Y)*i.Stride + (x - i.Rect.Min.X)
}

// Set implements draw.Image.
func (i *Gray14) Set(x, y int, c color.Color) {
	i.SetIntensity14(x, y, convertIntensity14(c))
}

// SetIntensity14 sets the Intensity14 value for the pixel at (x, y).
func (i *Gray14) SetIntensity14(x, y int, c Intensity14) {
	if !(image.Point{x, y}.In(i.Rect)) {
		return
	}
	i.Pix[i.PixOffset(x, y)] = uint16(c)
}

var _ draw.Image = &Gray14{}
