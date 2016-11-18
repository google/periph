// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ds248x controls a Maxim DS2483 or DS2482-100 1-wire interface chip over I²C.
//
// Datasheets
//
// https://www.maximintegrated.com/en/products/digital/one-wire/DS2483.html
//
// https://www.maximintegrated.com/en/products/interface/controllers-expanders/DS2482-100.html
package ds248x

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/periph/conn/i2c"
)

// PupOhm controls the strength of the passive pull-up resistor
// on the 1-wire data line. The default value is 1000 Ohm.
type PupOhm uint8

const (
	R500Ohm  = 4 // 500 Ohm passive pull-up resistor
	R1000Ohm = 6 // 1000 Ohm passive pull-up resistor
)

// Opts contains optional options to pass to the constructor.
type Opts struct {
	Address       uint16 // I2C address: must be 0x18 for a ds2483
	PassivePullup bool   // false:use active pull-up, true: disable active pullup

	// The following options are only available on the ds2483 (not ds2482-100).
	// The actual value used is the closest possible value (rounded up or down).
	ResetLowUs       int    // reset low time in us, range 440..740
	PresenceDetectUs int    // presence detect sample time in us, range 58..76
	Write0LowUs      int    // write zero low time in us, range 52..70
	Write0RecoveryNs int    // write zero recovery time in ns, range 2750..25250
	PullupRes        PupOhm // passive pull-up resistance, true: 500ohm, false: 1kohm
}

// NewI2C returns an object that communicates over I²C to the DS2482/DS2483 controller.
func NewI2C(i i2c.Conn, opts *Opts) (*Dev, error) {
	addr := uint16(0x18)
	if opts != nil {
		switch opts.Address {
		case 0x18, 0x19:
			addr = opts.Address
		case 0x00:
			// do not do anything
		default:
			return nil, errors.New("given address not supported by device")
		}
	}
	d := &Dev{i2c: &i2c.Dev{Conn: i, Addr: addr}}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// defaults holds default values for optional parameters.
var defaults = Opts{
	PassivePullup:    false,
	ResetLowUs:       560,
	PresenceDetectUs: 68,
	Write0LowUs:      64,
	Write0RecoveryNs: 5250,
	PullupRes:        R1000Ohm,
}

func (d *Dev) makeDev(opts *Opts) error {
	// Doctor the opts to apply default values.
	if opts == nil {
		opts = &defaults
	}
	if opts.ResetLowUs == 0 {
		opts.ResetLowUs = defaults.ResetLowUs
	}
	if opts.PresenceDetectUs == 0 {
		opts.PresenceDetectUs = defaults.PresenceDetectUs
	}
	if opts.Write0LowUs == 0 {
		opts.Write0LowUs = defaults.Write0LowUs
	}
	if opts.Write0RecoveryNs == 0 {
		opts.Write0RecoveryNs = defaults.Write0RecoveryNs
	}
	if opts.PullupRes == 0 {
		opts.PullupRes = defaults.PullupRes
	}
	d.tReset = time.Duration(2*opts.ResetLowUs) * microsecond
	d.tSlot = time.Duration(1000*opts.Write0LowUs + opts.Write0RecoveryNs)

	// Issue a reset command.
	if err := d.i2c.Tx([]byte{cmdReset}, nil); err != nil {
		return fmt.Errorf("%s while resetting ds248x", err)
	}

	// Read the status register to confirm that we have a responding ds248x
	stat := make([]byte, 1)
	if err := d.i2c.Tx([]byte{cmdSetReadPtr, regStatus}, stat); err != nil {
		return fmt.Errorf("%s while reading ds248x status register", err)
	}
	if stat[0] != 0x18 {
		return fmt.Errorf("invalid ds248x status register value: %#x, expected 0x18\n", stat[0])
	}

	// Write the device configuration register to get the chip out of reset state, immediately
	// read it back to get confirmation.
	d.confReg = 0xe1 // standard-speed, no strong pullup, no powerdown, active pull-up
	if opts.PassivePullup {
		d.confReg ^= 0x11
	}
	dcr := make([]byte, 1)
	if err := d.i2c.Tx([]byte{cmdWriteConfig, d.confReg}, dcr); err != nil {
		return fmt.Errorf("%s while writing ds248x device config register", err)
	}
	// When reading back we only get the bottom nibble
	if dcr[0] != d.confReg&0x0f {
		return fmt.Errorf("failure to write device config register, wrote %#x got %#x back",
			d.confReg, dcr[0])
	}

	// Set the read ptr to the port configuration register to determine whether we have a
	// ds2483 vs ds2482-100. This will fail on the ds2482-100.
	d.isDS2483 = d.i2c.Tx([]byte{cmdSetReadPtr, regPCR}, nil) == nil

	// Set the options for the ds2483.
	if d.isDS2483 {
		buf := []byte{cmdAdjPort,
			byte(0x00 + ((opts.ResetLowUs - 430) / 20 & 0x0f)),
			byte(0x20 + ((opts.PresenceDetectUs - 55) / 2 & 0x0f)),
			byte(0x40 + ((opts.Write0LowUs - 51) / 2 & 0x0f)),
			byte(0x60 + (((opts.Write0RecoveryNs-1250)/2500 + 5) & 0x0f)),
			byte(0x80 + (opts.PullupRes & 0x0f)),
		}
		if err := d.i2c.Tx(buf, nil); err != nil {
			return fmt.Errorf("%s while setting ds2483 port config values", err)
		}
	}

	return nil
}

const (
	microsecond = 1000 * time.Nanosecond

	cmdReset       = 0xf0 // reset ds248x
	cmdSetReadPtr  = 0xe1 // set the read pointer
	cmdWriteConfig = 0xd2 // write the device configuration
	cmdAdjPort     = 0xc3 // adjust 1-wire port
	cmd1WReset     = 0xb4 // reset the 1-wire bus
	cmd1WBit       = 0x87 // perform a single-bit transaction on the 1-wire bus
	cmd1WWrite     = 0xa5 // perform a byte write on the 1-wire bus
	cmd1WRead      = 0x96 // perform a byte read on the 1-wire bus
	cmd1WTriplet   = 0x78 // perform a triplet operation (2 bit reads, a bit write)

	regDCR    = 0xc3 // read ptr for device configuration register
	regStatus = 0xf0 // read ptr for status register
	regRDR    = 0xe1 // read ptr for read-data register
	regPCR    = 0xb4 // read ptr for port configuration register
)
