// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpio defines digital pins.
//
// The GPIO pins are described in their logical functionality, not in their
// physical position.
package gpio

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/pio/conn/pins"
)

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
	Down         Pull = 1 // Apply pull-down
	Up           Pull = 2 // Apply pull-up
	PullNoChange Pull = 3 // Do not change the previous pull resistor setting or an unknown value
)

const pullName = "FloatDownUpPullNoChange"

var pullIndex = [...]uint8{0, 5, 9, 11, 23}

func (i Pull) String() string {
	if i >= Pull(len(pullIndex)-1) {
		return fmt.Sprintf("Pull(%d)", i)
	}
	return pullName[pullIndex[i]:pullIndex[i+1]]
}

// Edge specifies if an input pin should have edge detection enabled.
//
// Only enable it when needed, since this causes system interrupts.
type Edge uint8

// Acceptable edge detection values.
const (
	None    Edge = 0
	Rising  Edge = 1
	Falling Edge = 2
	Both    Edge = 3
)

const edgeName = "NoneRisingFallingBoth"

var edgeIndex = [...]uint8{0, 4, 10, 17, 21}

func (i Edge) String() string {
	if i >= Edge(len(edgeIndex)-1) {
		return fmt.Sprintf("Edge(%d)", i)
	}
	return edgeName[edgeIndex[i]:edgeIndex[i+1]]
}

// PinIn is an input GPIO pin.
//
// It may optionally support internal pull resistor and edge based triggering.
type PinIn interface {
	pins.Pin
	// In setups a pin as an input.
	//
	// If WaitForEdge() is planned to be called, make sure to use one of the Edge
	// value. Otherwise, use None to not generated unneeded hardware interrupts.
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
	// occured since the last call.
	//
	// Only waits for the kind of edge as specified in a previous In() call.
	// Behavior is undefined if In() with a value other than None wasn't called
	// before.
	//
	// Returns true if an edge was detected during or before this call. Return
	// false if the timeout occured or In() was called while waiting, causing the
	// function to exit.
	//
	// Multiple edges may or may not accumulate between two calls to
	// WaitForEdge(). The behavior in this case is undefined.
	//
	// It is not required to call Read() to reset the edge detection.
	//
	// Specify -1 to effectively disable timeout.
	WaitForEdge(timeout time.Duration) bool
	// Pull returns the internal pull resistor if the pin is set as input pin.
	// Returns PullNoChange if the value cannot be read.
	Pull() Pull
}

const (
	// Max is the PWM fully at high. One should use Out(High) instead.
	Max = 65536
	// Half is a 50% PWM duty cycle.
	Half = Max / 2
)

// PinOut is an output GPIO pin.
type PinOut interface {
	pins.Pin
	// Out sets a pin as output if it wasn't already and sets the initial value.
	//
	// After the initial call to ensure that the pin has been set as output, it
	// is generally safe to ignore the error returned.
	Out(l Level) error
	// PWM sets a pin as output with a specified duty cycle between 0 and Max.
	//
	// The pin should use the highest frequency it can use.
	//
	// Use Half for a 50% duty cycle.
	PWM(duty int) error
}

// PinIO is a GPIO pin that supports both input and output.
//
// It may fail at either input and or output, for example ground, vcc and other
// similar pins.
type PinIO interface {
	pins.Pin
	In(pull Pull, edge Edge) error
	Read() Level
	WaitForEdge(timeout time.Duration) bool
	Pull() Pull
	Out(l Level) error
	PWM(duty int) error
}

// INVALID implements PinIO and fails on all access.
var INVALID PinIO = invalidPin{}

// BasicPin implements Pin as a non-functional pin.
type BasicPin struct {
	Name string
}

func (b *BasicPin) String() string {
	return b.Name
}

// Number implements pins.Pin.
func (b *BasicPin) Number() int {
	return -1
}

// Function implements pins.Pin.
func (b *BasicPin) Function() string {
	return ""
}

// In implements gpio.PinIn.
func (b *BasicPin) In(Pull, Edge) error {
	return fmt.Errorf("%s cannot be used as input", b.Name)
}

// Read implements gpio.PinIn.
func (b *BasicPin) Read() Level {
	return Low
}

