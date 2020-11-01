// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package tlv493d

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

// I2CAddr is the default I2C address for the TLV493D component.
const I2CAddr uint16 = 0x5e

// I2CAddr1 is an alternative I2C address for TLV493D components.
const I2CAddr1 uint16 = 0x1f

// Precision represents a request for a compromise between I2C bandwidth
// versus measurement precision.
type Precision int

const (
	// HighPrecisionWithTemperature reads the full 12-bit value for each axis
	// and the temperature
	HighPrecisionWithTemperature Precision = 0
	// LowPrecision reads only 8-bits for each axis. Temperature is not read.
	LowPrecision Precision = 1
)

const (
	numberOfReadRegisters            = 10
	numberOfMeasurementRegisters     = 7
	numberOfFastMeasurementRegisters = 3
	startupDelay                     = 40 * time.Millisecond
	commandRecovery                  = 0xff

	registerBx       = 0
	registerBy       = 1
	registerBz       = 2
	registerTemp     = 3
	registerBx2      = 4
	registerBz2      = 5
	registerTemp2    = 6
	registerFactSet1 = 7
	registerFactSet2 = 8
	registerFactSet3 = 9

	bitParity                 = 7
	bitFastMode               = 1
	bitLowPowerMode           = 0
	bitLowPowerPeriod         = 6
	bitInterruptPad           = 2
	bitTemperatureMeasurement = 7
	bitParityTest             = 5

	magneticFluxScaling   = 98 * physic.MicroTesla
	temperatureScaling    = 1100 * physic.MilliCelsius
	temperatureScalingDiv = 10
	referenceTemperature  = 25*physic.Celsius + physic.ZeroCelsius
)

// Mode reprents the various power modes described in the documentation
type Mode struct {
	fastMode             bool
	lowPowerMode         bool
	lowPowerPeriod       bool
	timeToMeasure        time.Duration
	measurementFrequency physic.Frequency
}

// PowerDownMode shuts down the sensor. It can still reply to I2C commands.
var PowerDownMode = Mode{
	fastMode:             false,
	lowPowerMode:         false,
	lowPowerPeriod:       false,
	timeToMeasure:        10 * time.Millisecond,
	measurementFrequency: 1 * physic.Hertz,
}

// FastMode is the mode using the most energy
var FastMode = Mode{
	fastMode:             true,
	lowPowerMode:         false,
	lowPowerPeriod:       false,
	timeToMeasure:        0,
	measurementFrequency: 3300 * physic.Hertz,
}

// LowPowerMode uses less energy than FastMode, with lower measurement rate
var LowPowerMode = Mode{
	fastMode:             false,
	lowPowerMode:         true,
	lowPowerPeriod:       true,
	timeToMeasure:        0,
	measurementFrequency: 100 * physic.Hertz,
}

// UltraLowPowerMode saves the most energy but with very low measurement rate
var UltraLowPowerMode = Mode{
	fastMode:             false,
	lowPowerMode:         true,
	lowPowerPeriod:       false,
	timeToMeasure:        0,
	measurementFrequency: 10 * physic.Hertz,
}

// MasterControlledMode refer to TLV493D documentation on how to use this mode
var MasterControlledMode = Mode{
	fastMode:             true,
	lowPowerMode:         true,
	lowPowerPeriod:       true,
	timeToMeasure:        0,
	measurementFrequency: 3300 * physic.Hertz,
}

// Opts holds the configuration options.
type Opts struct {
	I2cAddress                    uint16
	Reset                         bool
	Mode                          Mode
	InterruptPadEnabled           bool // If enabled, this can cause I2C failures. See documentation.
	EnableTemperatureMeasurement  bool // Disable to save power.
	ParityTestEnabled             bool
	TemperatureOffsetCompensation int
}

// DefaultOpts are the recommended default options.
var DefaultOpts = Opts{
	I2cAddress:                    I2CAddr,
	Reset:                         true,
	Mode:                          PowerDownMode,
	EnableTemperatureMeasurement:  true,
	InterruptPadEnabled:           false,
	ParityTestEnabled:             true,
	TemperatureOffsetCompensation: 340, // As per the documentation, can be calibrated for better precision
}

// Sample contains the metrics measured by the sensor
type Sample struct {
	Bx          physic.MagneticFluxDensity
	By          physic.MagneticFluxDensity
	Bz          physic.MagneticFluxDensity
	Temperature physic.Temperature
}

// Dev is an handle to a TLV493D hall effect sensor.
type Dev struct {
	mu               sync.Mutex
	i2c              i2c.Dev
	stop             chan struct{}
	continuousReadWG sync.WaitGroup

	registersBuffer []byte

	mode                         Mode
	enableTemperatureMeasurement bool
	interruptPadEnabled          bool
	parityTestEnabled            bool

	temperatureOffsetCompensation int
}

