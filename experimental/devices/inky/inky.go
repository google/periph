// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package inky

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/display"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

const (
	// Constants for an Inky pHAT
	cols = 104
	rows = 212
)

// Color is used to define which model of inky is being used, and also for
// setting the border color.
type Color int

// Valid Color.
const (
	Black Color = iota
	Red
	Yellow
	White
)

var borderColor = map[Color]byte{
	Black:  0x00,
	Red:    0x33,
	Yellow: 0x33,
	White:  0xff,
}

// Model lists the supported e-ink display models.
type Model int

// Supported Model.
const (
	PHAT Model = iota
	// TODO: Add wHAT here when supported.
)

// Opts is the options to specify which device is being controlled and its
// default settings.
type Opts struct {
	// Model being used.
	Model Model
	// Model color.
	ModelColor Color
	// Initial border color. Will be set on the first Draw().
	BorderColor Color
}

// New opens a handle to an Inky pHAT.
func New(p spi.Port, dc gpio.PinOut, reset gpio.PinOut, busy gpio.PinIn, o *Opts) (*Dev, error) {
	if o.ModelColor != Black && o.ModelColor != Red && o.ModelColor != Yellow {
		return nil, fmt.Errorf("unsupported color: %v", o.ModelColor)
	}

	c, err := p.Connect(488*physic.KiloHertz, spi.Mode0, 8)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inky over spi: %v", err)
	}

	d := &Dev{
		c:      c,
		dc:     dc,
		r:      reset,
		busy:   busy,
		color:  o.ModelColor,
		border: o.BorderColor,
	}

	return d, nil
}

// Dev is a handle to an Inky.
type Dev struct {
	c conn.Conn
	// Low when sending a command, high when sending data.
	dc gpio.PinOut
	// Reset pin, active low.
	r gpio.PinOut
	// High when device is busy.
	busy gpio.PinIn

	// Color of device screen (red, yellow or black).
	color Color
	// Modifiable color of border.
	border Color
}

// SetBorder changes the border color. This will not take effect until the next Draw().
func (d *Dev) SetBorder(c Color) {
	d.border = c
}

// String implements conn.Resource.
func (d *Dev) String() string {
	return "Inky pHAT"
}

// Halt implements conn.Resource
func (d *Dev) Halt() error {
	return nil
}

// ColorModel implements display.Drawer
// Maps white to white, black to black and anything else as red. Red is used as
// a placeholder for the display's third color, i.e., red or yellow.
func (d *Dev) ColorModel() color.Model {
	return color.ModelFunc(func(c color.Color) color.Color {
		r, g, b, _ := c.RGBA()
		if r == 0 && g == 0 && b == 0 {
			return color.RGBA{
				R: 0,
				G: 0,
				B: 0,
				A: 255,
			}
		} else if r == 0xffff && g == 0xffff && b == 0xffff {
			return color.RGBA{
				R: 255,
				G: 255,
				B: 255,
				A: 255,
			}
		}
		return color.RGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 255,
		}
	})
}

// Bounds implements display.Drawer
func (d *Dev) Bounds() image.Rectangle {
	return image.Rect(0, 0, rows, cols)
}

// Draw implements display.Drawer
func (d *Dev) Draw(dstRect image.Rectangle, src image.Image, srcPtrs image.Point) error {
	if dstRect != d.Bounds() {
		return fmt.Errorf("partial update not supported")
	}

	if src.Bounds() != d.Bounds() {
		return fmt.Errorf("image must be the same size as bounds: %v", d.Bounds())
	}

	b := src.Bounds()
	// Black/white pixels.
	white := make([]bool, rows*cols)
	// Red/Transparent pixels.
	red := make([]bool, rows*cols)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			i := x*cols + y
			srcX := x
			srcY := b.Max.Y - y - 1
			r, g, b, _ := d.ColorModel().Convert(src.At(srcX, srcY)).RGBA()
			if r >= 0x8000 && g >= 0x8000 && b >= 0x8000 {
				white[i] = true
			} else if r >= 0x8000 {
				// Red pixels also need white behind them.
				white[i] = true
				red[i] = true
			}
		}
	}

	bufA, _ := pack(white)
	bufB, _ := pack(red)
	return d.update(borderColor[d.border], bufA, bufB)
}

