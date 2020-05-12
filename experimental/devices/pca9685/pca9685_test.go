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

func initializationSequence() []i2ctest.IO {
	return []i2ctest.IO{
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
	}
}

func TestPCA9685_pin(t *testing.T) {
	scenario := &i2ctest.Playback{
		Ops: append(initializationSequence(),
			// Set PWM value of pin 0 to 50%
			i2ctest.IO{Addr: I2CAddr, W: []byte{led0OnL, 0, 0, 0, 0x08}, R: nil},
		),
	}

	dev, err := NewI2C(scenario, I2CAddr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dev.RegisterPins(); err != nil {
		t.Fatal(err)
	}
	defer dev.UnregisterPins()

	pin := gpioreg.ByName("PCA9685_40_0")
	pin.PWM(gpio.DutyHalf, 50*physic.Hertz)
}

func TestPCA9685_pin_fullOff(t *testing.T) {
	scenario := &i2ctest.Playback{
		Ops: append(initializationSequence(),
			// Set PWM value of pin 0 to 0%
			i2ctest.IO{Addr: I2CAddr, W: []byte{led0OnL + 3, 0x10}, R: nil},
		),
	}

	dev, err := NewI2C(scenario, I2CAddr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dev.RegisterPins(); err != nil {
		t.Fatal(err)
	}
	defer dev.UnregisterPins()

	pin := gpioreg.ByName("PCA9685_40_0")
	pin.PWM(0, 50*physic.Hertz)
}

func TestPCA9685_pin_fullOn(t *testing.T) {
	scenario := &i2ctest.Playback{
		Ops: append(initializationSequence(),
			// Set PWM value of pin 0 to 100%
			i2ctest.IO{Addr: I2CAddr, W: []byte{led0OnL + 1, 0x10, 0, 0}, R: nil},
		),
	}

	dev, err := NewI2C(scenario, I2CAddr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dev.RegisterPins(); err != nil {
		t.Fatal(err)
	}
	defer dev.UnregisterPins()

	pin := gpioreg.ByName("PCA9685_40_0")
	pin.PWM(gpio.DutyMax, 50*physic.Hertz)
}

func TestPCA9685(t *testing.T) {
	scenario := &i2ctest.Playback{
		Ops: append(initializationSequence(),
			// Set PWM value of pin 0 to 50%
			i2ctest.IO{Addr: I2CAddr, W: []byte{led0OnL, 0, 0, 0, 0x08}, R: nil},
		),
	}

	dev, err := NewI2C(scenario, I2CAddr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dev.SetPwm(0, 0, 0x800); err != nil {
		t.Fatal(err)
	}
}

func TestPCA9685_invalidCh(t *testing.T) {
	scenario := &i2ctest.Playback{
		Ops: append(initializationSequence()),
	}

	dev, err := NewI2C(scenario, I2CAddr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dev.SetPwm(16, 0, 0x800); err == nil {
		t.Fatal("Error expected")
	}
}
