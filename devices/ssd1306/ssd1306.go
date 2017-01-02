// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ssd1306 controls a 128x64 monochrome OLED display via a ssd1306
// controler.
//
// The SSD1306 is a write-only device. It can be driven on either I²C or SPI.
// Changing between protocol is likely done through resistor soldering, for
// boards that support both.
//
// Known issue
//
// The SPI version of this driver is not functional. To interface with the ssd1306
// in 3-wire SPI mode each byte must be transmitted using 9 bits where the 9th bit
// discriminates between command & data. To interface using 4-wire SPI a separate
// gpio is needed to drive a c/d input. Neither of these two mechanisms have been
// implemented yet.
// For more info, see
// https://drive.google.com/file/d/0B5lkVYnewKTGYzhyWWp0clBMR1E/view
// pages 17-18 (8.1.3, 8.1.4).
//
// Datasheets
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
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"

	"github.com/google/periph/conn/i2c"
	"github.com/google/periph/conn/spi"
	"github.com/google/periph/devices"
	"github.com/google/periph/devices/ssd1306/image1bit"
)

// FrameRate determines scrolling speed.
type FrameRate byte

// Possible frame rates.
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

// Dev is an open handle to the display controler.
type Dev struct {
	w io.Writer
	W int
	H int
}

// NewSPI returns a Dev object that communicates over SPI to SSD1306 display
// controler.
//
// If rotated, turns the display by 180°
//
// It's up to the caller to use the RES (reset) pin if desired. Simpler
// connection is to connect RES and DC to ground, CS to 3.3v, SDA to MOSI, SCK
// to SCLK.
//
// As per datasheet, maximum clock speed is 1/100ns = 10MHz.
func NewSPI(s spi.Conn, w, h int, rotated bool) (*Dev, error) {
	if err := s.Configure(spi.Mode3, 8); err != nil {
		return nil, err
	}
	return newDev(s, w, h, rotated)
}

// NewI2C returns a Dev object that communicates over I²C to SSD1306 display
// controler.
//
// If rotated, turns the display by 180°
//
// As per datasheet, maximum clock speed is 1/2.5µs = 400KHz. It's worth
// bumping up from default bus speed of 100KHz if possible.
func NewI2C(i i2c.Bus, w, h int, rotated bool) (*Dev, error) {
	return newDev(&i2c.Dev{Bus: i, Addr: 0x3C}, w, h, rotated)
}

// newDev is the common initialization code that is independent of the bus
// being used.
func newDev(dev io.Writer, w, h int, rotated bool) (*Dev, error) {
	if w < 8 || w > 128 || w&7 != 0 {
		return nil, fmt.Errorf("ssd1306: invalid width %d", w)
	}
	if h < 8 || h > 64 || h&7 != 0 {
		return nil, fmt.Errorf("ssd1306: invalid height %d", h)
	}
	d := &Dev{w: dev, W: w, H: h}

	// Set COM output scan direction; C0 means normal; C8 means reversed
	comScan := byte(0xC8)
	// See page 40.
	columnAddr := byte(0xA1)
	if rotated {
		// Change order both horizontally and vertically.
		comScan = 0xC0
		columnAddr = byte(0xA0)
	}
	// Initialize the device by fully resetting all values.
	// https://cdn-shop.adafruit.com/datasheets/SSD1306.pdf
	// Page 64 has the full recommended flow.
	// Page 28 lists all the commands.
	// Some values come from the DM-OLED096 datasheet p15.
	init := []byte{
		i2cCmd,
		0xAE,       // Display off
		0xD3, 0x00, // Set display offset; 0
		0x40,       // Start display start line; 0
		columnAddr, // Set segment remap; RESET is column 127.
		comScan,
		0xDA, 0x12, // Set COM pins hardware configuration; see page 40
		0x81, 0xff, // Set max contrast
		0xA4,       // Set display to use GDDRAM content
		0xA6,       // Set normal display (0xA7 for inverted 0=lit, 1=dark)
		0xD5, 0x80, // Set osc frequency and divide ratio; power on reset value is 0x3F.
		0x8D, 0x14, // Enable charge pump regulator; page 62
		0xD9, 0xf1, // Set pre-charge period; from adafruit driver
		0xDB, 0x40, // Set Vcomh deselect level; page 32
		0x20, 0x00, // Set memory addressing mode to horizontal
		0xB0,                // Set page start address
		0x2E,                // Deactivate scroll
		0x00,                // Set column offset (lower nibble)
		0x10,                // Set column offset (higher nibble)
		0xA8, byte(d.H - 1), // Set multiplex ratio (number of lines to display)
		0xAF, // Display on
	}
	if _, err := d.w.Write(init); err != nil {
		return nil, err
	}

	/* For reference, init sequence from adafruit driver:
	d.w.Write([]byte{i2cCmd,
		0xae, 0xd5, 0x80, 0xa8, 0x3f, 0xd3, 0x00, 0x40, 0x8d, 0x14,
		0x20, 0x00, 0xa1, 0xc8, 0xda, 0x12, 0x81, 0xcf, 0xd9, 0xf1,
		0xdb, 0x40, 0xa4, 0xa6, 0x2e, 0xaf})
	*/

	return d, nil
}

// ColorModel implements devices.Display. It is a one bit color model.
func (d *Dev) ColorModel() color.Model {
	return color.NRGBAModel
}

// Bounds implements devices.Display. Min is guaranteed to be {0, 0}.
func (d *Dev) Bounds() image.Rectangle {
	return image.Rectangle{Max: image.Point{X: d.W, Y: d.H}}
}

func colorToBit(c color.Color) byte {
	r, g, b, a := c.RGBA()
	if (r|g|b) >= 0x8000 && a >= 0x4000 {
		return 1
	}
	return 0
}

