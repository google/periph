// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package rainbowhat

import (
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/apa102"
	"periph.io/x/periph/devices/bmxx80"
	"periph.io/x/periph/host/rpi"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c/i2creg"
)

const (
	numPixels  = 7
	bmp280Addr = 0x77
)

// Dev represents a Rainbow HAT  (https://shop.pimoroni.com/products/rainbow-hat-for-android-things)
type Dev struct {
	ledstrip *apa102.Dev
	bmp280   *bmxx80.Dev
	buttonA  gpio.PinIn
	buttonB  gpio.PinIn
	buttonC  gpio.PinIn
	ledR     gpio.PinOut
	ledG     gpio.PinOut
	ledB     gpio.PinOut
	buzzer   gpio.PinOut
	servo    gpio.PinOut
	// TODO: Add support for HT16K33 display
}

// NewRainbowHat returns a rainbowhat driver.
func NewRainbowHat() (*Dev, error) {
	i2cPort, err := i2creg.Open("")
	if err != nil {
		return nil, err
	}

	spiPort, err := spireg.Open("")
	if err != nil {
		return nil, err
	}

	bmp280, err := bmxx80.NewI2C(i2cPort, bmp280Addr, &bmxx80.DefaultOpts)
	if err != nil {
		return nil, err
	}

	opts := apa102.DefaultOpts
	opts.NumPixels = numPixels
	ledstrip, err := apa102.New(spiPort, &opts)
	if err != nil {
		return nil, err
	}

	dev := &Dev{
		ledstrip: ledstrip,
		bmp280:   bmp280,
		buttonA:  rpi.P1_40, // GPIO21
		buttonB:  rpi.P1_38, // GPIO20
		buttonC:  rpi.P1_36, // GPIO16
		ledR:     rpi.P1_31, // GPIO06
		ledG:     rpi.P1_35, // GPIO19
		ledB:     rpi.P1_37, // GPIO26
		buzzer:   rpi.P1_33, // PWM1
		servo:    rpi.P1_32, // PWM0
	}

	if err := dev.buttonA.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		return nil, err
	}

	if err := dev.buttonB.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		return nil, err
	}

	if err := dev.buttonC.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		return nil, err
	}

	return dev, nil
}

// GetLedStrip returns apa102.Dev seven addressable led strip.
func (d *Dev) GetLedStrip() *apa102.Dev {
	return d.ledstrip
}

// GetBmp280 returns bmxx80.Dev handler.
func (d *Dev) GetBmp280() *bmxx80.Dev {
	return d.bmp280
}

// GetButtonA returns gpio.PinIn corresponding to the A capacitive button.
func (d *Dev) GetButtonA() gpio.PinIn {
	return d.buttonA
}

// GetButtonB returns gpio.PinIn corresponding to the B capacitive button.
func (d *Dev) GetButtonB() gpio.PinIn {
	return d.buttonB
}

// GetButtonC returns gpio.PinIn corresponding to the C capacitive button.
func (d *Dev) GetButtonC() gpio.PinIn {
	return d.buttonC
}

// GetLedR returns gpio.PinOut corresponding to the red LED.
func (d *Dev) GetLedR() gpio.PinOut {
	return d.ledR
}

// GetLedG returns gpio.PinOut corresponding to the green LED.
func (d *Dev) GetLedG() gpio.PinOut {
	return d.ledG
}

// GetLedB returns gpio.PinOut corresponding to the blue LED.
func (d *Dev) GetLedB() gpio.PinOut {
	return d.ledB
}

// GetBuzzer returns gpio.PinOut corresponding to the buzzer pin.
func (d *Dev) GetBuzzer() gpio.PinOut {
	return d.buzzer
}

// GetServo returns gpio.PinOut corresponding to the servo pin.
func (d *Dev) GetServo() gpio.PinOut {
	return d.servo
}

// Close halt all devices.
func (d *Dev) Close() error {
	err := d.bmp280.Halt()
	if err != nil {
		return err
	}

	err = d.ledstrip.Halt()
	if err != nil {
		return err
	}

	return nil
}