func (d *Dev) update(border byte, black []byte, red []byte) (err error) {
	if err := d.reset(); err != nil {
		return err
	}

	if err := d.sendCommand(0x74, []byte{0x54}); err != nil { // Set Analog Block Control.
		return err
	}
	if err := d.sendCommand(0x7e, []byte{0x3b}); err != nil { // Set Digital Block Control.
		return err
	}

	r := make([]byte, 3)
	binary.LittleEndian.PutUint16(r, rows)
	if err := d.sendCommand(0x01, r); err != nil { // Gate setting
		return err
	}

	init := []struct {
		cmd  byte
		data []byte
	}{
		{0x03, []byte{0x10, 0x01}}, // Gate Driving Voltage.
		{0x3a, []byte{0x07}},       // Dummy line period
		{0x3b, []byte{0x04}},       // Gate line width
		{0x11, []byte{0x03}},       // Data entry mode setting 0x03 = X/Y increment
		{0x04, nil},                // Power on
		{0x2c, []byte{0x3c}},       // VCOM Register, 0x3c = -1.5v?
		{0x3c, []byte{0x00}},
		{0x3c, []byte{byte(border)}}, // Border colour
	}

	for _, c := range init {
		if err := d.sendCommand(c.cmd, c.data); err != nil {
			return err
		}
	}

	switch d.color {
	case Black:
		if err := d.sendCommand(0x32, blackLUT[:]); err != nil {
			return err
		}
	case Red:
		if err := d.sendCommand(0x32, redLUT[:]); err != nil {
			return err
		}
	case Yellow:
		if err := d.sendCommand(0x04, []byte{0x07}); err != nil { // Set voltage of VSH and VSL.
			return err
		}
		if err := d.sendCommand(0x32, yellowLUT[:]); err != nil {
			return err
		}
	}

	h := make([]byte, 4)
	binary.LittleEndian.PutUint16(h[2:], rows)
	write := []struct {
		cmd  byte
		data []byte
	}{
		{0x44, []byte{0x00, cols/8 - 1}}, // Set RAM X Start/End
		{0x45, h},                        // Set RAM Y Start/End
		{0x43, []byte{0x00}},

		{0x4e, []byte{0x00}},
		{0x4f, []byte{0x00, 0x00}},
		{0x24, black},

		{0x43, []byte{0x00}},
		{0x4f, []byte{0x00, 0x00}},
		{0x26, red},

		{0x22, []byte{0xc7}},
	}

	for _, c := range write {
		if err := d.sendCommand(c.cmd, c.data); err != nil {
			return err
		}
	}

	if err := d.busy.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		return err
	}
	defer func() {
		if err2 := d.busy.In(gpio.PullUp, gpio.NoEdge); err2 != nil {
			err = err2
		}
	}()
	if err := d.sendCommand(0x20, nil); err != nil {
		return err
	}

	d.busy.WaitForEdge(-1)

	if err := d.sendCommand(0x10, []byte{0x01}); err != nil { // Enter deep sleep.
		return err
	}
	return
}

func (d *Dev) reset() (err error) {
	if err = d.r.Out(gpio.Low); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err = d.r.Out(gpio.High); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)

	if err = d.busy.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		return err
	}
	defer func() {
		if err2 := d.busy.In(gpio.PullUp, gpio.NoEdge); err2 != nil {
			err = err2
		}
	}()
	if err := d.sendCommand(0x12, nil); err != nil { // Soft Reset
		return fmt.Errorf("failed to reset inky: %v", err)
	}
	d.busy.WaitForEdge(-1)
	return
}

func (d *Dev) sendCommand(command byte, data []byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	if err := d.c.Tx([]byte{command}, nil); err != nil {
		return fmt.Errorf("failed to send command %x to inky: %v", command, err)
	}
	if data != nil {
		if err := d.sendData(data); err != nil {
			return fmt.Errorf("failed to send data for command %x to inky: %v", command, err)
		}
	}
	return nil
}

func (d *Dev) sendData(data []byte) error {
	if len(data) > 4096 {
		return fmt.Errorf("sending more data than chunk size: %d > 4096", len(data))
	}
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}
	if err := d.c.Tx(data, nil); err != nil {
		return fmt.Errorf("failed to send data to inky: %v", err)
	}
	return nil
}

func pack(bits []bool) ([]byte, error) {
	if len(bits)%8 != 0 {
		return nil, fmt.Errorf("len(bits) must be multiple of 8 but is %d", len(bits))
	}

	ret := make([]byte, len(bits)/8)
	for i, b := range bits {
		if b {
			ret[i/8] |= 1 << (7 - uint(i)%8)
		}
	}
	return ret, nil
}

var _ display.Drawer = &Dev{}
var _ conn.Resource = &Dev{}
