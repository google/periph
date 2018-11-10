// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ads1x15

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/conn/analog"
)

// I2CAddr is the default I2C address for the ADS1x15 components.
const I2CAddr uint16 = 0x48

// Channel is the analog reading to do. It can be either an absolute reading or
// a differential reading between two pins.
type Channel int

const (
	// Absolute reading.
	Channel0 Channel = 4
	Channel1 Channel = 5
	Channel2 Channel = 6
	Channel3 Channel = 7

	// Differential reading.
	Channel0Minus1 Channel = 0
	Channel0Minus3 Channel = 1
	Channel1Minus3 Channel = 2
	Channel2Minus3 Channel = 3
)

func (c Channel) String() string {
	switch c {
	case Channel0:
		return "0"
	case Channel1:
		return "1"
	case Channel2:
		return "2"
	case Channel3:
		return "3"
	case Channel0Minus1:
		return "0-1"
	case Channel0Minus3:
		return "0-3"
	case Channel1Minus3:
		return "1-3"
	case Channel2Minus3:
		return "2-3"
	default:
		return "Invalid"
	}
}

func (c Channel) number() int {
	switch c {
	case Channel0:
		return 0
	case Channel1:
		return 1
	case Channel2:
		return 2
	case Channel3:
		return 3
	case Channel0Minus1:
		return 4
	case Channel0Minus3:
		return 5
	case Channel1Minus3:
		return 6
	case Channel2Minus3:
		return 7
	default:
		return -1
	}
}

// ConversionQuality represents a request for a compromise between energy
// saving versus conversion quality.
type ConversionQuality int

const (
	// SaveEnergy optimizes the power consumption of the ADC, at the expense of
	// the quality by converting at the highest possible rate.
	SaveEnergy ConversionQuality = 0
	// BestQuality will use the lowest suitable data rate to reduce the impact of
	// the noise on the reading.
	BestQuality ConversionQuality = 1
)

// Opts holds the configuration options.
type Opts struct {
	I2cAddress uint16
}

// DefaultOpts are the recommended default options.
var DefaultOpts = Opts{
	I2cAddress: I2CAddr,
}

// PinADC represents a pin which is able to read an electric potential.
type PinADC interface {
	analog.PinADC
	// ReadContinuous opens a channel and reads continuously at the frequency the
	// pin was configured for.
	ReadContinuous() <-chan analog.Reading
}

// Dev is an handle to an ADS1015/ADS1115 ADC.
type Dev struct {
	c         i2c.Dev
	name      string
	dataRates map[int]uint16
	mu        sync.Mutex
}

// NewADS1015 creates a new driver for the ADS1015 (12-bit ADC).
func NewADS1015(i i2c.Bus, opts *Opts) (*Dev, error) {
	return &Dev{
		c:    i2c.Dev{Bus: i, Addr: opts.I2cAddress},
		name: "ADS1015",
		dataRates: map[int]uint16{
			128:  0x0000,
			250:  0x0020,
			490:  0x0040,
			920:  0x0060,
			1600: 0x0080,
			2400: 0x00A0,
			3300: 0x00C0,
		},
	}, nil
}

// NewADS1115 creates a new driver for the ADS1115 (16-bit ADC).
func NewADS1115(i i2c.Bus, opts *Opts) (*Dev, error) {
	return &Dev{
		c:    i2c.Dev{Bus: i, Addr: opts.I2cAddress},
		name: "ADS1115",
		dataRates: map[int]uint16{
			8:   0x0000,
			16:  0x0020,
			32:  0x0040,
			64:  0x0060,
			128: 0x0080,
			250: 0x00A0,
			475: 0x00C0,
			860: 0x00E0,
		},
	}, nil
}

// String implements conn.Resource.
func (d *Dev) String() string {
	return d.name
}

// Halt implements conn.Resource.
func (d *Dev) Halt() error {
	return nil
}

