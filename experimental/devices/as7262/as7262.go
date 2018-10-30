// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package as7262

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

// Opts holds the configuration options.
type Opts struct {
	InterruptPin gpio.PinIn
	Gain         Gain
}

// DefaultOpts are the recommended default options.
var DefaultOpts = Opts{
	InterruptPin: nil,
	Gain:         G1x,
}

// New opens a handle to an AS7262 sensor.
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	<-ctx.Done()
	return &Dev{
		c:         &i2c.Dev{Bus: bus, Addr: 0x49},
		gain:      opts.Gain,
		interrupt: opts.InterruptPin,
		order:     binary.BigEndian,
		timeout:   200 * time.Millisecond,
		cancel:    cancel,
		ctx:       ctx,
	}, nil
}

// Dev is a handle to the as7262 sensor.
type Dev struct {
	mu sync.Mutex

	c         conn.Conn
	timeout   time.Duration
	interrupt gpio.PinIn
	cancel    context.CancelFunc
	ctx       context.Context
	gain      Gain
	order     binary.ByteOrder
}

// Spectrum is the reading from the senor including the actual sensor state for
// the readings.
type Spectrum struct {
	Bands             []Band
	SensorTemperature physic.Temperature
	Gain              Gain
	LedDrive          physic.ElectricCurrent
	Integration       time.Duration
}

// Band has two types of measurement of relative spectral flux density.
//
// Value
//
// Value are the calibrated readings. The accuracy of the channel counts/μW/cm2
// is ±12%.
//
// Counts
//
// Counts are the raw readings, there are approximately 45 counts/μW/cm2 with a
// gain of 16 (Gx16).
//
// Wavelength
//
// Wavelength is the nominal center of a band, with a ±40nm bandwidth around the
// center. Wavelengths for the as7262 are: 450nm, 500nm, 550nm, 570nm, 600nm and
// 650nm.
type Band struct {
	Wavelength physic.Distance
	Value      float64
	Counts     uint16
	Name       string
}

// Sense preforms a reading of relative spectral radiance of all the sensor
// bands.
//
// Led Drive Current
//
// The AS7262 provides a current limated intergated led drive circuit. Valid
// limits for the drive current are 0mA, 12.5mA, 25mA, 50mA and 100mA. If non
// valid values are given the next lowest valid value is used.
//
// Resolution
//
// For best resolution it is recommended that for a specific led drive
// current that the senseTime or gain is increased until at least one of the
// bands returns a count above 10000. The maximum senseTime time is 714ms
// senseTime will be quantised into intervals of of 2.8ms. Actual time taken to
// make a reading is twice the senseTime.
func (d *Dev) Sense(ledDrive physic.ElectricCurrent, senseTime time.Duration) (Spectrum, error) {

	d.mu.Lock()
	defer d.mu.Unlock()

	d.ctx, d.cancel = context.WithCancel(context.Background())
	defer d.cancel()

	it, integration := calcSenseTime(senseTime)
	if err := d.writeVirtualRegister(intergrationReg, it); err != nil {
		return Spectrum{}, err
	}
	led, drive := calcLed(ledDrive)

	if err := d.writeVirtualRegister(ledControlReg, led); err != nil {
		return Spectrum{}, err
	}

	if err := d.writeVirtualRegister(controlReg, uint8(allOneShot)|uint8(d.gain)); err != nil {
		return Spectrum{}, err
	}

	if d.interrupt != nil {
		isEdge := make(chan bool)
		go func() {
			isEdge <- d.interrupt.WaitForEdge(integration*2 + d.timeout)
		}()
		select {
		case edge := <-isEdge:
			if !edge {
				return Spectrum{}, errTimeoutPin
			}
		case <-d.ctx.Done():
			return Spectrum{}, errHalted
		}
	} else {
		select {
		// WaitForSensor is time.After().
		case <-waitForSensor(integration * 2):
			if err := d.pollDataReady(); err != nil {
				return Spectrum{}, err
			}
		case <-d.ctx.Done():
			return Spectrum{}, errHalted
		}

	}

	if err := d.writeVirtualRegister(ledControlReg, 0x00); err != nil {
		return Spectrum{}, err
	}

	raw := make([]byte, 12)
	if err := d.readVirtualRegister(rawBase, raw); err != nil {
		return Spectrum{}, err
	}

	cal := make([]byte, 24)
	if err := d.readVirtualRegister(calBase, cal); err != nil {
		return Spectrum{}, err
	}

	v := d.order.Uint16(raw[0:2])
	b := d.order.Uint16(raw[2:4])
	g := d.order.Uint16(raw[4:6])
	y := d.order.Uint16(raw[6:8])
	o := d.order.Uint16(raw[8:10])
	r := d.order.Uint16(raw[10:12])

	vcal := float64(math.Float32frombits(d.order.Uint32(cal[0:4])))
	bcal := float64(math.Float32frombits(d.order.Uint32(cal[4:8])))
	gcal := float64(math.Float32frombits(d.order.Uint32(cal[8:12])))
	ycal := float64(math.Float32frombits(d.order.Uint32(cal[12:16])))
	ocal := float64(math.Float32frombits(d.order.Uint32(cal[16:20])))
	rcal := float64(math.Float32frombits(d.order.Uint32(cal[20:24])))

	traw := make([]byte, 1)
	if err := d.readVirtualRegister(deviceTemperatureReg, traw); err != nil {
		return Spectrum{}, err
	}
	temperature := physic.Temperature((int8(traw[0])))*physic.Kelvin + physic.ZeroCelsius
	return Spectrum{
		Bands: []Band{
			Band{Wavelength: 450 * physic.NanoMetre, Counts: v, Value: vcal, Name: "V"},
			Band{Wavelength: 500 * physic.NanoMetre, Counts: b, Value: bcal, Name: "B"},
			Band{Wavelength: 550 * physic.NanoMetre, Counts: g, Value: gcal, Name: "G"},
			Band{Wavelength: 570 * physic.NanoMetre, Counts: y, Value: ycal, Name: "Y"},
			Band{Wavelength: 600 * physic.NanoMetre, Counts: o, Value: ocal, Name: "O"},
			Band{Wavelength: 650 * physic.NanoMetre, Counts: r, Value: rcal, Name: "R"},
		},
		SensorTemperature: temperature,
		Gain:              d.gain,
		LedDrive:          drive,
		Integration:       integration,
	}, nil
}

