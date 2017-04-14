// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpio defines digital pins.
//
// While all GPIO implementations are expected to implement PinIO, they may
// expose more specific functionality like PinPWM, PinDefaultPull, etc.
package gpio

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"periph.io/x/periph/conn/pin"
)

// Interfaces

// Level is the level of the pin: Low or High.
type Level bool

const (
	// Low represents 0v.
	Low Level = false
	// High represents Vin, generally 3.3v or 5v.
	High Level = true
)

func (l Level) String() string {
	if l == Low {
		return "Low"
	}
	return "High"
}

// Pull specifies the internal pull-up or pull-down for a pin set as input.
type Pull uint8

// Acceptable pull values.
const (
	Float        Pull = 0 // Let the input float
	PullDown     Pull = 1 // Apply pull-down
	PullUp       Pull = 2 // Apply pull-up
	PullNoChange Pull = 3 // Do not change the previous pull resistor setting or an unknown value
)

const pullName = "FloatPullDownPullUpPullNoChange"

var pullIndex = [...]uint8{0, 5, 13, 19, 31}

func (i Pull) String() string {
	if i >= Pull(len(pullIndex)-1) {
		return fmt.Sprintf("Pull(%d)", i)
	}
	return pullName[pullIndex[i]:pullIndex[i+1]]
}

// Edge specifies if an input pin should have edge detection enabled.
//
// Only enable it when needed, since this causes system interrupts.
type Edge int

// Acceptable edge detection values.
const (
	NoEdge      Edge = 0
	RisingEdge  Edge = 1
	FallingEdge Edge = 2
	BothEdges   Edge = 3
)

const edgeName = "NoEdgeRisingEdgeFallingEdgeBothEdges"

var edgeIndex = [...]uint8{0, 6, 16, 27, 36}

func (i Edge) String() string {
	if i >= Edge(len(edgeIndex)-1) {
		return fmt.Sprintf("Edge(%d)", i)
	}
	return edgeName[edgeIndex[i]:edgeIndex[i+1]]
}

const (
	// DutyMax is a duty cycle of 100%.
	DutyMax Duty = 65535
	// DutyHalf is a 50% duty PWM, which boils down to a normal clock.
	DutyHalf Duty = DutyMax / 2
)

// Duty is the duty cycle for a PWM.
//
// Valid values are between 0 and DutyMax.
type Duty int32

func (d Duty) String() string {
	// TODO(maruel): Implement one fractional number.
	return fmt.Sprintf("%d%%", (d+50)/(DutyMax/100))
}

// Valid returns true if the Duty cycle value is valid.
func (d Duty) Valid() bool {
	return d >= 0 && d <= DutyMax
}

// ParseDuty parses a string and converts it to a Duty value.
func ParseDuty(s string) (Duty, error) {
	percent := strings.HasSuffix(s, "%")
	if percent {
		s = s[:len(s)-1]
	}
	i64, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	i := Duty(i64)
	if percent {
		// TODO(maruel): Add support for fractional number.
		if i < 0 {
			return 0, errors.New("duty must be >= 0%")
		}
		if i > 100 {
			return 0, errors.New("duty must be <= 100%")
		}
		return ((i * DutyMax) + 49) / 100, nil
	}
	if i < 0 {
		return 0, errors.New("duty must be >= 0")
	}
	if i > DutyMax {
		return 0, fmt.Errorf("duty must be <= %d", DutyMax)
	}
	return i, nil
}

// PinPWM exposes hardware PWM.
//
// The driver may uses DMA controller underneath for zero CPU implementation.
type PinPWM interface {
	// PWM sets the PWM output on supported pins.
	//
	// To use as a general purpose clock, set duty to DutyHalf. Some pins may
	// only support DutyHalf and no other value.
	//
	// Using 0 as period will use the optimal value as supported/preferred by the
	// pin.
	PWM(duty Duty, period time.Duration) error
}

