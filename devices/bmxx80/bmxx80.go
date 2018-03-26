// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bmxx80 controls a Bosch BMP180/BME280/BMP280 device over I²C, or SPI
// for the BMx280.
//
// BMx280
//
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280_DS001-11.pdf
//
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BMP280-DS001-18.pdf
//
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BMP180-DS000-121.pdf
//
// The font the official datasheet on page 15 is hard to read, a copy with
// readable text can be found here:
//
// https://cdn-shop.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf
//
// Notes on the BMP180 datasheet
//
// The results of the calculations in the algorithm on page 15 are partly
// wrong. It looks like the original authors used non-integer calculations and
// some nubers were rounded. Take the results of the calculations with a grain
// of salt.
package bmxx80

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

// Oversampling affects how much time is taken to measure each of temperature,
// pressure and humidity.
//
// Using high oversampling and low standby results in highest power
// consumption, but this is still below 1mA so we generally don't care.
type Oversampling uint8

// Possible oversampling values.
//
// The higher the more time and power it takes to take a measurement. Even at
// 16x for all 3 sensors, it is less than 100ms albeit increased power
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

func (o Oversampling) to180() uint8 {
	switch o {
	default:
		fallthrough
	case Off, O1x:
		return 0
	case O2x:
		return 1
	case O4x:
		return 2
	case O8x, O16x:
		return 3
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

// DefaultsOpts returns the default options used.
//
// Defaults to use O4x (4x oversampling) for all measurements.
var DefaultOpts = Opts{
	Temperature: O4x,
	Pressure:    O4x,
	Humidity:    O4x,
}

// Opts defines the options for the device.
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
	// Temperature can only be oversampled on BME280/BMP280.
	//
	// Temperature must be measured for pressure and humidity to be measured.
	Temperature Oversampling
	// Pressure can be oversampled up to 8x on BMP180 and 16x on BME280/BMP280.
	Pressure Oversampling
	// Humidity sensing is only supported on BME280. The value is ignored on other
	// devices.
	Humidity Oversampling
	// Filter is only used while using SenseContinuous() and is only supported on
	// BMx280.
	Filter Filter
}

func (o *Opts) delayTypical280() time.Duration {
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

// NewI2C returns an object that communicates over I²C to BMP180/BME280/BMP280
// environmental sensor.
//
// The address must be 0x76 or 0x77. BMP180 uses 0x77. BME280/BMP280 default to
// 0x76 and can optionally use 0x77. The value used depends on HW
// configuration of the sensor's SDO pin.
//
// It is recommended to call Halt() when done with the device so it stops
// sampling.
func NewI2C(b i2c.Bus, addr uint16, opts *Opts) (*Dev, error) {
	switch addr {
	case 0x76, 0x77:
	default:
		return nil, errors.New("bmxx80: given address not supported by device")
	}
	d := &Dev{d: &i2c.Dev{Bus: b, Addr: addr}, isSPI: false}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// NewSPI returns an object that communicates over SPI to either a BME280 or
// BMP280 environmental sensor.
//
// It is recommended to call Halt() when done with the device so it stops
// sampling.
//
// When using SPI, the CS line must be used.
func NewSPI(p spi.Port, opts *Opts) (*Dev, error) {
	// It works both in Mode0 and Mode3.
	c, err := p.Connect(10000000, spi.Mode3, 8)
	if err != nil {
		return nil, fmt.Errorf("bmxx80: %v", err)
	}
	d := &Dev{d: c, isSPI: true}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// Dev is a handle to an initialized BMxx80 device.
//
// The actual device type was auto detected.
type Dev struct {
	d         conn.Conn
	isSPI     bool
	is280     bool
	isBME     bool
	opts      Opts
	measDelay time.Duration
	name      string
	os        uint8
	cal180    calibration180
	cal280    calibration280

	mu   sync.Mutex
	stop chan struct{}
	wg   sync.WaitGroup
}

func (d *Dev) String() string {
	// d.dev.Conn
	return fmt.Sprintf("%s{%s}", d.name, d.d)
}

// Sense requests a one time measurement as °C, kPa and % of relative humidity.
//
// The very first measurements may be of poor quality.
func (d *Dev) Sense(e *physic.Env) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		return d.wrap(errors.New("already sensing continuously"))
	}

	if d.is280 {
		err := d.writeCommands([]byte{
			// ctrl_meas
			0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(forced),
		})
		if err != nil {
			return d.wrap(err)
		}
		doSleep(d.measDelay)
		for idle := false; !idle; {
			if idle, err = d.isIdle280(); err != nil {
				return d.wrap(err)
			}
		}
		return d.sense280(e)
	}
	return d.sense180(e)
}

// SenseContinuous returns measurements as °C, kPa and % of relative humidity
// on a continuous basis.
//
// The application must call Halt() to stop the sensing when done to stop the
// sensor and close the channel.
//
// It's the responsibility of the caller to retrieve the values from the
// channel as fast as possible, otherwise the interval may not be respected.
func (d *Dev) SenseContinuous(interval time.Duration) (<-chan physic.Env, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		// Don't send the stop command to the device.
		close(d.stop)
		d.stop = nil
		d.wg.Wait()
	}

	if d.is280 {
		s := chooseStandby(d.isBME, interval-d.measDelay)
		err := d.writeCommands([]byte{
			// config
			0xF5, byte(s)<<5 | byte(d.opts.Filter)<<2,
			// ctrl_meas
			0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(normal),
		})
		if err != nil {
			return nil, d.wrap(err)
		}
	}

	sensing := make(chan physic.Env)
	d.stop = make(chan struct{})
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		defer close(sensing)
		d.sensingContinuous(interval, sensing, d.stop)
	}()
	return sensing, nil
}

