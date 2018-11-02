package ads1x15

import (
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"time"

	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"

	"fmt"

	"periph.io/x/periph/conn/i2c"
)

const (

	// ADS1x15DefaultAddress is the default I2C address for the ADS1x15 components
	ADS1x15DefaultAddress = 0x48

	ads1x15PointerConversion    = 0x00
	ads1x15PointerConfig        = 0x01
	ads1x15PointerLowThreshold  = 0x02
	ads1x15PointerHighThreshold = 0x03
	// Write: Set to start a single-conversion
	ads1x15ConfigOsSingle       = 0x8000
	ads1x15ConfigMuxOffset      = 12
	ads1x15ConfigModeContinuous = 0x0000
	//Single shoot mode
	ads1x15ConfigModeSingle = 0x0100

	ads1x15ConfigCompWindow      = 0x0010
	ads1x15ConfigCompAactiveHigh = 0x0008
	ads1x15ConfigCompLatching    = 0x0004
	ads1x15ConfigCompQueDisable  = 0x0003

	Channel0 = 0
	Channel1 = 1
	Channel2 = 2
	Channel3 = 3
)

// Dev is the driver for the ADS1015/ADS1115 ADC
type Dev struct {
	// I2C Communication
	c i2c.Dev

	name string

	gainConfig      map[int]uint16
	dataRates       map[int]uint16
	gainVoltage     map[int]physic.ElectricPotential
	DefaultDataRate int
	mutex           *sync.Mutex
}

// AnalogPin represents a pin which is able to read an electric potential
type AnalogPin interface {
	pin.Pin
	Read() (physic.ElectricPotential, error)
}

type ads1x15AnalogPin struct {
	adc               *Dev
	preparedQuery     []byte
	voltageMultiplier physic.ElectricPotential
	waitTime          time.Duration
}

// NewADS1015 creates a new driver for the ADS1015 (12-bit ADC)
// Largely inspired by: https://github.com/adafruit/Adafruit_Python_ADS1x15
func NewADS1015(i i2c.Bus, options ...func(*Dev)) (l *Dev, err error) {
	l, err = newADS1x15(i, options...)

	l.dataRates = map[int]uint16{
		128:  0x0000,
		250:  0x0020,
		490:  0x0040,
		920:  0x0060,
		1600: 0x0080,
		2400: 0x00A0,
		3300: 0x00C0,
	}
	if l.DefaultDataRate == 0 {
		l.DefaultDataRate = 1600
	}

	l.name = "ADS1015"

	return
}

// NewADS1115 creates a new driver for the ADS1115 (16-bit ADC)
func NewADS1115(i i2c.Bus, options ...func(*Dev)) (l *Dev, err error) {
	l, err = newADS1x15(i, options...)

	l.dataRates = map[int]uint16{
		8:   0x0000,
		16:  0x0020,
		32:  0x0040,
		64:  0x0060,
		128: 0x0080,
		250: 0x00A0,
		475: 0x00C0,
		860: 0x00E0,
	}

	if l.DefaultDataRate == 0 {
		l.DefaultDataRate = 128
	}

	l.name = "ADS1115"

	return
}

func newADS1x15(i i2c.Bus, options ...func(*Dev)) (l *Dev, err error) {
	l = &Dev{
		c: i2c.Dev{Bus: i, Addr: ADS1x15DefaultAddress},
		// Mapping of gain values to config register values.
		gainConfig: map[int]uint16{
			2 / 3: 0x0000,
			1:     0x0200,
			2:     0x0400,
			4:     0x0600,
			8:     0x0800,
			16:    0x0A00,
		},
		gainVoltage: map[int]physic.ElectricPotential{
			2 / 3: 6144 * physic.MilliVolt,
			1:     4096 * physic.MilliVolt,
			2:     2048 * physic.MilliVolt,
			4:     1024 * physic.MilliVolt,
			8:     512 * physic.MilliVolt,
			16:    256 * physic.MilliVolt,
		},
		mutex: &sync.Mutex{},
	}

	for _, option := range options {
		option(l)
	}

	return
}

// Halt returns true if devices is halted successfully
func (d *Dev) Halt() (err error) { return }

// WithI2CAddress option sets the ADS1x15Driver I2C address, if it is not the default.
func WithI2CAddress(i2cAddress uint16) func(*Dev) {
	return func(d *Dev) {
		d.c.Addr = i2cAddress
	}
}

func (d *Dev) PinForChannel(channel int, maxVoltage physic.ElectricPotential, minimumFrequency physic.Frequency) (pin AnalogPin, err error) {
	if err = d.checkChannel(channel); err != nil {
		return
	}
	mux := channel + 0x04

	return d.prepareQuery(mux, maxVoltage, minimumFrequency)
}

// PinForDifferenceOfChannels reads the difference in volts between 2 inputs: channelA - channelB.
// diff can be:
// * Channel 0 - channel 1
// * Channel 0 - channel 3
// * Channel 1 - channel 3
// * Channel 2 - channel 3
func (d *Dev) PinForDifferenceOfChannels(channelA int, channelB int, maxVoltage physic.ElectricPotential, minimumFrequency physic.Frequency) (pin AnalogPin, err error) {
	var mux int

	if err = d.checkChannel(channelA); err != nil {
		return
	}
	if err = d.checkChannel(channelB); err != nil {
		return
	}

	if channelA == Channel0 && channelB == Channel1 {
		mux = 0
	} else if channelA == Channel0 && channelB == Channel3 {
		mux = 1
	} else if channelA == Channel1 && channelB == Channel3 {
		mux = 2
	} else if channelA == Channel2 && channelB == Channel3 {
		mux = 3
	} else {
		err = errors.New("Only some differences of channels are allowed:  0 - 1, 0 - 3, 1 - 3 or 2 - 3")
		return
	}

	return d.prepareQuery(mux, maxVoltage, minimumFrequency)
}

