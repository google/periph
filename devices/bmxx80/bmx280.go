// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bmxx80

import (
	"time"

	"periph.io/x/periph/conn/physic"
)

// sense280 reads the device's registers for bme280/bmp280.
//
// It must be called with d.mu lock held.
func (d *Dev) sense280(e *physic.Env) error {
	// All registers must be read in a single pass, as noted at page 21, section
	// 4.1.
	// Pressure: 0xF7~0xF9
	// Temperature: 0xFA~0xFC
	// Humidity: 0xFD~0xFE
	buf := [8]byte{}
	b := buf[:]
	if !d.isBME {
		b = buf[:6]
	}
	if err := d.readReg(0xF7, b); err != nil {
		return err
	}
	// These values are 20 bits as per doc.
	pRaw := int32(buf[0])<<12 | int32(buf[1])<<4 | int32(buf[2])>>4
	tRaw := int32(buf[3])<<12 | int32(buf[4])<<4 | int32(buf[5])>>4

	t, tFine := d.cal280.compensateTempInt(tRaw)
	// Convert CentiCelsius to Kelvin.
	e.Temperature = physic.Temperature(t)*10*physic.MilliCelsius + physic.ZeroCelsius

	if d.opts.Pressure != Off {
		p := d.cal280.compensatePressureInt64(pRaw, tFine)
		// It has 8 bits of fractional Pascal.
		e.Pressure = physic.Pressure(p) * 15625 * physic.MicroPascal / 4
	}

	if d.opts.Humidity != Off {
		// This value is 16 bits as per doc.
		hRaw := int32(buf[6])<<8 | int32(buf[7])
		h := physic.RelativeHumidity(d.cal280.compensateHumidityInt(hRaw, tFine))
		// Convert base 1024 to base 1000.
		e.Humidity = h * physic.MicroRH * 1000 / 1024
	}
	return nil
}

func (d *Dev) isIdle280() (bool, error) {
	// status
	v := [1]byte{}
	if err := d.readReg(0xF3, v[:]); err != nil {
		return false, err
	}
	// Make sure bit 3 is cleared. Bit 0 is only important at device boot up.
	return v[0]&8 == 0, nil
}

// mode is the operating mode.
type mode byte

const (
	sleep  mode = 0 // no operation, all registers accessible, lowest power, selected after startup
	forced mode = 1 // perform one measurement, store results and return to sleep mode
	normal mode = 3 // perpetual cycling of measurements and inactive periods
)

// standby is the time the BMx280 waits idle between measurements. This reduces
// power consumption when the host won't read the values as fast as the
// measurements are done.
type standby uint8

// Possible standby values, these determines the refresh rate.
const (
	s500us   standby = 0
	s10msBME standby = 6
	s20msBME standby = 7
	s62ms    standby = 1
	s125ms   standby = 2
	s250ms   standby = 3
	s500ms   standby = 4
	s1s      standby = 5
	s2sBMP   standby = 6
	s4sBMP   standby = 7
)

func chooseStandby(isBME bool, d time.Duration) standby {
	switch {
	case d < 10*time.Millisecond:
		return s500us
	case isBME && d < 20*time.Millisecond:
		return s10msBME
	case isBME && d < 62500*time.Microsecond:
		return s20msBME
	case d < 125*time.Millisecond:
		return s62ms
	case d < 250*time.Millisecond:
		return s125ms
	case d < 500*time.Millisecond:
		return s250ms
	case d < time.Second:
		return s500ms
	case d < 2*time.Second:
		return s1s
	case !isBME && d < 4*time.Second:
		return s2sBMP
	default:
		if isBME {
			return s1s
		}
		return s4sBMP
	}
}

