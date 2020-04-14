// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9685

import (
	"errors"
	"fmt"
	"math"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	gpiopin "periph.io/x/periph/conn/pin"
)

const (
	dutyMax gpio.Duty = math.MaxUint16
)

type pin struct {
	dev     *Dev
	channel int
}

// CreatePin creates a gpio handle for the given channel.
func (d *Dev) CreatePin(channel int) (gpio.PinIO, error) {
	if channel < 0 || channel >= 16 {
		return nil, errors.New("PCA9685: Valid channel range is 0..15")
	}
	return &pin{
		dev:     d,
		channel: channel,
	}, nil
}

// RegisterPins makes PWM channels available as PWM pins in the pin registry
//
// Pin names have the following format: PCA9685_<HexAddress>_<channel> (e.g. PCA9685_40_11)
func (d *Dev) RegisterPins() error {
	for i := 0; i < 16; i++ {
		pin, err := d.CreatePin(i)
		if err != nil {
			return err
		}
		if err = gpioreg.Register(pin); err != nil {
			return err
		}
	}
	return nil
}

func (p *pin) String() string {
	return p.Name()
}

func (p *pin) Halt() error {
	return p.Out(gpio.Low)
}

func (p *pin) Name() string {
	return fmt.Sprintf("PCA9685_%x_%d", p.dev.dev.Addr, p.channel)
}

func (p *pin) Number() int {
	return p.channel
}

func (p *pin) Function() string {
	return string(p.Func())
}

func (p *pin) In(pull gpio.Pull, edge gpio.Edge) error {
	return errors.New("PCA9685: Pin cannot be configured as input")
}

func (p *pin) Read() gpio.Level {
	return gpio.INVALID.Read()
}

func (p *pin) WaitForEdge(timeout time.Duration) bool {
	return false
}

func (p *pin) Pull() gpio.Pull {
	return gpio.Float
}

func (p *pin) DefaultPull() gpio.Pull {
	return gpio.Float
}

func (p *pin) Out(l gpio.Level) error {
	return p.PWM(gpio.DutyMax, 0)
}

func (p *pin) PWM(duty gpio.Duty, freq physic.Frequency) error {
	if err := p.dev.SetPwmFreq(freq); err != nil {
		return err
	}
	// PWM duty scaled down from 24 to 16 bits
	scaled := duty >> 8
	if scaled > dutyMax {
		scaled = dutyMax
	}
	return p.dev.SetPwm(p.channel, 0, scaled)
}

func (p *pin) Func() gpiopin.Func {
	return gpio.PWM
}

func (p *pin) SupportedFuncs() []gpiopin.Func {
	return []gpiopin.Func{gpio.PWM}
}

func (p *pin) SetFunc(f gpiopin.Func) error {
	if f != gpio.PWM {
		return fmt.Errorf("PCA9685: Function not supported: %s", f)
	}
	return nil
}