// Draw implements devices.Display.
//
// BUG(maruel): It discards any failure. Change devices.Display interface?
// BUG(maruel): Support r.Min.Y and r.Max.Y not divisible by 8.
// BUG(maruel): Support sp.Y not divisible by 8.
func (d *Dev) Draw(r image.Rectangle, src image.Image, sp image.Point) {
	r = r.Intersect(d.Bounds())
	srcR := src.Bounds()
	srcR.Min = srcR.Min.Add(sp)
	if dX := r.Dx(); dX < srcR.Dx() {
		srcR.Max.X = srcR.Min.X + dX
	}
	if dY := r.Dy(); dY < srcR.Dy() {
		srcR.Max.Y = srcR.Min.Y + dY
	}
	// Take 8 lines at a time.
	deltaX := r.Min.X - srcR.Min.X
	deltaY := r.Min.Y - srcR.Min.Y

	var pixels []byte
	if img, ok := src.(*image1bit.Image); ok {
		if srcR.Min.X == 0 && srcR.Dx() == d.W && srcR.Min.Y == 0 && srcR.Dy() == d.H {
			// Fast path.
			pixels = img.Buf
		}
	}
	if pixels == nil {
		pixels = make([]byte, d.W*d.H/8)
		for sY := srcR.Min.Y; sY < srcR.Max.Y; sY += 8 {
			rY := ((sY + deltaY) / 8) * d.W
			for sX := srcR.Min.X; sX < srcR.Max.X; sX++ {
				rX := sX + deltaX
				c0 := colorToBit(src.At(sX, sY))
				c1 := colorToBit(src.At(sX, sY+1)) << 1
				c2 := colorToBit(src.At(sX, sY+2)) << 2
				c3 := colorToBit(src.At(sX, sY+3)) << 3
				c4 := colorToBit(src.At(sX, sY+4)) << 4
				c5 := colorToBit(src.At(sX, sY+5)) << 5
				c6 := colorToBit(src.At(sX, sY+6)) << 6
				c7 := colorToBit(src.At(sX, sY+7)) << 7
				pixels[rX+rY] = c0 | c1 | c2 | c3 | c4 | c5 | c6 | c7
			}
		}
	}
	if _, err := d.Write(pixels); err != nil {
		log.Printf("ssd1306: Draw failed: %v", err)
	}
}

// Write writes a buffer of pixels to the display.
//
// The format is unsual as each byte represent 8 vertical pixels at a time. So
// the memory is effectively horizontal bands of 8 pixels high.
func (d *Dev) Write(pixels []byte) (int, error) {
	if len(pixels) != d.H*d.W/8 {
		return 0, errors.New("ssd1306: invalid pixel stream")
	}

	// Run as 2 big transactions to reduce downtime on the bus.
	// First tx is commands, second is data.

	// The following commands should not be needed, but then if the ssd1306 gets out of sync
	// for some reason the display ends up messed-up. Given the small overhead compared to
	// sending all the data might as well reset things a bit.
	hdr := []byte{
		i2cCmd,
		0xB0,       // Set page start addr just in case
		0x00, 0x10, // Set column start addr, lower & upper nibble
		0x20, 0x00, // Ensure addressing mode is horizontal
		0x21, 0x00, byte(d.W - 1), // Set start/end column
		0x22, 0x00, byte(d.H/8 - 1), // Set start/end page
	}
	if _, err := d.w.Write(hdr); err != nil {
		return 0, err
	}

	// Write the data.
	if _, err := d.w.Write(append([]byte{i2cData}, pixels...)); err != nil {
		return 0, err
	}

	return len(pixels), nil
}

// Scroll scrolls the entire screen.
func (d *Dev) Scroll(o Orientation, rate FrameRate) error {
	// TODO(maruel): Allow to specify page.
	// TODO(maruel): Allow to specify offset.
	if o == Left || o == Right {
		// page 28
		// STOP, <op>, dummy, <start page>, <rate>,  <end page>, <dummy>, <dummy>, <ENABLE>
		_, err := d.w.Write([]byte{i2cCmd, 0x2E, byte(o), 0x00, 0x00, byte(rate), 0x07, 0x00, 0xFF, 0x2F})
		return err
	}
	// page 29
	// STOP, <op>, dummy, <start page>, <rate>,  <end page>, <offset>, <ENABLE>
	// page 30: 0xA3 permits to set rows for scroll area.
	_, err := d.w.Write([]byte{i2cCmd, 0x2E, byte(o), 0x00, 0x00, byte(rate), 0x07, 0x01, 0x2F})
	return err
}

// StopScroll stops any scrolling previously set.
//
// It will only take effect after redrawing the ram.
func (d *Dev) StopScroll() error {
	_, err := d.w.Write([]byte{i2cCmd, 0x2E})
	return err
}

// SetContrast changes the screen contrast.
//
// Note: values other than 0xff do not seem useful...
func (d *Dev) SetContrast(level byte) error {
	_, err := d.w.Write([]byte{i2cCmd, 0x81, level})
	return err
}

// Enable or disable the display.
func (d *Dev) Enable(on bool) error {
	b := byte(0xAE)
	if on {
		b = 0xAF
	}
	_, err := d.w.Write([]byte{i2cCmd, b})
	return err
}

// Invert the display (black on white vs white on black).
func (d *Dev) Invert(blackOnWhite bool) error {
	b := byte(0xA6)
	if blackOnWhite {
		b = 0xA7
	}
	_, err := d.w.Write([]byte{i2cCmd, b})
	return err
}

const (
	i2cCmd  = 0x00 // i2c transaction has stream of command bytes
	i2cData = 0x40 // i2c transaction has stream of data bytes
)

var _ devices.Display = &Dev{}