// PinForChannel returns an AnalogPin for the requested channel at the
// requested frequency.
//
// The channel can either be an absolute reading or a differential one.
func (d *Dev) PinForChannel(c Channel, maxVoltage physic.ElectricPotential, f physic.Frequency, q ConversionQuality) (PinADC, error) {
	// Determine the most appropriate gain
	gain, err := d.bestGainForElectricPotential(maxVoltage)
	if err != nil {
		return nil, err
	}

	// Validate the gain.
	gainConf, ok := gainConfig[gain]
	if !ok {
		return nil, errors.New("gain must be one of: 2/3, 1, 2, 4, 8, 16")
	}

	// Determine the voltage multiplier for this gain.
	voltageMultiplier, ok := gainVoltage[gain]
	if !ok {
		return nil, errors.New("gain must be one of: 2/3, 1, 2, 4, 8, 16")
	}

	// Determine the most appropriate data rate.
	dataRate, err := d.bestDataRateForFrequency(f, q)
	if err != nil {
		return nil, err
	}

	dataRateConf, ok := d.dataRates[dataRate]
	if !ok {
		// Write a nice error message in case the data rate is not found.
		keys := []int{}
		for k := range d.dataRates {
			keys = append(keys, k)
		}
		return nil, fmt.Errorf("invalid data rate. Accepted values: %d", keys)
	}

	// Build the configuration value
	var config uint16
	config = ads1x15ConfigOsSingle // Go out of power-down mode for conversion.
	// Specify mux value.
	config |= uint16(c) << ads1x15ConfigMuxOffset
	// Validate the passed in gain and then set it in the config.
	config |= gainConf
	// Set the mode (continuous or single shot).
	config |= ads1x15ConfigModeSingle

	// Set the data rate (this is controlled by the subclass as it differs
	// between ADS1015 and ADS1115).
	config |= dataRateConf
	config |= ads1x15ConfigCompQueDisable // Disable comparator mode.

	// Build the query to the ADC.
	configBytes := [2]byte{}
	binary.BigEndian.PutUint16(configBytes[:], config)

	// The wait for the ADC sample to finish is based on the sample rate plus a
	// small offset to be sure (0.1 millisecond).
	waitTime := time.Second/time.Duration(dataRate) + 100*time.Microsecond

	return &analogPin{
		adc:                d,
		c:                  c,
		query:              [...]byte{ads1x15PointerConfig, configBytes[0], configBytes[1]},
		voltageMultiplier:  voltageMultiplier,
		waitTime:           waitTime,
		requestedFrequency: f,
	}, nil
}

func (d *Dev) executePreparedQuery(query []byte, waitTime time.Duration, voltageMultiplier physic.ElectricPotential) (analog.Reading, error) {
	// Lock the ADC converter to avoid multiple simultaneous readings.
	d.mu.Lock()
	defer d.mu.Unlock()

	// Send the config value to start the ADC conversion.
	// Explicitly break the 16-bit value down to a big endian pair of bytes.
	if err := d.c.Tx(query, nil); err != nil {
		return analog.Reading{}, err
	}

	// Wait for the ADC sample to finish.
	time.Sleep(waitTime)

	// Retrieve the result.
	data := []byte{0, 0}
	if err := d.c.Tx([]byte{ads1x15PointerConversion}, data); err != nil {
		return analog.Reading{}, err
	}

	// Convert the raw data into physical value.
	raw := int16(binary.BigEndian.Uint16(data))
	return analog.Reading{
		Raw: int32(raw),
		V:   physic.ElectricPotential(raw) * voltageMultiplier / physic.ElectricPotential(1<<15),
	}, nil
}

// bestGainForElectricPotential returns the gain the most adapted to read up to
// the specified difference of potential.
func (d *Dev) bestGainForElectricPotential(voltage physic.ElectricPotential) (int, error) {
	var max physic.ElectricPotential
	difference := physic.ElectricPotential(math.MaxInt64)
	currentBestGain := -1

	for key, value := range gainVoltage {
		// We compute the maximum in case we need to display an error
		if value > max {
			max = value
		}
		newDiff := value - voltage
		if newDiff >= 0 && newDiff < difference {
			difference = newDiff
			currentBestGain = key
		}
	}

	if currentBestGain < 0 {
		return 0, errors.New("maximum voltage which can be read is " + max.String())
	}
	return currentBestGain, nil
}

