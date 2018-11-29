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

// Dev is a handler to sn3218 controller.
type Dev struct {
	i2c        i2c.Dev
	on         [18]bool
	brightness [18]byte
}

// New returns a handle to a SN3218 LED driver.
func New(bus i2c.Bus) (*Dev, error) {
	d := &Dev{
		i2c: i2c.Dev{Bus: bus, Addr: i2cAddress},
	}
	if err := d.reset(); err != nil {
		return nil, err
	}
	return d, nil
}

// Halt resets the registers and switches the driver off.
func (d *Dev) Halt() error {
	return d.reset()
}

// WakeUp returns from sleep mode and switches the channels according to the states in the register of SN3218.
func (d *Dev) WakeUp() error {
	_, err := d.i2c.Write([]byte{cmdEnableOutput, 0x01})
	return err
}

// Sleep sends SN3218 to sleep mode while keeping the states in the registers.
func (d *Dev) Sleep() error {
	_, err := d.i2c.Write([]byte{cmdEnableOutput, 0x00})
	return err
}

// GetState returns the state (on/off) and the brightness (0..255) of the
// Channel 0..17.
func (d *Dev) GetState(channel int) (bool, byte, error) {
	if channel < 0 || channel >= 18 {
		return false, 0, errors.New("channel number out of range 0..17")
	}
	return d.on[channel], d.brightness[channel], nil
}

// Switch switched the channel (0..18) to state (on/off).
func (d *Dev) Switch(channel int, state bool) error {
	if channel < 0 || channel >= 18 {
		return errors.New("channel number out of range 0..17")
	}
	d.on[channel] = state
	return d.updateStates()
}

// SwitchAll switches all channels accoring to the state (on/off).
func (d *Dev) SwitchAll(state bool) error {
	for i := 0; i < 18; i++ {
		d.on[i] = state
	}
	return d.updateStates()
}

// Brightness sets the brightness of led (0..17) to value (0..255).
func (d *Dev) Brightness(channel int, value byte) error {
	if channel < 0 || channel >= 18 {
		return errors.New("channel number out of range 0..17")
	}
	d.brightness[channel] = value
	return d.updateBrightness()
}

// BrightnessAll sets the brightness of all channels to the value (0..255).
func (d *Dev) BrightnessAll(value byte) error {
	for i := 0; i < 18; i++ {
		d.brightness[i] = value
	}
	return d.updateBrightness()
}

// Reset resets the registers to the default values.
func (d *Dev) reset() error {
	_, err := d.i2c.Write([]byte{cmdReset, 0xFF})
	d.on = [18]bool{}
	d.brightness = [18]byte{}
	return err
}

func (d *Dev) stateArrayToInt() uint {
	var result uint = 0
	for i := uint(0); i < uint(18); i++ {
		state := uint(1)
		if !d.on[i] {
			state = uint(0)
		}
		result |= (state << i)
	}
	return result
}

func (d *Dev) update() error {
	_, err := d.i2c.Write([]byte{cmdUpdate, 0xFF})
	return err
}

func (d *Dev) updateStates() error {
	mask := d.stateArrayToInt()
	cmd := [...]byte{cmdEnableLeds, byte(mask & 0x3F), byte((mask >> 6) & 0x3F), byte((mask >> 12) & 0X3F)}
	if _, err := d.i2c.Write(cmd[:]); err != nil {
		return err
	}
	return d.update()
}

func (d *Dev) updateBrightness() error {
	_, err := d.i2c.Write(append([]byte{cmdSetBrightnessValues}, d.brightness[0:len(d.brightness)]...))
	if err != nil {
		return err
	}
	return d.update()
}
