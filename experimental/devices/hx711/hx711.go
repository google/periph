// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package hx711 implements an interface to the HX711 analog to digital converter.
//
// Datasheet
//
// http://www.aviaic.com/Download/hx711F_EN.pdf.pdf
package hx711

import (
	"errors"
	"time"

	"periph.io/x/periph/conn/gpio"
)

var (
	// TimeoutError is returned from Read and ReadAveraged when the ADC took too
	// long to indicate data was available.
	TimeoutError = errors.New("timed out waiting for HX711 to become ready")
)

// InputMode controls the voltage gain and the channel multiplexer on the HX711.
// Channel A can be used with a gain of 128 or 64, and Channel B can be used
// with a gain of 32.
type InputMode int

const (
	CHANNEL_A_GAIN_128 InputMode = 1
	CHANNEL_A_GAIN_64  InputMode = 3
	CHANNEL_B_GAIN_32  InputMode = 2
)

type Dev struct {
	inputMode InputMode
	clk       gpio.PinOut
	data      gpio.PinIn
	done      chan struct{}
}

// New creates a new HX711 device.
// The data pin must support edge detection.  If your pin doesn't natively
// support edge detection you can use PollEdge from
// periph.io/x/periph/experimental/conn/gpio/gpioutil
func New(clk gpio.PinOut, data gpio.PinIn) (*Dev, error) {
	if err := data.In(gpio.PullDown, gpio.FallingEdge); err != nil {
		return nil, err
	}

	if err := clk.Out(gpio.Low); err != nil {
		return nil, err
	}

	return &Dev{
		inputMode: CHANNEL_A_GAIN_128,
		clk:       clk,
		data:      data,
		done:      nil,
	}, nil
}

// SetInputMode changes the voltage gain and channel multiplexer mode.
func (d *Dev) SetInputMode(inputMode InputMode) {
	d.inputMode = inputMode
	d.readImmediately()
}

// IsReady returns true if there is data ready to be read from the ADC.
func (d *Dev) IsReady() bool {
	return d.data.Read() == gpio.Low
}

// Read reads a single value from the ADC.  It blocks until the ADC indicates
// there is data ready for retrieval.  If the ADC doesn't pull its Data pin low
// to indicate there is data ready before the timeout is reached, TimeoutError
// is returned.
func (d *Dev) Read(timeout time.Duration) (int32, error) {
	// Wait for the falling edge that indicates the ADC has data.
	if !d.IsReady() {
		if !d.data.WaitForEdge(timeout) {
			return 0, TimeoutError
		}
	}

	return d.readImmediately(), nil
}

func (d *Dev) readImmediately() int32 {
	// Shift the 24-bit 2's compliment value.
	var value uint32
	for i := 0; i < 24; i++ {
		d.clk.Out(gpio.High)
		level := d.data.Read()
		d.clk.Out(gpio.Low)

		value <<= 1
		if level {
			value |= 1
		}
	}

	// Pulse the clock 1-3 more times to set the new ADC mode.
	for i := 0; i < int(d.inputMode); i++ {
		d.clk.Out(gpio.High)
		d.clk.Out(gpio.Low)
	}
	// Convert the 24-bit 2's compliment value to a 32-bit signed value.
	return int32(value<<8) >> 8
}

// StartContinuousRead starts reading values continuously from the ADC.  It
// returns a channel that you can use to receive these values.
//
// You must call Halt to stop reading.
//
// Calling StartContinuousRead again before Halt is an error,
// and nil will be returned.
func (d *Dev) StartContinuousRead() <-chan int32 {
	if d.done != nil {
		return nil
	}
	done := make(chan struct{})
	ret := make(chan int32)

	go func() {
		for {
			select {
			case <-done:
				close(ret)
				return
			default:
				value, err := d.Read(time.Second)
				if err == nil {
					ret <- value
				}
			}
		}
	}()

	d.done = done
	return ret
}

// Halt stops a continuous read that was started with StartContinuousRead.
// This will close the channel that was returned by StartContinuousRead.
func (d *Dev) Halt() {
	if d.done != nil {
		close(d.done)
		d.done = nil
	}
}
