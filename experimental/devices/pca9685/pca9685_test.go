// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9685

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
)

func TestPCA9685_pin(t *testing.T) {
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// All leds cleared by init
			{Addr: I2CAddr, W: []byte{allLedOnL, 0, 0, 0, 0}, R: nil},
			// mode2 is set
			{Addr: I2CAddr, W: []byte{mode2, outDrv}, R: nil},
			// mode1 is set
			{Addr: I2CAddr, W: []byte{mode1, allCall}, R: nil},
			// mode1 is read and sleep bit is cleared
			{Addr: I2CAddr, W: []byte{mode1}, R: []byte{allCall | sleep}},
			{Addr: I2CAddr, W: []byte{mode1, allCall | ai}, R: nil},

			// SetPwmFreq 50 Hz
			// Read mode
			{Addr: I2CAddr, W: []byte{0x00}, R: []byte{allCall | ai}},
			// Set sleep
			{Addr: I2CAddr, W: []byte{0x00, allCall | ai | sleep}, R: nil},
			// Set prescale
			{Addr: I2CAddr, W: []byte{prescale, 122}, R: nil},
			// Clear sleep
			{Addr: I2CAddr, W: []byte{0x00, allCall | ai}, R: nil},
			// Set Restart
			{Addr: I2CAddr, W: []byte{0x00, allCall | ai | restart}, R: nil},

			// Set PWM value of pin 0 to 50%
			{Addr: I2CAddr, W: []byte{led0OnL, 0, 0, 0, 128}, R: nil},
		},
	}

	dev, err := NewI2C(scenario, I2CAddr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dev.RegisterPins(); err != nil {
		t.Fatal(err)
	}

	pin := gpioreg.ByName("PCA9685_40_0")
	pin.PWM(gpio.DutyHalf, 50*physic.Hertz)
}

func TestPCA9685(t *testing.T) {
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// All leds cleared by init
			{Addr: I2CAddr, W: []byte{allLedOnL, 0, 0, 0, 0}, R: nil},
			// mode2 is set
			{Addr: I2CAddr, W: []byte{mode2, outDrv}, R: nil},
			// mode1 is set
			{Addr: I2CAddr, W: []byte{mode1, allCall}, R: nil},
			// mode1 is read and sleep bit is cleared
			{Addr: I2CAddr, W: []byte{mode1}, R: []byte{allCall | sleep}},
			{Addr: I2CAddr, W: []byte{mode1, allCall | ai}, R: nil},

			// SetPwmFreq 50 Hz
			// Read mode
			{Addr: I2CAddr, W: []byte{0x00}, R: []byte{allCall | ai}},
			// Set sleep
			{Addr: I2CAddr, W: []byte{0x00, allCall | ai | sleep}, R: nil},
			// Set prescale
			{Addr: I2CAddr, W: []byte{prescale, 122}, R: nil},
			// Clear sleep
			{Addr: I2CAddr, W: []byte{0x00, allCall | ai}, R: nil},
			// Set Restart
			{Addr: I2CAddr, W: []byte{0x00, allCall | ai | restart}, R: nil},

			// Set PWM value of pin 0 to 50%
			{Addr: I2CAddr, W: []byte{led0OnL, 0, 0, 0, 128}, R: nil},
		},
	}

	dev, err := NewI2C(scenario, I2CAddr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dev.SetPwm(0, 0, 0x8000); err != nil {
		t.Fatal(err)
	}
}
