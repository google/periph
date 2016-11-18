// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds248x

import (
	"fmt"
	"time"

	"github.com/google/periph/conn"
)

// Dev is a handle to a ds248x device and it implements the onewire.Conn interface.
//
// Dev implements a persistent error model: if a fatal error is encountered it places
// itself into an error state and immediately returns the last error on all subsequent
// calls. A fresh Dev, which reinitializes the hardware, must be created to proceed.
//
// A persistent error is only set when there is a problem with the ds248x device itself
// (or the I²C bus used to access it). Errors on the 1-wire bus do not cause persistent
// errors and implement the Temporary interface to indicate this fact.
type Dev struct {
	i2c      conn.Conn     // i2c device handle for the ds248x
	isDS2483 bool          // true: ds2483, false: ds2482-100
	confReg  byte          // value written to configuration register
	tReset   time.Duration // time to perform a 1-wire reset
	tSlot    time.Duration // time to perform a 1-bit 1-wire read/write
	err      error         // persistent error, device will no longer operate
}

// OneWireBusError holds an error encountered on the 1-wire bus itself (as opposed
// to an error with the ds248x device). These errors are temporary, i.e. the ds248x
// can continue to be used.
type OneWireBusError string

// Temporary returns true.
func (e OneWireBusError) Temporary() bool { return true }

// Error implements the error interface.
func (e OneWireBusError) Error() string { return string(e) }

type Temporary interface {
	Temporary() bool // true if the ds248x Dev can continue to be used
}

// Tx performs a "match ROM" command on the bus, which selects at most
// one device and then transmits and receives the specified bytes.
func (d *Dev) Tx(addr uint64, w, r []byte) error {
	return d.tx(addr, w, r, false)
}

// TxPup performs a "match ROM" command on the bus, which selects at most
// one device, then transmits and receives the specified bytes, and finally
// turns on a strong pull-up to power devices on the bus.
func (d *Dev) TxPup(addr uint64, w, r []byte) error {
	return d.tx(addr, w, r, true)
}

// All performs a "skip ROM" command on the bus, which selects all devices,
// and then transmits the specified bytes.
func (d *Dev) All(w []byte) error {
	return d.all(w, false)
}

// AllPup performs a "skip ROM" command on the bus, which selects all devices,
// then transmits the specified bytes, and finally
// turns on a strong pull-up to power devices on the bus.
func (d *Dev) AllPup(w []byte) error {
	return d.all(w, true)
}

// tx performs a "match ROM" command on the bus, which selects at most
// one device and then transmits and receives the specified bytes.
func (d *Dev) tx(addr uint64, w, r []byte, pup bool) error {
	// Issue 1-wire bus reset.
	if present, err := d.reset(); err != nil {
		return err
	} else if !present {
		return OneWireBusError("1-wire: no device present")
	}

	// Issue ROM match command to select the device followed by the bytes being
	// written, we then switch to reading.
	ww := make([]byte, 9, len(w)+9)
	ww[0] = 0x55 // Match ROM
	ww[1] = byte(addr >> 0)
	ww[2] = byte(addr >> 8)
	ww[3] = byte(addr >> 16)
	ww[4] = byte(addr >> 24)
	ww[5] = byte(addr >> 32)
	ww[6] = byte(addr >> 40)
	ww[7] = byte(addr >> 48)
	ww[8] = byte(addr >> 56)
	ww = append(ww, w...)
	for i, b := range ww {
		if pup && i == len(ww)-1 && len(r) == 0 {
			// This is the last byte, need to activate strong-pull-up
			d.i2cTx([]byte{cmdWriteConfig, d.confReg&0xbf | 0x4}, nil)
		}
		d.i2cTx([]byte{cmd1WWrite, b}, nil)
		d.waitIdle(7 * d.tSlot)
	}

	// Now read bytes from one-wire bus
	for i, _ := range r {
		if pup && i == len(r)-1 {
			// This is the last byte, need to activate strong-pull-up
			d.i2cTx([]byte{cmdWriteConfig, d.confReg&0xbf | 0x4}, nil)
		}
		d.i2cTx([]byte{cmd1WRead}, r[i:i+1])
		d.waitIdle(7 * d.tSlot)
		d.i2cTx([]byte{cmdSetReadPtr, regRDR}, r[i:i+1])
	}

	return d.err
}

func (d *Dev) all(w []byte, pup bool) error {
	// Issue 1-wire bus reset.
	if present, err := d.reset(); err != nil {
		return err
	} else if !present {
		return OneWireBusError("1-wire: no device present")
	}

	// Issue Skip ROM command to select all devices followed by the bytes being written.
	ww := make([]byte, 1, len(w)+1)
	ww[0] = 0xCC // Skip ROM
	ww = append(ww, w...)
	for i, b := range ww {
		if pup && i == len(ww)-1 {
			// This is the last byte, need to activate strong-pull-up
			d.i2cTx([]byte{cmdWriteConfig, d.confReg&0xbf | 0x4}, nil)
		}
		d.i2cTx([]byte{cmd1WWrite, b}, nil)
		d.waitIdle(7 * d.tSlot)
	}

	return d.err
}

// reset issues a reset signal on the 1-wire bus and returns true if any device responded with a
// presence pulse.
func (d *Dev) reset() (bool, error) {
	// Issue reset.
	d.i2cTx([]byte{cmd1WReset}, nil)

	// Wait for reset to complete.
	status := d.waitIdle(d.tReset)
	if d.err != nil {
		return false, d.err
	}
	// Detect bus short and turn into 1-wire error
	if (status & 4) != 0 {
		return false, OneWireBusError("1-wire: bus has a short")
	}
	return (status & 2) != 0, nil
}

// i2cTx is a helper function to call i2c.Tx and handle the error by persisting it.
func (d *Dev) i2cTx(w, r []byte) {
	if d.err != nil {
		return
	}
	if err := d.i2c.Tx(w, r); err != nil {
		d.err = err
	}
}

// waitIdle waits for the one wire bus to be idle.
//
// It initially sleeps for the delay and then polls the status register and sleeps for a tenth of
// the delay each time the status register indicates that the bus is still busy. The last read
// status byte is returned. An overall timeout of 3ms is applied to the whole procedure.  waitIdle
// uses the persistent error model and returns 0 if there is an error.
func (d *Dev) waitIdle(delay time.Duration) byte {
	if d.err != nil {
		return 0
	}
	// Overall timeout.
	tOut := time.Now().Add(3 * time.Millisecond)
	time.Sleep(delay)
	for {
		// Read status register.
		var status [1]byte
		d.i2cTx(nil, status[:])
		// If bus idle complete, return status.
		// No explicit error check needed here because status[0]==0 on error.
		if (status[0] & 1) == 0 {
			return status[0]
		}
		// If we're timing out return error. This is an error with the ds248x, not with
		// devices on the 1-wire bus, hence it is persistent.
		if time.Now().After(tOut) {
			d.err = fmt.Errorf("ds248x: timeout waiting for ds248x to finish bus cycle")
			return 0
		}
		// Try not to hog the kernel thread.
		time.Sleep(delay / 10)
	}
}