// New creates a new TLV493D driver for a 3D hall effect sensors
func New(i i2c.Bus, opts *Opts) (*Dev, error) {
	switch opts.I2cAddress {
	case I2CAddr, I2CAddr1:
	default:
		return nil, errors.New("TLV493D: given address not supported by device")
	}

	d := &Dev{
		i2c:                           i2c.Dev{Bus: i, Addr: opts.I2cAddress},
		mode:                          opts.Mode,
		enableTemperatureMeasurement:  opts.EnableTemperatureMeasurement,
		interruptPadEnabled:           opts.InterruptPadEnabled,
		parityTestEnabled:             opts.ParityTestEnabled,
		temperatureOffsetCompensation: opts.TemperatureOffsetCompensation,
		registersBuffer:               make([]byte, numberOfReadRegisters),
	}
	if err := d.initialize(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// String implements conn.Resource.
func (d *Dev) String() string {
	return "TLV493D"
}

// Halt implements conn.Resource.
func (d *Dev) Halt() error {
	// Stop any continuous read
	d.StopContinousRead()

	return d.SetMode(PowerDownMode)
}

func (d *Dev) initialize(opts *Opts) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	time.Sleep(startupDelay)

	// Send recovery
	if err := d.i2c.Tx([]byte{commandRecovery}, nil); err != nil {
		return err
	}

	if opts.Reset {
		// Reset I2C address
		var resetAddress byte = 0x00
		if d.i2c.Addr == I2CAddr1 {
			resetAddress = 0xff
		}
		if err := d.i2c.Tx([]byte{resetAddress}, nil); err != nil {
			return err
		}
	}

	// Read all 10 registers and store it as the initial data
	if err := d.i2c.Tx([]byte{registerBx}, d.registersBuffer); err != nil {
		return err
	}

	// Configure sensor
	return d.configure()
}

func (d *Dev) configure() error {

	// Configure sensor
	config1 := d.registersBuffer[registerFactSet1]
	config2 := d.registersBuffer[registerFactSet3]

	// Unset parity bit first
	config1 = setBit(config1, bitParity, false)

	// Mode
	config1 = setBit(config1, bitFastMode, d.mode.fastMode)
	config1 = setBit(config1, bitLowPowerMode, d.mode.lowPowerMode)
	config2 = setBit(config2, bitLowPowerPeriod, d.mode.lowPowerPeriod)

	// Temperature: set to 0 to enable it
	config2 = setBit(config2, bitTemperatureMeasurement, !d.enableTemperatureMeasurement)

	// Other configuration bits
	config1 = setBit(config1, bitInterruptPad, d.interruptPadEnabled)
	config2 = setBit(config2, bitParityTest, d.parityTestEnabled)

	configBuffer := []byte{
		0x00,
		config1,
		d.registersBuffer[registerFactSet2],
		config2,
	}

	// Parity: the number of bits set must be odd
	// As we unset the parity bit first, if the number of bits is currently even, we have to set it
	// to make the number of bits set odd
	configBuffer[1] = setBit(config1, 7, isNumberOfBitsEven(configBuffer))

	if err := d.i2c.Tx(configBuffer, nil); err != nil {
		return fmt.Errorf("unable to read configuration: %#v", err)
	}

	return nil
}

// SetMode sets the power mode of the sensor
func (d *Dev) SetMode(mode Mode) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.mode = mode
	return d.configure()
}

// EnableTemperatureMeasurement controls the temperature sensor activation
func (d *Dev) EnableTemperatureMeasurement(enable bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.enableTemperatureMeasurement = enable
	return d.configure()
}

// EnableInterruptions controls if the sensor should send interruption of new measurement
func (d *Dev) EnableInterruptions(enable bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.interruptPadEnabled = enable
	return d.configure()
}

// EnableParityTest controls if the sensor should control the parity of the data transmitted
func (d *Dev) EnableParityTest(enable bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.parityTestEnabled = enable
	return d.configure()
}

// Read returns a sample from the last measurement of the sensor
func (d *Dev) Read(precision Precision) (Sample, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if precision == LowPrecision {
		return d.readLowPrecision()
	}

	return d.readHighPrecision()
}

