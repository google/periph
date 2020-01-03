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

func (c *Color) String() string {
	switch *c {
	case Black:
		return "black"
	case Red:
		return "red"
	case Yellow:
		return "yellow"
	case White:
		return "white"
	default:
		return "unknown"
	}
}

// Set sets the Color to a value represented by the string s. Set implements the flag.Value interface.
func (c *Color) Set(s string) error {
	switch s {
	case "black":
		*c = Black
	case "red":
		*c = Red
	case "yellow":
		*c = Yellow
	case "white":
		*c = White
	default:
		return fmt.Errorf("Unknown color %q: expected either black, red, yellow or white", s)
	}
	return nil
}

// Model lists the supported e-ink display models.
type Model int

// Supported Model.
const (
	PHAT Model = iota
	WHAT
)

func (m *Model) String() string {
	switch *m {
	case PHAT:
		return "PHAT"
	case WHAT:
		return "WHAT"
	default:
		return "Unknown"
	}
}

// Set sets the Model to a value represented by the string s. Set implements the flag.Value interface.
func (m *Model) Set(s string) error {
	switch s {
	case "PHAT":
		*m = PHAT
	case "WHAT":
		*m = WHAT
	default:
		return fmt.Errorf("Unknown model %q: expected either PHAT or WHAT", s)
	}
	return nil
}

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

var borderColor = map[Color]byte{
	Black:  0x00,
	Red:    0x73,
	Yellow: 0x33,
	White:  0x31,
}

// New opens a handle to an Inky pHAT or wHAT.
func New(p spi.Port, dc gpio.PinOut, reset gpio.PinOut, busy gpio.PinIn, o *Opts) (*Dev, error) {
	if o.ModelColor != Black && o.ModelColor != Red && o.ModelColor != Yellow {
		return nil, fmt.Errorf("unsupported color: %v", o.ModelColor)
	}

	c, err := p.Connect(488*physic.KiloHertz, spi.Mode0, 8)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inky over spi: %v", err)
	}

	// Get the maxTxSize from the conn if it implements the conn.Limits interface,
	// otherwise use 4096 bytes.
	maxTxSize := 0
	if limits, ok := c.(conn.Limits); ok {
		maxTxSize = limits.MaxTxSize()
	}
	if maxTxSize == 0 {
		maxTxSize = 4096 // Use a conservative default.
	}

	d := &Dev{
		c:         c,
		maxTxSize: maxTxSize,
		dc:        dc,
		r:         reset,
		busy:      busy,
		color:     o.ModelColor,
		border:    o.BorderColor,
	}

	switch o.Model {
	case PHAT:
		d.bounds = image.Rect(0, 0, 104, 212)
		d.flipVertically = true
	case WHAT:
		d.bounds = image.Rect(0, 0, 400, 300)
	}

	return d, nil
}

// Dev is a handle to an Inky.
type Dev struct {
	c conn.Conn
	// Maximum number of bytes allowed to be sent as a single I/O on c.
	maxTxSize int
	// Low when sending a command, high when sending data.
	dc gpio.PinOut
	// Reset pin, active low.
	r gpio.PinOut
	// High when device is busy.
	busy gpio.PinIn
	// Size of this model's display.
	bounds image.Rectangle
	// Whether this model needs the image flipped vertically.
	flipVertically bool

	// Color of device screen (red, yellow or black).
	color Color
	// Modifiable color of border.
	border Color
}

// SetBorder changes the border color. This will not take effect until the next Draw().
func (d *Dev) SetBorder(c Color) {
	d.border = c
}

// SetModelColor changes the model color. This will not take effect until the next Draw().
// Useful if you want to switch between two-color and three-color drawing.
func (d *Dev) SetModelColor(c Color) error {
	if c != Black && c != Red && c != Yellow {
		return fmt.Errorf("unsupported color: %v", c)
	}
	d.color = c
	return nil
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
	return d.bounds
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
	white := make([]bool, b.Size().Y*b.Size().X)
	// Red/Transparent pixels.
	red := make([]bool, b.Size().Y*b.Size().X)
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			i := y*b.Size().X + x
			srcX := x
			srcY := y
			if d.flipVertically {
				srcY = b.Max.Y - y - 1
			}
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

// DrawAll redraws the whole display.
func (d *Dev) DrawAll(src image.Image) error {
	return d.Draw(d.Bounds(), src, image.ZP)
}

func (d *Dev) update(border byte, black []byte, red []byte) (err error) {
	if err := d.reset(); err != nil {
		return err
	}

	r := make([]byte, 3)
	binary.LittleEndian.PutUint16(r, uint16(d.Bounds().Size().Y))
	h := make([]byte, 4)
	binary.LittleEndian.PutUint16(h[2:], uint16(d.Bounds().Size().Y))

	type cmdData struct {
		cmd  byte
		data []byte
	}
	cmds := []cmdData{
		{0x01, r},                        // Gate setting
		{0x74, []byte{0x54}},             // Set Analog Block Control.
		{0x7e, []byte{0x3b}},             // Set Digital Block Control.
		{0x03, []byte{0x17}},             // Gate Driving Voltage.
		{0x04, []byte{0x41, 0xac, 0x32}}, // Gate Driving Voltage.
		{0x3a, []byte{0x07}},             // Dummy line period
		{0x3b, []byte{0x04}},             // Gate line width
		{0x11, []byte{0x03}},             // Data entry mode setting 0x03 = X/Y increment
		{0x2c, []byte{0x3c}},             // VCOM Register, 0x3c = -1.5v?
		{0x3c, []byte{0x00}},
		{0x3c, []byte{byte(border)}}, // Border colour
		{0x32, modelLUT[d.color]},    // Set LUTs.
		{0x44, []byte{0x00, byte(d.Bounds().Size().X/8) - 1}}, // Set RAM Y Start/End
		{0x45, h},                  // Set RAM X Start/End
		{0x4e, []byte{0x00}},       // Set RAM X Pointer Start
		{0x4f, []byte{0x00, 0x00}}, // Set RAM Y Pointer Start
		{0x24, black},
		{0x4e, []byte{0x00}},       // Set RAM X Pointer Start
		{0x4f, []byte{0x00, 0x00}}, // Set RAM Y Pointer Start
		{0x26, red},
	}
	if d.color == Yellow {
		cmds = append(cmds, cmdData{0x04, []byte{0x07, 0xac, 0x32}}) // Set voltage of VSH and VSL
	}
	cmds = append(cmds, cmdData{0x22, []byte{0xc7}}) // Update the image.

	for _, c := range cmds {
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
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}
	for len(data) != 0 {
		var chunk []byte
		if len(data) > d.maxTxSize {
			chunk, data = data[:d.maxTxSize], data[d.maxTxSize:]
		} else {
			chunk, data = data, nil
		}
		if err := d.c.Tx(chunk, nil); err != nil {
			return fmt.Errorf("failed to send data to inky: %v", err)
		}
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
