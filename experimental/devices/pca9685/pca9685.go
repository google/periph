// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9685

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

// I2CAddr i2c default address.
const I2CAddr uint16 = 0x40

// PCA9685 registers.
const (
	mode1    byte = 0x00
	mode2    byte = 0x01
	prescale byte = 0xFE
	// Each channel has two 12-bit registers (on & off time).
	led0OnL   byte = 0x06 // Start address for setting channel 0.
	allLedOnL byte = 0xFA // Start address for setting all channels.
)

// Mode register 1, mode1.
const (
	restart byte = 0x80
	ai      byte = 0x20 // Auto-increment register after each read and write.
	sleep   byte = 0x10
	allCall byte = 0x01
)

// Mode register 2, mode2.
const (
	invrt  byte = 0x10
	outDrv byte = 0x04
)

// Dev is a handler to pca9685 controller.
type Dev struct {
	dev  *i2c.Dev
	freq physic.Frequency
}

// NewI2C returns a Dev object that communicates over I2C.
//
// To use on the default address, pca9685.I2CAddr must be passed as argument.
func NewI2C(bus i2c.Bus, address uint16) (*Dev, error) {
	dev := &Dev{
		dev: &i2c.Dev{Bus: bus, Addr: address},
	}
	err := dev.init()
	if err != nil {
		return nil, err
	}

	return dev, nil
}

func (d *Dev) init() error {
	if err := d.SetAllPwm(0, 0); err != nil {
		return err
	}

	if _, err := d.dev.Write([]byte{mode2, outDrv}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{mode1, allCall}); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	modeRead := [1]byte{}
	if err := d.dev.Tx([]byte{mode1}, modeRead[:]); err != nil {
		return err
	}

	mode := (modeRead[0] & ^sleep) | ai
	if _, err := d.dev.Write([]byte{mode1, mode}); err != nil {
		return err
	}

	time.Sleep(5 * time.Millisecond)

	return d.SetPwmFreq(50 * physic.Hertz)
}

// SetPwmFreq set the PWM frequency.
func (d *Dev) SetPwmFreq(freqHz physic.Frequency) error {
	if d.freq == freqHz {
		// Don't need to write frequency if it's not changed.
		// Note: this is required to avoid setting it each time
		// when PWM value is changed via gpio.PinOut.PWM() API
		return nil
	}

	p := (25*physic.MegaHertz/4096 + freqHz/2) / freqHz

	modeRead := [1]byte{}
	if err := d.dev.Tx([]byte{mode1}, modeRead[:]); err != nil {
		return err
	}

	oldmode := modeRead[0]
	if _, err := d.dev.Write([]byte{mode1, (oldmode & ^restart) | sleep}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{prescale, byte(p)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{mode1, oldmode}); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	_, err := d.dev.Write([]byte{mode1, oldmode | restart})
	d.freq = freqHz
	return err
}

// setPWM writes a PWM value in a specific register.
func (d *Dev) setPWM(register uint8, on, off gpio.Duty) error {
	// Chained writes are possible due to auto-increment.
	_, err := d.dev.Write([]byte{
		register,
		byte(on),
		byte(on >> 8),
		byte(off),
		byte(off >> 8),
	})
	return err
}

// SetAllPwm set a PWM value for all outputs.
func (d *Dev) SetAllPwm(on, off gpio.Duty) error {
	return d.setPWM(allLedOnL, on, off)
}

// SetPwm set a PWM value for a given PCA9685 channel.
func (d *Dev) SetPwm(channel int, on, off gpio.Duty) error {
	err := verifyChannel(channel)
	if err != nil {
		return err
	}
	return d.setPWM(led0OnL+byte(4*channel), on, off)
}

// SetFullOff sets PWM duty to 0%.
//
// This function uses the dedicated bit to reduce bus traffic.
func (d *Dev) SetFullOff(channel int) error {
	err := verifyChannel(channel)
	if err != nil {
		return err
	}
	_, err = d.dev.Write([]byte{
		led0OnL + byte(4*channel) + 3, // LEDX_OFF_H
		0x10,                          // bit 4 is full-off
	})
	return err
}

// SetFullOn sets PWM duty to 100%.
//
// This function uses the dedicated FULL_ON bit.
func (d *Dev) SetFullOn(channel int) error {
	err := verifyChannel(channel)
	if err != nil {
		return err
	}
	_, err = d.dev.Write([]byte{
		led0OnL + byte(4*channel) + 1, // LEDX_ON_H
		0x10,                          // bit 4 is full-on
		0,
		0, // LEDX_OFF_H is cleared because full-off has a priority over full-on
	})
	return err
}

func verifyChannel(channel int) error {
	if channel < 0 || channel > 15 {
		return fmt.Errorf("PCA9685: invalid channel: %d", channel)
	}
	return nil
}
