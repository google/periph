// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds248x

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/periph/conn"
	"github.com/google/periph/experimental/conn/onewire"
)

// Dev is a handle to a ds248x device and it implements the onewire.Bus interface.
//
// Dev implements a persistent error model: if a fatal error is encountered it places
// itself into an error state and immediately returns the last error on all subsequent
// calls. A fresh Dev, which reinitializes the hardware, must be created to proceed.
//
// A persistent error is only set when there is a problem with the ds248x device itself
// (or the I²C bus used to access it). Errors on the 1-wire bus do not cause persistent
// errors and implement the onewire.BusError interface to indicate this fact.
type Dev struct {
	sync.Mutex               // lock for the bus while a transaction is in progress
	i2c        conn.Conn     // i2c device handle for the ds248x
	isDS2483   bool          // true: ds2483, false: ds2482-100
	confReg    byte          // value written to configuration register
	tReset     time.Duration // time to perform a 1-wire reset
	tSlot      time.Duration // time to perform a 1-bit 1-wire read/write
	err        error         // persistent error, device will no longer operate
}

// String
func (d *Dev) String() string {
	return fmt.Sprintf("ds248x")
}

// Close drops the I²C bus handle and sets a persistent error.
func (d *Dev) Close() error {
	d.i2c = nil
	d.err = fmt.Errorf("ds248x: invalid operation on closed bus")
	return nil
}

// Tx performs a bus transaction, sending and receiving bytes, and
// ending by pulling the bus high either weakly or strongly depending
// on the value of power.
//
// A strong pull-up is typically required to power temperature conversion
// or EEPROM writes.
func (d *Dev) Tx(w, r []byte, power onewire.Pullup) error {
	d.Lock()
	defer d.Unlock()

	// Issue 1-wire bus reset.
	if present, err := d.reset(); err != nil {
		return err
	} else if !present {
		return busError("ds248x: no device present")
	}

	// Send bytes onto 1-wire bus.
	for i, b := range w {
		if power == onewire.StrongPullup && i == len(w)-1 && len(r) == 0 {
			// This is the last byte, need to activate strong pull-up.
			d.i2cTx([]byte{cmdWriteConfig, d.confReg&0xbf | 0x4}, nil)
		}
		d.i2cTx([]byte{cmd1WWrite, b}, nil)
		d.waitIdle(7 * d.tSlot)
	}

	// Read bytes from one-wire bus.
	for i, _ := range r {
		if power == onewire.StrongPullup && i == len(r)-1 {
			// This is the last byte, need to activate strong-pull-up
			d.i2cTx([]byte{cmdWriteConfig, d.confReg&0xbf | 0x4}, nil)
		}
		d.i2cTx([]byte{cmd1WRead}, r[i:i+1])
		d.waitIdle(7 * d.tSlot)
		d.i2cTx([]byte{cmdSetReadPtr, regRDR}, r[i:i+1])
	}

	return d.err
}

// Search performs a "search" cycle on the 1-wire bus and returns the
// addresses of all devices on the bus if alarmOnly is false and of all
// devices in alarm state if alarmOnly is true.
//
// If an error occurs during the search the already-discovered devices are
// returned with the error.
func (d *Dev) Search(alarmOnly bool) ([]onewire.Address, error) {
	return onewire.Search(d, alarmOnly)
}

// SearchTriplet performs a single bit search triplet command on the bus,
// waits for it to complete and returs the outcome.
//
// SearchTriplet should not be used directly, use Search instead.
func (d *Dev) SearchTriplet(direction byte) (onewire.TripletResult, error) {
	// Send one-wire triplet command.
	var dir byte
	if direction != 0 {
		dir = 0x80
	}
	d.i2cTx([]byte{cmd1WTriplet, dir}, nil)
	// Wait and read status register, concoct result from there.
	status := d.waitIdle(0 * d.tSlot) // in theory 3*tSlot but it's actually overlapped
	tr := onewire.TripletResult{
		GotZero: status&0x20 == 0,
		GotOne:  status&0x40 == 0,
		Taken:   status >> 7,
	}
	return tr, d.err
}

//

// reset issues a reset signal on the 1-wire bus and returns true if any device
// responded with a presence pulse.
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
		return false, shortedBusError("onewire/ds248x: bus has a short")
	}
	return (status & 2) != 0, nil
}

// i2cTx is a helper function to call i2c.Tx and handle the error by persisting it.
func (d *Dev) i2cTx(w, r []byte) {
	if d.err != nil {
		return
	}
	d.err = d.i2c.Tx(w, r)
}

// waitIdle waits for the one wire bus to be idle.
//
// It initially sleeps for the delay and then polls the status register and
// sleeps for a tenth of the delay each time the status register indicates
// that the bus is still busy. The last read status byte is returned.
//
// An overall timeout of 3ms is applied to the whole procedure. waitIdle
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
		// If bus idle complete, return status. This also returns if d.err!=nil
		// because in that case status[0]==0.
		if (status[0] & 1) == 0 {
			return status[0]
		}
		// If we're timing out return error. This is an error with the ds248x, not with
		// devices on the 1-wire bus, hence it is persistent.
		if time.Now().After(tOut) {
			d.err = fmt.Errorf("ds248x: timeout waiting for bus cycle to finish")
			return 0
		}
		// Try not to hog the kernel thread.
		time.Sleep(delay / 10)
	}
}

// shortedBusError implements error and onewire.ShortedBusError.
type shortedBusError string

func (e shortedBusError) Error() string   { return string(e) }
func (e shortedBusError) IsShorted() bool { return true }
func (e shortedBusError) BusError() bool  { return true }

// busError implements error and onewire.BusError.
type busError string

func (e busError) Error() string  { return string(e) }
func (e busError) BusError() bool { return true }
