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
	"log"
	"sync"
	"time"

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
//
// The higher the more time and power it takes to take a measurement. Even at
// 16x for all 3 sensor, it is less than 100ms albeit increased power
// consumption may increase the temperature reading.
const (
	Off  Oversampling = 0
	O1x  Oversampling = 1
	O2x  Oversampling = 2
	O4x  Oversampling = 3
	O8x  Oversampling = 4
	O16x Oversampling = 5
)

const oversamplingName = "Off1x2x4x8x16x"

var oversamplingIndex = [...]uint8{0, 3, 5, 7, 9, 11, 14}

func (o Oversampling) String() string {
	if o >= Oversampling(len(oversamplingIndex)-1) {
		return fmt.Sprintf("Oversampling(%d)", o)
	}
	return oversamplingName[oversamplingIndex[o]:oversamplingIndex[o+1]]
}

func (o Oversampling) asValue() int {
	switch o {
	case O1x:
		return 1
	case O2x:
		return 2
	case O4x:
		return 4
	case O8x:
		return 8
	case O16x:
		return 16
	default:
		return 0
	}
}

// Filter specifies the internal IIR filter to get steadier measurements.
//
// Oversampling will get better measurements than filtering but at a larger
// power consumption cost, which may slightly affect temperature measurement.
type Filter uint8

// Possible filtering values.
//
// The higher the filter, the slower the value converges but the more stable
// the measurement is.
const (
	NoFilter Filter = 0
	F2       Filter = 1
	F4       Filter = 2
	F8       Filter = 3
	F16      Filter = 4
)

// Dev is a handle to an initialized bme280.
type Dev struct {
	d         conn.Conn
	isSPI     bool
	opts      Opts
	measDelay time.Duration
	c         calibration

	mu   sync.Mutex
	stop chan struct{}
}

func (d *Dev) String() string {
	return fmt.Sprintf("BME280{%s}", d.d)
}

// Sense requests a one time measurement as °C, kPa and % of relative humidity.
//
// The very first measurements may be of poor quality.
func (d *Dev) Sense(env *devices.Environment) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		return errors.New("bme280: already sensing continuously")
	}
	err := d.writeCommands([]byte{
		// ctrl_meas
		0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(forced),
	})
	if err != nil {
		return err
	}
	time.Sleep(d.measDelay)
	for idle := false; !idle; {
		if idle, err = d.isIdle(); err != nil {
			return err
		}
	}
	return d.sense(env)
}

// SenseContinuous returns measurements as °C, kPa and % of relative humidity
// on a continuous basis.
//
// The application must call Halt() to stop the sensing when done to stop the
// sensor and close the channel.
//
// It's the responsibility of the caller to retrieve the values from the
// channel as fast as possible, otherwise the interval may not be respected.
func (d *Dev) SenseContinuous(interval time.Duration) (<-chan devices.Environment, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		// Don't send the stop command to the device.
		close(d.stop)
		d.stop = nil
	}
	s := chooseStandby(interval - d.measDelay)
	err := d.writeCommands([]byte{
		// config
		0xF5, byte(s)<<5 | byte(d.opts.Filter)<<2,
		// ctrl_meas
		0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(normal),
	})
	if err != nil {
		return nil, err
	}
	sensing := make(chan devices.Environment)
	d.stop = make(chan struct{})
	go func() {
		defer close(sensing)
		d.sensingContinuous(interval, sensing, d.stop)
	}()
	return sensing, nil
}

// Halt stops the bme280 from acquiring measurements as initiated by Sense().
//
// It is recommended to call this function before terminating the process to
// reduce idle power usage.
func (d *Dev) Halt() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop == nil {
		return nil
	}
	close(d.stop)
	d.stop = nil
	// Page 27 (for register) and 12~13 section 3.3.
	return d.writeCommands([]byte{
		// config
		0xF5, byte(s1s)<<5 | byte(NoFilter)<<2,
		// ctrl_meas
		0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
	})
}

