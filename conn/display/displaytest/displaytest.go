// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package displaytest

import (
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/periph/conn/display"
)

// Drawer is a fake display.Drawer.
type Drawer struct {
	Img *image.NRGBA
}

func (d *Drawer) String() string {
	return "Drawer"
}

// Halt implements conn.Resource. It is a noop.
func (d *Drawer) Halt() error {
	return nil
}

// ColorModel implements image.Image.
func (d *Drawer) ColorModel() color.Model {
	return d.Img.ColorModel()
}

// Bounds implements image.Image.
func (d *Drawer) Bounds() image.Rectangle {
	return d.Img.Bounds()
}

// Draw implements draw.Image.
func (d *Drawer) Draw(dstRect image.Rectangle, src image.Image, sp image.Point) error {
	draw.Draw(d.Img, dstRect, src, sp, draw.Src)
	return nil
}

var _ display.Drawer = &Drawer{}
