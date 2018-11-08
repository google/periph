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

// Opts holds the configuration options.
//
// Slave Address
//
// Depending which pins the A1, A0 pins are connected to will change the slave
// address. Default configuration is address 0x40 (both pins to GND). For a full
// address table see datasheet.
type Opts struct {
	Address       int
	SenseResistor physic.ElectricResistance
	MaxCurrent    physic.ElectricCurrent
}

// DefaultOpts is the recommended default options.
var DefaultOpts = Opts{
	Address:       0x40,
	SenseResistor: 100 * physic.MilliOhm,
	MaxCurrent:    3200 * physic.MilliAmpere,
}

// New opens a handle to an ina219 sensor.
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {

	i2cAddress := DefaultOpts.Address
	if opts.Address != 0 {
		if opts.Address < 0x40 || opts.Address > 0x4f {
			return nil, errAddressOutOfRange
		}
		i2cAddress = opts.Address
	}

	senseResistor := DefaultOpts.SenseResistor
	if opts.SenseResistor != 0 {
		if opts.SenseResistor < 1 {
			return nil, errSenseResistorValueInvalid
		}
		senseResistor = opts.SenseResistor
	}

	maxCurrent := DefaultOpts.MaxCurrent
	if opts.MaxCurrent != 0 {
		if opts.MaxCurrent < 1 {
			return nil, errMaxCurrentInvalid
		}
		maxCurrent = opts.MaxCurrent
	}

	dev := &Dev{
		m: mmr.Dev8{
			Conn:  &i2c.Dev{Bus: bus, Addr: uint16(i2cAddress)},
			Order: binary.BigEndian,
		},
	}

	if err := dev.calibrate(senseResistor, maxCurrent); err != nil {
		return nil, err
	}

	if err := dev.m.WriteUint16(configRegister, 0x1FFF); err != nil {
		return nil, errWritingToConfigRegister
	}

	return dev, nil
}

// Dev is a handle to the ina219 sensor.
type Dev struct {
	m mmr.Dev8

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

// Sense reads the power values from the ina219 sensor.
func (d *Dev) Sense() (PowerMonitor, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var pm PowerMonitor

	shunt, err := d.m.ReadUint16(shuntVoltageRegister)
	if err != nil {
		return PowerMonitor{}, errReadShunt
	}
	// Least significant bit is 10µV.
	pm.Shunt = physic.ElectricPotential(shunt) * 10 * physic.MicroVolt

	bus, err := d.m.ReadUint16(busVoltageRegister)
	if err != nil {
		return PowerMonitor{}, errReadBus
	}
	// Check if bit zero is set, if set the ADC has overflowed.
	if bus&1 > 0 {
		return PowerMonitor{}, errRegisterOverflow
	}

	// Least significant bit is 4mV.
	pm.Voltage = physic.ElectricPotential(bus>>3) * 4 * physic.MilliVolt

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

	return pm, nil
}

// Since physic electrical is in nano units we need to scale taking care to not
// overflow int64 or loose resolution.
const calibratescale int64 = ((int64(physic.Ampere) * int64(physic.Ohm)) / 100000) << 12

// calibrate sets the scaling factor of the current and power registers for the
// maximum resolution. calibrate is run on init.
func (d *Dev) calibrate(sense physic.ElectricResistance, maxCurrent physic.ElectricCurrent) error {
	// TODO: Check calibration with float implementation in tests.
	if sense <= 0 {
		return errSenseResistorValueInvalid
	}
	if maxCurrent <= 0 {
		return errMaxCurrentInvalid
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.currentLSB = maxCurrent / (1 << 15)
	d.powerLSB = physic.Power((maxCurrent*20 + (1 << 14)) / (1 << 15))
	// Calibration Register = 0.04096 / (current LSB * Shunt Resistance)
	// Where lsb is in Amps and resistance is in ohms.
	// Calibration register is 16 bits.
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
