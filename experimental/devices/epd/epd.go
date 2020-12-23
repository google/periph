// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package epd

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"

	"periph.io/x/periph/host/rpi"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/display"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/devices/ssd1306/image1bit"
)

// EPD commands
const (
	driverOutputControl            byte = 0x01
	boosterSoftStartControl        byte = 0x0C
	gateScanStartPosition          byte = 0x0F
	deepSleepMode                  byte = 0x10
	dataEntryModeSetting           byte = 0x11
	swReset                        byte = 0x12
	temperatureSensorControl       byte = 0x1A
	masterActivation               byte = 0x20
	displayUpdateControl1          byte = 0x21
	displayUpdateControl2          byte = 0x22
	writeRAM                       byte = 0x24
	writeVcomRegister              byte = 0x2C
	writeLutRegister               byte = 0x32
	setDummyLinePeriod             byte = 0x3A
	setGateTime                    byte = 0x3B
	borderWaveformControl          byte = 0x3C
	setRAMXAddressStartEndPosition byte = 0x44
	setRAMYAddressStartEndPosition byte = 0x45
	setRAMXAddressCounter          byte = 0x4E
	setRAMYAddressCounter          byte = 0x4F
	terminateFrameReadWrite        byte = 0xFF
)

// LUT contains the display specific waveform for the pixel programming of the display.
type LUT []byte

// PartialUpdate represents if updates to the display should be full or partial.
type PartialUpdate bool

const (
	// Full LUT config to update all the display
	Full PartialUpdate = false
	// Partial LUT config only a part of the display
	Partial PartialUpdate = true
)

// EPD2in13 is the config for the 2.13 inch display.
var EPD2in13 = Opts{
	W: 128,
	H: 250,
	FullUpdate: LUT{
		0x22, 0x55, 0xAA, 0x55, 0xAA, 0x55, 0xAA, 0x11,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x1E, 0x1E, 0x1E, 0x1E, 0x1E, 0x1E, 0x1E, 0x1E,
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
	},
	PartialUpdate: LUT{
		0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x0F, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	},
}

// EPD1in54 is the config for the 1.54 inch display.
var EPD1in54 = Opts{
	W: 200,
	H: 200,
	FullUpdate: LUT{
		0x02, 0x02, 0x01, 0x11, 0x12, 0x12, 0x22, 0x22,
		0x66, 0x69, 0x69, 0x59, 0x58, 0x99, 0x99, 0x88,
		0x00, 0x00, 0x00, 0x00, 0xF8, 0xB4, 0x13, 0x51,
		0x35, 0x51, 0x51, 0x19, 0x01, 0x00,
	},
	PartialUpdate: LUT{
		0x10, 0x18, 0x18, 0x08, 0x18, 0x18, 0x08, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x13, 0x14, 0x44, 0x12,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	},
}

// Opts defines the options for the ePaper Device.
type Opts struct {
	W             int
	H             int
	FullUpdate    LUT
	PartialUpdate LUT
}

// NewSPI returns a Dev object that communicates over SPI to a E-Paper display controller.
func NewSPI(p spi.Port, dc, cs, rst gpio.PinOut, busy gpio.PinIO, opts *Opts) (*Dev, error) {
	if dc == gpio.INVALID {
		return nil, errors.New("epd: use nil for dc to use 3-wire mode, do not use gpio.INVALID")
	}

	if err := dc.Out(gpio.Low); err != nil {
		return nil, err
	}

	c, err := p.Connect(5*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return nil, err
	}

	d := &Dev{
		c:      c,
		dc:     dc,
		cs:     cs,
		rst:    rst,
		busy:   busy,
		update: Full,
		opts:   opts,
		rect:   image.Rect(0, 0, opts.W, opts.H),
	}

	d.Reset()

	if err := d.Init(); err != nil {
		return nil, err
	}

	return d, nil
}

// NewSPIHat returns a Dev object that communicates over SPI
// and have the default config for the e-paper hat for raspberry pi
func NewSPIHat(p spi.Port, opts *Opts) (*Dev, error) {
	dc := rpi.P1_22
	cs := rpi.P1_24
	rst := rpi.P1_11
	busy := rpi.P1_18
	return NewSPI(p, dc, cs, rst, busy, opts)
}

// Dev is an open handle to the display controller.
type Dev struct {
	// Communication
	c    conn.Conn
	dc   gpio.PinOut
	cs   gpio.PinOut
	rst  gpio.PinOut
	busy gpio.PinIO

	// Display size controlled by the e-paper display.
	rect image.Rectangle

	update PartialUpdate
	opts   *Opts
}

func (d *Dev) String() string {
	return fmt.Sprintf("epd.Dev{%s, %s, %s}", d.c, d.dc, d.rect.Max)
}

// ColorModel implements display.Drawer.
// It is a one bit color model, as implemented by image1bit.Bit.
func (d *Dev) ColorModel() color.Model {
	return image1bit.BitModel
}

// Bounds implements display.Drawer. Min is guaranteed to be {0, 0}.
func (d *Dev) Bounds() image.Rectangle {
	return d.rect
}

