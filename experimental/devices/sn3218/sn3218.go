// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sn3218

import (
	"errors"

	"periph.io/x/periph/conn/i2c"
)

const (
	i2cAddress             = 0x54
	cmdEnableOutput        = 0x00
	cmdSetBrightnessValues = 0x01
	cmdEnableLeds          = 0x13
	cmdUpdate              = 0x16
	cmdReset               = 0x17
)

type dev struct {
	i2c        i2c.Dev
	states     [18]bool
	brightness [18]byte
}

// New returns a handle to a SN3218 LED driver.
func New(bus i2c.Bus) (dev, error) {
	d := i2c.Dev{Bus: bus, Addr: i2cAddress}
	dev := dev{}
	dev.i2c = d
	dev.reset()
	return dev, nil
}

// Halt resets the registers and switches the driver off.
func (d *dev) Halt() error {
	if err := d.Disable(); err != nil {
		return err
	}
	return d.reset()
}

// Enable enables the SN3218.
func (d *dev) Enable() error {
	_, err := d.i2c.Write([]byte{cmdEnableOutput, 0x01})
	return err
}

// Disable disables the SN3218.
func (d *dev) Disable() error {
	_, err := d.i2c.Write([]byte{cmdEnableOutput, 0x00})
	return err
}

// GetLedState returns the state (on/off) and the brightness (0..255) of the
// LED 0..17.
func (d *dev) GetLedState(led int) (bool, byte, error) {
	if led < 0 || led >= 18 {
		return false, 0, errors.New("LED number out of range 0..17")
	}
	return d.states[led], d.brightness[led], nil
}

// SwitchLed switched the LED led (0..18) to state (on/off).
func (d *dev) SwitchLed(led int, state bool) error {
	if led < 0 || led >= 18 {
		return errors.New("LED number out of range 0..17")
	}
	d.states[led] = state
	return d.updateLeds()
}

// SetGlobalBrightness sets the brightness of all LEDs to the value (0..255).
func (d *dev) SetGlobalBrightness(value byte) {
	for i := 0; i < 18; i++ {
		d.brightness[i] = value
	}
	d.updateBrightness()
}

// SetBrightness sets the brightness of led (0..17) to value (0..255).
func (d *dev) SetBrightness(led int, value byte) error {
	if led < 0 || led >= 18 {
		return errors.New("LED number out of range 0..17")
	}
	d.brightness[led] = value
	d.updateBrightness()
	return nil
}

// SwitchAllLeds switches all LEDs accoring to the state (on/off).
func (d *dev) SwitchAllLeds(state bool) {
	for i := 0; i < 18; i++ {
		d.states[i] = state
	}
	d.updateLeds()
}

// Reset resets the registers to the default values.
func (d *dev) reset() error {
	_, err := d.i2c.Write([]byte{cmdReset, 0xFF})
	d.states = [18]bool{}
	d.brightness = [18]byte{}
	return err
}

func boolArrayToInt(states [18]bool) uint {
	var result uint = 0
	for i := uint(0); i < uint(18); i++ {
		state := uint(1)
		if !states[i] {
			state = uint(0)
		}
		result |= (state << i)
	}
	return result
}

func (d *dev) update() error {
	_, err := d.i2c.Write([]byte{cmdUpdate, 0xFF})
	return err
}

func (d *dev) updateLeds() error {
	mask := boolArrayToInt(d.states)
	_, err := d.i2c.Write([]byte{cmdEnableLeds, byte(mask & 0x3F), byte((mask >> 6) & 0x3F), byte((mask >> 12) & 0X3F)})
	if err != nil {
		return err
	}
	return d.update()
}

func (d *dev) updateBrightness() error {
	_, err := d.i2c.Write(append([]byte{cmdSetBrightnessValues}, d.brightness[0:len(d.brightness)]...))
	if err != nil {
		return err
	}
	return d.update()
}
