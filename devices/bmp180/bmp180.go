// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bmp180 controls a Bosch BMP180 device over I²C.
//
// Datasheet
//
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BMP180-DS000-121.pdf
//
// The font the official datasheet on page 15 is hard to read, a copy with
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
	"log"
	"sync"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/devices"
)

// Oversampling affects how much time is taken to measure pressure.
type Oversampling uint8

// Possible oversampling values.
//
// The higher the more time and power it takes to take a measurement. Even at
// 8x for pressure sensor, it is less than 30ms albeit at small increased power
// consumption, which may increase the temperature reading.
const (
	O1x Oversampling = 0
	O2x Oversampling = 1
	O4x Oversampling = 2
	O8x Oversampling = 3
)

const oversamplingName = "1x2x4x8x"

var oversamplingIndex = [...]uint8{0, 2, 4, 6, 8}

func (i Oversampling) String() string {
	if i >= Oversampling(len(oversamplingIndex)-1) {
		return fmt.Sprintf("Oversampling(%d)", i)
	}
	return oversamplingName[oversamplingIndex[i]:oversamplingIndex[i+1]]
}

// Dev is a handle to a bmp180.
type Dev struct {
	dev mmr.Dev8
	os  Oversampling
	cal calibration

	mu   sync.Mutex
	stop chan struct{}
	wg   sync.WaitGroup
}

func (d *Dev) String() string {
	return fmt.Sprintf("BMP180{%s}", d.dev.Conn)
}

// Sense returns measurements as °C and kPa.
func (d *Dev) Sense(env *devices.Environment) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		return wrap(errors.New("already sensing continuously"))
	}
	return d.sense(env)
}

// SenseContinuous implements devices.Environmental.
func (d *Dev) SenseContinuous(interval time.Duration) (<-chan devices.Environment, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
		d.wg.Wait()
	}

	sensing := make(chan devices.Environment)
	d.stop = make(chan struct{})
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		defer close(sensing)
		d.sensingContinuous(interval, sensing, d.stop)
	}()
	return sensing, nil
}

// Halt stops continuous reading.
func (d *Dev) Halt() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		close(d.stop)
		d.stop = nil
		d.wg.Wait()
	}
	return nil
}

// New returns an object that communicates over I²C to BMP180 environmental
// sensor.
//
// The I²C bus frequency can be up to 3.4MHz.
func New(b i2c.Bus, os Oversampling) (d *Dev, err error) {
	bus := &i2c.Dev{Bus: b, Addr: 0x77}
	d = &Dev{dev: mmr.Dev8{Conn: bus, Order: binary.BigEndian}, os: os}

	// Confirm the chip ID.
	id, err := d.dev.ReadUint8(0xD0)
	if err != nil {
		return nil, wrap(err)
	}
	if id != 0x55 {
		return nil, wrap(fmt.Errorf("unexpected chip id 0x%x; is this a BMP180?", id))
	}

	// Read calibration data.
	if err := d.dev.ReadStruct(0xAA, &d.cal); err != nil {
		return nil, wrap(err)
	}
	if !d.cal.isValid() {
		return nil, wrap(errors.New("calibration data is invalid"))
	}
	return d, nil
}

//

func (d *Dev) sense(env *devices.Environment) error {
	// Request temperature convertion and read measurement.
	if err := d.dev.WriteUint8(0xF4, 0x20|0x0E); err != nil {
		return wrap(err)
	}
	time.Sleep(4500 * time.Microsecond)
	ut, err := d.dev.ReadUint16(0xF6)
	if err != nil {
		return wrap(err)
	}
	temp := d.cal.compensateTemp(ut)

	// Request pressure conversion and read measurement.
	if err := d.dev.WriteUint8(0xF4, 0x20|0x14|(uint8(d.os)<<6)); err != nil {
		return wrap(err)
	}
	time.Sleep(pressureConvTime[d.os])
	var pressureBuf [3]byte
	if err := d.dev.ReadStruct(0xF6, pressureBuf[:]); err != nil {
		return wrap(err)
	}
	up := (int32(pressureBuf[0])<<16 + int32(pressureBuf[1])<<8 | int32(pressureBuf[2])) >> (8 - d.os)
	pressure := d.cal.compensatePressure(up, int32(ut), d.os)
	env.Temperature = devices.Celsius(temp * 100)
	env.Pressure = devices.KPascal(pressure)
	return nil
}

func (d *Dev) sensingContinuous(interval time.Duration, sensing chan<- devices.Environment, stop <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		// Do one initial sensing right away.
		var e devices.Environment
		d.mu.Lock()
		err := d.sense(&e)
		d.mu.Unlock()
		if err != nil {
			log.Printf("bmp180: failed to sense: %v", err)
			return
		}
		select {
		case sensing <- e:
		case <-stop:
			return
		}
		select {
		case <-stop:
			return
		case <-t.C:
		}
	}
}

//

// Maximum conversion time for pressure.
var pressureConvTime = [...]time.Duration{
	4500 * time.Microsecond,
	7500 * time.Microsecond,
	13500 * time.Microsecond,
	25500 * time.Microsecond,
}

// calibration data read from the internal EEPROM (datasheet page 13).
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

func wrap(err error) error {
	return fmt.Errorf("bmp180: %v", err)
}

var _ devices.Environmental = &Dev{}
var _ devices.Device = &Dev{}
var _ fmt.Stringer = &Dev{}
