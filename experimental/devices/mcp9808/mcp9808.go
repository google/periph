// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mcp9808

import (
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/physic"
)

// Opts holds the configuration options.
//
// Slave Address
//
// Depending which pins the A0, A1 and A2 pins are connected to will change the
// slave address. Default configuration is address 0x18 (Ax pins to GND). For a
// full address table see datasheet.
type Opts struct {
	Addr int
	Res  resolution
}

// DefaultOpts is the recommended default options.
var DefaultOpts = Opts{
	Addr: 0x18,
	Res:  Maximum,
}

// New opens a handle to an mcp9808 sensor.
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	i2cAddress := DefaultOpts.Addr
	if opts.Addr != 0 {
		if opts.Addr < 0x18 || opts.Addr > 0x1f {
			return nil, errAddressOutOfRange
		}
		i2cAddress = opts.Addr
	}

	dev := &Dev{
		m: mmr.Dev8{
			Conn:  &i2c.Dev{Bus: bus, Addr: uint16(i2cAddress)},
			Order: binary.BigEndian,
		},
		stop:    make(chan struct{}, 1),
		res:     opts.Res,
		enabled: false,
	}

	if err := dev.setResolution(opts.Res); err != nil {
		return nil, err
	}
	if err := dev.enable(); err != nil {
		return nil, err
	}
	return dev, nil
}

// Dev is a handle to the mcp9808 sensor.
type Dev struct {
	m    mmr.Dev8
	stop chan struct{}
	res  resolution

	mu       sync.Mutex
	sensing  bool
	critical physic.Temperature
	upper    physic.Temperature
	lower    physic.Temperature
	enabled  bool
}

// Sense reads the current temperature.
func (d *Dev) Sense(e *physic.Env) error {
	t, _, err := d.readTemperature()
	e.Temperature = t
	return err
}

// SenseContinuous returns measurements as °C, on a continuous basis.
// The application must call Halt() to stop the sensing when done to stop the
// sensor and close the channel.
// It's the responsibility of the caller to retrieve the values from the channel
// as fast as possible, otherwise the interval may not be respected.
func (d *Dev) SenseContinuous(interval time.Duration) (<-chan physic.Env, error) {
	switch d.res {
	case Maximum:
		if interval < 250*time.Millisecond {
			return nil, errTooShortInterval
		}
	case High:
		if interval < 130*time.Millisecond {
			return nil, errTooShortInterval
		}
	case Medium:
		if interval < 65*time.Millisecond {
			return nil, errTooShortInterval
		}
	case Low:
		if interval < 30*time.Millisecond {
			return nil, errTooShortInterval
		}
	}

	env := make(chan physic.Env)
	d.mu.Lock()
	d.sensing = true
	d.mu.Unlock()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			select {
			case <-time.After(interval):
				t, _, _ := d.readTemperature()
				env <- physic.Env{Temperature: t}
			case <-d.stop:
				wg.Done()
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(env)
		d.mu.Lock()
		d.sensing = false
		d.mu.Unlock()
	}()
	return env, nil
}

func (d *Dev) Precision(e *physic.Env) {
	switch d.res {
	case Maximum:
		e.Temperature = 62500 * physic.MicroKelvin
	case High:
		e.Temperature = 125 * physic.MilliKelvin
	case Medium:
		e.Temperature = 250 * physic.MilliKelvin
	case Low:
		e.Temperature = 500 * physic.MilliKelvin
	}
}

// SenseTemp reads the current temperature.
func (d *Dev) SenseTemp() (physic.Temperature, error) {
	t, _, err := d.readTemperature()
	return t, err
}

// SenseWithAlerts reads the ambient temperature and returns an slice of any
// alerts that have been tripped. Lower must be less than upper which must be
// less than critical.
func (d *Dev) SenseWithAlerts(lower, upper, critical physic.Temperature) (physic.Temperature, []Alert, error) {
	if critical > upper && upper > lower {
		if err := d.setCriticalAlert(critical); err != nil {
			return 0, nil, err
		}
		if err := d.setUpperAlert(upper); err != nil {
			return 0, nil, err
		}
		if err := d.setLowerAlert(lower); err != nil {
			return 0, nil, err
		}
	} else {
		return 0, nil, errAlertInvalid
	}

	t, alertBits, err := d.readTemperature()
	if err != nil {
		return 0, nil, err
	}

	// Check for Alerts.
	if alertBits&0xe0 > 0 {
		var as []Alert
		if alertBits&0x80 > 0 {
			// Critical Alert bit set.
			crit, err := d.m.ReadUint16(critAlert)
			if err != nil {
				return t, nil, errReadCriticalAlert
			}
			t := bitsToTemperature(crit)
			as = append(as, Alert{"critical", t})
		}

		if alertBits&0x40 > 0 {
			// Upper Alert bit set.
			upper, err := d.m.ReadUint16(upperAlert)
			if err != nil {
				return t, nil, errReadUpperAlert
			}
			t := bitsToTemperature(upper)
			as = append(as, Alert{"upper", t})
		}

		if alertBits&0x20 > 0 {
			// Lower Alert bit set.
			lower, err := d.m.ReadUint16(lowerAlert)
			if err != nil {
				return t, nil, errReadLowerAlert
			}
			t := bitsToTemperature(lower)
			as = append(as, Alert{"lower", t})
		}

		return t, as, nil
	}
	return t, nil, nil
}

// Halt put the mcp9808 into shutdown mode. It will not read temperatures while
// in shutdown mode.
func (d *Dev) Halt() error {
	d.mu.Lock()
	if d.sensing {
		d.stop <- struct{}{}
	}
	d.mu.Unlock()

	if err := d.m.WriteUint16(configuration, 0x0100); err != nil {
		return errWritingConfiguration
	}

	d.mu.Lock()
	d.enabled = false
	d.mu.Unlock()
	return nil
}

