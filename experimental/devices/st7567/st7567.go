// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package st7567 implements an interface to the single-chip dot matrix LCD
//
// Datasheet
//
// https://www.newhavendisplay.com/appnotes/datasheets/LCDs/ST7567.pdf

package st7567

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

const (
	// Width the max pixel width
	Width = 128
	// Height the max pixel height
	Height = 64

	pageSize = 128
	// displayOff (0xae): Display OFF (sleep mode)
	displayOff = 0xae
	// displayOn (0xaf): Display ON in normal mode
	displayOn = 0xaf
	// setStartLine (0x40-7f): Set display start line
	setStartLine = 0x40
	// setPageStart (0xb0-b7): Set page start address
	setPageStart = 0xb0
	// setColl (0x00-0x0f): Set lower column address
	setColl = 0x00
	// setCollH (0x10-0x1f): Set higher column address
	setCollH = 0x10
	// displayRAM (0xa4): Resume to RAM content display
	displayRAM = 0xa4
	// displayEntire (0xa5): Entire display ON
	displayEntire = 0xa5
	// enterRMWMode (0xe0): Enter the Read Modify Write mode
	enterRMWMode = 0xe0
	// exitRMWMode (0xee): Leave the Read Modify Write mode
	exitRMWMode = 0xee
	// powerControl (0x2c): Control built-in power circuit
	powerControl = 0x2f
	// setContrast (0x81): Set contrast control
	setContrast = 0x81
)

// Dev is a handle to a ST7567.
type Dev struct {
	c conn.Conn

	//dc low when sending a command, high when sending data.
	dc gpio.PinOut
	//rst reset pin, active low.
	rst gpio.PinOut
	//cs chip select pin
	cs gpio.PinIn

	//pixels the array containing the pixel map
	pixels [1024]byte
}

// Bias selects the LCD bias ratio of the voltage required for driving the LCD
type Bias byte

const (
	// Bias17 (0xa3): Select BIAS setting 1/7
	Bias17 Bias = 0xa3
	// Bias19 (0xa2): Select BIAS setting 1/9
	Bias19 Bias = 0xa2
)

func (b *Bias) Set(s string) error {
	switch s {
	case "17":
		*b = Bias17
	case "19":
		*b = Bias19
	default:
		return fmt.Errorf("unknown Bias %q: expected either 17 or 19", s)
	}
	return nil
}

func (b *Bias) String() string {
	switch *b {
	case Bias17:
		return "Bias 1/7"
	case Bias19:
		return "Bias 1/9"
	default:
		return "Unknown"
	}
}

// SegmentDirection is the direction of the segments
type SegmentDirection byte

const (
	// SegmentDirNormal (0xa0): Column address 0 is mapped to SEG0
	SegmentDirNormal SegmentDirection = 0xa0
	// SegmentDirReverse (0xa1): Column address 128 is mapped to SEG0
	SegmentDirReverse SegmentDirection = 0xa1
)

func (sd *SegmentDirection) Set(s string) error {
	switch s {
	case "normal":
		*sd = SegmentDirNormal
	case "reverse":
		*sd = SegmentDirReverse
	default:
		return fmt.Errorf("unknown SegmentDirection %q: expected either 'normal' or 'reverse'", s)
	}
	return nil
}

func (sd *SegmentDirection) String() string {
	switch *sd {
	case SegmentDirNormal:
		return "Normal segment direction"
	case SegmentDirReverse:
		return "Reverse segment direction"
	default:
		return "Unknown"
	}
}

// CommonDirection controls the common output status which changes the vertical display direction.
type CommonDirection byte

const (
	// CommonDirNormal (0xc0): Column address 0 is mapped to SEG0
	CommonDirNormal CommonDirection = 0xc0
	// CommonDirReverse (0xc8): Column address 128 is mapped to SEG0
	CommonDirReverse CommonDirection = 0xc8
)

func (cd *CommonDirection) Set(s string) error {
	switch s {
	case "normal":
		*cd = CommonDirNormal
	case "reverse":
		*cd = CommonDirReverse
	default:
		return fmt.Errorf("unknown CommonDirection %q: expected either 'normal' or 'reverse'", s)
	}
	return nil
}

func (cd *CommonDirection) String() string {
	switch *cd {
	case CommonDirNormal:
		return "Normal common direction"
	case CommonDirReverse:
		return "Reverse common direction"
	default:
		return "Unknown"
	}
}

// Display contains if the display is in normal or inverse mode (black will be white and vice versa)
type Display byte

const (
	// DisplayNormal (0xa6): Normal display
	DisplayNormal Display = 0xa6
	// DisplayInverse (0xa7): Inverse display
	DisplayInverse Display = 0xa7
)

func (d *Display) Set(s string) error {
	switch s {
	case "normal":
		*d = DisplayNormal
	case "inverse":
		*d = DisplayInverse
	default:
		return fmt.Errorf("unknown Display %q: expected either 'normal' or 'inverse'", s)
	}
	return nil
}

func (d *Display) String() string {
	switch *d {
	case DisplayNormal:
		return "Normal display"
	case DisplayInverse:
		return "Inverse display"
	default:
		return "Unknown"
	}
}

// RegulationResistor is the single regulation resistor value
type RegulationResistor byte