func (d *Dev) prepareQuery(mux int, maxVoltage physic.ElectricPotential, minimumFrequency physic.Frequency) (pin AnalogPin, err error) {
	// Determine the most appropriate gain
	gain, err := d.bestGainForElectricPotential(maxVoltage)
	if err != nil {
		return
	}

	// Validate the gain.
	gainConf, ok := d.gainConfig[gain]
	if !ok {
		err = errors.New("Gain must be one of: 2/3, 1, 2, 4, 8, 16")
		return
	}

	// Determine the voltage multiplier for this gain
	voltageMultiplier, ok := d.gainVoltage[gain]
	if !ok {
		err = errors.New("Gain must be one of: 2/3, 1, 2, 4, 8, 16")
		return
	}

	// Determine the most appropriate data rate
	dataRate, err := d.bestDataRateForFrequency(minimumFrequency)
	if err != nil {
		return
	}

	dataRateConf, ok := d.dataRates[dataRate]

	if !ok {
		// Write a nice error message in case the data rate is not found
		keys := []int{}
		for k := range d.dataRates {
			keys = append(keys, k)
		}

		err = fmt.Errorf("Invalid data rate. Accepted values: %d", keys)
		return
	}

	// Build the configuration value
	var config uint16
	config = ads1x15ConfigOsSingle // Go out of power-down mode for conversion.
	// Specify mux value.
	config |= uint16((mux & 0x07) << ads1x15ConfigMuxOffset)
	// Validate the passed in gain and then set it in the config.
	config |= gainConf
	// Set the mode (continuous or single shot).
	config |= ads1x15ConfigModeSingle

	// Set the data rate (this is controlled by the subclass as it differs
	// between ADS1015 and ADS1115).
	config |= dataRateConf
	config |= ads1x15ConfigCompQueDisable // Disable comparator mode.

	// Build the query to the ADC
	configBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(configBytes, config)
	preparedQuery := append([]byte{ads1x15PointerConfig}, configBytes...)

	// The wait for the ADC sample to finish is based on the sample rate plus a
	// small offset to be sure (0.1 millisecond).
	waitTime := time.Duration(1000000/dataRate+100) * time.Microsecond

	pin = &ads1x15AnalogPin{
		adc:               d,
		preparedQuery:     preparedQuery,
		voltageMultiplier: voltageMultiplier,
		waitTime:          waitTime,
	}

	return
}

func (d *Dev) executePreparedQuery(query []byte, waitTime time.Duration, voltageMultiplier physic.ElectricPotential) (value physic.ElectricPotential, err error) {
	// Lock the ADC converter to avoid multiple simultaneous readings.
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Send the config value to start the ADC conversion.
	// Explicitly break the 16-bit value down to a big endian pair of bytes.
	if err = d.c.Tx(query, nil); err != nil {
		return
	}

	// Wait for the ADC sample to finish.
	time.Sleep(waitTime)

	// Retrieve the result.
	data := []byte{0, 0}
	if err = d.c.Tx([]byte{ads1x15PointerConversion}, data); err != nil {
		return
	}

	// Convert the raw data into physical value.
	result := physic.ElectricPotential(binary.BigEndian.Uint16(data))

	if result&0x8000 != 0 {
		result -= 1 << 16
	}

	value = result * voltageMultiplier / physic.ElectricPotential(1<<15)

	return
}

// bestGainForElectricPotential returns the gain the most adapted to read up to the specified difference of potential.
func (d *Dev) bestGainForElectricPotential(voltage physic.ElectricPotential) (bestGain int, err error) {
	var max physic.ElectricPotential
	difference := physic.ElectricPotential(math.MaxInt64)
	currentBestGain := -1

	for key, value := range d.gainVoltage {
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
		err = fmt.Errorf("The maximum voltage which can be read is %s", max.String())
		return
	}

	bestGain = currentBestGain
	return
}

// bestDataRateForFrequency returns the gain the most data rate to read samples at least at the requested frequency.
func (d *Dev) bestDataRateForFrequency(minimumFrequency physic.Frequency) (bestDataRate int, err error) {
	var max physic.Frequency
	difference := physic.Frequency(math.MaxInt64)
	currentBestDataRate := -1

	for key := range d.dataRates {
		freq := physic.Frequency(key) * physic.Hertz

		// We compute the minimum in case we need to display an error
		if freq > max {
			max = freq
		}

		newDiff := freq - minimumFrequency
		if newDiff >= 0 && newDiff < difference {
			difference = newDiff
			currentBestDataRate = key
		}
	}

	if currentBestDataRate < 0 {
		err = fmt.Errorf("The maximum frequency which can be read is %s", max.String())
		return
	}

	bestDataRate = currentBestDataRate
	return
}

func (d *Dev) checkChannel(channel int) (err error) {
	if channel < 0 || channel > 3 {
		err = errors.New("Invalid channel, must be between 0 and 3")
	}
	return
}

func (p *ads1x15AnalogPin) Read() (physic.ElectricPotential, error) {
	return p.adc.executePreparedQuery(p.preparedQuery, p.waitTime, p.voltageMultiplier)
}

func (p *ads1x15AnalogPin) Name() string {
	return fmt.Sprintf("%s pin", p.adc.name)
}

func (p *ads1x15AnalogPin) Number() int {
	return -1
}

func (p *ads1x15AnalogPin) Function() string {
	return "DEPRECATED"
}

func (p *ads1x15AnalogPin) Halt() (err error) {
	return
}

func (p *ads1x15AnalogPin) String() string {
	return p.Name()
}