func (d *Dev) String() string {
	return "MCP9808"
}

func (d *Dev) enable() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.enabled {
		if err := d.m.WriteUint16(configuration, 0x0000); err != nil {
			return errWritingConfiguration
		}
		d.enabled = true
	}
	return nil
}

func (d *Dev) readTemperature() (physic.Temperature, uint8, error) {
	if err := d.enable(); err != nil {
		return 0, 0, err
	}

	tbits, err := d.m.ReadUint16(temperature)
	if err != nil {
		return 0, 0, errReadTemperature
	}

	return bitsToTemperature(tbits), uint8(tbits>>8) & 0xe0, nil
}

func (d *Dev) setResolution(r resolution) error {
	switch r {
	case Low:
		if err := d.m.WriteUint8(resolutionConfig, 0x00); err != nil {
			return errWritingResolution
		}
	case Medium:
		if err := d.m.WriteUint8(resolutionConfig, 0x01); err != nil {
			return errWritingResolution
		}
	case High:
		if err := d.m.WriteUint8(resolutionConfig, 0x02); err != nil {
			return errWritingResolution
		}
	case Maximum:
		if err := d.m.WriteUint8(resolutionConfig, 0x03); err != nil {
			return errWritingResolution
		}
	default:
		return errInvalidResolution
	}
	return nil
}

func (d *Dev) setCriticalAlert(t physic.Temperature) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t == d.critical {
		return nil
	}
	crit, err := alertTemperatureToBits(t)
	if err != nil {
		return err
	}
	if err := d.m.WriteUint16(critAlert, crit); err != nil {
		return errWritingCritAlert
	}
	d.critical = t
	return nil
}

func (d *Dev) setUpperAlert(t physic.Temperature) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t == d.upper {
		return nil
	}
	upper, err := alertTemperatureToBits(t)
	if err != nil {
		return err
	}
	if err := d.m.WriteUint16(upperAlert, upper); err != nil {
		return errWritingUpperAlert
	}
	d.upper = t
	return nil
}

func (d *Dev) setLowerAlert(t physic.Temperature) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t == d.lower {
		return nil
	}
	lower, err := alertTemperatureToBits(t)
	if err != nil {
		return err
	}
	if err := d.m.WriteUint16(lowerAlert, lower); err != nil {
		return errWritingLowerAlert
	}
	d.lower = t
	return nil
}

type Alert struct {
	AlertMode  string
	AlertLevel physic.Temperature
}

const (
	// Register addresses.
	configuration    byte = 0x01
	upperAlert       byte = 0x02
	lowerAlert       byte = 0x03
	critAlert        byte = 0x04
	temperature      byte = 0x05
	manifactureID    byte = 0x06
	deviceID         byte = 0x07
	resolutionConfig byte = 0x08
)

var (
	errReadTemperature      = errors.New("failed to read ambient temperature")
	errReadCriticalAlert    = errors.New("failed to read critical temperature")
	errReadUpperAlert       = errors.New("failed to read upper temperature")
	errReadLowerAlert       = errors.New("failed to read lower temperature")
	errAddressOutOfRange    = errors.New("i2c address out of range")
	errInvalidResolution    = errors.New("invalid resolution")
	errWritingResolution    = errors.New("failed to write resolution configuration")
	errWritingConfiguration = errors.New("failed to write configuration")
	errWritingCritAlert     = errors.New("failed to write critical alert configuration")
	errWritingUpperAlert    = errors.New("failed to write upper alert configuration")
	errWritingLowerAlert    = errors.New("failed to write lower alert configuration")
	errAlertOutOfRange      = errors.New("alert setting exceeds operating conditions")
	errAlertInvalid         = errors.New("invalid alert temperature configuration")
	errTooShortInterval     = errors.New("too short interval for resolution")
)

// bitsToTemperature converts the given bits to a physic.Temperature, assuming the
// bit layout common to the ambient temperature register and the alert registers.
// This works for the alert registers because while they do not make use of the 2
// least significant bits (i.e. they have resolution of 0.25°C vs. 0.0625°C for the
// ambient temp register) those 2 bits are always read as 0. See page 22 of the
// datasheet.
func bitsToTemperature(b uint16) physic.Temperature {
	t := physic.Temperature(b&0x0fff) * 62500 * physic.MicroKelvin
	if b&0x1000 > 0 {
		// Account for sign bit.
		t -= 256 * physic.Celsius
	}
	return t + physic.ZeroCelsius
}

func alertTemperatureToBits(t physic.Temperature) (uint16, error) {
	const maxAlert = 125*physic.Kelvin + physic.ZeroCelsius
	const minAlert = -40*physic.Kelvin + physic.ZeroCelsius

	if t > maxAlert || t < minAlert {
		return 0, errAlertOutOfRange
	}
	t -= physic.ZeroCelsius
	// 0.25°C per bit.
	t /= 250 * physic.MilliKelvin

	// We don't need to explicitly handle negative temperatures because both Go and the MCP9808
	// store negative values using two's complement. We can rely on Go's implementation since
	// we know that the bits of a negative value are already in two's complement, implying that
	// the sign bit will already be set to 1 due to the range check. We need to be sure to mask
	// off the 3 most significant bits after shifting though: they will all be set to 1 if the
	// value is negative. Also mask off the 2 least significant bits. While not strictly necessary,
	// the MCP9808 doesn't use them.
	bits := (uint16(t) << 2) & 0x1ffc
	return bits, nil
}

type resolution uint8

const (
	Maximum resolution = 0
	Low     resolution = 1
	Medium  resolution = 2
	High    resolution = 3
)

var _ conn.Resource = &Dev{}
var _ physic.SenseEnv = &Dev{}
