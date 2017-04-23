// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bmp180 controls a Bosch BMP180 device over I²C.
//
// Datasheet
//
// https://cdn-shop.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf
package bmp180

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/devices"
)

// Oversampling affects how much time is taken to measure pressure.
type Oversampling uint8

// Possible oversampling values.
const (
	No  Oversampling = 0
	O2x Oversampling = 1
	O4x Oversampling = 2
	O8x Oversampling = 3
)

// Dev is an handle to a bmp180.
type Dev struct {
	d  conn.Conn
	c  calibration
	os Oversampling
}

func (d *Dev) String() string {
	return fmt.Sprintf("BMP180{%s}", d.d)
}

const (
	ctrlCommandTemperature = 0x0E
	ctrlCommandPressure    = 0x14
)

// bit offsets in regCtrlMeas
const (
	ctrlMeasurementControl = 0
	ctrlStartConversion    = 5
	ctrlOversampling       = 6
)

// Sense returns measurements as °C and kPa.
func (d *Dev) Sense(env *devices.Environment) error {
	// start conversion for temperature
	var cmd byte
	cmd = 1 << ctrlStartConversion
	cmd |= ctrlCommandTemperature << ctrlMeasurementControl

	if err := d.d.Tx([]byte{regCtrlMeas, cmd}, nil); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	// read value
	var tempBuf [2]byte
	if err := d.readReg(regOutMSB, tempBuf[:]); err != nil {
		return err
	}

	UT := uint16(tempBuf[0])<<8 | uint16(tempBuf[1])
	temp := d.c.compensateTemp(UT)

	// start conversion for pressure
	cmd = 1 << ctrlStartConversion
	cmd |= ctrlCommandPressure << ctrlMeasurementControl
	cmd |= byte(d.os) << ctrlOversampling

	if err := d.d.Tx([]byte{regCtrlMeas, cmd}, nil); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	// read value
	var pressureBuf [3]byte
	if err := d.readReg(regOutMSB, pressureBuf[:]); err != nil {
		return err
	}

	UP := (int32(pressureBuf[0])<<16 + int32(pressureBuf[1])<<8 | int32(pressureBuf[2])) >> (8 - d.os)

	pressure := d.c.compensatePressure(UP, int32(UT), d.os)

	env.Temperature = devices.Celsius(temp * 100)
	env.Pressure = devices.KPascal(pressure)
	return nil
}

// Halt is a noop for the BMP180.
func (d *Dev) Halt() error {
	return nil
}

// Opts is optional options to pass to the constructor.
type Opts struct {
	Pressure Oversampling
}

const (
	bmp180Address = 0x77 // the address of a BMP180 is fixed at 0x77
	chipID        = 0x55 // contents of the ID register, always 0x55
)

// New returns an object that communicates over I²C to BMP180 environmental
// sensor.
func New(b i2c.Bus, opts *Opts) (*Dev, error) {
	d := &Dev{
		d:  &i2c.Dev{Bus: b, Addr: bmp180Address},
		os: No,
	}

	if opts != nil {
		d.os = opts.Pressure
	}

	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

var defaults = Opts{
	Pressure: O4x,
}

// registers
const (
	regChipID           = 0xD0 // register contains the chip id
	regCalibrationStart = 0xAA // first calibration register address
	regSoftReset        = 0xE0 // soft reset register
	regCtrlMeas         = 0xF4 // control measurement
	regOutMSB           = 0xF6 // MSB of measurement
	regOutLSB           = 0xF7 // LSB of measurement
	regOutXLSB          = 0xF8 // extended LSB of measurement
)

const softResetValue = 0xB6

func (d *Dev) makeDev(opts *Opts) error {
	if opts == nil {
		opts = &defaults
	}
	var id [1]byte
	if err := d.readReg(regChipID, id[:]); err != nil {
		return err
	}

	if id[0] != chipID {
		return fmt.Errorf("bmp180: unexpected chip id 0x%x; is this a BMP180?", id[0])
	}

	// issue soft reset to initialize device
	if err := d.d.Tx([]byte{regSoftReset, softResetValue}, nil); err != nil {
		return err
	}

	// wait for restart
	time.Sleep(10 * time.Millisecond)

	// read calibration data from internal EEPROM, 11 registers with two bytes each
	var cal [22]byte
	if err := d.readReg(regCalibrationStart, cal[:]); err != nil {
		return err
	}

	// consistency check as suggested by the datasheet (page 13)
	for i := 0; i < len(cal)/2; i++ {
		val := uint16(cal[2*i])<<8 | uint16(cal[2*i+1])
		if val == 0 || val == 0xffff {
			return fmt.Errorf("calibration byte %v has invalid value 0x%x", i, val)
		}
	}

	d.c.AC1 = int16(cal[0])<<8 | int16(cal[1])
	d.c.AC2 = int16(cal[2])<<8 | int16(cal[3])
	d.c.AC3 = int16(cal[4])<<8 | int16(cal[5])
	d.c.AC4 = uint16(cal[6])<<8 | uint16(cal[7])
	d.c.AC5 = uint16(cal[8])<<8 | uint16(cal[9])
	d.c.AC6 = uint16(cal[10])<<8 | uint16(cal[11])
	d.c.B1 = int16(cal[12])<<8 | int16(cal[13])
	d.c.B2 = int16(cal[14])<<8 | int16(cal[15])
	d.c.MB = int16(cal[16])<<8 | int16(cal[17])
	d.c.MC = int16(cal[18])<<8 | int16(cal[19])
	d.c.MD = int16(cal[20])<<8 | int16(cal[21])

	return nil
}

func (d *Dev) readReg(reg uint8, b []byte) error {
	return d.d.Tx([]byte{reg}, b)
}

// calibration data read from the internal EEPROM (datasheet page 13)
type calibration struct {
	AC1, AC2, AC3 int16
	AC4, AC5, AC6 uint16
	B1, B2        int16
	MB, MC, MD    int16
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
func (c *calibration) compensatePressure(UP, UT int32, os Oversampling) uint32 {
	x1 := ((int64(UT) - int64(c.AC6)) * int64(c.AC5)) >> 15
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
	b7 := (int64(UP) - b3) * (50000 >> uint(os))

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
