// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bh1750

import (
	"encoding/binary"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

// I2CAddr i2c default address.
const I2CAddr uint16 = 0x23

// AlternativeI2CAddr i2c alternative address.
const AlternativeI2CAddr uint16 = 0x5c

// Resolution represents the measurements modes.
type Resolution uint8

const (
	// ContinuousHighResMode start measurement at 1lx resolution. Measurement time is approx 120ms.
	ContinuousHighResMode Resolution = 0x10

	// ContinuousHighResMode2 start measurement at 0.5lx resolution. Measurement time is approx 120ms.
	ContinuousHighResMode2 Resolution = 0x11

	// ContinuousLowResMode start measurement at 4lx resolution. Measurement time is approx 16ms.
	ContinuousLowResMode Resolution = 0x13

	// OneTimeHighResMode measurement at 1lx resolution and the device is automatically set to Power Down after measurement.
	OneTimeHighResMode Resolution = 0x20

	// OneTimeHighResMode2 measurement at 0.5lx resolution and the device is automatically set to Power Down after measurement.
	OneTimeHighResMode2 Resolution = 0x21

	// OneTimeLowResMode measurement at 4lx resolution and the device is automatically set to Power Down after measurement.
	OneTimeLowResMode Resolution = 0x23

	safetyTimeout = 20 * time.Millisecond
)

// Map of timeouts for measurement type.
var timeout = map[Resolution]time.Duration{
	ContinuousHighResMode:  120*time.Millisecond + safetyTimeout,
	ContinuousHighResMode2: 120*time.Millisecond + safetyTimeout,
	ContinuousLowResMode:   16*time.Millisecond + safetyTimeout,
	OneTimeHighResMode:     120*time.Millisecond + safetyTimeout,
	OneTimeHighResMode2:    120*time.Millisecond + safetyTimeout,
	OneTimeLowResMode:      16*time.Millisecond + safetyTimeout,
}

// Mode - device mode.
type Mode uint8

const (
	// PowerDown - no active state.
	PowerDown Mode = 0x00

	// PowerOn - waiting for measurement command.
	PowerOn Mode = 0x01
)

const (
	dataRegister            uint8 = 0x00
	measurementTimeRegister uint8 = 0x45
	reset                   uint8 = 0x07
)

// Dev is a handler to bh1750 controller
type Dev struct {
	dev  i2c.Dev
	res  Resolution
	mode Mode
}

// NewI2C opens a handle to an bh1750 sensor.
//
// To use on the default address, bh1750.I2CAddr must be passed as argument.
func NewI2C(bus i2c.Bus, address uint16) (*Dev, error) {
	dev := &Dev{dev: i2c.Dev{Bus: bus, Addr: address}}

	if err := dev.init(); err != nil {
		return nil, err
	}
	return dev, nil
}

func (d *Dev) init() error {
	if err := d.SetMode(PowerOn); err != nil {
		return err
	}

	if err := d.SetResolution(ContinuousHighResMode); err != nil {
		return err
	}

	_, err := d.dev.Write([]byte{reset})
	return err
}

// SetMode set the sleep mode of the sensor.
func (d *Dev) SetMode(m Mode) error {
	d.mode = m
	_, err := d.dev.Write([]byte{reset})
	return err
}

// SetResolution set the resolution mode of the sensor.
func (d *Dev) SetResolution(r Resolution) error {
	d.res = r
	_, err := d.dev.Write([]byte{byte(r)})
	return err
}

// Sense reads the light value from the bh1750 sensor.
func (d *Dev) Sense() (physic.LuminousFlux, error) {
	if err := d.SetResolution(d.res); err != nil {
		return 0, err
	}

	time.Sleep(timeout[d.res])

	buf := [2]byte{}
	if err := d.dev.Tx([]byte{byte(d.res)}, buf[:]); err != nil {
		return 0, err
	}

	rawValue := binary.BigEndian.Uint16(buf[:])
	lux := physic.LuminousFlux(float64(rawValue) / 1.2)
	return lux * physic.Lumen, nil
}

// Halt turn off device.
func (d *Dev) Halt() error {
	return d.SetMode(PowerDown)
}
