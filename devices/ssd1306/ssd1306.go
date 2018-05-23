// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ssd1306 controls a 128x64 monochrome OLED display via a SSD1306
// controller.
//
// The driver does differential updates: it only sends modified pixels for the
// smallest rectangle, to economize bus bandwidth. This is especially important
// when using I²C as the bus default speed (often 100kHz) is slow enough to
// saturate the bus at less than 10 frames per second.
//
// The SSD1306 is a write-only device. It can be driven on either I²C or SPI
// with 4 wires. Changing between protocol is likely done through resistor
// soldering, for boards that support both.
//
// Some boards expose a RES / Reset pin. If present, it must be normally be
// High. When set to Low (Ground), it enables the reset circuitry. It can be
// used externally to this driver, if used, the driver must be reinstantiated.
//
// Datasheets
//
// Product page:
// http://www.solomon-systech.com/en/product/display-ic/oled-driver-controller/ssd1306/
//
// https://cdn-shop.adafruit.com/datasheets/SSD1306.pdf
//
// "DM-OLED096-624": https://drive.google.com/file/d/0B5lkVYnewKTGaEVENlYwbDkxSGM/view
//
// "ssd1306": https://drive.google.com/file/d/0B5lkVYnewKTGYzhyWWp0clBMR1E/view
package ssd1306

// Some have SPI enabled;
// https://hallard.me/adafruit-oled-display-driver-for-pi/
// https://learn.adafruit.com/ssd1306-oled-displays-with-raspberry-pi-and-beaglebone-black?view=all

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/devices"
	"periph.io/x/periph/devices/ssd1306/image1bit"
)

// FrameRate determines scrolling speed.
type FrameRate byte

// Possible frame rates. The value determines the number of refreshes between
// movement. The lower value, the higher speed.
const (
	FrameRate2   FrameRate = 7
	FrameRate3   FrameRate = 4
	FrameRate4   FrameRate = 5
	FrameRate5   FrameRate = 0
	FrameRate25  FrameRate = 6
	FrameRate64  FrameRate = 1
	FrameRate128 FrameRate = 2
	FrameRate256 FrameRate = 3
)

// Orientation is used for scrolling.
type Orientation byte

// Possible orientations for scrolling.
const (
	Left    Orientation = 0x27
	Right   Orientation = 0x26
	UpRight Orientation = 0x29
	UpLeft  Orientation = 0x2A
)

// DefaultOpts is the recommended default options.
var DefaultOpts = Opts{
	W:       128,
	H:       64,
	Rotated: false,
}

// Opts defines the options for the device.
type Opts struct {
	W int
	H int
	// Rotated determines if the display is rotated by 180°.
	Rotated bool
}

// NewSPI returns a Dev object that communicates over SPI to a SSD1306 display
// controller.
//
// The SSD1306 can operate at up to 3.3Mhz, which is much higher than I²C. This
// permits higher refresh rates.
//
// Wiring
//
// Connect SDA to MOSI, SCK to SCLK, CS to CS.
//
// In 3-wire SPI mode, pass nil for 'dc'. In 4-wire SPI mode, pass a GPIO pin
// to use.
//
// The RES (reset) pin can be used outside of this driver but is not supported
// natively. In case of external reset via the RES pin, this device drive must
// be reinstantiated.
func NewSPI(p spi.Port, dc gpio.PinOut, opts *Opts) (*Dev, error) {
	if dc == gpio.INVALID {
		return nil, errors.New("ssd1306: use nil for dc to use 3-wire mode, do not use gpio.INVALID")
	}
	bits := 8
	if dc == nil {
		// 3-wire SPI uses 9 bits per word.
		bits = 9
	} else if err := dc.Out(gpio.Low); err != nil {
		return nil, err
	}
	c, err := p.Connect(3300000, spi.Mode0, bits)
	if err != nil {
		return nil, err
	}
	return newDev(c, opts, true, dc)
}

// NewI2C returns a Dev object that communicates over I²C to a SSD1306 display
// controller.
func NewI2C(i i2c.Bus, opts *Opts) (*Dev, error) {
	// Maximum clock speed is 1/2.5µs = 400KHz.
	return newDev(&i2c.Dev{Bus: i, Addr: 0x3C}, opts, false, nil)
}