// Opts is optional options to pass to the constructor.
//
// Recommended (and default) values are O4x for oversampling.
//
// Address can only used on creation of an I²C-device. Its default value is
// 0x76. It can be set to 0x77. Both values depend on HW configuration of the
// sensor's SDO pin.
//
// Filter is only used while using SenseContinuous().
//
// Recommended sensing settings as per the datasheet:
//
// → Weather monitoring: manual sampling once per minute, all sensors O1x.
// Power consumption: 0.16µA, filter NoFilter. RMS noise: 3.3Pa / 30cm, 0.07%RH.
//
// → Humidity sensing: manual sampling once per second, pressure Off, humidity
// and temperature O1X, filter NoFilter. Power consumption: 2.9µA, 0.07%RH.
//
// → Indoor navigation: continuous sampling at 40ms with filter F16, pressure
// O16x, temperature O2x, humidity O1x, filter F16. Power consumption 633µA.
// RMS noise: 0.2Pa / 1.7cm.
//
// → Gaming: continuous sampling at 40ms with filter F16, pressure O4x,
// temperature O1x, humidity Off, filter F16. Power consumption 581µA. RMS
// noise: 0.3Pa / 2.5cm.
//
// See the datasheet for more details about the trade offs.
type Opts struct {
	Temperature Oversampling
	Pressure    Oversampling
	Humidity    Oversampling
	Filter      Filter
	Address     uint16
}

func (o *Opts) delayTypical() time.Duration {
	// Page 51.
	µs := 1000
	if o.Temperature != Off {
		µs += 2000 * o.Temperature.asValue()
	}
	if o.Pressure != Off {
		µs += 2000*o.Pressure.asValue() + 500
	}
	if o.Humidity != Off {
		µs += 2000*o.Humidity.asValue() + 500
	}
	return time.Microsecond * time.Duration(µs)
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
// It is recommended to call Halt() when done with the device so it stops
// sampling.
//
// When using SPI, the CS line must be used.
func NewSPI(p spi.Port, opts *Opts) (*Dev, error) {
	if opts != nil && opts.Address != 0 {
		return nil, errors.New("bme280: do not use Address in SPI")
	}
	// It works both in Mode0 and Mode3.
	c, err := p.Connect(10000000, spi.Mode3, 8)
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

func (d *Dev) makeDev(opts *Opts) error {
	if opts == nil {
		opts = &defaults
	}
	if opts.Temperature == Off {
		return errors.New("temperature measurement is required, use at least O1x")
	}
	d.opts = *opts
	d.measDelay = d.opts.delayTypical()

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

	// TODO(maruel): We may want to wait for isIdle().
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
	d.c = newCalibration(tph[:], h[:])

	config := []byte{
		// ctrl_meas; put it to sleep otherwise the config update may be ignored.
		// This is really just in case the device was somehow put into normal but
		// was not Halt'ed.
		0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
		// ctrl_hum
		0xF2, byte(opts.Humidity),
		// config
		0xF5, byte(s1s)<<5 | byte(NoFilter)<<2,
		// As per page 25, ctrl_meas must be re-written last.
		// ctrl_meas
		0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
	}
	if err := d.writeCommands(config[:]); err != nil {
		return err
	}
	return nil
}

// sense reads the device's registers.
//
// It must be called with d.mu lock held.
func (d *Dev) sense(env *devices.Environment) error {
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
			log.Printf("bme280: failed to sense: %v", err)
			return
		}
		sensing <- e

		select {
		case <-stop:
			return
		case <-t.C:
		}
	}
}

func (d *Dev) isIdle() (bool, error) {
	// status
	v := [1]byte{}
	if err := d.readReg(0xF3, v[:]); err != nil {
		return false, err
	}
	// Make sure bit 3 is cleared. Bit 0 is only important at device boot up.
	return v[0]&8 == 0, nil
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

//

// mode is the operating mode.
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
	Address:     0x76,
}

// standby is the time the BME280 waits idle between measurements. This reduces
// power consumption when the host won't read the values as fast as the
// measurements are done.
type standby uint8

// Possible standby values, these determines the refresh rate.
const (
	s500us standby = 0
	s10ms  standby = 6
	s20ms  standby = 7
	s62ms  standby = 1
	s125ms standby = 2
	s250ms standby = 3
	s500ms standby = 4
	s1s    standby = 5
)

func chooseStandby(d time.Duration) standby {
	switch {
	case d < 10*time.Millisecond:
		return s500us
	case d < 20*time.Millisecond:
		return s10ms
	case d < 62500*time.Microsecond:
		return s20ms
	case d < 125*time.Millisecond:
		return s62ms
	case d < 250*time.Millisecond:
		return s125ms
	case d < 500*time.Millisecond:
		return s250ms
	case d < time.Second:
		return s500ms
	default:
		return s1s
	}
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
var _ fmt.Stringer = &Dev{}