// WaitForEdge implements gpio.PinIn.
func (b *BasicPin) WaitForEdge(timeout time.Duration) bool {
	return false
}

// Pull implements gpio.PinIn.
func (b *BasicPin) Pull() Pull {
	return PullNoChange
}

// Out implements gpio.PinOut.
func (b *BasicPin) Out(Level) error {
	return fmt.Errorf("%s cannot be used as output", b.Name)
}

// PWM implements gpio.PinOut.
func (b *BasicPin) PWM(duty int) error {
	return fmt.Errorf("%s cannot be used as PWM", b.Name)
}

//

// ByNumber returns a GPIO pin from its number.
//
// Returns nil in case the pin is not present.
func ByNumber(number int) PinIO {
	lock.Lock()
	defer lock.Unlock()
	pin, _ := byNumber[number]
	return pin
}

// ByName returns a GPIO pin from its name.
//
// This can be strings like GPIO2, PB8, etc.
//
// Returns nil in case the pin is not present.
func ByName(name string) PinIO {
	lock.Lock()
	defer lock.Unlock()
	pin, _ := byName[name]
	return pin
}

// ByFunction returns a GPIO pin from its function.
//
// This can be strings like I2C1_SDA, SPI0_MOSI, etc.
//
// Returns nil in case there is no pin setup with this function.
func ByFunction(fn string) PinIO {
	lock.Lock()
	defer lock.Unlock()
	pin, _ := byFunction[fn]
	return pin
}

// All returns all the GPIO pins available on this host.
//
// The list is guaranteed to be in order of number.
//
// This list excludes non-GPIO pins like GROUND, V3_3, etc.
func All() []PinIO {
	lock.Lock()
	defer lock.Unlock()
	out := make(pinList, 0, len(byNumber))
	for _, p := range byNumber {
		out = append(out, p)
	}
	sort.Sort(out)
	return out
}

// Functional returns a map of all pins implementing hardware provided
// special functionality, like I²C, SPI, ADC.
func Functional() map[string]PinIO {
	lock.Lock()
	defer lock.Unlock()
	out := make(map[string]PinIO, len(byFunction))
	for k, v := range byFunction {
		out[k] = v
	}
	return out
}

// Register registers a GPIO pin.
//
// Registering the same pin number or name twice is an error.
func Register(pin PinIO) error {
	lock.Lock()
	defer lock.Unlock()
	number := pin.Number()
	if _, ok := byNumber[number]; ok {
		return fmt.Errorf("registering the same pin %d twice", number)
	}
	name := pin.String()
	if _, ok := byName[name]; ok {
		return fmt.Errorf("registering the same pin %s twice", name)
	}

	byNumber[number] = pin
	byName[name] = pin
	return nil
}

// Unregister removes a previously registered pin.
//
// This can happen when a pin is exposed via an USB device and the device is
// unplugged.
func Unregister(name string, number int, function string) error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := byName[name]; !ok {
		return errors.New("unknown name")
	}
	if _, ok := byNumber[number]; !ok {
		return errors.New("unknown number")
	}
	if function != "" {
		if _, ok := byFunction[function]; !ok {
			return errors.New("unknown function")
		}
	}

	delete(byName, name)
	delete(byNumber, number)
	if function != "" {
		delete(byFunction, function)
	}
	return nil
}

// MapFunction registers a GPIO pin for a specific function.
func MapFunction(function string, pin PinIO) {
	lock.Lock()
	defer lock.Unlock()
	byFunction[function] = pin
}

//

// errInvalidPin is returned when trying to use INVALID.
var errInvalidPin = errors.New("invalid pin")

// invalidPin implements PinIO for compability but fails on all access.
type invalidPin struct {
}

func (invalidPin) Number() int {
	return -1
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

func (invalidPin) PWM(duty int) error {
	return errInvalidPin
}

var (
	lock       sync.Mutex
	byNumber   = map[int]PinIO{}
	byName     = map[string]PinIO{}
	byFunction = map[string]PinIO{}
)

type pinList []PinIO

func (p pinList) Len() int           { return len(p) }
func (p pinList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pinList) Less(i, j int) bool { return p[i].Number() < p[j].Number() }

var _ PinIn = INVALID
var _ PinOut = INVALID
var _ PinIO = INVALID