var waitForSensor = time.After

// Halt stops any pending operations
func (d *Dev) Halt() error {
	d.cancel()
	return nil
}

// String implaments the stringer interface
func (d *Dev) String() string {
	return fmt.Sprintf("AMS AS7262 6 channel visible spectrum sensor")
}

// Gain is the sensor gain for all bands
type Gain int

const (
	// Gain1x is gain of 1
	G1x Gain = 0x00
	// Gain4x is gain of 3.7
	G4x Gain = 0x10
	// Gain16x is a gain of 16
	G16x Gain = 0x20
	// Gain64x us a gain of 64
	G64x Gain = 0x30
)

// Gain sets the gain of the sensor.
func (d *Dev) Gain(gain Gain) error {

	d.mu.Lock()
	defer d.mu.Unlock()

	d.ctx, d.cancel = context.WithCancel(context.Background())
	defer d.cancel()

	if err := d.writeVirtualRegister(controlReg, uint8(gain)); err != nil {
		return err
	}
	d.gain = gain
	return nil
}

func (d *Dev) writeVirtualRegister(register, data byte) error {

	// Check for pending writes.
	if err := d.pollStatus(writing); err != nil {
		return err
	}

	// Set virtual register that is being written to.
	if err := d.c.Tx([]byte{writeReg, register | 0x80}, nil); err != nil {
		return &IOError{"setting virtual register", err}
	}

	// Check for pending writes again.
	if err := d.pollStatus(writing); err != nil {
		return err
	}

	// Write data to register that is being written to.
	if err := d.c.Tx([]byte{writeReg, data}, nil); err != nil {
		return &IOError{"writing virtual register", err}
	}

	return nil

}

func (d *Dev) readVirtualRegister(register byte, data []byte) error {
	rx := make([]byte, 1)
	for i := 0; i < len(data); i++ {
		// Check for pending reads.
		if err := d.pollStatus(clearBuffer); err != nil {
			return err
		}

		// Set virtual register that is being read from plus offset.
		if err := d.c.Tx([]byte{writeReg, register + byte(i)}, nil); err != nil {
			return &IOError{"setting virtual register", err}
		}

		// Check if read buffer is ready.
		if err := d.pollStatus(reading); err != nil {
			return err
		}

		// Read byte from register that is being read from into our buffer with
		// offset.
		if err := d.c.Tx([]byte{readReg}, rx); err != nil {
			return &IOError{"reading virtual register", err}
		}
		data[i] = rx[0]
	}
	return nil
}

func (d *Dev) pollDataReady() error {
	pollctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	for {
		if err := d.pollStatus(clearBuffer); err != nil {
			return err
		}

		// Set virtual register that is being read from plus offset.
		if err := d.c.Tx([]byte{writeReg, controlReg}, nil); err != nil {
			return &IOError{"setting virtual register", err}
		}

		// Check if read buffer is ready.
		if err := d.pollStatus(reading); err != nil {
			return err
		}

		// Read byte from register that is being read from into our buffer with
		// offset.
		data := make([]byte, 1)
		if err := d.c.Tx([]byte{readReg}, data); err != nil {
			return &IOError{"reading virtual register", err}
		}
		if data[0]&0x02 > 0 {
			return nil
		}
		select {
		case <-time.After(5 * time.Millisecond):
			// Polling interval.
		case <-pollctx.Done():
			// Return error if it takes too long.
			return errStatusDeadline
		case <-d.ctx.Done():
			return errHalted
		}
	}

}

type direction byte

const (
	// Reading is a bit mask for the status register.
	reading direction = 1
	// Writing is a bit mask for the status register.
	writing direction = 2
	// ClearBuffer clears any data left in the buffer and then checks the reading
	clearBuffer direction = 3
)

