// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bmxx80

import (
	"encoding/binary"
	"time"

	"periph.io/x/periph/conn/physic"
)

// sense180 reads the device's registers for bmp180.
//
// It must be called with d.mu lock held.
func (d *Dev) sense180(e *physic.Env) error {
	// Request temperature conversion and read measurement.
	if err := d.writeCommands([]byte{0xF4, 0x20 | 0x0E}); err != nil {
		return d.wrap(err)
	}
	doSleep(4500 * time.Microsecond)
	var tempBuf [2]byte
	if err := d.readReg(0xF6, tempBuf[:]); err != nil {
		return d.wrap(err)
	}
	rawTemp := binary.BigEndian.Uint16(tempBuf[:])
	temp := d.cal180.compensateTemp(rawTemp)

	// Request pressure conversion and read measurement.
	if err := d.writeCommands([]byte{0xF4, 0x20 | 0x14 | d.os<<6}); err != nil {
		return d.wrap(err)
	}
	doSleep(pressureConvTime180[d.os])
	var pressureBuf [3]byte
	if err := d.readReg(0xF6, pressureBuf[:]); err != nil {
		return d.wrap(err)
	}
	up := (int32(pressureBuf[0])<<16 + int32(pressureBuf[1])<<8 | int32(pressureBuf[2])) >> (8 - d.os)
	pressure := d.cal180.compensatePressure(up, int32(rawTemp), uint(d.os))
	// Convert DeciCelsius to Kelvin.
	e.Temperature = physic.Temperature(temp)*100*physic.MilliCelsius + physic.ZeroCelsius
	e.Pressure = physic.Pressure(pressure) * physic.Pascal
	return nil
}

// pressureConvTime180 is the maximum conversion time for pressure.
var pressureConvTime180 = [...]time.Duration{
	4500 * time.Microsecond,
	7500 * time.Microsecond,
	13500 * time.Microsecond,
	25500 * time.Microsecond,
}

// calibration180 data read from the internal EEPROM (datasheet page 13).
type calibration180 struct {
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
func (c *calibration180) isValid() bool {
	return isValid(c.AC1) && isValid(c.AC2) && isValid(c.AC3) && isValidU(c.AC4) && isValidU(c.AC5) && isValidU(c.AC6) && isValid(c.B1) && isValid(c.B2) && isValid(c.MB) && isValid(c.MC) && isValid(c.MD)
}

// compensateTemp returns temperature in °C, resolution is 0.1 °C.
// Output value of 512 equals 51.2 C.
func (c *calibration180) compensateTemp(raw uint16) int32 {
	x1 := ((int64(raw) - int64(c.AC6)) * int64(c.AC5)) >> 15
	x2 := (int64(c.MC) << 11) / (x1 + int64(c.MD))
	b5 := x1 + x2
	t := (b5 + 8) >> 4
	return int32(t)
}

// compensatePressure returns pressure in Pa.
func (c *calibration180) compensatePressure(up, ut int32, os uint) uint32 {
	x1 := ((int64(ut) - int64(c.AC6)) * int64(c.AC5)) >> 15
	x2 := (int64(c.MC) * 2048) / (x1 + int64(c.MD))
	b5 := x1 + x2

	b6 := b5 - 4000
	x1 = (int64(c.B2) * ((b6 * b6) >> 12)) >> 11
	x2 = int64(c.AC2) * b6 >> 11
	x3 := x1 + x2
	b3 := (((int64(c.AC1)*4 + x3) << os) + 2) / 4

	x1 = (int64(c.AC3) * b6) >> 13
	x2 = (int64(c.B1) * ((b6 * b6) >> 12)) >> 16
	x3 = ((x1 + x2) + 2) / 4
	b4 := (int64(c.AC4) * (x3 + 32768)) >> 15
	b7 := (int64(up) - b3) * (50000 >> os)

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