func (d *Dev) readLowPrecision() (Sample, error) {
	// The information we need is in the first 3 registers
	if err := d.i2c.Tx([]byte{registerBx}, d.registersBuffer[:numberOfFastMeasurementRegisters]); err != nil {
		return Sample{}, err
	}
	buf := d.registersBuffer

	// The values are signed:
	// convert uint8 to int8, then convert to int to preserve the sign
	rawBx := int(int8(buf[registerBx])) << 4
	rawBy := int(int8(buf[registerBy])) << 4
	rawBz := int(int8(buf[registerBz])) << 4

	return Sample{
		Bx: magneticFluxScaling * physic.MagneticFluxDensity(rawBx),
		By: magneticFluxScaling * physic.MagneticFluxDensity(rawBy),
		Bz: magneticFluxScaling * physic.MagneticFluxDensity(rawBz),
	}, nil
}

func (d *Dev) readHighPrecision() (Sample, error) {
	// The information we need is in the first 7 registers
	if err := d.i2c.Tx([]byte{registerBx}, d.registersBuffer[:numberOfMeasurementRegisters]); err != nil {
		return Sample{}, err
	}
	buf := d.registersBuffer

	// The values are signed:
	// convert uint8 to int8, then convert to int to preserve the sign
	rawBx := (int(int8(buf[registerBx])) << 4) | (int(buf[registerBx2]&0xf0) >> 4)
	rawBy := (int(int8(buf[registerBy])) << 4) | int(buf[registerBx2]&0x0f)
	rawBz := (int(int8(buf[registerBz])) << 4) | int(buf[registerBz2]&0x0f)
	rawTemp := (int(int8(buf[registerTemp]&0xf0)) << 4) | int(buf[registerTemp2])

	// Compute measurement based upon reference documentation
	temp := physic.Temperature(rawTemp-d.temperatureOffsetCompensation)*temperatureScaling + referenceTemperature

	return Sample{
		Bx:          magneticFluxScaling * physic.MagneticFluxDensity(rawBx),
		By:          magneticFluxScaling * physic.MagneticFluxDensity(rawBy),
		Bz:          magneticFluxScaling * physic.MagneticFluxDensity(rawBz),
		Temperature: temp,
	}, nil
}

// ReadContinuous returns a channel which will receive readings at regular intervals
func (d *Dev) ReadContinuous(frequency physic.Frequency, precision Precision) (<-chan Sample, error) {
	// First release the current continuous reading if there is one
	d.StopContinousRead()
	reading := make(chan Sample, 16)
	d.stop = make(chan struct{})

	// Choose the best operating mode for the sensor
	newMode, err := bestModeForFrequency(frequency)
	if err != nil {
		return nil, err
	}

	previousMode := d.mode
	d.SetMode(newMode)

	t := time.NewTicker(frequency.Period())

	d.continuousReadWG.Add(1)

	go func(s <-chan struct{}) {
		defer d.SetMode(previousMode)
		defer t.Stop()
		defer close(reading)
		defer d.continuousReadWG.Done()

		for {
			select {
			case <-s:
				return
			case <-t.C:
				value, err := d.Read(precision)
				if err != nil {
					// In continuous mode, we'll ignore errors silently.
					continue
				}
				reading <- value
			}
		}
	}(d.stop)

	return reading, nil
}

// StopContinousRead stops a currently running continuous read
func (d *Dev) StopContinousRead() {
	if d.stop == nil {
		return
	}

	d.stop <- struct{}{}
	d.stop = nil
	d.continuousReadWG.Wait()
}

func bestModeForFrequency(frequency physic.Frequency) (Mode, error) {

	allowed := []*Mode{&FastMode, &LowPowerMode, &UltraLowPowerMode}
	var minAbove *Mode = nil

	for _, m := range allowed {
		if m.measurementFrequency >= frequency {
			if minAbove == nil || minAbove.measurementFrequency > m.measurementFrequency {
				minAbove = m
			}
		}
	}

	if minAbove == nil {
		return PowerDownMode, fmt.Errorf("frequency too high, no mode available for %s", frequency)
	}

	return *minAbove, nil
}

// Sets the bit at pos in the integer n.
func setBit(n byte, pos uint, isSet bool) byte {
	var bit byte = (1 << pos)
	if isSet {
		return n | bit
	}
	return n &^ bit
}

func isNumberOfBitsEven(array []byte) bool {
	var accumulator byte = 0

	// The keys is to use XOR
	// 1 ^ 1 -> 0 (even)
	// 0 ^ 1 -> 1 (odd)
	// 1 ^ 0 -> 1 (odd)
	// 0 ^ 0 -> 0 (even)

	// Combine all bytes
	for _, b := range array {
		accumulator ^= b
	}

	// Combine adjacent bits
	accumulator ^= (accumulator >> 1)
	accumulator ^= (accumulator >> 2)
	accumulator ^= (accumulator >> 4)

	// Parity is in the LSB
	return !((accumulator & 0x01) == 1)
}
