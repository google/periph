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

	"periph.io/x/periph/conn/spi"
)

const (
	Height = 16
	Width  = 16
	speed  = 9000000
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
	err       error
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

// Returns a string with the driver name and the width and height of the display.
func (device *Dev) String() string {
	return fmt.Sprintf("UnicornHD{%d, %d}", Width, Height)
}

// Implement devices.Display
//
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
func (device *Dev) Draw(area image.Rectangle, img image.Image, origin image.Point) {
	// Use stdlib image copying functionality.
	draw.Draw(device.pixels, area, img, origin, draw.Src)
	// And then copy the image into the transmission buffer, where it is sent via SPI.
	device.flush()
}

func (device *Dev) flush() {
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

	device.err = device.connector.Tx(device.txBuffer, nil)
}

func (device *Dev) Err() error {
	return device.err
}
