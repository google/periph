// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package rwio

import (
	"image"
	"image/color"
	"io"
	"time"

	"periph.io/x/periph/conn"
)

// Display represents a pixel output device. It is a write-only interface.
//
// What Display represents can be as varied as a 1 bit OLED display or a strip
// of LED lights.
type Display interface {
	conn.Resource

	// Writer can be used when the native display pixel format is known. Each
	// write must cover exactly the whole screen as a single packed stream of
	// pixels.
	io.Writer
	// ColorModel returns the device native color model.
	//
	// It is generally color.NRGBA for a color display.
	ColorModel() color.Model
	// Bounds returns the size of the output device.
	//
	// Generally displays should have Min at {0, 0} but this is not guaranteed in
	// multiple displays setup or when an instance of this interface represents a
	// section of a larger logical display.
	Bounds() image.Rectangle
	// Draw updates the display with this image starting at 'sp' offset into the
	// display into 'r'. The code will likely be faster if the image is in the
	// display's native color format.
	//
	// To be compatible with draw.Drawer, this function doesn't return an error.
	Draw(r image.Rectangle, src image.Image, sp image.Point)
}

// Env represents measurements from an environmental sensor.
type Env struct {
	Temperature Celsius
	Pressure    KPascal
	Humidity    RelativeHumidity
}

// SenseEnv represents an environmental sensor.
type SenseEnv interface {
	conn.Resource

	// Sense returns the value read from the sensor. Unsupported metrics are not
	// modified.
	Sense(env *Env) error
	// SenseContinuous initiates a continuous sensing at the specified interval.
	//
	// It is important to call Halt() once done with the sensing, which will turn
	// the device off and will close the channel.
	SenseContinuous(interval time.Duration) (<-chan Env, error)
}
