// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ina219

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

// Option is a configuration function
type Option func(*Ina219) error

// Address is used to set the I²C address if not the default address of 0x40.
func Address(address uint8) Option {
	return func(i *Ina219) error {
		if address < 0x40 || address > 0x4f {
			return errAddressOutOfRange
		}
		i.mu.Lock()
		i.addr = uint16(address)
		if i.c != nil {
			i.c = &i2c.Dev{
				Bus:  i.bus,
				Addr: i.addr,
			}
		}
		i.mu.Unlock()
		return nil
	}
}

// MaxCurrent is a configuration function to set the maximum expected current.
// Not required if the device has the default maximum current 3.2A
func MaxCurrent(max physic.ElectricCurrent) Option {
	return func(i *Ina219) error {
		if max < physic.NanoAmpere {
			return errMaxCurrentInvalid
		}
		i.mu.Lock()
		i.maxCurrent = max
		i.mu.Unlock()
		return nil
	}
}

// SenseResistor is a configuration function to set the actual measured value of
// the sense resistor.
// Not required if the device has the default value of 100mΩ
func SenseResistor(sense physic.ElectricResistance) Option {
	return func(i *Ina219) error {
		if sense < physic.NanoOhm {
			return errSenseResistorValueInvalid
		}
		i.mu.Lock()
		i.sense = sense
		i.mu.Unlock()
		return nil
	}
}

const (
	// DefaultSenseResistor is 100 mΩ and can be changed using SenseResistor(value)
	DefaultSenseResistor = 100 * physic.MilliOhm
	// DefaultI2CAddress is 0x40 and can be changed using Address(address)
	DefaultI2CAddress = 0x40
	// DefaultMaxCurrent is 3.2 and can be changed using MaxCurrent(value)
	DefaultMaxCurrent = 3200 * physic.MilliAmpere
)

// New opens a handle to an ina219
func New(bus i2c.Bus, opts ...Option) (*Ina219, error) {

	dev := &Ina219{
		caibrated:  false,
		sense:      DefaultSenseResistor,
		maxCurrent: DefaultMaxCurrent,
		addr:       DefaultI2CAddress,
		bus:        bus,
	}

	for _, opt := range opts {
		if err := opt(dev); err != nil {
			return nil, err
		}
	}
	dev.c = &i2c.Dev{
		Bus:  bus,
		Addr: dev.addr,
	}

	if err := dev.Calibrate(); err != nil {
		return nil, err
	}

	return dev, nil
}

// Ina219 is a handle to the ina219 sensor.
type Ina219 struct {
	bus  i2c.Bus
	c    conn.Conn
	addr uint16

	mu         sync.Mutex
	caibrated  bool
	sense      physic.ElectricResistance
	maxCurrent physic.ElectricCurrent
	currentLSB physic.ElectricCurrent
	powerLSB   physic.Power
	stop       chan struct{}
	wg         sync.WaitGroup
}

const (
	shuntVoltageRegister = 0x01
	busVoltageRegister   = 0x02
	powerRegister        = 0x03
	currentRegister      = 0x04
	calibrationRegister  = 0x05
)

// Sense reads the power values from the ina219 sensor
func (d *Ina219) Sense() (PowerMonitor, error) {
	// One rx buffer for entire transaction
	rx := make([]byte, 2)
	d.mu.Lock()
	defer d.mu.Unlock()
	if err := d.c.Tx([]byte{shuntVoltageRegister}, rx); err != nil {
		return PowerMonitor{}, errReadShunt
	}
	var pm PowerMonitor
	pm.Shunt = physic.ElectricPotential(int16(binary.BigEndian.Uint16(rx)))
	// Least significant bit is 10µV
	pm.Shunt *= (10 * physic.MicroVolt)

	if err := d.c.Tx([]byte{busVoltageRegister}, rx); err != nil {
		return PowerMonitor{}, errReadBus
	}
	// check bit 0 of bus voltage register, if set data is invalid
	if hasBit(rx[1], 0) {
		return PowerMonitor{}, errRegisterOverflow
	}
	pm.Voltage = physic.ElectricPotential(int16(binary.BigEndian.Uint16(rx) >> 3))
	// Least significant bit is 4mV
	pm.Voltage *= (4 * physic.MilliVolt)

	// if calibration register is not set then current and power readings are
	// meaningless
	if d.caibrated {
		if err := d.c.Tx([]byte{currentRegister}, rx); err != nil {
			return PowerMonitor{}, errReadCurrent
		}
		pm.Current = physic.ElectricCurrent(binary.BigEndian.Uint16(rx))
		pm.Current *= d.currentLSB

		if err := d.c.Tx([]byte{powerRegister}, rx); err != nil {
			return PowerMonitor{}, errReadPower
		}
		pm.Power = physic.Power(binary.BigEndian.Uint16(rx))
		pm.Power *= d.powerLSB
	}

	return pm, nil
}