// The as7262 registers are implemented as virtual registers pollStatus
// provides a way to repeatedly check if there are any pending reads or writes
// in the relevent buffer before a transaction while with a timeout of 200
// milliseconds. Direction is used to set which buffer is being polled to be
// ready.
func (d *Dev) pollStatus(dir direction) error {
	pollctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	// Check if already canceled first
	select {
	case <-d.ctx.Done():
		return errHalted
	default:
		// Proceed.
	}
	status := make([]byte, 1)
	for {

		// Read status register.
		err := d.c.Tx([]byte{statusReg}, status)
		if err != nil {
			return &IOError{"reading status register", err}
		}

		switch dir {
		case reading:
			// Bit 0: rx valid bit.
			//    0 → No data is ready to be read in READ register.
			//    1 → Data byte available in READ register.
			if status[0]&byte(dir) == 1 {
				return nil
			}
		case writing:
			// Bit 1: tx valid bit.
			//    0 → New data may be written to WRITE register.
			//    1 → WRITE register occupied. Do NOT write.
			if status[0]&byte(dir) == 0 {
				return nil
			}
		case clearBuffer:
			// If there is data left in the buffer read it.
			if status[0]&0x01 == 1 {
				discard := make([]byte, 1)
				if err := d.c.Tx([]byte{readReg}, discard); err != nil {
					return &IOError{"clearing buffer", err}
				}
			}
			if status[0]&0x01 == 0 {
				return nil
			}
		}
		select {
		case <-time.After(5 * time.Millisecond):
			// Polling interval.
		case <-pollctx.Done():
			// Return error if it takes too long.
			return errStatusDeadline
		case <-d.ctx.Done():
			return errHalted
		}
	}
}

const (
	maxSenseTime time.Duration = 714 * time.Millisecond
	minSenseTime               = 2800 * time.Microsecond
)

// calculateIntergrationTime converts a time.Duration into a value between 0 and
// 256
func calcSenseTime(t time.Duration) (uint8, time.Duration) {
	if t > maxSenseTime {
		return 255, maxSenseTime
	}
	if t < minSenseTime {
		return 1, minSenseTime
	}
	// Minimum step is 2.8ms
	quantizedTime := (t / minSenseTime)
	return uint8(quantizedTime), quantizedTime * minSenseTime
}

func calcLed(drive physic.ElectricCurrent) (uint8, physic.ElectricCurrent) {
	switch {
	case drive < 12500*physic.MicroAmpere:
		return 0x00, 0
	case drive >= 12500*physic.MicroAmpere && drive < 25*physic.MilliAmpere:
		return 0x08, 12500 * physic.MicroAmpere
	case drive >= 25*physic.MilliAmpere && drive < 50*physic.MilliAmpere:
		return 0x18, 25 * physic.MilliAmpere
	case drive >= 50*physic.MilliAmpere && drive < 100*physic.MilliAmpere:
		return 0x28, 50 * physic.MilliAmpere
	default:
		return 0x38, 100 * physic.MilliAmpere
	}
}

type mode uint8

const (
	// Bank 1 consists of data from the V, G, B, Y photodiodes.
	bank1 mode = 0x00
	// Bank 2 consists of data from the G, Y, O, R photodiodes.
	bank2 mode = 0x04
	// AllContinuously gets data from both banks continuously, requires 2x
	// the intergration time.
	allContinuously mode = 0x08
	// AllOneShot gets data from both banks once, and set the data ready bit in
	// the status control register when complete requires 2x the intergration
	// time.
	allOneShot mode = 0x0c
)

// IOError is a I/O specific error.
type IOError struct {
	Op  string
	Err error
}

// Error implements the Error interface.
func (e *IOError) Error() string {
	if e.Err != nil {
		return "ioerror while " + e.Op + ": " + e.Err.Error()
	}
	return "ioerror while " + e.Op
}

var (
	errStatusDeadline  = errors.New("deadline exceeded reading status register")
	errTimeoutPin      = errors.New("timeout waiting for interrupt signal on pin")
	errHalted          = errors.New("recived halt command")
	errHaltTimeExeeded = errors.New("halt timeout")
	errNothingToHalt   = error(nil)
)

const (
	statusReg            = 0x00
	writeReg             = 0x01
	readReg              = 0x02
	hardwareVersion      = 0x00
	firmwareVersion      = 0x02
	controlReg           = 0x04
	intergrationReg      = 0x05
	deviceTemperatureReg = 0x06
	ledControlReg        = 0x07
	rawBase              = 0x08
	rawVReg              = 0x08
	rawBReg              = 0x0a
	rawGReg              = 0x0c
	rawYReg              = 0x0e
	rawOReg              = 0x10
	rawRReg              = 0x12
	calBase              = 0x14
	calibratedVReg       = 0x14
	calibratedBReg       = 0x18
	calibratedGReg       = 0x1c
	calibratedYReg       = 0x20
	calibratedOReg       = 0x24
	calibratedRReg       = 0x28
)
