// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9685

import (
	"math"
	"time"

	"periph.io/x/periph/conn/physic"

	"periph.io/x/periph/conn/i2c"
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

	var modeRead []byte
	modeRead = make([]byte, 1)
	err := d.dev.Tx([]byte{mode1}, modeRead)

	if err != nil {
		return err
	}

	mode := modeRead[0]
	mode = mode & ^sleep
	if _, err := d.dev.Write([]byte{mode1, mode & 0xFF}); err != nil {
		return err
	}

	time.Sleep(5 * time.Millisecond)

	return d.SetPwmFreq(50 * physic.Hertz)
}

// SetPwmFreq set the pwm frequency
func (d *Dev) SetPwmFreq(freqHz physic.Frequency) error {
	prescaleVal := float32(25 * physic.MegaHertz)
	prescaleVal /= 4096.0 //# 12-bit
	prescaleVal /= float32(freqHz)
	prescaleVal -= 1.0

	prescaleToSend := int(math.Floor(float64(prescaleVal + 0.5)))

	var modeRead []byte
	modeRead = make([]byte, 1)
	err := d.dev.Tx([]byte{mode1}, modeRead)

	if err != nil {
		return err
	}

	oldmode := modeRead[0]
	newmode := (byte)((oldmode & 0x7F) | 0x10)                     // sleep
	if _, err := d.dev.Write([]byte{mode1, newmode}); err != nil { // go to sleep;
		return err
	}
	if _, err := d.dev.Write([]byte{prescale, byte(prescaleToSend)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{mode1, oldmode}); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := d.dev.Write([]byte{mode1, (byte)(oldmode | 0x80)}); err != nil {
		return err
	}
	return nil
}

// SetAllPwm set a pwm value for all outputs
func (d *Dev) SetAllPwm(on uint16, off uint16) error {
	if _, err := d.dev.Write([]byte{allLedOnL, byte(on) & 0xFF}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{allLedOnH, byte(on >> 8)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{allLedOffL, byte(off) & 0xFF}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{allLedOffH, byte(off >> 8)}); err != nil {
		return err
	}
	return nil
}

// SetPwm set a pwm value for given pca9685 channel
func (d *Dev) SetPwm(channel int, on uint16, off uint16) error {
	if _, err := d.dev.Write([]byte{led0OnL + byte(4*channel), byte(on) & 0xFF}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{led0OnH + byte(4*channel), byte(on >> 8)}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{led0OffL + byte(4*channel), byte(off) & 0xFF}); err != nil {
		return err
	}
	if _, err := d.dev.Write([]byte{led0OffH + byte(4*channel), byte(off >> 8)}); err != nil {
		return err
	}
	return nil
}
