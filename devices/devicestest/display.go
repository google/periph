// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package devicestest

import (
	"errors"
	"image"
	"image/color"
	"image/draw"

	"github.com/google/periph/devices"
)

// Display is a fake devices.Display.
type Display struct {
	Img *image.NRGBA
}

// Write implements devices.Display.
func (d *Display) Write(pixels []byte) (int, error) {
	if len(pixels)%3 != 0 {
		return 0, errors.New("invalid RGB stream length")
	}
	copy(d.Img.Pix, pixels)
	return len(pixels), nil
}

// ColorModel implements image.Image.
func (d *Display) ColorModel() color.Model {
	return d.Img.ColorModel()
}

// Bounds implements image.Image.
func (d *Display) Bounds() image.Rectangle {
	return d.Img.Bounds()
}

// Draw implements draw.Image.
func (d *Display) Draw(r image.Rectangle, src image.Image, sp image.Point) {
	draw.Draw(d.Img, r, src, sp, draw.Src)
}

var _ devices.Display = &Display{}
