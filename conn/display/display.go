// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package display

import (
	"image"
	"image/color"

	"periph.io/x/periph/conn"
)

// Drawer represents a context to display pixels on an output device. It is a
// write-only interface.
//
// What Drawer represents can be as varied as a 1 bit OLED display or a strip
// of LED lights. The implementation keeps a single frame buffer, so that
// partial updates can be done.
type Drawer interface {
	conn.Resource

	// ColorModel returns the device native color model.
	ColorModel() color.Model
	// Bounds returns the size of the output device.
	//
	// Generally displays should have Min at {0, 0} but this is not guaranteed in
	// multiple displays setup or when an instance of this interface represents a
	// section of a larger logical display.
	Bounds() image.Rectangle
	// Draw updates the display with this image.
	//
	// Only the pixels within the display boundary are updated. Partial update is
	// supported.
	//
	// Coordinates are top-left 0,0.
	//
	// dstRect aligns the the drawing operation in the display, enabling partial
	// update.
	//
	// srcPts aligns the image at this offset, enabling using a subset of the
	// source image. uUe image.Point{} to take the image at its origin.
	Draw(dstRect image.Rectangle, src image.Image, srcPts image.Point) error
}