// Since physic electrical is in nano units we need
// to scale taking care to not overflow int64 or loose resolution.
const calibratescale int64 = ((int64(physic.Ampere) * int64(physic.Ohm)) / 100000) << 12

// Calibrate sets the scaling factor of the current and power registers for the
// maximum resolution. Calibrate is run on init. here it allows you to make
// tune the measured value with actual value
func (d *Ina219) Calibrate() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sense < 1 {
		return errSenseResistorValueInvalid
	}
	if d.maxCurrent < 1 {
		return errMaxCurrentInvalid
	}
	d.currentLSB = d.maxCurrent / (2 << 15)
	d.powerLSB = physic.Power(d.currentLSB / 20)
	// cal = 0.04096 / (current LSB * Shunt Resistance) where lsb is in Amps and
	// resistance is in ohms.
	// calibration register is 16 bits wide.
	cal := uint16(calibratescale / (int64(d.currentLSB) * int64(d.sense)))
	if err := d.WriteRegister(calibrationRegister, cal); err != nil {
		return err
	}

	d.caibrated = true
	return nil
}

// Reset performs a power on reset.
func (d *Ina219) reset() error {
	// reset Ina219ice by writing 1 to the 15th bit in register 0x00
	if err := d.c.Tx([]byte{0x00, 0x80, 0x00}, nil); err != nil {
		return err
	}
	// give ina219 time to preform reset.
	time.After(time.Millisecond * 20)
	rx := make([]byte, 2)
	if err := d.c.Tx([]byte{0x00}, rx); err != nil {
		return err
	}

	const powerOnReset = 0x399F // register 0x00 on POR
	if binary.BigEndian.Uint16(rx) != powerOnReset {
		return errResetError
	}
	return nil
}

// WriteRegister is used to write to register address directly
func (d *Ina219) WriteRegister(register uint8, data uint16) error {
	databytes := make([]byte, 2)
	binary.BigEndian.PutUint16(databytes, data)
	tx := []byte{register, databytes[0], databytes[1]}
	return d.c.Tx(tx, nil)
}

// ReadRegister is used to read to register address directly
func (d *Ina219) ReadRegister(register uint8) (uint16, error) {
	rx := make([]byte, 2)
	err := d.c.Tx([]byte{register}, rx)
	data := binary.BigEndian.Uint16(rx)
	return data, err
}

// PowerMonitor represents measurements from ina219 sensor.
type PowerMonitor struct {
	Shunt   physic.ElectricPotential
	Voltage physic.ElectricPotential
	Current physic.ElectricCurrent
	Power   physic.Power
}

// String inplment the stringer interface.
func (p PowerMonitor) String() string {
	return fmt.Sprintf("Bus: %s, Current: %s, Power: %s, Shunt: %s", p.Voltage, p.Current, p.Power, p.Shunt)
}

var (
	errResetError                = errors.New("failed to reset Ina219")
	errReadShunt                 = errors.New("failed to read shunt voltage")
	errReadBus                   = errors.New("failed to read bus voltage")
	errReadPower                 = errors.New("failed to read power")
	errReadCurrent               = errors.New("failed to read current")
	errAddressOutOfRange         = errors.New("i2c address out of range")
	errSenseResistorValueInvalid = errors.New("sense resistor value cannot be negative or zero")
	errMaxCurrentInvalid         = errors.New("max current cannot be negative or zero")
	errRegisterOverflow          = errors.New("bus voltage register overflow")
)

func clearBit(n byte, pos uint8) byte {
	mask := ^(1 << pos)
	n &= byte(mask)
	return n
}

func setBit(n byte, pos uint8) byte {
	n |= (1 << pos)
	return n
}

func hasBit(n byte, pos uint8) bool {
	val := n & (1 << pos)
	return (val > 0)
}
