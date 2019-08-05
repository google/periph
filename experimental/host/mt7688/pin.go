// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mt7688

import (
	"errors"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/host/sysfs"
)

// function specifies the active functionality of a pin. The alternative
// function is GPIO pin dependent.
type function uint8

// Each pin can have one of 5 functions.
const (
	in  function = 0
	out function = 1
	// TODO: work out how to change pin functions
	alt0 function = 2
	alt1 function = 3
	alt2 function = 4
)

// cpuPins are all the pins as supported by the CPU. There is no guarantee that
// they are actually connected to anything on the board.
var cpuPins = []Pin{
	// TODO: discover default pull
	{number: 0, name: "GPIO0", defaultPull: gpio.Float},
	{number: 1, name: "GPIO1", defaultPull: gpio.Float},
	{number: 2, name: "GPIO2", defaultPull: gpio.Float},
	{number: 3, name: "GPIO3", defaultPull: gpio.Float},
	{number: 4, name: "GPIO4", defaultPull: gpio.Float},
	{number: 5, name: "GPIO5", defaultPull: gpio.Float},
	{number: 6, name: "GPIO6", defaultPull: gpio.Float},
	{number: 7, name: "GPIO7", defaultPull: gpio.Float},
	{number: 8, name: "GPIO8", defaultPull: gpio.Float},
	{number: 9, name: "GPIO9", defaultPull: gpio.Float},
	{number: 10, name: "GPIO10", defaultPull: gpio.Float},
	{number: 11, name: "GPIO11", defaultPull: gpio.Float},
	{number: 12, name: "GPIO12", defaultPull: gpio.Float},
	{number: 13, name: "GPIO13", defaultPull: gpio.Float},
	{number: 14, name: "GPIO14", defaultPull: gpio.Float},
	{number: 15, name: "GPIO15", defaultPull: gpio.Float},
	{number: 16, name: "GPIO16", defaultPull: gpio.Float},
	{number: 17, name: "GPIO17", defaultPull: gpio.Float},
	{number: 18, name: "GPIO18", defaultPull: gpio.Float},
	{number: 19, name: "GPIO19", defaultPull: gpio.Float},
	{number: 20, name: "GPIO20", defaultPull: gpio.Float},
	{number: 21, name: "GPIO21", defaultPull: gpio.Float},
	{number: 22, name: "GPIO22", defaultPull: gpio.Float},
	{number: 23, name: "GPIO23", defaultPull: gpio.Float},
	{number: 24, name: "GPIO24", defaultPull: gpio.Float},
	{number: 25, name: "GPIO25", defaultPull: gpio.Float},
	{number: 26, name: "GPIO26", defaultPull: gpio.Float},
	{number: 27, name: "GPIO27", defaultPull: gpio.Float},
	{number: 28, name: "GPIO28", defaultPull: gpio.Float},
	{number: 29, name: "GPIO29", defaultPull: gpio.Float},
	{number: 30, name: "GPIO30", defaultPull: gpio.Float},
	{number: 31, name: "GPIO31", defaultPull: gpio.Float},
	{number: 32, name: "GPIO32", defaultPull: gpio.Float},
	{number: 33, name: "GPIO33", defaultPull: gpio.Float},
	{number: 34, name: "GPIO34", defaultPull: gpio.Float},
	{number: 35, name: "GPIO35", defaultPull: gpio.Float},
	{number: 36, name: "GPIO36", defaultPull: gpio.Float},
	{number: 37, name: "GPIO37", defaultPull: gpio.Float},
	{number: 38, name: "GPIO38", defaultPull: gpio.Float},
	{number: 39, name: "GPIO39", defaultPull: gpio.Float},
	{number: 40, name: "GPIO40", defaultPull: gpio.Float},
	{number: 41, name: "GPIO41", defaultPull: gpio.Float},
	{number: 42, name: "GPIO42", defaultPull: gpio.Float},
	{number: 43, name: "GPIO43", defaultPull: gpio.Float},
	{number: 44, name: "GPIO44", defaultPull: gpio.Float},
	{number: 45, name: "GPIO45", defaultPull: gpio.Float},
	{number: 46, name: "GPIO46", defaultPull: gpio.Float},
}

// Pin is a GPIO number (GPIOnn) on MT7688(AN/KN).
type Pin struct {
	// Immutable.
	number      int
	name        string
	defaultPull gpio.Pull // Default pull at system boot, as per datasheet.

	// Immutable after driver initialization.
	sysfsPin *sysfs.Pin // Set to the corresponding sysfs.Pin, if any.
}

// String implements conn.Resource.
func (p Pin) String() string {
	return p.name
}

// Halt implements conn.Resource.
func (*Pin) Halt() error {
	return nil
}

// Name implements pin.Pin.
func (p *Pin) Name() string {
	return p.name
}

// Number implements pin.Pin.
//
// This is the GPIO number, not the pin number on a header.
func (p *Pin) Number() int {
	return p.number
}

// Function implements pin.Pin.
func (p *Pin) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *Pin) Func() pin.Func {
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return pin.Func("ERR")
		}
		return p.sysfsPin.Func()
	}
	switch f := p.function(); f {
	case in:
		// TODO: implement FastRead
		//if p.FastRead() {
		//	return gpio.IN_HIGH
		//}
		return gpio.IN_LOW
	case out:
		//if p.FastRead() {
		//	return gpio.OUT_HIGH
		//}
		return gpio.OUT_LOW
	case alt0:
		if s := mapping[p.number][0]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT0")
	case alt1:
		if s := mapping[p.number][1]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT1")
	case alt2:
		if s := mapping[p.number][2]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT2")
	default:
		return pin.Func("ERR")
	}
}

// SupportedFuncs implements pin.PinFunc.
//
// Not fully implemented yet.
func (p *Pin) SupportedFuncs() []pin.Func {
	return []pin.Func{gpio.IN, gpio.OUT}
}

// SetFunc implements pin.PinFunc.
//
// Not implemented yet.
func (p *Pin) SetFunc(pin.Func) error {
	return errors.New("not implemented")
}

// In implements gpio.PinIn.
//
// Not implemented yet.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	return errors.New("not implemented")
}

// Read implements gpio.PinIn.
//
// Not implemented yet.
func (p *Pin) Read() gpio.Level {
	return gpio.Low
}

// WaitForEdge implements gpio.PinIn.
//
// Not implemented yet.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	return false
}

// Pull implements gpio.PinIn.
//
// Not implemented yet.
func (p *Pin) Pull() gpio.Pull {
	return p.defaultPull
}

// DefaultPull implements gpio.PinIn.
func (p *Pin) DefaultPull() gpio.Pull {
	return p.defaultPull
}

// Out implements gpio.PinOut.
//
// Not implemented yet.
func (p *Pin) Out(l gpio.Level) error {
	return errors.New("not implemented")
}

// PWM implements gpio.PinOut.
//
// Not implemented yet.
func (p *Pin) PWM(duty gpio.Duty, f physic.Frequency) error {
	return errors.New("not implemented")
}

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	// TODO: implement function
	return out
}

var _ gpio.PinIO = &Pin{}
var _ pin.PinFunc = &Pin{}