// Draw implements display.Drawer.
func (d *Dev) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	xStart := sp.X
	yStart := sp.Y
	imageW := r.Dx() & 0xF8
	imageH := r.Dy()
	w := d.rect.Dx()
	h := d.rect.Dy()

	xEnd := xStart + imageW - 1
	if xStart+imageW >= w {
		xEnd = w - 1
	}

	yEnd := yStart + imageH - 1
	if yStart+imageH >= h {
		yEnd = h - 1
	}

	if err := d.setMemoryArea(xStart, yStart, xEnd, yEnd); err != nil {
		return err
	}

	next := image1bit.NewVerticalLSB(d.rect)
	draw.Src.Draw(next, r, src, sp)
	var byteToSend byte = 0x00
	for y := yStart; y < yEnd+1; y++ {
		if err := d.setMemoryPointer(xStart, y); err != nil {
			return err
		}
		if err := d.sendCommand([]byte{writeRAM}); err != nil {
			return err
		}
		for x := xStart; x < xEnd+1; x++ {
			bit := next.BitAt(x-xStart, y-yStart)
			if bit {
				byteToSend |= 0x80 >> (uint32(x) % 8)
			}
			if x%8 == 7 {
				if err := d.sendData([]byte{byteToSend}); err != nil {
					return err
				}
				byteToSend = 0x00
			}
		}
	}

	return nil
}

// ClearFrameMemory clear the frame memory with the specified color.
// this won't update the display.
func (d *Dev) ClearFrameMemory(color byte) error {
	w := d.rect.Dx()
	h := d.rect.Dy()
	if err := d.setMemoryArea(0, 0, w-1, h-1); err != nil {
		return err
	}
	if err := d.setMemoryPointer(0, 0); err != nil {
		return err
	}
	if err := d.sendCommand([]byte{writeRAM}); err != nil {
		return err
	}
	// send the color data
	for i := 0; i < (w / 8 * h); i++ {
		if err := d.sendData([]byte{color}); err != nil {
			return err
		}
	}
	return nil
}

// DisplayFrame update the display.
//
// There are 2 memory areas embedded in the e-paper display but once
// this function is called, the next action of SetFrameMemory or ClearFrame
// will set the other memory area.
func (d *Dev) DisplayFrame() error {
	if err := d.sendCommand([]byte{displayUpdateControl2}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0xC4)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{masterActivation}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{terminateFrameReadWrite}); err != nil {
		return err
	}

	d.waitUntilIdle()
	return nil
}

// Halt turns off the display.
func (d *Dev) Halt() error {
	return d.ClearFrameMemory(0xFF)
}

// Sleep after this command is transmitted, the chip would enter the
// deep-sleep mode to save power.
//
// The deep sleep mode would return to standby by hardware reset.
// You can use Reset() to awaken and Init to re-initialize the device.
func (d *Dev) Sleep() error {
	if err := d.sendCommand([]byte{deepSleepMode}); err != nil {
		return err
	}

	d.waitUntilIdle()
	return nil
}

// Init initialize the display config. This method is already called when creating
// a device using NewSPI and NewSPIHat methods.
//
// It should be only used when you put the device to sleep and need to re-init the device.
func (d *Dev) Init() error {
	if err := d.sendCommand([]byte{driverOutputControl}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte((d.opts.H - 1) & 0xFF)}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(((d.opts.H - 1) >> 8) & 0xFF)}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0x00)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{boosterSoftStartControl}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0xD7)}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0xD6)}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0x9D)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{writeVcomRegister}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0xA8)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{setDummyLinePeriod}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0x1A)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{setGateTime}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0x08)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{dataEntryModeSetting}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte(0x03)}); err != nil {
		return err
	}

	return d.setLut(Full)
}

// Reset can be also used to awaken the device
func (d *Dev) Reset() {
	_ = d.rst.Out(gpio.Low)
	time.Sleep(200 * time.Millisecond)
	_ = d.rst.Out(gpio.High)
	time.Sleep(200 * time.Millisecond)
}

func (d *Dev) setMemoryPointer(x, y int) error {
	if err := d.sendCommand([]byte{setRAMXAddressCounter}); err != nil {
		return err
	}

	// x point must be the multiple of 8 or the last 3 bits will be ignored
	if err := d.sendData([]byte{byte((x >> 3) & 0xFF)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{setRAMYAddressCounter}); err != nil {
		return err
	}
	if err := d.sendData([]byte{byte(y & 0xFF)}); err != nil {
		return err
	}
	if err := d.sendData([]byte{byte((y >> 8) & 0xFF)}); err != nil {
		return err
	}

	d.waitUntilIdle()

	return nil
}

func (d *Dev) waitUntilIdle() {
	for d.busy.Read() == gpio.High {
		time.Sleep(100 * time.Millisecond)
	}
}

func (d *Dev) setMemoryArea(xStart, yStart, xEnd, yEnd int) error {
	if err := d.sendCommand([]byte{setRAMXAddressStartEndPosition}); err != nil {
		return err
	}

	if err := d.sendData([]byte{byte((xStart >> 3) & 0xFF)}); err != nil {
		return err
	}
	if err := d.sendData([]byte{byte((xEnd >> 3) & 0xFF)}); err != nil {
		return err
	}

	if err := d.sendCommand([]byte{setRAMYAddressStartEndPosition}); err != nil {
		return err
	}
	if err := d.sendData([]byte{byte(yStart & 0xFF)}); err != nil {
		return err
	}
	if err := d.sendData([]byte{byte((yStart >> 8) & 0xFF)}); err != nil {
		return err
	}
	if err := d.sendData([]byte{byte(yEnd & 0xFF)}); err != nil {
		return err
	}
	return d.sendData([]byte{byte((yEnd >> 8) & 0xFF)})
}

func (d *Dev) setLut(update PartialUpdate) error {
	d.update = update
	lut := d.opts.FullUpdate
	if d.update == Partial {
		lut = d.opts.PartialUpdate
	}

	if err := d.sendCommand([]byte{writeLutRegister}); err != nil {
		return err
	}

	for i := range lut {
		if err := d.sendData([]byte{lut[i]}); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dev) sendData(c []byte) error {
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}
	return d.c.Tx(c, nil)
}

func (d *Dev) sendCommand(c []byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	return d.c.Tx(c, nil)
}

var _ display.Drawer = &Dev{}