// Halt stops the BMxx80 from acquiring measurements as initiated by
// SenseContinuous().
//
// It is recommended to call this function before terminating the process to
// reduce idle power usage and a goroutine leak.
func (d *Dev) Halt() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop == nil {
		return nil
	}
	close(d.stop)
	d.stop = nil
	d.wg.Wait()

	if d.is280 {
		// Page 27 (for register) and 12~13 section 3.3.
		return d.writeCommands([]byte{
			// config
			0xF5, byte(s1s)<<5 | byte(NoFilter)<<2,
			// ctrl_meas
			0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
		})
	}
	return nil
}

//

func (d *Dev) makeDev(opts *Opts) error {
	d.opts = *opts
	d.measDelay = d.opts.delayTypical280()

	// The device starts in 2ms as per datasheet. No need to wait for boot to be
	// finished.

	var chipID [1]byte
	// Read register 0xD0 to read the chip id.
	if err := d.readReg(0xD0, chipID[:]); err != nil {
		return err
	}
	switch chipID[0] {
	case 0x55:
		d.name = "BMP180"
		d.os = opts.Pressure.to180()
	case 0x58:
		d.name = "BMP280"
		d.is280 = true
		d.opts.Humidity = Off
	case 0x60:
		d.name = "BME280"
		d.is280 = true
		d.isBME = true
	default:
		return fmt.Errorf("bmxx80: unexpected chip id %x", chipID[0])
	}

	if d.is280 && opts.Temperature == Off {
		// Ignore the value for BMP180, since it's not controllable.
		return d.wrap(errors.New("temperature measurement is required, use at least O1x"))
	}

	if d.is280 {
		// TODO(maruel): We may want to wait for isIdle280().
		// Read calibration data t1~3, p1~9, 8bits padding, h1.
		var tph [0xA2 - 0x88]byte
		if err := d.readReg(0x88, tph[:]); err != nil {
			return err
		}
		// Read calibration data h2~6
		var h [0xE8 - 0xE1]byte
		if d.isBME {
			if err := d.readReg(0xE1, h[:]); err != nil {
				return err
			}
		}
		d.cal280 = newCalibration(tph[:], h[:])
		var b []byte
		if d.isBME {
			b = []byte{
				// ctrl_meas; put it to sleep otherwise the config update may be
				// ignored. This is really just in case the device was somehow put
				// into normal but was not Halt'ed.
				0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
				// ctrl_hum
				0xF2, byte(d.opts.Humidity),
				// config
				0xF5, byte(s1s)<<5 | byte(NoFilter)<<2,
				// As per page 25, ctrl_meas must be re-written last.
				0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
			}
		} else {
			// BMP280 doesn't have humidity to control.
			b = []byte{
				// ctrl_meas; put it to sleep otherwise the config update may be
				// ignored. This is really just in case the device was somehow put
				// into normal but was not Halt'ed.
				0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
				// config
				0xF5, byte(s1s)<<5 | byte(NoFilter)<<2,
				// As per page 25, ctrl_meas must be re-written last.
				0xF4, byte(d.opts.Temperature)<<5 | byte(d.opts.Pressure)<<2 | byte(sleep),
			}
		}
		return d.writeCommands(b)
	}
	// Read calibration data.
	dev := mmr.Dev8{Conn: d.d, Order: binary.BigEndian}
	if err := dev.ReadStruct(0xAA, &d.cal180); err != nil {
		return d.wrap(err)
	}
	if !d.cal180.isValid() {
		return d.wrap(errors.New("calibration data is invalid"))
	}
	return nil
}

func (d *Dev) sensingContinuous(interval time.Duration, sensing chan<- physic.Env, stop <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()

	var err error
	for {
		// Do one initial sensing right away.
		e := physic.Env{}
		d.mu.Lock()
		if d.is280 {
			err = d.sense280(&e)
		} else {
			err = d.sense180(&e)
		}
		d.mu.Unlock()
		if err != nil {
			log.Printf("%s: failed to sense: %v", d, err)
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

func (d *Dev) readReg(reg uint8, b []byte) error {
	// Page 32-33
	if d.isSPI {
		// MSB is 0 for write and 1 for read.
		read := make([]byte, len(b)+1)
		write := make([]byte, len(read))
		// Rest of the write buffer is ignored.
		write[0] = reg
		if err := d.d.Tx(write, read); err != nil {
			return d.wrap(err)
		}
		copy(b, read[1:])
		return nil
	}
	if err := d.d.Tx([]byte{reg}, b); err != nil {
		return d.wrap(err)
	}
	return nil
}

// writeCommands writes a command to the device.
//
// Warning: b may be modified!
func (d *Dev) writeCommands(b []byte) error {
	if d.isSPI {
		// Page 33; set RW bit 7 to 0.
		for i := 0; i < len(b); i += 2 {
			b[i] &^= 0x80
		}
	}
	if err := d.d.Tx(b, nil); err != nil {
		return d.wrap(err)
	}
	return nil
}

func (d *Dev) wrap(err error) error {
	return fmt.Errorf("%s: %v", strings.ToLower(d.name), err)
}

var doSleep = time.Sleep

var _ conn.Resource = &Dev{}
var _ physic.SenseEnv = &Dev{}
var _ fmt.Stringer = &Dev{}
