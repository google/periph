// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bme280 controls a Bosch BME280 device over I²C.
//
// Datasheet
//
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280_DS001-11.pdf
package bme280

import (
	"errors"
	"fmt"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/devices"
)

// Oversampling affects how much time is taken to measure each of temperature,
// pressure and humidity.
//
// Temperature must be measured for pressure and humidity to be measured. The
// duration is approximatively:
//     duration_in_ms = 1 + 2*temp + 2*press+0.5 + 2*humidy+0.5
//
// Using high oversampling and low standby results in highest power
// consumption, but this is still below 1mA so we generally don't care.
type Oversampling uint8

// Possible oversampling values.
const (
	No   Oversampling = 0
	O1x  Oversampling = 1
	O2x  Oversampling = 2
	O4x  Oversampling = 3
	O8x  Oversampling = 4
	O16x Oversampling = 5
)

// Standby is the time the BME280 waits idle between measurements. This reduces
// power consumption when the host won't read the values as fast as the
// measurements are done.
type Standby uint8

// Possible standby values, these determines the refresh rate.
const (
	S500us Standby = 0
	S10ms  Standby = 6
	S20ms  Standby = 7
	S62ms  Standby = 1
	S125ms Standby = 2
	S250ms Standby = 3
	S500ms Standby = 4
	S1s    Standby = 5
)

// Filter specifies the internal IIR filter to get steady measurements without
// using oversampling. This is mainly used to reduce power consumption.
type Filter uint8

// Possible filtering values.
const (
	FOff Filter = 0
	F2   Filter = 1
	F4   Filter = 2
	F8   Filter = 3
	F16  Filter = 4
)

// Dev is an handle to a bme280.
type Dev struct {
	d     conn.Conn
	isSPI bool
	c     calibration
}

func (d *Dev) String() string {
	return fmt.Sprintf("BME280{%s}", d.d)
}

// Sense returns measurements as °C, kPa and % of relative humidity.
func (d *Dev) Sense(env *devices.Environment) error {
	// All registers must be read in a single pass, as noted at page 21, section
	// 4.1.
	// Pressure: 0xF7~0xF9
	// Temperature: 0xFA~0xFC
	// Humidity: 0xFD~0xFE
	buf := [0xFF - 0xF7]byte{}
	if err := d.readReg(0xF7, buf[:]); err != nil {
		return err
	}
	// These values are 20 bits as per doc.
	pRaw := int32(buf[0])<<12 | int32(buf[1])<<4 | int32(buf[2])>>4
	tRaw := int32(buf[3])<<12 | int32(buf[4])<<4 | int32(buf[5])>>4
	// This value is 16 bits as per doc.
	hRaw := int32(buf[6])<<8 | int32(buf[7])

	t, tFine := d.c.compensateTempInt(tRaw)
	env.Temperature = devices.Celsius(t * 10)

	p := d.c.compensatePressureInt64(pRaw, tFine)
	env.Pressure = devices.KPascal((int32(p) + 127) / 256)

	h := d.c.compensateHumidityInt(hRaw, tFine)
	env.Humidity = devices.RelativeHumidity((int32(h)*100 + 511) / 1024)
	return nil
}

// Halt stops the bme280 from acquiring measurements.
//
// It is recommended to call to reduce idle power usage.
func (d *Dev) Halt() error {
	// Page 27 (for register) and 12~13 section 3.3.
	return d.writeCommands([]byte{0xF4, byte(sleep)})
}

// Opts is optional options to pass to the constructor.
//
// Recommended (and default) values are O4x for oversampling, S20ms for standby
// and FOff for filter if planing to call frequently, else use S500ms to get a
// bit more than one reading per second.
//
// Address is only used on creation of an I²C-device. Its default value is 0x76.
// It can be set to 0x77. Both values depend on HW configuration of the sensor's
// SDO pin. This has no effect with NewSPI()
//
// BUG(maruel): Remove the Standby flag and replace with a
// WaitForNextSample(time.Duration). Then use the closest value automatically.
type Opts struct {
	Temperature Oversampling
	Pressure    Oversampling
	Humidity    Oversampling
	Standby     Standby
	Filter      Filter
	Address     uint16
}

