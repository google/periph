// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package hx711 implements an interface to the 24-bits HX711 analog to digital
// converter.
//
// Datasheet
//
// http://www.aviaic.com/Download/hx711F_EN.pdf.pdf
package hx711

import (
	"errors"
	"sync"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/experimental/conn/analog"
)

var (
	// ErrTimeout is returned from Read and ReadAveraged when the ADC took too
	// long to indicate data was available.
	ErrTimeout = errors.New("timed out waiting for HX711 to become ready")
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

// Dev is a handle to a hx711.
type Dev struct {
	// Immutable.
	name string
	clk  gpio.PinOut
	data gpio.PinIn

	// Mutable.
	mu        sync.Mutex
	inputMode InputMode
	done      chan struct{}
}

// New creates a new HX711 device.
//
// The data pin must support edge detection. If your pin doesn't natively
// support edge detection you can use PollEdge from gpioutil.
func New(clk gpio.PinOut, data gpio.PinIn) (*Dev, error) {
	if err := data.In(gpio.PullDown, gpio.FallingEdge); err != nil {
		return nil, err
	}
	if err := clk.Out(gpio.Low); err != nil {
		return nil, err
	}
	return &Dev{
		name:      "hx711{" + clk.Name() + ", " + data.Name() + "}",
		inputMode: CHANNEL_A_GAIN_128,
		clk:       clk,
		data:      data,
		done:      nil,
	}, nil
}

// String implements analog.PinADC.
func (d *Dev) String() string {
	return d.name
}

// Name implements analog.PinADC.
func (d *Dev) Name() string {
	return d.String()
}

// Number implements analog.PinADC.
func (d *Dev) Number() int {
	return -1
}

// Function implements analog.PinADC.
func (d *Dev) Function() string {
	return "ADC"
}

// SetInputMode changes the voltage gain and channel multiplexer mode.
func (d *Dev) SetInputMode(inputMode InputMode) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.inputMode = inputMode
	_, err := d.readRaw()
	return err
}

// Range implements analog.PinADC.
func (d *Dev) Range() (analog.Sample, analog.Sample) {
	return analog.Sample{Raw: -(1 << 23)}, analog.Sample{Raw: 1 << 23}
}

// Read implements analog.PinADC.
func (d *Dev) Read() (analog.Sample, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	raw, err := d.readRaw()
	return analog.Sample{Raw: raw}, err
}

// ReadContinuous starts reading values continuously from the ADC. It
// returns a channel that you can use to receive these values.
//
// You must call Halt to stop reading.
//
// Calling ReadContinuous again before Halt is an error,
// and nil will be returned.
func (d *Dev) ReadContinuous() <-chan analog.Sample {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.done != nil {
		return nil
	}
	done := make(chan struct{})
	ret := make(chan analog.Sample)

	go func() {
		for {
			select {
			case <-done:
				close(ret)
				return
			default:
				value, err := d.ReadTimeout(time.Second)
				if err == nil {
					ret <- analog.Sample{Raw: value}
				}
			}
		}
	}()

	d.done = done
	return ret
}

// Halt stops a continuous read that was started with ReadContinuous.
//
// This will close the channel that was returned by ReadContinuous.
func (d *Dev) Halt() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.done != nil {
		close(d.done)
		d.done = nil
	}
	return nil
}

// IsReady returns true if there is data ready to be read from the ADC.
func (d *Dev) IsReady() bool {
	return d.data.Read() == gpio.Low
}

// ReadTimeout reads a single value from the ADC.
//
// It blocks until the ADC indicates there is data ready for retrieval. If the
// ADC doesn't pull its Data pin low to indicate there is data ready before the
// timeout is reached, ErrTimeout is returned.
func (d *Dev) ReadTimeout(timeout time.Duration) (int32, error) {
	// Wait for the falling edge that indicates the ADC has data.
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.IsReady() {
		if !d.data.WaitForEdge(timeout) {
			return 0, ErrTimeout
		}
	}
	return d.readRaw()
}

func (d *Dev) readRaw() (int32, error) {
	// Shift the 24-bit 2's compliment value.
	var value uint32
	for i := 0; i < 24; i++ {
		if err := d.clk.Out(gpio.High); err != nil {
			return 0, err
		}
		level := d.data.Read()
		if err := d.clk.Out(gpio.Low); err != nil {
			return 0, err
		}

		value <<= 1
		if level {
			value |= 1
		}
	}

	// Pulse the clock 1-3 more times to set the new ADC mode.
	for i := 0; i < int(d.inputMode); i++ {
		if err := d.clk.Out(gpio.High); err != nil {
			return 0, err
		}
		if err := d.clk.Out(gpio.Low); err != nil {
			return 0, err
		}
	}
	// Convert the 24-bit 2's compliment value to a 32-bit signed value.
	return int32(value<<8) >> 8, nil
}

var _ analog.PinADC = &Dev{}