const (
	// RegResistorRR0 (0x21): Regulation Resistor ratio
	RegResistorRR0 RegulationResistor = 0x21
	// RegResistorRR1 (0x22): Regulation Resistor ratio
	RegResistorRR1 RegulationResistor = 0x22
	// RegResistorRR2 (0x24): Regulation Resistor ratio
	RegResistorRR2 RegulationResistor = 0x24
)

func (rr *RegulationResistor) Set(s string) error {
	switch s {
	case "RR0":
		*rr = RegResistorRR0
	case "RR1":
		*rr = RegResistorRR1
	case "RR2":
		*rr = RegResistorRR2
	default:
		return fmt.Errorf("unknown RegulataionRatio %q: expected either 'RR0' or 'RR1' or 'RR2'", s)
	}
	return nil
}

func (rr *RegulationResistor) String() string {
	switch *rr {
	case RegResistorRR0:
		return "Regulation resistor RR0"
	case RegResistorRR1:
		return "Regulation resistor RR1"
	case RegResistorRR2:
		return "Regulation resistor RR2"
	default:
		return "Unknown"
	}
}

//RegulationRatio selects the regulation resistor ratio
type RegulationRatio []RegulationResistor

func (rrs *RegulationRatio) String() string {
	return "Regulation resistor ratio"
}

func (rrs *RegulationRatio) Set(value string) error {
	values := strings.Split(value, ",")

	for _, v := range values {
		rr := new(RegulationResistor)

		if err := rr.Set(v); err != nil {
			return err
		}

		*rrs = append(*rrs, *rr)
	}

	return nil
}

func (rrs RegulationRatio) getValue() RegulationResistor {
	var out RegulationResistor
	for _, v := range rrs {
		out |= v
	}
	return out
}

// Opts contains the configuration for the S7567 device.
type Opts struct {
	// Bias selects the LCD bias ratio of the voltage required for driving the LCD.
	Bias Bias
	// SegmentDirection is the direction of the segments.
	SegmentDirection SegmentDirection
	// CommonDirection controls the common output status which changes the vertical display direction.
	CommonDirection CommonDirection
	// Display changes the selected and non-selected voltage of SEG.
	Display Display
	// RegulationRatio controls the regulation ratio of the built-in regulator.
	RegulationRatio RegulationRatio
	// StartLine sets the line address of the Display Data RAM to determine the initial display line.
	StartLine byte
	// Contrast the value to adjust the display contrast.
	Contrast byte
}

// New opens a handle to a ST7567 LCD.
func New(p spi.Port, dc gpio.PinOut, rst gpio.PinOut, cs gpio.PinIn, o *Opts) (*Dev, error) {
	c, err := p.Connect(1000*physic.KiloHertz, spi.Mode0, 8)

	if err != nil {
		return nil, errors.New("could not connect to device")
	}

	d := &Dev{
		c:   c,
		dc:  dc,
		rst: rst,
		cs:  cs,
	}

	var cmd []byte
	cmd = append(cmd, byte(o.Bias))
	cmd = append(cmd, byte(o.SegmentDirection))
	cmd = append(cmd, byte(o.CommonDirection))
	cmd = append(cmd, byte(o.Display))
	cmd = append(cmd, setStartLine|o.StartLine)
	cmd = append(cmd, powerControl)
	cmd = append(cmd, byte(o.RegulationRatio.getValue()))
	cmd = append(cmd, displayOn)
	cmd = append(cmd, setContrast)
	cmd = append(cmd, o.Contrast)

	if err := d.sendCommand(cmd); err != nil {
		return nil, err
	}

	return d, nil
}

// Halt resets the registers and switches the driver off.
func (d *Dev) Halt() error {
	return d.reset()
}

// SetContrast sets the contrast
func (d *Dev) SetContrast(value byte) error {
	return d.sendCommand([]byte{setContrast, value})
}

// SetPixel sets a pixel in the pixels array
func (d *Dev) SetPixel(x, y int, active bool) {
	offset := (y / 8 * Width) + x
	pageAddress := y % 8
	d.pixels[offset] &= ^(1 << byte(pageAddress))
	d.pixels[offset] |= bTob(active) & 1 << byte(pageAddress)
}

// Update updates the display
func (d *Dev) Update() error {
	if err := d.sendCommand([]byte{enterRMWMode}); err != nil {
		return err
	}
	for i := 0; i < 8; i++ {
		offset := i * pageSize
		if err := d.sendCommand([]byte{setPageStart | byte(i), setColl, setCollH}); err != nil {
			return err
		}
		if err := d.sendData(d.pixels[offset : offset+pageSize]); err != nil {
			return err
		}
	}
	if err := d.sendCommand([]byte{exitRMWMode}); err != nil {
		return err
	}
	return nil
}

// PowerSave turning the display into sleep
func (d *Dev) PowerSave() error {
	return d.sendCommand([]byte{displayOff, displayEntire})
}

// WakeUp wakes the display up from power saving mode
func (d *Dev) WakeUp() error {
	return d.sendCommand([]byte{displayRAM, displayOn})
}

func (d *Dev) sendCommand(c []byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	return d.c.Tx(c, nil)
}

func (d *Dev) sendData(c []byte) error {
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}
	return d.c.Tx(c, nil)
}

func (d *Dev) reset() error {
	if err := d.rst.Out(gpio.Low); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := d.rst.Out(gpio.High); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}

func bTob(b bool) byte {
	if b {
		return 1
	}
	return 0
}
