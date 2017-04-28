// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bmp180 controls a Bosch BMP180 device over I²C.
//
// Datasheet
//
// The official data sheet can be found here:
//
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BMP180-DS000-121.pdf
//
// The font the official datasheet on page 15 is unreadable, a copy with
// readable text can be found here:
//
// https://cdn-shop.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf
//
// Notes on the datasheet
//
// The results of the calculations in the algorithm on page 15 are partly
// wrong. It looks like the original authors used non-integer calculations and
// some nubers were rounded. Take the results of the calculations with a grain
// of salt.
package bmp180

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/devices"
)

// Oversampling affects how much time is taken to measure pressure.
type Oversampling uint8

const (
	// Possible oversampling values
	No  Oversampling = 0
	O2x Oversampling = 1
	O4x Oversampling = 2
	O8x Oversampling = 3

	// bit offsets in regCtrlMeas
	ctrlMeasurementControlShift = 0
	ctrlStartConversionShift    = 5
	ctrlOversamplingShift       = 6

	// commands
	cmdStartTempConv     uint8 = (1 << ctrlStartConversionShift) | (0x0E << ctrlMeasurementControlShift)
	cmdStartPressureConv uint8 = (1 << ctrlStartConversionShift) | (0x14 << ctrlMeasurementControlShift)

	chipAddress = 0x77 // the address of a BMP180 is fixed at 0x77
	chipID      = 0x55 // contents of the ID register, always 0x55

	// registers
	regChipID           = 0xD0 // register contains the chip id
	regCalibrationStart = 0xAA // first calibration register address
	regSoftReset        = 0xE0 // soft reset register
	regCtrlMeas         = 0xF4 // control measurement
	regOut              = 0xF6 // 3 bytes register with measurement data

	softResetValue = 0xB6

	tempConvTime = 4500 * time.Microsecond // maximum conversion time for temperature
)

// maximum conversion time for pressure
var pressureConvTime = [...]time.Duration{
	4500 * time.Microsecond,
	7500 * time.Microsecond,
	13500 * time.Microsecond,
	25500 * time.Microsecond,
}

// Dev is a handle to a bmp180.
type Dev struct {
	dev *mmr.Dev8
	cal calibration
	os  Oversampling
}

func (d *Dev) String() string {
	return fmt.Sprintf("BMP180{%s}", d.dev.Conn)
}

// Sense returns measurements as °C and kPa.
func (d *Dev) Sense(env *devices.Environment) error {
	// start conversion for temperature
	if err := d.dev.WriteUint8(regCtrlMeas, cmdStartTempConv); err != nil {
		return err
	}

	time.Sleep(tempConvTime)

	// read value
	ut, err := d.dev.ReadUint16(regOut)
	if err != nil {
		return err
	}

	temp := d.cal.compensateTemp(ut)

	// start conversion for pressure
	cmd := cmdStartPressureConv
	cmd |= uint8(d.os) << ctrlOversamplingShift

	if err := d.dev.WriteUint8(regCtrlMeas, cmd); err != nil {
		return err
	}

	time.Sleep(pressureConvTime[d.os])

	// read value
	var pressureBuf [3]byte
	if err := d.dev.ReadStruct(regOut, pressureBuf[:]); err != nil {
		return err
	}

	up := (int32(pressureBuf[0])<<16 + int32(pressureBuf[1])<<8 | int32(pressureBuf[2])) >> (8 - d.os)

	pressure := d.cal.compensatePressure(up, int32(ut), d.os)

	env.Temperature = devices.Celsius(temp * 100)
	env.Pressure = devices.KPascal(pressure)
	return nil
}

// Halt is a noop for the BMP180.
func (d *Dev) Halt() error {
	return nil
}

// Reset issues a soft reset to the device.
func (d *Dev) Reset() error {
	// issue soft reset to initialize device
	if err := d.dev.WriteUint8(regSoftReset, softResetValue); err != nil {
		return err
	}

	// wait for restart
	time.Sleep(10 * time.Millisecond)

	return nil
}

// New returns an object that communicates over I²C to BMP180 environmental
// sensor. The frequency for the bus can be up to 3.4MHz.
func New(b i2c.Bus, os Oversampling) (d *Dev, err error) {
	bus := &i2c.Dev{Bus: b, Addr: chipAddress}
	d = &Dev{
		os: os,
		dev: &mmr.Dev8{
			Conn:  bus,
			Order: binary.BigEndian,
		},
	}

	id, err := d.dev.ReadUint8(regChipID)
	if err != nil {
		return nil, err
	}

	if id != chipID {
		return nil, fmt.Errorf("bmp180: unexpected chip id 0x%x; is this a BMP180?", id)
	}

	// read calibration data from internal EEPROM, 11 registers with two bytes each
	if err := d.dev.ReadStruct(regCalibrationStart, &d.cal); err != nil {
		return nil, err
	}

	if !d.cal.isValid() {
		return nil, errors.New("calibration data is invalid")
	}

	return d, nil
}

// calibration data read from the internal EEPROM (datasheet page 13)
type calibration struct {
	AC1, AC2, AC3 int16
	AC4, AC5, AC6 uint16
	B1, B2        int16
	MB, MC, MD    int16
}

func isValid(i int16) bool {
	return i != 0 && i != ^int16(0)
}

func isValidU(i uint16) bool {
	return i != 0 && i != 0xFFFF
}

// valid checks whether the calibration data is valid.
func (c *calibration) isValid() bool {
	return isValid(c.AC1) && isValid(c.AC2) && isValid(c.AC3) && isValidU(c.AC4) && isValidU(c.AC5) && isValidU(c.AC6) && isValid(c.B1) && isValid(c.B2) && isValid(c.MB) && isValid(c.MC) && isValid(c.MD)
}

// compensateTemp returns temperature in °C, resolution is 0.1 °C.
// Output value of 512 equals 51.2 C.
func (c *calibration) compensateTemp(raw uint16) int32 {
	x1 := ((int64(raw) - int64(c.AC6)) * int64(c.AC5)) >> 15
	x2 := (int64(c.MC) << 11) / (x1 + int64(c.MD))
	b5 := x1 + x2
	t := (b5 + 8) >> 4
	return int32(t)
}

// compensatePressure returns pressure in Pa.
func (c *calibration) compensatePressure(up, ut int32, os Oversampling) uint32 {
	x1 := ((int64(ut) - int64(c.AC6)) * int64(c.AC5)) >> 15
	x2 := (int64(c.MC) * 2048) / (x1 + int64(c.MD))
	b5 := x1 + x2

	b6 := b5 - 4000
	x1 = (int64(c.B2) * ((b6 * b6) >> 12)) >> 11
	x2 = int64(c.AC2) * b6 >> 11
	x3 := x1 + x2
	b3 := (((int64(c.AC1)*4 + x3) << uint(os)) + 2) / 4

	x1 = (int64(c.AC3) * b6) >> 13
	x2 = (int64(c.B1) * ((b6 * b6) >> 12)) >> 16
	x3 = ((x1 + x2) + 2) / 4
	b4 := (int64(c.AC4) * (x3 + 32768)) >> 15
	b7 := (int64(up) - b3) * (50000 >> uint(os))

	var p int64
	if b7 < 0x80000000 {
		p = (b7 * 2) / b4
	} else {
		p = (b7 / b4) * 2
	}

	x1 = (p >> 8) * (p >> 8)
	x1 = (x1 * 3038) >> 16
	x2 = (-7357 * p) >> 16
	p = p + (x1+x2+3791)>>4
	return uint32(p)
}

var _ devices.Environmental = &Dev{}
var _ devices.Device = &Dev{}
