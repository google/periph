// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package rwiotest

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/rwio"
)

// Display is a fake rwio.Display.
type Display struct {
	Img *image.NRGBA
}

func (d *Display) String() string {
	return "Display"
}

// Halt implements conn.Resource. It is a noop.
func (d *Display) Halt() error {
	return nil
}

// Write implements rwio.Display.
func (d *Display) Write(pixels []byte) (int, error) {
	if len(pixels)%3 != 0 {
		return 0, errors.New("devicetest: invalid RGB stream length")
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

var _ conn.Resource = &Display{}
var _ rwio.Display = &Display{}
var _ fmt.Stringer = &Display{}