// bestDataRateForFrequency returns the gain the most data rate to read samples
// at least at the requested frequency.
func (d *Dev) bestDataRateForFrequency(f physic.Frequency, q ConversionQuality) (int, error) {
	var max physic.Frequency
	currentBestDataRate := -1

	// In order to save energy, we are going to select the fastest conversion
	// rate, as explained in the ADS1115 specifications: 9.4.3 Duty Cycling For
	// Low Power.
	// When searching for the best quality, we will select the slowest conversion
	// rate which is still faster than the requested frequency.
	var comparator func(physic.Frequency, physic.Frequency) bool
	var difference physic.Frequency

	switch q {
	case SaveEnergy:
		// Saving energy requires the maximum data rate
		difference = physic.Frequency(-1)
		comparator = func(newDiff physic.Frequency, difference physic.Frequency) bool { return newDiff > difference }
	case BestQuality:
		// Best quality requires the minimum difference between the target and the capability
		difference = physic.Frequency(math.MaxInt64)
		comparator = func(newDiff physic.Frequency, difference physic.Frequency) bool { return newDiff < difference }
	default:
		return 0, errors.New("unknown value for ConversionQuality")
	}

	for key := range d.dataRates {
		freq := physic.Frequency(key) * physic.Hertz

		// We compute the minimum in case we need to display an error
		if freq > max {
			max = freq
		}

		newDiff := freq - f
		// Conversion rate slower than the requested frequency is not suitable
		if newDiff < 0 {
			continue
		}

		if comparator(newDiff, difference) {
			difference = newDiff
			currentBestDataRate = key
		}
	}

	if currentBestDataRate < 0 {
		return 0, errors.New("maximum frequency which can be read is " + max.String())
	}
	return currentBestDataRate, nil
}

//

const (
	ads1x15PointerConversion    = 0x00
	ads1x15PointerConfig        = 0x01
	ads1x15PointerLowThreshold  = 0x02
	ads1x15PointerHighThreshold = 0x03
	// Write: Set to start a single-conversion.
	ads1x15ConfigOsSingle       = 0x8000
	ads1x15ConfigMuxOffset      = 12
	ads1x15ConfigModeContinuous = 0x0000
	// Single shoot mode.
	ads1x15ConfigModeSingle = 0x0100

	ads1x15ConfigCompWindow      = 0x0010
	ads1x15ConfigCompAactiveHigh = 0x0008
	ads1x15ConfigCompLatching    = 0x0004
	ads1x15ConfigCompQueDisable  = 0x0003
)

var (
	// Mapping of gain values to config register values.
	gainConfig = map[int]uint16{
		2 / 3: 0x0000,
		1:     0x0200,
		2:     0x0400,
		4:     0x0600,
		8:     0x0800,
		16:    0x0A00,
	}
	gainVoltage = map[int]physic.ElectricPotential{
		2 / 3: 6144 * physic.MilliVolt,
		1:     4096 * physic.MilliVolt,
		2:     2048 * physic.MilliVolt,
		4:     1024 * physic.MilliVolt,
		8:     512 * physic.MilliVolt,
		16:    256 * physic.MilliVolt,
	}
)

type analogPin struct {
	adc                *Dev
	c                  Channel
	query              [3]byte
	voltageMultiplier  physic.ElectricPotential
	waitTime           time.Duration
	requestedFrequency physic.Frequency
	stop               chan struct{}
	mu                 sync.Mutex
}

// Range returns the maximum supported range [min, max] of the values.
func (p *analogPin) Range() (analog.Reading, analog.Reading) {
	max := analog.Reading{Raw: math.MaxInt16, V: p.voltageMultiplier}
	min := analog.Reading{Raw: -math.MaxInt16, V: -p.voltageMultiplier}
	return min, max
}

// Read returns the current pin level.
func (p *analogPin) Read() (analog.Reading, error) {
	return p.adc.executePreparedQuery(p.query[:], p.waitTime, p.voltageMultiplier)
}

func (p *analogPin) ReadContinuous() <-chan analog.Reading {
	// We need to lock if there are multiple Halt or ReadContinuous
	// calls simultaneously.
	p.mu.Lock()
	defer p.mu.Unlock()

	// First release the current continuous reading if there is one
	if p.stop != nil {
		close(p.stop)
		p.stop = nil
	}
	reading := make(chan analog.Reading)
	p.stop = make(chan struct{})

	go func() {
		t := time.NewTicker(p.requestedFrequency.Duration())
		defer t.Stop()
		defer p.Halt()
		defer close(reading)

		for {
			select {
			case <-p.stop:
				return
			case <-t.C:
				value, err := p.Read()
				if err != nil {
					// In continous mode, we'll ignore errors silently.
					continue
				}
				reading <- value
			}
		}
	}()

	return reading
}

func (p *analogPin) Name() string {
	return p.adc.name + "(" + p.c.String() + ")"
}

func (p *analogPin) Number() int {
	return p.c.number()
}

func (p *analogPin) Function() string {
	return "ADC"
}

func (p *analogPin) Halt() error {
	// We need to lock if there are multiple Halt or ReadContinuous
	// calls simultaneously.
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stop != nil {
		close(p.stop)
		p.stop = nil
	}
	return nil
}

func (p *analogPin) String() string {
	return p.Name()
}
