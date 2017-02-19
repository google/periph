// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package tm1637 controls a TM1637 device over GPIO pins.
//
// Datasheet
//
// http://olimex.cl/website_MCI/static/documents/Datasheet_TM1637.pdf
package tm1637

import (
	"errors"
	"runtime"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/cpu"
)

// Clock converts time to a slice of bytes as segments.
func Clock(hour, minute int, showDots bool) []byte {
	seg := make([]byte, 4)
	seg[0] = byte(digitToSegment[hour/10])
	seg[1] = byte(digitToSegment[hour%10])
	seg[2] = byte(digitToSegment[minute/10])
	seg[3] = byte(digitToSegment[minute%10])
	if showDots {
		seg[1] |= 0x80
	}
	return seg[:]
}

// Digits converts hex numbers to a slice of bytes as segments.
//
// Numbers outside the range [0, 15] are displayed as blank. Use -1 to mark it
// as blank.
func Digits(n ...int) []byte {
	seg := make([]byte, len(n))
	for i := range n {
		if n[i] >= 0 && n[i] < 16 {
			seg[i] = byte(digitToSegment[n[i]])
		}
	}
	return seg
}

// Brightness defines the screen brightness as controlled by the internal PWM.
type Brightness uint8

// Valid brightness values.
const (
	Off          Brightness = 0x80 // Completely off.
	Brightness1  Brightness = 0x88 // 1/16 PWM
	Brightness2  Brightness = 0x89 // 2/16 PWM
	Brightness4  Brightness = 0x8A // 4/16 PWM
	Brightness10 Brightness = 0x8B // 10/16 PWM
	Brightness11 Brightness = 0x8C // 11/16 PWM
	Brightness12 Brightness = 0x8D // 12/16 PWM
	Brightness13 Brightness = 0x8E // 13/16 PWM
	Brightness14 Brightness = 0x8F // 14/16 PWM
)

// Dev represents an handle to a tm1637.
type Dev struct {
	clk  gpio.PinOut
	data gpio.PinIO
}

// SetBrightness changes the brightness and/or turns the display on and off.
func (d *Dev) SetBrightness(b Brightness) error {
	// This helps reduce jitter a little.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	d.start()
	d.writeByte(byte(b))
	d.stop()
	return nil
}

// Write writes raw segments, while implementing io.Writer.
//
// P can be a dot or ':' following a digit. Otherwise it is likely
// disconnected. Each byte is encoded as PGFEDCBA.
//
//     -A-
//    F   B
//     -G-
//    E   C
//     -D-   P
func (d *Dev) Write(seg []byte) (int, error) {
	if len(seg) > 6 {
		return 0, errors.New("tm1637: up to 6 segment groups are supported")
	}
	// This helps reduce jitter a little.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// Use auto-incrementing address. It is possible to write to a single
	// segment but there isn't much point.
	d.start()
	d.writeByte(0x40)
	d.stop()
	d.start()
	d.writeByte(0xC0)
	for i := 0; i < 6; i++ {
		if len(seg) <= i {
			d.writeByte(0)
		} else {
			d.writeByte(seg[i])
		}
	}
	d.stop()
	return len(seg), nil
}

// New returns an object that communicates over two pins to a TM1637.
func New(clk gpio.PinOut, data gpio.PinIO) (*Dev, error) {
	// Spec calls to idle at high.
	if err := clk.Out(gpio.High); err != nil {
		return nil, err
	}
	if err := data.Out(gpio.High); err != nil {
		return nil, err
	}
	d := &Dev{clk: clk, data: data}
	return d, nil
}

//

// Page 10 states the max clock frequency is 500KHz but page 3 states 250KHz.
//
// Writing the complete display is 8 bytes, totalizing 9*8+2 = 74 cycles.
// At 250KHz, this is 296µs.
const clockHalfCycle = time.Second / 250000 / 2

// Hex digits from 0 to F.
var digitToSegment = []byte{
	0x3f, 0x06, 0x5b, 0x4f, 0x66, 0x6d, 0x7d, 0x07, 0x7f, 0x6f, 0x77, 0x7c, 0x39, 0x5e, 0x79, 0x71,
}

func (d *Dev) start() {
	d.data.Out(gpio.Low)
	d.sleepHalfCycle()
	d.clk.Out(gpio.Low)
}

func (d *Dev) stop() {
	d.sleepHalfCycle()
	d.clk.Out(gpio.High)
	d.sleepHalfCycle()
	d.data.Out(gpio.High)
	d.sleepHalfCycle()
}

// writeByte starts with d.data low and d.clk high and ends with d.data low and
// d.clk high.
func (d *Dev) writeByte(b byte) (bool, error) {
	for i := 0; i < 8; i++ {
		// LSB (!)
		d.data.Out(b&(1<<byte(i)) != 0)
		d.sleepHalfCycle()
		d.clk.Out(gpio.High)
		d.sleepHalfCycle()
		d.clk.Out(gpio.Low)
	}
	// 9th clock is ACK.
	d.data.Out(gpio.Low)
	time.Sleep(clockHalfCycle)
	// TODO(maruel): Add.
	//if err := d.data.In(gpio.PullUp, gpio.NoEdge); err != nil {
	//	return false, err
	//}
	d.clk.Out(gpio.High)
	d.sleepHalfCycle()
	//ack := d.data.Read() == gpio.Low
	//d.sleepHalfCycle()
	//if err := d.data.Out(); err != nil {
	//	return false, err
	//}
	d.clk.Out(gpio.Low)
	return true, nil
}

// sleep does a busy loop to act as fast as possible.
func (d *Dev) sleepHalfCycle() {
	cpu.Nanospin(clockHalfCycle)
}
