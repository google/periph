// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package hd44780 controls the Hitachi LCD display chipset HD-44780
//
// Datasheet
//
// https://www.sparkfun.com/datasheets/LCD/HD44780.pdf
package hd44780

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
)

// lineTwo offset for the second line in the LCD buffer.
const lineTwo = 0x40

// Dev is the 4-bit addressing device for HD-44780
type Dev struct {
	// data pins
	dataPins []gpio.PinOut

	// register select pin
	rsPin gpio.PinOut

	// enable pin
	enablePin gpio.PinOut
}

// New creates and initializes the LCD device
//	data - references to data pins
//	rs - rs pin
//	e - strobe pin
func New(data []gpio.PinOut, rs, e gpio.PinOut) (*Dev, error) {
	if len(data) != 4 {
		return nil, fmt.Errorf("expected 4 data pins, passed %d", len(data))
	}
	dev := &Dev{
		dataPins:  data,
		enablePin: e,
		rsPin:     rs,
	}
	if err := dev.Reset(); err != nil {
		return nil, err
	}
	return dev, nil
}

// Reset resets the HC-44780 chipset, clears the screen buffer and moves cursor to the
// home of screen (line 0, column 0).
func (r *Dev) Reset() error {
	if err := r.clearBits(); err != nil {
		return err
	}

	delayMs(15)

	if err := r.rsPin.Out(gpio.Low); err != nil {
		return err
	}
	if err := r.enablePin.Out(gpio.Low); err != nil {
		return err
	}

	if err := r.bulkSendData(resetSequence, r.write4Bits); err != nil {
		return err
	}

	if err := r.bulkSendData(initSequence, r.writeInstruction); err != nil {
		return err
	}

	return nil
}

func (r *Dev) String() string {
	return "HD44870, 4 bit mode"
}

// Halt clears the LCD screen
func (r *Dev) Halt() error {
	if err := r.writeInstruction(0x01); err != nil {
		return err
	}
	delayMs(2)
	return nil
}

// SetCursor positions the cursor
//	line - screen line, 0-based
//	column - column, 0-based
func (r *Dev) SetCursor(line uint8, column uint8) error {
	if err := r.writeInstruction(0x80 | (line*lineTwo + column)); err != nil {
		return err
	}
	return nil
}

// Print the data string
//	data string to display
func (r *Dev) Print(data string) error {
	for _, v := range []byte(data) {
		if err := r.WriteChar(v); err != nil {
			return err
		}
	}
	return nil
}

// WriteChar writes a single byte (character) at the cursor position.
//	data - character code
func (r *Dev) WriteChar(data uint8) error {
	if err := r.sendData(); err != nil {
		return err
	}
	if err := r.write4Bits(data >> 4); err != nil {
		return err
	}
	if err := r.write4Bits(data); err != nil {
		return err
	}
	delayUs(10)
	return nil
}

// service methods

var resetSequence = [][]uint{
	{0x03, 50}, // init 1-st cycle
	{0x03, 10}, // init 2-nd cycle
	{0x03, 10}, // init 3-rd cycle
	{0x02, 10}, // init finish
}

var initSequence = [][]uint{
	{0x14, 0},    // 4-bit mode, 2 lines, 5x7 chars high
	{0x10, 0},    // disable display
	{0x01, 2000}, // clear screen
	{0x06, 0},    // cursor shift right, no display move
	{0x0c, 0},    // enable display no cursor
	{0x01, 2000}, // clear screen
	{0x02, 2000}, // cursor home
}

func (r *Dev) bulkSendData(seq [][]uint, f func(_data uint8) error) error {
	for _, v := range seq {
		if err := f(uint8(v[0])); err != nil {
			return err
		}
		if v[1] > 0 {
			delayUs(v[1])
		}
	}
	return nil
}

func (r *Dev) clearBits() error {
	for _, v := range r.dataPins {
		if err := v.Out(gpio.Low); err != nil {
			return err
		}
	}
	return nil
}

func (r *Dev) write4Bits(data uint8) error {
	for i, v := range r.dataPins {
		if data&(1<<uint(i)) > 0 {
			if err := v.Out(gpio.High); err != nil {
				return err
			}
		} else {
			if err := v.Out(gpio.Low); err != nil {
				return err
			}
		}
	}
	return r.strobe()
}

func (r *Dev) sendInstruction() error {
	if err := r.rsPin.Out(gpio.Low); err != nil {
		return err
	}
	if err := r.enablePin.Out(gpio.Low); err != nil {
		return err
	}
	return nil
}

func (r *Dev) sendData() error {
	if err := r.rsPin.Out(gpio.High); err != nil {
		return err
	}
	if err := r.enablePin.Out(gpio.Low); err != nil {
		return err
	}
	return nil
}

func (r *Dev) writeInstruction(data uint8) error {
	if err := r.sendInstruction(); err != nil {
		return err
	}
	// write high 4 bits
	if err := r.write4Bits(data >> 4); err != nil {
		return err
	}
	// write low  bits
	if err := r.write4Bits(data); err != nil {
		return err
	}
	delayUs(50)
	return nil
}

func (r *Dev) strobe() error {
	if err := r.enablePin.Out(gpio.High); err != nil {
		return err
	}
	delayUs(2)
	if err := r.enablePin.Out(gpio.Low); err != nil {
		return err
	}
	return nil
}

func delayUs(ms uint) {
	time.Sleep(time.Duration(ms) * time.Microsecond)
}

func delayMs(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

var _ conn.Resource = &Dev{}