// Dev is an open handle to the display controller.
type Dev struct {
	// Communication
	c   conn.Conn
	dc  gpio.PinOut
	spi bool

	// Display size controlled by the SSD1306.
	rect image.Rectangle

	// Mutable
	// See page 25 for the GDDRAM pages structure.
	// Narrow screen will waste the end of each page.
	// Short screen will ignore the lower pages.
	// There is 8 pages, each covering an horizontal band of 8 pixels high (1
	// byte) for 128 bytes.
	// 8*128 = 1024 bytes total for 128x64 display.
	buffer []byte
	// next is lazy initialized on first Draw(). Write() skips this buffer.
	next               *image1bit.VerticalLSB
	startPage, endPage int
	startCol, endCol   int
	scrolled           bool
	halted             bool
	err                error
}

func (d *Dev) String() string {
	if d.spi {
		return fmt.Sprintf("ssd1360.Dev{%s, %s, %s}", d.c, d.dc, d.rect.Max)
	}
	return fmt.Sprintf("ssd1360.Dev{%s, %s}", d.c, d.rect.Max)
}

// ColorModel implements devices.Display.
//
// It is a one bit color model, as implemented by image1bit.Bit.
func (d *Dev) ColorModel() color.Model {
	return image1bit.BitModel
}

// Bounds implements devices.Display. Min is guaranteed to be {0, 0}.
func (d *Dev) Bounds() image.Rectangle {
	return d.rect
}

// Draw implements devices.Display.
//
// It draws synchronously, once this function returns, the display is updated.
// It means that on slow bus  (I²C), it may be preferable to defer Draw() calls
// to a background goroutine.
//
// It discards any failure.
func (d *Dev) Draw(r image.Rectangle, src image.Image, sp image.Point) {
	var next []byte
	if img, ok := src.(*image1bit.VerticalLSB); ok && r == d.Bounds() && src.Bounds() == d.rect && sp.X == 0 && sp.Y == 0 {
		// Exact size, full frame, image1bit encoding: fast path!
		next = img.Pix
	} else {
		// Double buffering.
		if d.next == nil {
			d.next = image1bit.NewVerticalLSB(d.rect)
		}
		next = d.next.Pix
		draw.Src.Draw(d.next, r, src, sp)
	}
	d.err = d.drawInternal(next)
}

// Err returns the last error that occurred
func (d *Dev) Err() error {
	return d.err
}

// Write writes a buffer of pixels to the display.
//
// The format is unsual as each byte represent 8 vertical pixels at a time. The
// format is horizontal bands of 8 pixels high.
//
// This function accepts the content of image1bit.VerticalLSB.Pix.
func (d *Dev) Write(pixels []byte) (int, error) {
	if len(pixels) != len(d.buffer) {
		return 0, fmt.Errorf("ssd1306: invalid pixel stream length; expected %d bytes, got %d bytes", len(d.buffer), len(pixels))
	}
	// Write() skips d.next so it saves 1kb of RAM.
	if err := d.drawInternal(pixels); err != nil {
		return 0, err
	}
	return len(pixels), nil
}

// Scroll scrolls an horizontal band.
//
// Only one scrolling operation can happen at a time.
//
// Both startLine and endLine must be multiples of 8.
//
// Use -1 for endLine to extend to the bottom of the display.
func (d *Dev) Scroll(o Orientation, rate FrameRate, startLine, endLine int) error {
	h := d.rect.Dy()
	if endLine == -1 {
		endLine = h
	}
	if startLine >= endLine {
		return fmt.Errorf("startLine (%d) must be lower than endLine (%d)", startLine, endLine)
	}
	if startLine&7 != 0 || startLine < 0 || startLine >= h {
		return fmt.Errorf("invalid startLine %d", startLine)
	}
	if endLine&7 != 0 || endLine < 0 || endLine > h {
		return fmt.Errorf("invalid endLine %d", endLine)
	}

	startPage := uint8(startLine / 8)
	endPage := uint8(endLine / 8)
	d.scrolled = true
	if o == Left || o == Right {
		// page 28
		// <op>, dummy, <start page>, <rate>,  <end page>, <dummy>, <dummy>, <ENABLE>
		return d.sendCommand([]byte{byte(o), 0x00, startPage, byte(rate), endPage - 1, 0x00, 0xFF, 0x2F})
	}
	// page 29
	// <op>, dummy, <start page>, <rate>,  <end page>, <offset>, <ENABLE>
	// page 30: 0xA3 permits to set rows for scroll area.
	return d.sendCommand([]byte{byte(o), 0x00, startPage, byte(rate), endPage - 1, 0x01, 0x2F})
}

