// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package devices

import (
	"fmt"
	"image"
	"image/color"
	"io"
)

// Display represents a pixel output device. It is a write-only interface.
//
// What Display represents can be as varied as a 1 bit OLED display or a strip
// of LED lights.
type Display interface {
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

// Milli is a fixed point value with 0.001 precision.
type Milli int32

// Float64 returns the value as float64 with 0.001 precision.
func (m Milli) Float64() float64 {
	return float64(m) * .001
}

// String returns the value formatted as a string.
func (m Milli) String() string {
	return fmt.Sprintf("%d.%03d", m/1000, m%1000)
}

// Celcius is a temperature at a precision of 0.001°C.
//
// Expected range is [-273150, >1000000]
//
// BUG(maruel): Add function to convert to Fahrenheit for my American friends.
type Celcius Milli

// Float64 returns the value as float64 with 0.001 precision.
func (c Celcius) Float64() float64 {
	return Milli(c).Float64()
}

// String returns the temperature formatted as a string.
func (c Celcius) String() string {
	return Milli(c).String() + "°C"
}

// ToF returns the temperature as Fahrenheit, a unit used in the United States.
func (c Celcius) ToF() Fahrenheit {
	return Fahrenheit((c*9+2)/5 + 32000)
}

// Fahrenheit is a unit used in the United States.
type Fahrenheit Milli

// Float64 returns the value as float64 with 0.001 precision.
func (f Fahrenheit) Float64() float64 {
	return Milli(f).Float64()
}

// String returns the temperature formatted as a string.
func (f Fahrenheit) String() string {
	return Milli(f).String() + "°F"
}

// KPascal is pressure at precision of 1Pa.
//
// Expected range is [0, >1000000].
type KPascal Milli

// Float64 returns the value as float64 with 0.001 precision.
func (k KPascal) Float64() float64 {
	return Milli(k).Float64()
}

// String returns the pressure formatted as a string.
func (k KPascal) String() string {
	return Milli(k).String() + "KPa"
}

// RelativeHumidity is humidity level in %rH with 0.01%rH precision.
type RelativeHumidity int32

// Float64 returns the value in %.
func (r RelativeHumidity) Float64() float64 {
	return float64(r) * .01
}

// String returns the humidity formatted as a string.
func (r RelativeHumidity) String() string {
	return fmt.Sprintf("%d.%02d%%rH", r/100, r%100)
}

// Environment represents measurements from an environmental sensor.
type Environment struct {
	Temperature Celcius
	Pressure    KPascal
	Humidity    RelativeHumidity
}

// Environmental represents an environmental sensor.
type Environmental interface {
	// Sense returns the value read from the sensor. Unsupported metrics are not
	// modified.
	Sense(env *Environment) error
}
