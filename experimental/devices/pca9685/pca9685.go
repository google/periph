// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9685

import (
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

// I2CAddr i2c default address.
const I2CAddr uint16 = 0x40

// PCA9685 Commands
const (
	mode1      byte = 0x00
	mode2      byte = 0x01
	subAdr1    byte = 0x02
	subAdr2    byte = 0x03
	subAdr3    byte = 0x04
	prescale   byte = 0xFE
	led0OnL    byte = 0x06
	led0OnH    byte = 0x07
	led0OffL   byte = 0x08
	led0OffH   byte = 0x09
	allLedOnL  byte = 0xFA
	allLedOnH  byte = 0xFB
	allLedOffL byte = 0xFC
	allLedOffH byte = 0xFD

	// Bits
	restart byte = 0x80
	sleep   byte = 0x10
	allCall byte = 0x01
	invrt   byte = 0x10
	outDrv  byte = 0x04
)

// Dev is a handler to pca9685 controller
type Dev struct {
	dev *i2c.Dev
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

	mode := modeRead[0] & ^sleep
	if _, err := d.dev.Write([]byte{mode1, mode}); err != nil {
		return err
	}

	time.Sleep(5 * time.Millisecond)

	return d.SetPwmFreq(50 * physic.Hertz)
}

// SetPwmFreq set the pwm frequency
func (d *Dev) SetPwmFreq(freqHz physic.Frequency) error {
	p := (25*physic.MegaHertz/4096 + freqHz/2) / freqHz

	modeRead := [1]byte{}
	if err := d.dev.Tx([]byte{mode1}, modeRead[:]); err != nil {
		return err
	}

	oldmode := modeRead[0]
	if _, err := d.dev.Write([]byte{mode1, byte((oldmode & 0x7F) | 0x10)}); err != nil { // go to sleep;
		return err
	}
	if _, err := d.dev.Write([]byte{prescale, byte(p)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{mode1, oldmode}); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	_, err := d.dev.Write([]byte{mode1, (byte)(oldmode | 0x80)})
	return err
}

// SetAllPwm set a pwm value for all outputs
func (d *Dev) SetAllPwm(on, off gpio.Duty) error {
	if _, err := d.dev.Write([]byte{allLedOnL, byte(on)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{allLedOnH, byte(on >> 8)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{allLedOffL, byte(off)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{allLedOffH, byte(off >> 8)}); err != nil {
		return err
	}
	return nil
}

// SetPwm set a pwm value for given pca9685 channel
func (d *Dev) SetPwm(channel int, on, off gpio.Duty) error {
	if _, err := d.dev.Write([]byte{led0OnL + byte(4*channel), byte(on)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{led0OnH + byte(4*channel), byte(on >> 8)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{led0OffL + byte(4*channel), byte(off)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{led0OffH + byte(4*channel), byte(off >> 8)}); err != nil {
		return err
	}
	return nil
}