// PinIn is an input GPIO pin.
//
// It may optionally support internal pull resistor and edge based triggering.
type PinIn interface {
	pin.Pin
	// In setups a pin as an input.
	//
	// If WaitForEdge() is planned to be called, make sure to use one of the Edge
	// value. Otherwise, use None to not generated unneeded hardware interrupts.
	//
	// Calling In() will try to empty the accumulated edges but it cannot be 100%
	// reliable due to the OS (linux) and its driver. It is possible that on a
	// gpio that is as input, doing a quick Out(), In() may return an edge that
	// occurred before the Out() call.
	In(pull Pull, edge Edge) error
	// Read return the current pin level.
	//
	// Behavior is undefined if In() wasn't used before.
	//
	// In some rare case, it is possible that Read() fails silently. This happens
	// if another process on the host messes up with the pin after In() was
	// called. In this case, call In() again.
	Read() Level
	// WaitForEdge() waits for the next edge or immediately return if an edge
	// occurred since the last call.
	//
	// Only waits for the kind of edge as specified in a previous In() call.
	// Behavior is undefined if In() with a value other than None wasn't called
	// before.
	//
	// Returns true if an edge was detected during or before this call. Return
	// false if the timeout occurred or In() was called while waiting, causing the
	// function to exit.
	//
	// Multiple edges may or may not accumulate between two calls to
	// WaitForEdge(). The behavior in this case is undefined and is OS driver
	// specific.
	//
	// It is not required to call Read() to reset the edge detection.
	//
	// Specify -1 to effectively disable timeout.
	WaitForEdge(timeout time.Duration) bool
	// Pull returns the internal pull resistor if the pin is set as input pin.
	//
	// Returns PullNoChange if the value cannot be read.
	Pull() Pull
}

// PinOut is an output GPIO pin.
type PinOut interface {
	pin.Pin
	// Out sets a pin as output if it wasn't already and sets the initial value.
	//
	// After the initial call to ensure that the pin has been set as output, it
	// is generally safe to ignore the error returned.
	//
	// Out() tries to empty the accumulated edges detected if the gpio was
	// previously set as input but this is not 100% guaranteed due to the OS.
	Out(l Level) error
}

// PinIO is a GPIO pin that supports both input and output. It matches both
// interfaces PinIn and PinOut.
//
// A GPIO pin implementing PinIO may fail at either input or output or both.
//
// The GPIO pin may optionally support more interfaces, like PinPWM,
// PinDefaultPull.
type PinIO interface {
	pin.Pin
	// PinIn
	In(pull Pull, edge Edge) error
	Read() Level
	WaitForEdge(timeout time.Duration) bool
	Pull() Pull
	// PinOut
	Out(l Level) error
}

// PinDefaultPull is optionally implemented to return the default pull at boot
// time. This is useful to determine if the pin is acceptable for operation
// with certain devices.
type PinDefaultPull interface {
	// DefaultPull returns the pull that is initialized on CPU reset.
	DefaultPull() Pull
}

// INVALID implements PinIO and fails on all access.
var INVALID PinIO

// RealPin is implemented by aliased pin and allows the retrieval of the real
// pin underlying an alias.
//
// The purpose of the RealPin is to be able to cleanly test whether an arbitrary
// gpio.PinIO returned by ByName is really an alias for another pin.
type RealPin interface {
	Real() PinIO // Real returns the real pin behind an Alias
}

//

// errInvalidPin is returned when trying to use INVALID.
var errInvalidPin = errors.New("gpio: invalid pin")

func init() {
	INVALID = invalidPin{}
}

// invalidPin implements PinIO for compatibility but fails on all access.
type invalidPin struct {
}

func (invalidPin) Number() int {
	return -1
}

func (invalidPin) Name() string {
	return "INVALID"
}

func (invalidPin) String() string {
	return "INVALID"
}

func (invalidPin) Function() string {
	return ""
}

func (invalidPin) In(Pull, Edge) error {
	return errInvalidPin
}

func (invalidPin) Read() Level {
	return Low
}

func (invalidPin) WaitForEdge(timeout time.Duration) bool {
	return false
}

func (invalidPin) Pull() Pull {
	return PullNoChange
}

func (invalidPin) Out(Level) error {
	return errInvalidPin
}

var _ PinIn = INVALID
var _ PinOut = INVALID
var _ PinIO = INVALID
