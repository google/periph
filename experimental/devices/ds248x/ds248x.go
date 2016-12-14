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
// on the 1-wire data line. The default value is 1000Ω.
type PupOhm uint8

const (
	R500Ω  = 4 // 500Ω passive pull-up resistor
	R1000Ω = 6 // 1000Ω passive pull-up resistor
)

// Opts contains options to pass to the constructor.
type Opts struct {
	Addr          uint16 // I²C address, default 0x18
	PassivePullup bool   // false:use active pull-up, true: disable active pullup

	// The following options are only available on the ds2483 (not ds2482-100).
	// The actual value used is the closest possible value (rounded up or down).
	ResetLow       time.Duration // reset low time, range 440μs..740μs
	PresenceDetect time.Duration // presence detect sample time, range 58μs..76μs
	Write0Low      time.Duration // write zero low time, range 52μs..70μs
	Write0Recovery time.Duration // write zero recovery time, range 2750ns..25250ns
	PullupRes      PupOhm        // passive pull-up resistance, true: 500Ω, false: 1kΩ
}

// New returns a device object that communicates over I²C to the DS2482/DS2483
// controller. This device object implements onewire.Bus and can be used to
// access devices on the bus.
func New(i i2c.Bus, opts *Opts) (*Dev, error) {
	addr := uint16(0x18)
	if opts != nil {
		switch opts.Addr {
		case 0x18, 0x19, 0x20, 0x21:
			addr = opts.Addr
		case 0x00:
		default:
			return nil, errors.New("ds248x: given address not supported by device")
		}
	}
	d := &Dev{i2c: &i2c.Dev{Bus: i, Addr: addr}}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

//

// defaults holds default values for optional parameters.
var defaults = Opts{
	PassivePullup:  false,
	ResetLow:       560 * time.Microsecond,
	PresenceDetect: 68 * time.Microsecond,
	Write0Low:      64 * time.Microsecond,
	Write0Recovery: 5250 * time.Nanosecond,
	PullupRes:      R1000Ω,
}

func (d *Dev) makeDev(opts *Opts) error {
	// Doctor the opts to apply default values.
	if opts == nil {
		opts = &defaults
	}
	if opts.ResetLow == 0 {
		opts.ResetLow = defaults.ResetLow
	}
	if opts.PresenceDetect == 0 {
		opts.PresenceDetect = defaults.PresenceDetect
	}
	if opts.Write0Low == 0 {
		opts.Write0Low = defaults.Write0Low
	}
	if opts.Write0Recovery == 0 {
		opts.Write0Recovery = defaults.Write0Recovery
	}
	if opts.PullupRes == 0 {
		opts.PullupRes = defaults.PullupRes
	}
	d.tReset = 2 * opts.ResetLow
	d.tSlot = opts.Write0Low + opts.Write0Recovery

	// Issue a reset command.
	if err := d.i2c.Tx([]byte{cmdReset}, nil); err != nil {
		return fmt.Errorf("ds248x: error while resetting: %s", err)
	}

	// Read the status register to confirm that we have a responding ds248x
	var stat [1]byte
	if err := d.i2c.Tx([]byte{cmdSetReadPtr, regStatus}, stat[:]); err != nil {
		return fmt.Errorf("ds248x: error while reading status register: %s", err)
	}
	if stat[0] != 0x18 {
		return fmt.Errorf("ds248x: invalid status register value: %#x, expected 0x18\n", stat[0])
	}

	// Write the device configuration register to get the chip out of reset state, immediately
	// read it back to get confirmation.
	d.confReg = 0xe1 // standard-speed, no strong pullup, no powerdown, active pull-up
	if opts.PassivePullup {
		d.confReg ^= 0x11
	}
	var dcr [1]byte
	if err := d.i2c.Tx([]byte{cmdWriteConfig, d.confReg}, dcr[:]); err != nil {
		return fmt.Errorf("ds248x: error while writing device config register: %s", err)
	}
	// When reading back we only get the bottom nibble
	if dcr[0] != d.confReg&0x0f {
		return fmt.Errorf("ds248x: failure to write device config register, wrote %#x got %#x back",
			d.confReg, dcr[0])
	}

	// Set the read ptr to the port configuration register to determine whether we have a
	// ds2483 vs ds2482-100. This will fail on devices that do not have a port config
	// register, such as the ds2482-100.
	d.isDS2483 = d.i2c.Tx([]byte{cmdSetReadPtr, regPCR}, nil) == nil

	// Set the options for the ds2483.
	if d.isDS2483 {
		buf := []byte{cmdAdjPort,
			byte(0x00 + ((opts.ResetLow/time.Microsecond - 430) / 20 & 0x0f)),
			byte(0x20 + ((opts.PresenceDetect/time.Microsecond - 55) / 2 & 0x0f)),
			byte(0x40 + ((opts.Write0Low/time.Microsecond - 51) / 2 & 0x0f)),
			byte(0x60 + (((opts.Write0Recovery-1250)/2500 + 5) & 0x0f)),
			byte(0x80 + (opts.PullupRes & 0x0f)),
		}
		if err := d.i2c.Tx(buf, nil); err != nil {
			return fmt.Errorf("ds248x: error while setting port config values: %s", err)
		}
	}

	return nil
}

const (
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
