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

// PCA9685Address i2c default address
const PCA9685Address uint16 = 0x40

// PCA9685 Commands
const (
	Mode1      byte = 0x00
	Mode2      byte = 0x01
	SubAdr1    byte = 0x02
	SubAdr2    byte = 0x03
	SubAdr3    byte = 0x04
	Prescale   byte = 0xFE
	Led0OnL    byte = 0x06
	Led0OnH    byte = 0x07
	Led0OffL   byte = 0x08
	Led0OffH   byte = 0x09
	AllLedOnL  byte = 0xFA
	AllLedOnH  byte = 0xFB
	AllLedOffL byte = 0xFC
	AllLedOffH byte = 0xFD

	// Bits
	Restart byte = 0x80
	Sleep   byte = 0x10
	AllCall byte = 0x01
	Invrt   byte = 0x10
	OutDrv  byte = 0x04
)

// Dev is a handler to pca9685 controller
type Dev struct {
	dev *i2c.Dev
}

// NewI2CAddress returns a Dev object that communicates over I2C on a alternate address
func NewI2CAddress(bus i2c.Bus, address uint16) (*Dev, error) {
	dev := &Dev{
		dev: &i2c.Dev{Bus: bus, Addr: address},
	}
	err := dev.init()
	return dev, err
}

// NewI2C returns a Dev object that communicates over I2C on the default address
func NewI2C(bus i2c.Bus) (*Dev, error) {
	return NewI2CAddress(bus, PCA9685Address)
}

func (d *Dev) init() error {
	d.SetAllPwm(0, 0)
	d.dev.Write([]byte{Mode2, OutDrv})
	d.dev.Write([]byte{Mode1, AllCall})

	time.Sleep(100 * time.Millisecond)

	var mode1 byte
	err := d.dev.Tx([]byte{Mode1}, []byte{mode1})

	if err != nil {
		return err
	}

	mode1 = mode1 & ^Sleep
	d.dev.Write([]byte{Mode1, mode1 & 0xFF})

	time.Sleep(5 * time.Millisecond)

	err = d.SetPwmFreq(50)
	return err
}

// SetPwmFreq set the pwm frequency
func (d *Dev) SetPwmFreq(freqHz float32) error {
	prescaleval := float32(25 * physic.MegaHertz)
	prescaleval /= 4096.0 //# 12-bit
	prescaleval /= (freqHz * float32(physic.Hertz))
	prescaleval -= 1.0

	prescale := int(math.Floor(float64(prescaleval + 0.5)))

	var oldmode byte
	err := d.dev.Tx([]byte{Mode1}, []byte{oldmode})

	if err != nil {
		return err
	}

	newmode := (byte)((oldmode & 0x7F) | 0x10) // sleep
	d.dev.Write([]byte{Mode1, newmode})        // go to sleep
	d.dev.Write([]byte{Prescale, byte(prescale)})
	d.dev.Write([]byte{Mode1, oldmode})

	time.Sleep(100 * time.Millisecond)

	d.dev.Write([]byte{Mode1, (byte)(oldmode | 0x80)})
	return nil
}

// SetAllPwm set a pwm value for all outputs
func (d *Dev) SetAllPwm(on uint16, off uint16) {
	d.dev.Write([]byte{AllLedOnL, byte(on) & 0xFF})
	d.dev.Write([]byte{AllLedOnH, byte(on >> 8)})
	d.dev.Write([]byte{AllLedOffL, byte(off) & 0xFF})
	d.dev.Write([]byte{AllLedOffH, byte(off >> 8)})
}

// SetPwm set a pwm value for given pca9685 channel
func (d *Dev) SetPwm(channel int, on uint16, off uint16) {
	d.dev.Write([]byte{Led0OnL + byte(4*channel), byte(on) & 0xFF})
	d.dev.Write([]byte{Led0OnH + byte(4*channel), byte(on >> 8)})
	d.dev.Write([]byte{Led0OffL + byte(4*channel), byte(off) & 0xFF})
	d.dev.Write([]byte{Led0OffH + byte(4*channel), byte(off >> 8)})
}