// NewI2C returns an object that communicates over I²C to BME280 environmental
// sensor.
//
// It is recommended to call Halt() when done with the device so it stops
// sampling.
func NewI2C(b i2c.Bus, opts *Opts) (*Dev, error) {
	addr := uint16(0x76)
	if opts != nil {
		switch opts.Address {
		case 0x76, 0x77:
			addr = opts.Address
		case 0x00:
			// do not do anything
		default:
			return nil, errors.New("bme280: given address not supported by device")
		}
	}
	d := &Dev{d: &i2c.Dev{Bus: b, Addr: addr}, isSPI: false}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// NewSPI returns an object that communicates over SPI to BME280 environmental
// sensor.
//
// Recommended values are O4x for oversampling, S20ms for standby and FOff for
// filter if planing to call frequently, else use S500ms to get a bit more than
// one reading per second.
//
// It is recommended to call Halt() when done with the device so it stops
// sampling.
//
// When using SPI, the CS line must be used.
//
// BUG(maruel): This code was not tested yet, still waiting for a SPI enabled
// device in the mail.
func NewSPI(p spi.Port, opts *Opts) (*Dev, error) {
	// It works both in Mode0 and Mode3.
	c, err := p.DevParams(10000000, spi.Mode3, 8)
	if err != nil {
		return nil, err
	}
	d := &Dev{d: c, isSPI: true}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

//

// mode is stored in config
type mode byte

const (
	sleep  mode = 0 // no operation, all registers accessible, lowest power, selected after startup
	forced mode = 1 // perform one measurement, store results and return to sleep mode
	normal mode = 3 // perpetual cycling of measurements and inactive periods
)

type status byte

const (
	measuring status = 8 // set when conversion is running
	imUpdate  status = 1 // set when NVM data are being copied to image registers
)

var defaults = Opts{
	Temperature: O4x,
	Pressure:    O4x,
	Humidity:    O4x,
	Standby:     S20ms,
	Filter:      FOff,
}

func (d *Dev) makeDev(opts *Opts) error {
	if opts == nil {
		opts = &defaults
	}
	config := []byte{
		// ctrl_meas; put it to sleep otherwise the config update may be ignored.
		0xF4, byte(opts.Temperature)<<5 | byte(opts.Pressure)<<2 | byte(sleep),
		// ctrl_hum
		0xF2, byte(opts.Humidity),
		// config
		0xF5, byte(opts.Standby)<<5 | byte(opts.Filter)<<2,
		// ctrl_meas
		0xF4, byte(opts.Temperature)<<5 | byte(opts.Pressure)<<2 | byte(normal),
	}

	// The device starts in 2ms as per datasheet. No need to wait for boot to be
	// finished.

	var chipID [1]byte
	// Read register 0xD0 to read the chip id.
	if err := d.readReg(0xD0, chipID[:]); err != nil {
		return err
	}
	if chipID[0] != 0x60 {
		return fmt.Errorf("bme280: unexpected chip id %x; is this a BME280?", chipID[0])
	}
	// Read calibration data t1~3, p1~9, 8bits padding, h1.
	var tph [0xA2 - 0x88]byte
	if err := d.readReg(0x88, tph[:]); err != nil {
		return err
	}
	// Read calibration data h2~6
	var h [0xE8 - 0xE1]byte
	if err := d.readReg(0xE1, h[:]); err != nil {
		return err
	}
	if err := d.writeCommands(config[:]); err != nil {
		return err
	}

	d.c = newCalibration(tph[:], h[:])

	return nil
}

func (d *Dev) readReg(reg uint8, b []byte) error {
	// Page 32-33
	if d.isSPI {
		// MSB is 0 for write and 1 for read.
		read := make([]byte, len(b)+1)
		write := make([]byte, len(read))
		// Rest of the write buffer is ignored.
		write[0] = reg
		if err := d.d.Tx(write, read); err != nil {
			return err
		}
		copy(b, read[1:])
		return nil
	}
	return d.d.Tx([]byte{reg}, b)
}

// writeCommands writes a command to the bme280.
//
// Warning: b may be modified!
func (d *Dev) writeCommands(b []byte) error {
	if d.isSPI {
		// Page 33; set RW bit 7 to 0.
		for i := 0; i < len(b); i += 2 {
			b[i] &^= 0x80
		}
	}
	return d.d.Tx(b, nil)
}

// Register table:
// 0x00..0x87  --
// 0x88..0xA1  Calibration data
// 0xA2..0xCF  --
// 0xD0        Chip id; reads as 0x60
// 0xD1..0xDF  --
// 0xE0        Reset by writing 0xB6 to it
// 0xE1..0xF0  Calibration data
// 0xF1        --
// 0xF2        ctrl_hum; ctrl_meas must be written to after for change to this register to take effect
// 0xF3        status
// 0xF4        ctrl_meas
// 0xF5        config
// 0xF6        --
// 0xF7        press_msb
// 0xF8        press_lsb
// 0xF9        press_xlsb
// 0xFA        temp_msb
// 0xFB        temp_lsb
// 0xFC        temp_xlsb
// 0xFD        hum_msb
// 0xFE        hum_lsb

// https://cdn-shop.adafruit.com/datasheets/BST-BME280_DS001-10.pdf
// Page 23

// newCalibration parses calibration data from both buffers.
func newCalibration(tph, h []byte) (c calibration) {
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

type calibration struct {
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
func (c *calibration) compensateTempInt(raw int32) (int32, int32) {
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
func (c *calibration) compensatePressureInt64(raw, tFine int32) uint32 {
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
func (c *calibration) compensateHumidityInt(raw, tFine int32) uint32 {
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

var _ devices.Environmental = &Dev{}
var _ devices.Device = &Dev{}