// StopScroll stops any scrolling previously set and resets the screen.
func (d *Dev) StopScroll() error {
	return d.sendCommand([]byte{0x2E})
}

// SetContrast changes the screen contrast.
//
// Note: values other than 0xff do not seem useful...
func (d *Dev) SetContrast(level byte) error {
	return d.sendCommand([]byte{0x81, level})
}

// Halt turns off the display.
//
// Sending any other command afterward reenables the display.
func (d *Dev) Halt() error {
	d.halted = false
	err := d.sendCommand([]byte{0xAE})
	if err == nil {
		d.halted = true
	}
	return err
}

// Invert the display (black on white vs white on black).
func (d *Dev) Invert(blackOnWhite bool) error {
	b := []byte{0xA6}
	if blackOnWhite {
		b[0] = 0xA7
	}
	return d.sendCommand(b)
}

//

// newDev is the common initialization code that is independent of the
// communication protocol (I²C or SPI) being used.
func newDev(c conn.Conn, opts *Opts, usingSPI bool, dc gpio.PinOut) (*Dev, error) {
	if opts.W < 8 || opts.W > 128 || opts.W&7 != 0 {
		return nil, fmt.Errorf("ssd1306: invalid width %d", opts.W)
	}
	if opts.H < 8 || opts.H > 64 || opts.H&7 != 0 {
		return nil, fmt.Errorf("ssd1306: invalid height %d", opts.H)
	}

	nbPages := opts.H / 8
	pageSize := opts.W
	d := &Dev{
		c:         c,
		spi:       usingSPI,
		dc:        dc,
		rect:      image.Rect(0, 0, opts.W, opts.H),
		buffer:    make([]byte, nbPages*pageSize),
		startPage: 0,
		endPage:   nbPages,
		startCol:  0,
		endCol:    opts.W,
		// Signal that the screen must be redrawn on first draw().
		scrolled: true,
	}
	if err := d.sendCommand(getInitCmd(opts.W, opts.H, opts.Rotated)); err != nil {
		return nil, err
	}
	return d, nil
}

func getInitCmd(w, h int, rotated bool) []byte {
	// Set COM output scan direction; C0 means normal; C8 means reversed
	comScan := byte(0xC8)
	// See page 40.
	columnAddr := byte(0xA1)
	if rotated {
		// Change order both horizontally and vertically.
		comScan = 0xC0
		columnAddr = byte(0xA0)
	}
	// Set the max frequency. The problem with I²C is that it creates visible
	// tear down. On SPI at high speed this is not visible. Page 23 pictures how
	// to avoid tear down. For now default to max frequency.
	freq := byte(0xF0)

	// Initialize the device by fully resetting all values.
	// Page 64 has the full recommended flow.
	// Page 28 lists all the commands.
	return []byte{
		0xAE,       // Display off
		0xD3, 0x00, // Set display offset; 0
		0x40,       // Start display start line; 0
		columnAddr, // Set segment remap; RESET is column 127.
		comScan,    //
		0xDA, 0x12, // Set COM pins hardware configuration; see page 40
		0x81, 0xFF, // Set max contrast
		0xA4,       // Set display to use GDDRAM content
		0xA6,       // Set normal display (0xA7 for inverted 0=lit, 1=dark)
		0xD5, freq, // Set osc frequency and divide ratio; power on reset value is 0x80.
		0x8D, 0x14, // Enable charge pump regulator; page 62
		0xD9, 0xF1, // Set pre-charge period; from adafruit driver
		0xDB, 0x40, // Set Vcomh deselect level; page 32
		0x2E,              // Deactivate scroll
		0xA8, byte(h - 1), // Set multiplex ratio (number of lines to display)
		0x20, 0x00, // Set memory addressing mode to horizontal
		0x21, 0, uint8(w - 1), // Set column address (Width)
		0x22, 0, uint8(h/8 - 1), // Set page address (Pages)
		0xAF, // Display on
	}
}