// newCalibration parses calibration data from both buffers.
func newCalibration(tph, h []byte) (c calibration280) {
	c.t1 = uint16(tph[0]) | uint16(tph[1])<<8
	c.t2 = int16(tph[2]) | int16(tph[3])<<8
	c.t3 = int16(tph[4]) | int16(tph[5])<<8
	c.p1 = uint16(tph[6]) | uint16(tph[7])<<8
	c.p2 = int16(tph[8]) | int16(tph[9])<<8
	c.p3 = int16(tph[10]) | int16(tph[11])<<8
	c.p4 = int16(tph[12]) | int16(tph[13])<<8
	c.p5 = int16(tph[14]) | int16(tph[15])<<8
	c.p6 = int16(tph[16]) | int16(tph[17])<<8
	c.p7 = int16(tph[18]) | int16(tph[19])<<8
	c.p8 = int16(tph[20]) | int16(tph[21])<<8
	c.p9 = int16(tph[22]) | int16(tph[23])<<8
	c.h1 = uint8(tph[25])

	c.h2 = int16(h[0]) | int16(h[1])<<8
	c.h3 = uint8(h[2])
	c.h4 = int16(h[3])<<4 | int16(h[4])&0xF
	c.h5 = int16(h[4])>>4 | int16(h[5])<<4
	c.h6 = int8(h[6])

	return c
}

type calibration280 struct {
	t1                             uint16
	t2, t3                         int16
	p1                             uint16
	p2, p3, p4, p5, p6, p7, p8, p9 int16
	h2                             int16 // Reordered for packing
	h1, h3                         uint8
	h4, h5                         int16
	h6                             int8
}

// Pages 23-24

// compensateTempInt returns temperature in °C, resolution is 0.01 °C.
// Output value of 5123 equals 51.23 C.
//
// raw has 20 bits of resolution.
func (c *calibration280) compensateTempInt(raw int32) (int32, int32) {
	x := ((raw>>3 - int32(c.t1)<<1) * int32(c.t2)) >> 11
	y := ((((raw>>4 - int32(c.t1)) * (raw>>4 - int32(c.t1))) >> 12) * int32(c.t3)) >> 14
	tFine := x + y
	return (tFine*5 + 128) >> 8, tFine
}

// compensatePressureInt64 returns pressure in Pa in Q24.8 format (24 integer
// bits and 8 fractional bits). Output value of 24674867 represents
// 24674867/256 = 96386.2 Pa = 963.862 hPa.
//
// raw has 20 bits of resolution.
func (c *calibration280) compensatePressureInt64(raw, tFine int32) uint32 {
	x := int64(tFine) - 128000
	y := x * x * int64(c.p6)
	y += (x * int64(c.p5)) << 17
	y += int64(c.p4) << 35
	x = (x*x*int64(c.p3))>>8 + ((x * int64(c.p2)) << 12)
	x = ((int64(1)<<47 + x) * int64(c.p1)) >> 33
	if x == 0 {
		return 0
	}
	p := ((((1048576 - int64(raw)) << 31) - y) * 3125) / x
	x = (int64(c.p9) * (p >> 13) * (p >> 13)) >> 25
	y = (int64(c.p8) * p) >> 19
	return uint32(((p + x + y) >> 8) + (int64(c.p7) << 4))
}

// compensateHumidityInt returns humidity in %RH in Q22.10 format (22 integer
// and 10 fractional bits). Output value of 47445 represents 47445/1024 =
// 46.333%
//
// raw has 16 bits of resolution.
func (c *calibration280) compensateHumidityInt(raw, tFine int32) uint32 {
	x := tFine - 76800
	x1 := raw<<14 - int32(c.h4)<<20 - int32(c.h5)*x
	x2 := (x1 + 16384) >> 15
	x3 := (x * int32(c.h6)) >> 10
	x4 := (x * int32(c.h3)) >> 11
	x5 := (x3 * (x4 + 32768)) >> 10
	x6 := ((x5+2097152)*int32(c.h2) + 8192) >> 14
	x = x2 * x6
	x = x - ((((x>>15)*(x>>15))>>7)*int32(c.h1))>>4
	if x < 0 {
		return 0
	}
	if x > 419430400 {
		return 419430400 >> 12
	}
	return uint32(x >> 12)
}
