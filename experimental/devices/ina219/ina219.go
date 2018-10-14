// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ina219

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/physic"
)

// Config is a Configuration
type Config struct {
	Address       int
	SenseResistor physic.ElectricResistance
	MaxCurrent    physic.ElectricCurrent
}

const (
	// DefaultSenseResistor is 100 mΩ and can be changed using SenseResistor(value)
	DefaultSenseResistor = 100 * physic.MilliOhm
	// DefaultI2CAddress is 0x40 and can be changed using Address(address)
	DefaultI2CAddress = 0x40
	// DefaultMaxCurrent is 3.2 and can be changed using MaxCurrent(value)
	DefaultMaxCurrent = 3200 * physic.MilliAmpere
)

// New opens a handle to an ina219 sensor
func New(bus i2c.Bus, config Config) (*Dev, error) {

	i2cAddress := DefaultI2CAddress
	if config.Address != 0 {
		if config.Address < 0x40 || config.Address > 0x4f {
			return nil, errAddressOutOfRange
		}
		i2cAddress = config.Address
	}

	senseResistor := DefaultSenseResistor
	if config.SenseResistor != 0 {
		if config.SenseResistor < 1 {
			return nil, errSenseResistorValueInvalid
		}
		senseResistor = config.SenseResistor
	}

	maxCurrent := DefaultMaxCurrent
	if config.MaxCurrent != 0 {
		if config.MaxCurrent < 1 {
			return nil, errMaxCurrentInvalid
		}
		maxCurrent = config.MaxCurrent
	}

	dev := &Dev{
		m: &mmr.Dev8{
			Conn:  &i2c.Dev{Bus: bus, Addr: uint16(i2cAddress)},
			Order: binary.BigEndian,
		},
	}

	if err := dev.Calibrate(senseResistor, maxCurrent); err != nil {
		return nil, err
	}

	if err := dev.m.WriteUint16(configRegister, 0x1FFF); err != nil {
		return nil, errWritingToConfigRegister
	}

	return dev, nil
}

// Dev is a handle to the ina219 sensor.
type Dev struct {
	m *mmr.Dev8

	mu         sync.Mutex
	currentLSB physic.ElectricCurrent
	powerLSB   physic.Power
}

const (
	configRegister       = 0x00
	shuntVoltageRegister = 0x01
	busVoltageRegister   = 0x02
	powerRegister        = 0x03
	currentRegister      = 0x04
	calibrationRegister  = 0x05
)

// Sense reads the power values from the ina219 sensor
func (d *Dev) Sense() (PowerMonitor, error) {
	// One rx buffer for entire transaction
	d.mu.Lock()
	defer d.mu.Unlock()

	var pm PowerMonitor

	shunt, err := d.m.ReadUint16(shuntVoltageRegister)
	if err != nil {
		return PowerMonitor{}, errReadShunt
	}
	// Least significant bit is 10µV
	pm.Shunt = physic.ElectricPotential(shunt) * 10 * physic.MicroVolt

	bus, err := d.m.ReadUint16(busVoltageRegister)
	if err != nil {
		return PowerMonitor{}, errReadBus
	}
	// check if bit zero is set, if set ADC has overflowed
	if bus&1 > 0 {
		return PowerMonitor{}, errRegisterOverflow
	}
	pm.Voltage = physic.ElectricPotential(bus>>3) * 4 * physic.MilliVolt
	// Least significant bit is 4mV

	// if calibration register is not set then current and power readings are
	// meaningless
	// if d.caibrated {
	current, err := d.m.ReadUint16(currentRegister)
	if err != nil {
		return PowerMonitor{}, errReadCurrent
	}
	pm.Current = physic.ElectricCurrent(current) * d.currentLSB

	power, err := d.m.ReadUint16(powerRegister)
	if err != nil {
		return PowerMonitor{}, errReadPower
	}
	pm.Power = physic.Power(power) * d.powerLSB
	// }

	return pm, nil
}

// Since physic electrical is in nano units we need
// to scale taking care to not overflow int64 or loose resolution.
const calibratescale int64 = ((int64(physic.Ampere) * int64(physic.Ohm)) / 100000) << 12

// Calibrate sets the scaling factor of the current and power registers for the
// maximum resolution. Calibrate is run on init. here it allows you to make
// tune the measured value with actual value
func (d *Dev) Calibrate(sense physic.ElectricResistance, maxCurrent physic.ElectricCurrent) error {
	if sense <= 0 {
		return errSenseResistorValueInvalid
	}
	if maxCurrent <= 0 {
		return errMaxCurrentInvalid
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.currentLSB = maxCurrent / (2 << 15)
	d.powerLSB = physic.Power(d.currentLSB * 20)
	// cal = 0.04096 / (current LSB * Shunt Resistance) where lsb is in Amps and
	// resistance is in ohms.
	// calibration register is 16 bits wide.
	cal := uint16(calibratescale / (int64(d.currentLSB) * int64(sense)))
	return d.m.WriteUint16(calibrationRegister, cal)
}

// PowerMonitor represents measurements from ina219 sensor.
type PowerMonitor struct {
	Shunt   physic.ElectricPotential
	Voltage physic.ElectricPotential
	Current physic.ElectricCurrent
	Power   physic.Power
}

// String returns a PowerMonitor as string
func (p PowerMonitor) String() string {
	return fmt.Sprintf("Bus: %s, Current: %s, Power: %s, Shunt: %s", p.Voltage, p.Current, p.Power, p.Shunt)
}

var (
	errReadShunt                 = errors.New("failed to read shunt voltage")
	errReadBus                   = errors.New("failed to read bus voltage")
	errReadPower                 = errors.New("failed to read power")
	errReadCurrent               = errors.New("failed to read current")
	errAddressOutOfRange         = errors.New("i2c address out of range")
	errSenseResistorValueInvalid = errors.New("sense resistor value cannot be negative or zero")
	errMaxCurrentInvalid         = errors.New("max current cannot be negative or zero")
	errRegisterOverflow          = errors.New("bus voltage register overflow")
	errWritingToConfigRegister   = errors.New("failed to write to configuration register")
)
