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
	"time"

	"periph.io/x/periph/conn/gpio"
)

// LineTwo offset for the second line in the LCD buffer.
const LineTwo = 0x40

// LCD4 is the 4-bit addressing device for HD-44780
type LCD4 struct {
	// data pins
	dataPins []gpio.PinOut

	// register select pin
	rsPin gpio.PinOut

	// enable pin
	enablePin gpio.PinOut
}

// NewLCD4Bit creates and initializes the LCD device
//	data - references to data pins
//	rs - rs pin
//	e - strobe pin
func NewLCD4Bit(data []gpio.PinOut, rs, e gpio.PinOut) (*LCD4, error) {
	dev := LCD4{
		dataPins:  data,
		enablePin: e,
		rsPin:     rs,
	}
	if err := dev.Init(); err != nil {
		return nil, err
	}
	return &dev, nil
}

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

// Init initializes the HC-44780 chipset, clears the screen buffer and moves cursor to the
// home of screen (line 0, column 0)
func (r *LCD4) Init() error {
	if err := clearBits(r); err != nil {
		return err
	}

	delayMs(15)

	if err := r.rsPin.Out(gpio.Low); err != nil {
		return err
	}
	if err := r.enablePin.Out(gpio.Low); err != nil {
		return err
	}

	wrapper := func(seq [][]uint, f func(_lcd *LCD4, _data uint8) error) error {
		for _, v := range seq {
			if err := f(r, uint8(v[0])); err != nil {
				return err
			}
			if v[1] > 0 {
				delayUs(v[1])
			}
		}
		return nil
	}

	if err := wrapper(resetSequence, write4Bits); err != nil {
		return err
	}

	if err := wrapper(initSequence, writeInstruction); err != nil {
		return err
	}

	return nil
}

// Cls clears the LCD screen
func (r *LCD4) Cls() error {
	if err := writeInstruction(r, 0x01); err != nil {
		return err
	}
	delayMs(2)
	return nil
}

// SetCursor positions the cursor
//	line - screen line, 0-based
//	column - column, 0-based
func (r *LCD4) SetCursor(line uint8, column uint8) error {
	if err := writeInstruction(r, 0x80|(line*LineTwo+column)); err != nil {
		return err
	}
	return nil
}

// Print the data string
//	data string to display
func (r *LCD4) Print(data string) error {
	for _, v := range []byte(data) {
		if err := r.WriteChar(v); err != nil {
			return err
		}
	}
	return nil
}

// WriteChar writes a single byte (character)
//	data - character code
func (r *LCD4) WriteChar(data uint8) error {
	if err := sendData(r); err != nil {
		return err
	}
	if err := write4Bits(r, data>>4); err != nil {
		return err
	}
	if err := write4Bits(r, data); err != nil {
		return err
	}
	delayUs(10)
	return nil
}

// service methods

func clearBits(r *LCD4) error {
	for _, v := range r.dataPins {
		if err := v.Out(gpio.Low); err != nil {
			return err
		}
	}
	return nil
}

func write4Bits(ref *LCD4, data uint8) error {
	for i, v := range ref.dataPins {
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
	return strobe(ref)
}

func sendInstruction(ref *LCD4) error {
	if err := ref.rsPin.Out(gpio.Low); err != nil {
		return err
	}
	if err := ref.enablePin.Out(gpio.Low); err != nil {
		return err
	}
	return nil
}

func sendData(ref *LCD4) error {
	if err := ref.rsPin.Out(gpio.High); err != nil {
		return err
	}
	if err := ref.enablePin.Out(gpio.Low); err != nil {
		return err
	}
	return nil
}

func writeInstruction(ref *LCD4, data uint8) error {
	if err := sendInstruction(ref); err != nil {
		return err
	}
	// write high 4 bits
	if err := write4Bits(ref, data>>4); err != nil {
		return err
	}
	// write low  bits
	if err := write4Bits(ref, data); err != nil {
		return err
	}
	delayUs(50)
	return nil
}

func strobe(ref *LCD4) error {
	if err := ref.enablePin.Out(gpio.High); err != nil {
		return err
	}
	delayUs(2)
	if err := ref.enablePin.Out(gpio.Low); err != nil {
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
