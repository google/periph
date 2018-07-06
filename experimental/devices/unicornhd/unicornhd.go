// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package unicornhd implements interfacing code to Pimoroni's Unicorn HD hat.
package unicornhd

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/periph/conn/display"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

const (
	Height = 16
	Width  = 16
	speed  = 9 * physic.MegaHertz
	bits   = 8
	prefix = 0x72
)

// Dev represents a Unicorn HAT HD (https://shop.pimoroni.com/products/unicorn-hat-hd)
// connected over a SPI port.
type Dev struct {
	// Communication
	connector spi.Conn
	pixels    *image.NRGBA
	txBuffer  []byte
}

// New returns a unicornHD driver that communicates over SPI.
//
// The SPI port speed must be 9MHz and the SPI mode, 0, as in the
// python example library.
func NewUnicornhd(port spi.Port) (*Dev, error) {
	connector, err := port.Connect(speed, spi.Mode0, bits)
	if err != nil {
		return nil, err
	}
	return &Dev{
		connector: connector,
		pixels:    image.NewNRGBA(image.Rect(0, 0, Width, Height)),
		txBuffer:  make([]byte, Width*Height*3+1),
	}, nil
}

// Implement display.Drawer
//
// Returns a string with the driver name and the width and height of the display.
func (device *Dev) String() string {
	return fmt.Sprintf("UnicornHD{%d, %d}", Width, Height)
}

// Halting the unicorn HD sets all the pixels to black. Error is always nil.
func (device *Dev) Halt() error {
	black := color.RGBA{0, 0, 0, 0}
	device.Draw(device.Bounds(), &image.Uniform{black}, image.ZP)
	return nil
}

// ColorModel implements devices.Display. There's no surprise, it is
// color.RGBAModel.
func (device *Dev) ColorModel() color.Model {
	return color.NRGBAModel
}

// Bounds implements devices.Display.
//
// Min is guaranteed to be {0, 0}.
func (device *Dev) Bounds() image.Rectangle {
	return device.pixels.Bounds()
}

// Draw implements devices.Display.
//
// Using something else than image.NRGBA is 10x slower. When using image.NRGBA,
// the alpha channel is ignored.
func (device *Dev) Draw(dstRect image.Rectangle, src image.Image, srcPts image.Point) error {
	// Use stdlib image copying functionality.
	draw.Draw(device.pixels, dstRect, src, srcPts, draw.Src)
	// And then copy the image into the transmission buffer, where it is sent via SPI.
	return device.flush()
}

func (device *Dev) flush() error {
	device.txBuffer[0] = prefix
	x := 0
	y := 0
	for i := 0; i < Width*Height; i++ {
		color := device.pixels.NRGBAAt(x, y)
		x++
		if x >= Width {
			x = 0
			y++
		}
		red := color.R
		green := color.G
		blue := color.B

		k := 3*i + 1
		device.txBuffer[k] = red
		device.txBuffer[k+1] = green
		device.txBuffer[k+2] = blue
	}

	return device.connector.Tx(device.txBuffer, nil)
}

// Test that driver implements display.Drawer interface.  This is
// enforced at compile time.
var _ display.Drawer = (*Dev)(nil)