func (d *Dev) calculateSubset(next []byte) (int, int, int, int, bool) {
	w := d.rect.Dx()
	h := d.rect.Dy()
	startPage := 0
	endPage := h / 8
	startCol := 0
	endCol := w
	if d.scrolled {
		// Painting disable scrolling but if scrolling was enabled, this requires a
		// full screen redraw.
		d.scrolled = false
	} else {
		// Calculate the smallest square that need to be sent.
		pageSize := w

		// Top.
		for ; startPage < endPage; startPage++ {
			x := pageSize * startPage
			y := pageSize * (startPage + 1)
			if !bytes.Equal(d.buffer[x:y], next[x:y]) {
				break
			}
		}
		// Bottom.
		for ; endPage > startPage; endPage-- {
			x := pageSize * (endPage - 1)
			y := pageSize * endPage
			if !bytes.Equal(d.buffer[x:y], next[x:y]) {
				break
			}
		}
		if startPage == endPage {
			// Early exit, the image is exactly the same.
			return 0, 0, 0, 0, true
		}
		// TODO(maruel): This currently corrupts the screen. Likely a small error
		// in the way the commands are sent.
		/*
				// Left.
				for ; startCol < endCol; startCol++ {
					for i := startPage; i < endPage; i++ {
						x := i*pageSize + startCol
						if d.buffer[x] != next[x] {
							goto breakLeft
						}
					}
				}
			breakLeft:
				// Right.
				for ; endCol > startCol; endCol-- {
					for i := startPage; i < endPage; i++ {
						x := i*pageSize + endCol - 1
						if d.buffer[x] != next[x] {
							goto breakRight
						}
					}
				}
			breakRight:
		*/
	}
	return startPage, endPage, startCol, endCol, false
}

// drawInternal sends image data to the controller.
func (d *Dev) drawInternal(next []byte) error {
	startPage, endPage, startCol, endCol, skip := d.calculateSubset(next)
	if skip {
		return nil
	}
	copy(d.buffer, next)

	if d.startPage != startPage || d.endPage != endPage || d.startCol != startCol || d.endCol != endCol {
		d.startPage = startPage
		d.endPage = endPage
		d.startCol = startCol
		d.endCol = endCol
		cmd := []byte{
			0x21, uint8(d.startCol), uint8(d.endCol - 1), // Set column address (Width)
			0x22, uint8(d.startPage), uint8(d.endPage - 1), // Set page address (Pages)
		}
		if err := d.sendCommand(cmd); err != nil {
			return err
		}
	}

	// Write the subset of the data as needed.
	pageSize := d.rect.Dx()
	return d.sendData(d.buffer[startPage*pageSize+startCol : (endPage-1)*pageSize+endCol])
}

func (d *Dev) sendData(c []byte) error {
	if d.halted {
		// Transparently enable the display.
		if err := d.sendCommand(nil); err != nil {
			return err
		}
	}
	if d.spi {
		// 4-wire SPI.
		if err := d.dc.Out(gpio.High); err != nil {
			return err
		}
		return d.c.Tx(c, nil)
	}
	return d.c.Tx(append([]byte{i2cData}, c...), nil)
}

func (d *Dev) sendCommand(c []byte) error {
	if d.halted {
		// Transparently enable the display.
		c = append([]byte{0xAF}, c...)
		d.halted = false
	}
	if d.spi {
		if d.dc == nil {
			// 3-wire SPI.
			return errors.New("ssd1306: 3-wire SPI mode is not yet implemented")
		}
		// 4-wire SPI.
		if err := d.dc.Out(gpio.Low); err != nil {
			return err
		}
		return d.c.Tx(c, nil)
	}
	return d.c.Tx(append([]byte{i2cCmd}, c...), nil)
}

const (
	i2cCmd  = 0x00 // I²C transaction has stream of command bytes
	i2cData = 0x40 // I²C transaction has stream of data bytes
)

var _ conn.Resource = &Dev{}
var _ devices.Display = &Dev{}
var _ fmt.Stringer = &Dev{}
