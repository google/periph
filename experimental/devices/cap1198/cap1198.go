// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package cap1198 controls a Microchip cap1198 device over I²C.

// The device is a 8 Channel Capacitive Touch Sensor with 8 LED Drivers
//
// Datasheet
//
// The official data sheet can be found here:
//
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1188%20.pdf
//
package cap1198

import (
	"errors"
	"fmt"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/devices"
)

// Dev is a handle to a cap1198.
type Dev struct {
	d     conn.Conn
	isSPI bool
	opts  *Opts
	mu    sync.Mutex
}

func (d *Dev) String() string {
	return fmt.Sprintf("CAP1198{%s}", d.d)
}

// Halt is a noop for the CAP1198.
func (d *Dev) Halt() error {
	return nil
}

// Reset issues a soft reset to the device.
func (d *Dev) Reset() error {
	panic("not implemented")

	return nil
}

// Standby turns off the capacitive touch sensor inputs. The status registers will
// not be cleared until read. LEDs that are linked to capacitive touch sensor
// inputs will remain linked and active. Sensor inputs that are no longer
// sampled will flag a release and then remain in a non-touched state. LEDs that
// are manually controlled will be unaffected.
func (d *Dev) Standby() error {
	panic("not implemented")

	return nil
}

// WakeUp takes the device out of standby or deep sleep mode.
func (d *Dev) WakeUp() error {
	panic("not implemented")
}

// DeepSleep puts the device in a deep sleep mode. All sensor input scanning is
// disabled. All LEDs are driven to their programmed non-actuated
// state and no PWM operations will be done.
func (d *Dev) DeepSleep() error {
	panic("not implemented")
}

// NewI2C returns a new device that communicates over I²C to CAP1198.
func NewI2C(b i2c.Bus, opts *Opts) (*Dev, error) {
	addr := uint16(0x28)
	if opts != nil {
		switch opts.Address {
		case 0x28, 0x29, 0x2a, 0x2b, 0x2c:
			addr = opts.Address
		case 0x00:
			// do not do anything
		default:
			return nil, errors.New("cap1198: given address not supported by device")
		}
	}
	d := &Dev{d: &i2c.Dev{Bus: b, Addr: addr}, isSPI: false}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// NewSPI returns an object that communicates over SPI to CAP1198 environmental
// sensor.
func NewSPI(p spi.Port, opts *Opts) (*Dev, error) {
	panic("not implemented")
}

func (d *Dev) makeDev(opts *Opts) error {
	// Use default options if none are passed.
	if opts == nil {
		opts = DefaultOpts()
	}
	d.opts = opts
	return nil
}

var _ devices.Device = &Dev{}
