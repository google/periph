// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Specification
//
// http://www.nxp.com/documents/user_manual/UM10204.pdf

package bitbang

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/i2c"
	"github.com/google/periph/host"
)

// Use SkipAddr to skip the address from being sent.
const SkipAddr uint16 = 0xFFFF

// I2C represents an I²C master implemented as bit-banging on 2 GPIO pins.
type I2C struct {
	mu        sync.Mutex
	scl       gpio.PinIO // Clock line
	sda       gpio.PinIO // Data line
	halfCycle time.Duration
}

func (i *I2C) String() string {
	return fmt.Sprintf("bitbang/i2c(%s, %s)", i.scl, i.sda)
}

// Close implements i2c.ConnCloser.
func (i *I2C) Close() error {
	return nil
}

// Tx implements i2c.Conn.
func (i *I2C) Tx(addr uint16, w, r []byte) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	//syscall.Setpriority(which, who, prio)

	i.start()
	defer i.stop()
	if addr != SkipAddr {
		if addr > 0xFF {
			// Page 15, section 3.1.11 10-bit addressing
			// TOOD(maruel): Implement if desired; prefix 0b11110xx.
			return errors.New("invalid address")
		}
		// Page 13, section 3.1.10 The slave address and R/W bit
		addr <<= 1
		if len(r) == 0 {
			addr |= 1
		}
		ack, err := i.writeByte(byte(addr))
		if err != nil {
			return err
		}
		if !ack {
			return errors.New("i2c: got NACK")
		}
	}
	for _, b := range w {
		ack, err := i.writeByte(b)
		if err != nil {
			return err
		}
		if !ack {
			return errors.New("i2c: got NACK")
		}
	}
	for x := range r {
		var err error
		r[x], err = i.readByte()
		if err != nil {
			return err
		}
	}
	return nil
}

// Speed implements i2c.Conn.
func (i *I2C) Speed(hz int64) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.halfCycle = time.Second / time.Duration(hz) / time.Duration(2)
	return nil
}

// SCL implements i2c.Pins.
func (i *I2C) SCL() gpio.PinIO {
	return i.scl
}

// SDA implements i2c.Pins.
func (i *I2C) SDA() gpio.PinIO {
	return i.sda
}

// New returns an object that communicates I²C over two pins.
//
// BUG(maruel): It is close to working but not yet, the signal is incorrect
// during ACK.
//
// It has two special features:
// - Special address SkipAddr can be used to skip the address from being
//   communicated
// - An arbitrary speed can be used
func New(clk gpio.PinIO, data gpio.PinIO, speedHz int) (*I2C, error) {
	// Spec calls to idle at high. Page 8, section 3.1.1.
	// Set SCL as pull-up.
	if err := clk.In(gpio.Up, gpio.None); err != nil {
		return nil, err
	}
	if err := clk.Out(gpio.High); err != nil {
		return nil, err
	}
	// Set SDA as pull-up.
	if err := data.In(gpio.Up, gpio.None); err != nil {
		return nil, err
	}
	if err := data.Out(gpio.High); err != nil {
		return nil, err
	}
	i := &I2C{
		scl:       clk,
		sda:       data,
		halfCycle: time.Second / time.Duration(speedHz) / time.Duration(2),
	}
	return i, nil
}

//

// "When CLK is a high level and DIO changes from high to low level, data input
// starts."
//
// Ends with SDA and SCL low.
//
// Lasts 1/2 cycle.
func (i *I2C) start() {
	// Page 9, section 3.1.4 START and STOP conditions
	// In multi-master mode, it would have to sense SDA first and after the sleep.
	i.sda.Out(gpio.Low)
	i.sleepHalfCycle()
	i.scl.Out(gpio.Low)
}

// "When CLK is a high level and DIO changes from low level to high level, data
// input ends."
//
// Lasts 3/2 cycle.
func (i *I2C) stop() {
	// Page 9, section 3.1.4 START and STOP conditions
	i.scl.Out(gpio.Low)
	i.sleepHalfCycle()
	i.scl.Out(gpio.High)
	i.sleepHalfCycle()
	i.sda.Out(gpio.High)
	// TODO(maruel): This sleep could be skipped, assuming we wait for the next
	// transfer if too quick to happen.
	i.sleepHalfCycle()
}

// writeByte writes 8 bits then waits for ACK.
//
// Expects SDA and SCL low.
//
// Ends with SDA low and SCL high.
//
// Lasts 9 cycles.
func (i *I2C) writeByte(b byte) (bool, error) {
	// Page 9, section 3.1.3 Data validity
	// "The data on te SDA line must be stable during the high period of the
	// clock."
	// Page 10, section 3.1.5 Byte format
	for x := 0; x < 8; x++ {
		i.sda.Out(b&byte(1<<byte(7-x)) != 0)
		i.sleepHalfCycle()
		// Let the device read SDA.
		// TODO(maruel): Support clock stretching, the device may keep the line low.
		i.scl.Out(gpio.High)
		i.sleepHalfCycle()
		i.scl.Out(gpio.Low)
	}
	// Page 10, section 3.1.6 ACK and NACK
	// 9th clock is ACK.
	i.sleepHalfCycle()
	// SCL was already set as pull-up. PullNoChange
	if err := i.scl.In(gpio.Up, gpio.None); err != nil {
		return false, err
	}
	// SDA was already set as pull-up.
	if err := i.sda.In(gpio.Up, gpio.None); err != nil {
		return false, err
	}
	// Implement clock stretching, the device may keep the line low.
	for i.scl.Read() == gpio.Low {
		i.sleepHalfCycle()
	}
	// ACK == Low.
	ack := i.sda.Read() == gpio.Low
	if err := i.scl.Out(gpio.Low); err != nil {
		return false, err
	}
	if err := i.sda.Out(gpio.Low); err != nil {
		return false, err
	}
	return ack, nil
}

// readByte reads 8 bits and an ACK.
//
// Expects SDA and SCL low.
//
// Ends with SDA low and SCL high.
//
// Lasts 9 cycles.
func (i *I2C) readByte() (byte, error) {
	var b byte
	if err := i.sda.In(gpio.Up, gpio.None); err != nil {
		return b, err
	}
	for x := 0; x < 8; x++ {
		i.sleepHalfCycle()
		// TODO(maruel): Support clock stretching, the device may keep the line low.
		i.scl.Out(gpio.High)
		i.sleepHalfCycle()
		if i.sda.Read() == gpio.High {
			b |= byte(1) << byte(7-x)
		}
		i.scl.Out(gpio.Low)
	}
	if err := i.sda.Out(gpio.Low); err != nil {
		return 0, err
	}
	i.sleepHalfCycle()
	i.scl.Out(gpio.High)
	i.sleepHalfCycle()
	return b, nil
}

// sleep does a busy loop to act as fast as possible.
func (i *I2C) sleepHalfCycle() {
	host.Nanospin(i.halfCycle)
}

var _ i2c.Conn = &I2C{}
