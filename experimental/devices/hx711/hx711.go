// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package hx711 implements an interface to the HX711 analog to digital converter.
//
// Datasheet: https://www.mouser.com/ds/2/813/hx711_english-1022875.pdf
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
//
// Changing the InputMode on an HX711 will only take effect *after* the next
// read.
type InputMode int

const (
	CHANNEL_A_GAIN_128 InputMode = 1
	CHANNEL_A_GAIN_64  InputMode = 3
	CHANNEL_B_GAIN_32  InputMode = 2

	dataBits         = 24
	readPollInterval = 50 * time.Millisecond
)

type HX711 struct {
	InputMode InputMode

	clk     gpio.PinOut
	data    gpio.PinIn
	useEdge bool
	done    chan struct{}
}

// New creates a new HX711 device.
func New(clk gpio.PinOut, data gpio.PinIn) (*HX711, error) {
	// Try enabling edge detection on the data pin.
	var useEdge bool
	if err := data.In(gpio.PullDown, gpio.FallingEdge); err != nil {
		if err := data.In(gpio.PullDown, gpio.NoEdge); err != nil {
			return nil, err
		}
		useEdge = false
	} else {
		useEdge = true
	}

	clk.Out(gpio.Low)

	return &HX711{
		InputMode: CHANNEL_A_GAIN_128,
		clk:       clk,
		data:      data,
		useEdge:   useEdge,
		done:      nil,
	}, nil
}

// IsReady returns true if there is data ready to be read from the ADC.
func (d *HX711) IsReady() bool {
	return d.data.Read() == gpio.Low
}

// Read reads a single value from the ADC.  It blocks until the ADC indicates
// there is data ready for retrieval.  If the ADC doesn't pull its Data pin low
// to indicate there is data ready before the timeout is reached, TimeoutError
// is returned.
func (d *HX711) Read(timeout time.Duration) (int32, error) {
	if d.useEdge {
		// If the clock pin supports edge detection, wait for the falling edge that
		// indicates the ADC has data.
		if !d.IsReady() {
			if !d.data.WaitForEdge(timeout) {
				return 0, TimeoutError
			}
		}
	} else {
		// If the clock pin doesn't support edge detection just poll every few
		// milliseconds.
		startTime := time.Now()
		for !d.IsReady() {
			if time.Now().Sub(startTime) > timeout {
				return 0, TimeoutError
			}
			time.Sleep(readPollInterval)
		}
	}

	// Shift the 24-bit 2's compliment value.
	var value uint32
	for i := 0; i < dataBits; i++ {
		d.clk.Out(gpio.High)
		level := d.data.Read()
		d.clk.Out(gpio.Low)

		if level {
			value = (value << 1) | 1
		} else {
			value = (value << 1)
		}
	}

	// Convert the 24-bit value to a 32-bit value by moving the MSB and padding
	// the top byte with 1s if the value was negative.
	if value&0x00800000 != 0 {
		value |= 0xFF000000
	}

	// Pulse the clock 1-3 more times to set the new ADC mode.
	for i := 0; i < int(d.InputMode); i++ {
		d.clk.Out(gpio.High)
		d.clk.Out(gpio.Low)
	}
	return int32(value), nil
}

// StartContinuousRead starts reading values continuously from the ADC.  It
// returns a channel that you can use to receive these values.
//
// You must call StopContinuousRead to stop reading.
//
// Calling StartContinuousRead again before StopContinuousRead is an error,
// and nil will be returned.
func (d *HX711) StartContinuousRead() <-chan int32 {
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
				d.done = nil
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

// StopContinuousRead stops a continuous read that was started with
// StartContinuousRead.  This will close the channel that was returned by
// StartContinuousRead.
func (d *HX711) StopContinuousRead() {
	close(d.done)
}

// ReadAveraged reads several samples from the ADC and returns the mean value.
func (d *HX711) ReadAveraged(timeout time.Duration, samples int) (int32, error) {
	var sum int64
	for i := 0; i < samples; i++ {
		value, err := d.Read(timeout)
		if err != nil {
			return 0, err
		}
		sum += int64(value)
	}
	return int32(sum / int64(samples)), nil
}
