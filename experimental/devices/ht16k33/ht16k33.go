// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ht16k33

import (
	"errors"

	"periph.io/x/periph/conn/i2c"
)

// I2CAddr i2c default address.
const I2CAddr uint16 = 0x70

const (
	cmdRAM        = 0x00
	cmdKeys       = 0x40
	displaySetup  = 0x80
	displayOff    = 0x00
	displayOn     = 0x01
	systemSetup   = 0x20
	oscillatorOff = 0x00
	oscillatorOn  = 0x01
	cmdBrightness = 0xE0
)

// BlinkFrequency display frequency must be a value allowed by the HT16K33.
type BlinkFrequency byte

// Blinking frequencies.
const (
	BlinkOff    = 0x00
	Blink2Hz    = 0x02
	Blink1Hz    = 0x04
	BlinkHalfHz = 0x06
)

// Dev is a handler to ht16k33 controller
type Dev struct {
	dev i2c.Dev
}

// NewI2C returns a Dev object that communicates over I2C.
//
// To use on the default address, ht16k33.I2CAddr must be passed as argument.
func NewI2C(bus i2c.Bus, address uint16) (*Dev, error) {
	dev := &Dev{dev: i2c.Dev{Bus: bus, Addr: address}}

	if err := dev.init(); err != nil {
		return nil, err
	}

	return dev, nil
}

func (d *Dev) init() error {
	// Turn on the oscillator.
	if _, err := d.dev.Write([]byte{systemSetup | oscillatorOn}); err != nil {
		return err
	}

	// Turn on display
	if _, err := d.dev.Write([]byte{displaySetup | displayOn}); err != nil {
		return err
	}

	// Set no blinking.
	if err := d.SetBlink(BlinkOff); err != nil {
		return err
	}

	// Set display to full brightness.
	return d.SetBrightness(15)
}

// SetBlink Blink display at specified frequency.
func (d *Dev) SetBlink(freq BlinkFrequency) error {
	_, err := d.dev.Write([]byte{displaySetup | displayOn | byte(freq)})
	return err
}

// SetBrightness of entire display to specified value.
//
// Supports 16 levels, from 0 to 15.
func (d *Dev) SetBrightness(brightness int) error {
	if brightness < 0 || brightness > 15 {
		return errors.New("ht16k33: brightness must be between 0 and 15")
	}
	_, err := d.dev.Write([]byte{cmdBrightness | byte(brightness)})
	return err
}

// WriteColumn set data in a given column.
func (d *Dev) WriteColumn(column int, data uint16) error {
	_, err := d.dev.Write([]byte{byte(column * 2), byte(data & 0xFF), byte(data >> 8)})
	return err
}

// Halt clear the contents of display buffer.
func (d *Dev) Halt() error {
	for i := 0; i < 4; i++ {
		if err := d.WriteColumn(i, 0); err != nil {
			return err
		}
	}
	return nil
}
