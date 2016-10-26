// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains the definitions of all possible generic Allwinner pins and their
// implementation using a combination of sysfs and memory-mapped I/O.

package allwinner

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/host/sysfs"
)

// cpupins that may be implemented by a generic Allwinner CPU. Not all pins will be present on all
// models and even if the CPU model supports them they may not be connected to anything on the
// board. The net effect is that it may look like more pins are available than really are, but
// trying to get the pin list 100% correct on all platforms seems futile, hence periph errs on the
// side of caution.
//
// Group/offset calculation from http://forum.pine64.org/showthread.php?tid=474
var cpupins = map[string]*Pin{
	"PB0":  {group: 1, offset: 0, name: "PB0", defaultPull: gpio.Float},
	"PB1":  {group: 1, offset: 1, name: "PB1", defaultPull: gpio.Float},
	"PB2":  {group: 1, offset: 2, name: "PB2", defaultPull: gpio.Float},
	"PB3":  {group: 1, offset: 3, name: "PB3", defaultPull: gpio.Float},
	"PB4":  {group: 1, offset: 4, name: "PB4", defaultPull: gpio.Float},
	"PB5":  {group: 1, offset: 5, name: "PB5", defaultPull: gpio.Float},
	"PB6":  {group: 1, offset: 6, name: "PB6", defaultPull: gpio.Float},
	"PB7":  {group: 1, offset: 7, name: "PB7", defaultPull: gpio.Float},
	"PB8":  {group: 1, offset: 8, name: "PB8", defaultPull: gpio.Float},
	"PB9":  {group: 1, offset: 9, name: "PB9", defaultPull: gpio.Float},
	"PB10": {group: 1, offset: 10, name: "PB10", defaultPull: gpio.Float},
	"PB11": {group: 1, offset: 11, name: "PB11", defaultPull: gpio.Float},
	"PB12": {group: 1, offset: 12, name: "PB12", defaultPull: gpio.Float},
	"PB13": {group: 1, offset: 13, name: "PB13", defaultPull: gpio.Float},
	"PB14": {group: 1, offset: 14, name: "PB14", defaultPull: gpio.Float},
	"PB15": {group: 1, offset: 15, name: "PB15", defaultPull: gpio.Float},
	"PB16": {group: 1, offset: 16, name: "PB16", defaultPull: gpio.Float},
	"PB17": {group: 1, offset: 17, name: "PB17", defaultPull: gpio.Float},
	"PB18": {group: 1, offset: 18, name: "PB18", defaultPull: gpio.Float},
	"PC0":  {group: 2, offset: 0, name: "PC0", defaultPull: gpio.Float},
	"PC1":  {group: 2, offset: 1, name: "PC1", defaultPull: gpio.Float},
	"PC2":  {group: 2, offset: 2, name: "PC2", defaultPull: gpio.Float},
	"PC3":  {group: 2, offset: 3, name: "PC3", defaultPull: gpio.Up},
	"PC4":  {group: 2, offset: 4, name: "PC4", defaultPull: gpio.Up},
	"PC5":  {group: 2, offset: 5, name: "PC5", defaultPull: gpio.Float},
	"PC6":  {group: 2, offset: 6, name: "PC6", defaultPull: gpio.Up},
	"PC7":  {group: 2, offset: 7, name: "PC7", defaultPull: gpio.Up},
	"PC8":  {group: 2, offset: 8, name: "PC8", defaultPull: gpio.Float},
	"PC9":  {group: 2, offset: 9, name: "PC9", defaultPull: gpio.Float},
	"PC10": {group: 2, offset: 10, name: "PC10", defaultPull: gpio.Float},
	"PC11": {group: 2, offset: 11, name: "PC11", defaultPull: gpio.Float},
	"PC12": {group: 2, offset: 12, name: "PC12", defaultPull: gpio.Float},
	"PC13": {group: 2, offset: 13, name: "PC13", defaultPull: gpio.Float},
	"PC14": {group: 2, offset: 14, name: "PC14", defaultPull: gpio.Float},
	"PC15": {group: 2, offset: 15, name: "PC15", defaultPull: gpio.Float},
	"PC16": {group: 2, offset: 16, name: "PC16", defaultPull: gpio.Float},
	"PD0":  {group: 3, offset: 0, name: "PD0", defaultPull: gpio.Float},
	"PD1":  {group: 3, offset: 1, name: "PD1", defaultPull: gpio.Float},
	"PD2":  {group: 3, offset: 2, name: "PD2", defaultPull: gpio.Float},
	"PD3":  {group: 3, offset: 3, name: "PD3", defaultPull: gpio.Float},
	"PD4":  {group: 3, offset: 4, name: "PD4", defaultPull: gpio.Float},
	"PD5":  {group: 3, offset: 5, name: "PD5", defaultPull: gpio.Float},
	"PD6":  {group: 3, offset: 6, name: "PD6", defaultPull: gpio.Float},
	"PD7":  {group: 3, offset: 7, name: "PD7", defaultPull: gpio.Float},
	"PD8":  {group: 3, offset: 8, name: "PD8", defaultPull: gpio.Float},
	"PD9":  {group: 3, offset: 9, name: "PD9", defaultPull: gpio.Float},
	"PD10": {group: 3, offset: 10, name: "PD10", defaultPull: gpio.Float},
	"PD11": {group: 3, offset: 11, name: "PD11", defaultPull: gpio.Float},
	"PD12": {group: 3, offset: 12, name: "PD12", defaultPull: gpio.Float},
	"PD13": {group: 3, offset: 13, name: "PD13", defaultPull: gpio.Float},
	"PD14": {group: 3, offset: 14, name: "PD14", defaultPull: gpio.Float},
	"PD15": {group: 3, offset: 15, name: "PD15", defaultPull: gpio.Float},
	"PD16": {group: 3, offset: 16, name: "PD16", defaultPull: gpio.Float},
	"PD17": {group: 3, offset: 17, name: "PD17", defaultPull: gpio.Float},
	"PD18": {group: 3, offset: 18, name: "PD18", defaultPull: gpio.Float},
	"PD19": {group: 3, offset: 19, name: "PD19", defaultPull: gpio.Float},
	"PD20": {group: 3, offset: 20, name: "PD20", defaultPull: gpio.Float},
	"PD21": {group: 3, offset: 21, name: "PD21", defaultPull: gpio.Float},
	"PD22": {group: 3, offset: 22, name: "PD22", defaultPull: gpio.Float},
	"PD23": {group: 3, offset: 23, name: "PD23", defaultPull: gpio.Float},
	"PD24": {group: 3, offset: 24, name: "PD24", defaultPull: gpio.Float},
	"PD25": {group: 3, offset: 25, name: "PD25", defaultPull: gpio.Float},
	"PD26": {group: 3, offset: 26, name: "PD26", defaultPull: gpio.Float},
	"PD27": {group: 3, offset: 27, name: "PD27", defaultPull: gpio.Float},
	"PE0":  {group: 4, offset: 0, name: "PE0", defaultPull: gpio.Float},
	"PE1":  {group: 4, offset: 1, name: "PE1", defaultPull: gpio.Float},
	"PE2":  {group: 4, offset: 2, name: "PE2", defaultPull: gpio.Float},
	"PE3":  {group: 4, offset: 3, name: "PE3", defaultPull: gpio.Float},
	"PE4":  {group: 4, offset: 4, name: "PE4", defaultPull: gpio.Float},
	"PE5":  {group: 4, offset: 5, name: "PE5", defaultPull: gpio.Float},
	"PE6":  {group: 4, offset: 6, name: "PE6", defaultPull: gpio.Float},
	"PE7":  {group: 4, offset: 7, name: "PE7", defaultPull: gpio.Float},
	"PE8":  {group: 4, offset: 8, name: "PE8", defaultPull: gpio.Float},
	"PE9":  {group: 4, offset: 9, name: "PE9", defaultPull: gpio.Float},
	"PE10": {group: 4, offset: 10, name: "PE10", defaultPull: gpio.Float},
	"PE11": {group: 4, offset: 11, name: "PE11", defaultPull: gpio.Float},
	"PE12": {group: 4, offset: 12, name: "PE12", defaultPull: gpio.Float},
	"PE13": {group: 4, offset: 13, name: "PE13", defaultPull: gpio.Float},
	"PE14": {group: 4, offset: 14, name: "PE14", defaultPull: gpio.Float},
	"PE15": {group: 4, offset: 15, name: "PE15", defaultPull: gpio.Float},
	"PE16": {group: 4, offset: 16, name: "PE16", defaultPull: gpio.Float},
	"PE17": {group: 4, offset: 17, name: "PE17", defaultPull: gpio.Float},
	"PF0":  {group: 5, offset: 0, name: "PF0", defaultPull: gpio.Float},
	"PF1":  {group: 5, offset: 1, name: "PF1", defaultPull: gpio.Float},
	"PF2":  {group: 5, offset: 2, name: "PF2", defaultPull: gpio.Float},
	"PF3":  {group: 5, offset: 3, name: "PF3", defaultPull: gpio.Float},
	"PF4":  {group: 5, offset: 4, name: "PF4", defaultPull: gpio.Float},
	"PF5":  {group: 5, offset: 5, name: "PF5", defaultPull: gpio.Float},
	"PF6":  {group: 5, offset: 6, name: "PF6", defaultPull: gpio.Float},
	"PG0":  {group: 6, offset: 0, name: "PG0", defaultPull: gpio.Float},
	"PG1":  {group: 6, offset: 1, name: "PG1", defaultPull: gpio.Float},
	"PG2":  {group: 6, offset: 2, name: "PG2", defaultPull: gpio.Float},
	"PG3":  {group: 6, offset: 3, name: "PG3", defaultPull: gpio.Float},
	"PG4":  {group: 6, offset: 4, name: "PG4", defaultPull: gpio.Float},
	"PG5":  {group: 6, offset: 5, name: "PG5", defaultPull: gpio.Float},
	"PG6":  {group: 6, offset: 6, name: "PG6", defaultPull: gpio.Float},
	"PG7":  {group: 6, offset: 7, name: "PG7", defaultPull: gpio.Float},
	"PG8":  {group: 6, offset: 8, name: "PG8", defaultPull: gpio.Float},
	"PG9":  {group: 6, offset: 9, name: "PG9", defaultPull: gpio.Float},
	"PG10": {group: 6, offset: 10, name: "PG10", defaultPull: gpio.Float},
	"PG11": {group: 6, offset: 11, name: "PG11", defaultPull: gpio.Float},
	"PG12": {group: 6, offset: 12, name: "PG12", defaultPull: gpio.Float},
	"PG13": {group: 6, offset: 13, name: "PG13", defaultPull: gpio.Float},
	"PH0":  {group: 7, offset: 0, name: "PH0", defaultPull: gpio.Float},
	"PH1":  {group: 7, offset: 1, name: "PH1", defaultPull: gpio.Float},
	"PH2":  {group: 7, offset: 2, name: "PH2", defaultPull: gpio.Float},
	"PH3":  {group: 7, offset: 3, name: "PH3", defaultPull: gpio.Float},
	"PH4":  {group: 7, offset: 4, name: "PH4", defaultPull: gpio.Float},
	"PH5":  {group: 7, offset: 5, name: "PH5", defaultPull: gpio.Float},
	"PH6":  {group: 7, offset: 6, name: "PH6", defaultPull: gpio.Float},
	"PH7":  {group: 7, offset: 7, name: "PH7", defaultPull: gpio.Float},
	"PH8":  {group: 7, offset: 8, name: "PH8", defaultPull: gpio.Float},
	"PH9":  {group: 7, offset: 9, name: "PH9", defaultPull: gpio.Float},
	"PH10": {group: 7, offset: 10, name: "PH10", defaultPull: gpio.Float},
	"PH11": {group: 7, offset: 11, name: "PH11", defaultPull: gpio.Float},
}

// ===== PinIO implementation.
// Page 73 for memory mapping overview.
// Page 194 for PWM.
// Page 230 for crypto engine.
// Page 278 audio including ADC.
// Page 376 GPIO PB to PH
// Page 410 GPIO PL
// Page 536 I²C (I2C)
// Page 545 SPI
// Page 560 UART
// Page 621 I2S/PCM

// Pin implements the gpio.PinIO interface for generic Allwinner CPU pins using memory mapping
// for gpio in/out functionality.
type Pin struct {
	group       uint8      // as per register offset calculation
	offset      uint8      // as per register offset calculation
	name        string     // name as per datasheet
	defaultPull gpio.Pull  // default pull at startup
	altFunc     [5]string  // alternate functions
	isOut       bool       // whether the pin is currently an output
	edge        *sysfs.Pin // mutable, set once, then never set back to nil
}

// Number returns the GPIO pin number as represented by gpio sysfs.
func (p *Pin) Number() int {
	if p == nil {
		return -1
	}
	return int(p.group)*32 + int(p.offset)
}

// Name returns the pin name.
func (p *Pin) Name() string {
	if p == nil {
		return ""
	}
	return p.name
}

// String returns the name of the pin in the processor and the GPIO pin number.
func (p *Pin) String() string {
	if p == nil {
		return "INVALID"
	}
	return fmt.Sprintf("%s(%d)", p.name, p.Number())
}

// Function returns the current function of the pin in printable form.
func (p *Pin) Function() string {
	if p == nil {
		return ""
	}
	switch f := p.function(); f {
	case in:
		return "In/" + p.Read().String() + "/" + p.Pull().String()
	case out:
		return "Out/" + p.Read().String()
	case alt1:
		if p.altFunc[0] != "" {
			return p.altFunc[0]
		}
		return "<Alt1>"
	case alt2:
		if p.altFunc[1] != "" {
			return p.altFunc[1]
		}
		return "<Alt2>"
	case alt3:
		if p.altFunc[2] != "" {
			return p.altFunc[2]
		}
		return "<Alt3>"
	case alt4:
		if p.altFunc[3] != "" {
			return p.altFunc[3]
		}
		return "<Alt4>"
	case alt5:
		if p.altFunc[4] != "" {
			return p.altFunc[4]
		}
		return "<Alt5>"
	case disabled:
		return "<Disabled>"
	default:
		return "<Internal error>"
	}
}

// In sets the pin direction to input and optionally enables a pull-up/down resistor as well as edge
// detection. Not all pins support edge detection on Allwinner processors!
//
// Edge detection requires opening a gpio sysfs file handle. The pin will be
// exported at /sys/class/gpio/gpio*/. Note that the pin will not be unexported
// at shutdown.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	if gpioMemory == nil {
		return errors.New("subsystem not initialized")
	}
	if !p.setFunction(in) {
		return fmt.Errorf("failed to set pin %s as input", p.name)
	}
	if pull != gpio.PullNoChange {
		off := p.offset / 16
		shift := 2 * (p.offset % 16)
		// Do it in a way that is concurrency safe.
		gpioMemory.groups[p.group].pull[off] &^= 3 << shift
		switch pull {
		case gpio.Down:
			gpioMemory.groups[p.group].pull[off] = 2 << shift
		case gpio.Up:
			gpioMemory.groups[p.group].pull[off] = 1 << shift
		default:
		}
	}
	if edge != gpio.None {
		switch p.group {
		case 1, 6, 7:
			// TODO(maruel): Some pins do not support Alt5 in these groups.
		default:
			return errors.New("only groups PB, PG, PH (and PL if available) support edge based triggering")
		}
		if p.edge == nil {
			ok := false
			if p.edge, ok = sysfs.Pins[p.Number()]; !ok {
				return fmt.Errorf("pin %s is not exported by sysfs", p)
			}
		}
	}
	if p.edge != nil {
		if err := p.edge.In(gpio.PullNoChange, edge); err != nil {
			return err
		}
	}
	return nil
}

// Read returns the current level of the pin. Due to the way the Allwinner hardware functions it
// will do this regardless of the pin's function but this should not be relied upon.
func (p *Pin) Read() gpio.Level {
	if p == nil {
		return gpio.Low
	}
	return gpio.Level(gpioMemory.groups[p.group].data&(1<<p.offset) != 0)
}

// WaitForEdge waits for an edge as previously set using In() or the expiration of a timeout.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	if p != nil && p.edge != nil {
		return p.edge.WaitForEdge(timeout)
	}
	return false
}

// Pull returns the current pull-up/down registor setting.
func (p *Pin) Pull() gpio.Pull {
	if p == nil {
		return gpio.PullNoChange
	}
	v := gpioMemory.groups[p.group].pull[p.offset/16]
	switch (v >> (2 * (p.offset % 16))) & 3 {
	case 0:
		return gpio.Float
	case 1:
		return gpio.Up
	case 2:
		return gpio.Down
	default:
		// Confused.
		return gpio.PullNoChange
	}
}

// Out ensures that the pin is configured as an output and outputs the value.
func (p *Pin) Out(l gpio.Level) error {
	if gpioMemory == nil {
		return errors.New("subsystem not initialized")
	}
	if !(p.isOut || p.setFunction(out)) {
		return fmt.Errorf("failed to set pin %s as output", p.name)
	}
	// TODO(maruel): Set the value *before* changing the pin to be an output, so
	// there is no glitch.
	bit := uint32(1 << p.offset)
	// Pn_DAT  n*0x24+0x10  Port n Data Register (n from 1(B) to 7(H))
	if l {
		gpioMemory.groups[p.group].data |= bit
	} else {
		gpioMemory.groups[p.group].data &^= bit
	}
	return nil
}

// PWM is not supported.
func (p *Pin) PWM(duty int) error {
	return errors.New("pwm is not supported")
}

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	if gpioMemory == nil {
		return disabled
	}
	shift := 4 * (p.offset % 8)
	return function((gpioMemory.groups[p.group].cfg[p.offset/8] >> shift) & 7)
}

// setFunction changes the GPIO pin function to in or out if the current function is in, out or
// alt5. It returns false if the function could not be changed.
func (p *Pin) setFunction(f function) bool {
	if f != in && f != out {
		return false
	}
	if p == nil {
		return false
	}
	// Interrupt based edge triggering is Alt5 but this is only supported on some pins.
	// TODO(maruel): This check should use a whitelist of pins.
	if actual := p.function(); actual != in && actual != out && actual != disabled && actual != alt5 {
		// Pin is in special mode.
		return false
	}
	off := p.offset / 8
	shift := 4 * (p.offset % 8)
	mask := uint32(disabled) << shift
	v := (uint32(f) << shift) ^ mask
	// First disable, then setup. This is concurrent safe.
	gpioMemory.groups[p.group].cfg[off] |= mask
	gpioMemory.groups[p.group].cfg[off] &^= v
	if p.function() != f {
		panic(f)
	}
	p.isOut = f == out
	return true
}

var (
	PB0  gpio.PinIO = cpupins["PB0"]
	PB1  gpio.PinIO = cpupins["PB1"]
	PB2  gpio.PinIO = cpupins["PB2"]
	PB3  gpio.PinIO = cpupins["PB3"]
	PB4  gpio.PinIO = cpupins["PB4"]
	PB5  gpio.PinIO = cpupins["PB5"]
	PB6  gpio.PinIO = cpupins["PB6"]
	PB7  gpio.PinIO = cpupins["PB7"]
	PB8  gpio.PinIO = cpupins["PB8"]
	PB9  gpio.PinIO = cpupins["PB9"]
	PB10 gpio.PinIO = cpupins["PB10"]
	PB11 gpio.PinIO = cpupins["PB11"]
	PB12 gpio.PinIO = cpupins["PB12"]
	PB13 gpio.PinIO = cpupins["PB13"]
	PB14 gpio.PinIO = cpupins["PB14"]
	PB15 gpio.PinIO = cpupins["PB15"]
	PB16 gpio.PinIO = cpupins["PB16"]
	PB17 gpio.PinIO = cpupins["PB17"]
	PB18 gpio.PinIO = cpupins["PB18"]
	PC0  gpio.PinIO = cpupins["PC0"]
	PC1  gpio.PinIO = cpupins["PC1"]
	PC2  gpio.PinIO = cpupins["PC2"]
	PC3  gpio.PinIO = cpupins["PC3"]
	PC4  gpio.PinIO = cpupins["PC4"]
	PC5  gpio.PinIO = cpupins["PC5"]
	PC6  gpio.PinIO = cpupins["PC6"]
	PC7  gpio.PinIO = cpupins["PC7"]
	PC8  gpio.PinIO = cpupins["PC8"]
	PC9  gpio.PinIO = cpupins["PC9"]
	PC10 gpio.PinIO = cpupins["PC10"]
	PC11 gpio.PinIO = cpupins["PC11"]
	PC12 gpio.PinIO = cpupins["PC12"]
	PC13 gpio.PinIO = cpupins["PC13"]
	PC14 gpio.PinIO = cpupins["PC14"]
	PC15 gpio.PinIO = cpupins["PC15"]
	PC16 gpio.PinIO = cpupins["PC16"]
	PD0  gpio.PinIO = cpupins["PD0"]
	PD1  gpio.PinIO = cpupins["PD1"]
	PD2  gpio.PinIO = cpupins["PD2"]
	PD3  gpio.PinIO = cpupins["PD3"]
	PD4  gpio.PinIO = cpupins["PD4"]
	PD5  gpio.PinIO = cpupins["PD5"]
	PD6  gpio.PinIO = cpupins["PD6"]
	PD7  gpio.PinIO = cpupins["PD7"]
	PD8  gpio.PinIO = cpupins["PD8"]
	PD9  gpio.PinIO = cpupins["PD9"]
	PD10 gpio.PinIO = cpupins["PD10"]
	PD11 gpio.PinIO = cpupins["PD11"]
	PD12 gpio.PinIO = cpupins["PD12"]
	PD13 gpio.PinIO = cpupins["PD13"]
	PD14 gpio.PinIO = cpupins["PD14"]
	PD15 gpio.PinIO = cpupins["PD15"]
	PD16 gpio.PinIO = cpupins["PD16"]
	PD17 gpio.PinIO = cpupins["PD17"]
	PD18 gpio.PinIO = cpupins["PD18"]
	PD19 gpio.PinIO = cpupins["PD19"]
	PD20 gpio.PinIO = cpupins["PD20"]
	PD21 gpio.PinIO = cpupins["PD21"]
	PD22 gpio.PinIO = cpupins["PD22"]
	PD23 gpio.PinIO = cpupins["PD23"]
	PD24 gpio.PinIO = cpupins["PD24"]
	PD25 gpio.PinIO = cpupins["PD25"]
	PD26 gpio.PinIO = cpupins["PD26"]
	PD27 gpio.PinIO = cpupins["PD27"]
	PE0  gpio.PinIO = cpupins["PE0"]
	PE1  gpio.PinIO = cpupins["PE1"]
	PE2  gpio.PinIO = cpupins["PE2"]
	PE3  gpio.PinIO = cpupins["PE3"]
	PE4  gpio.PinIO = cpupins["PE4"]
	PE5  gpio.PinIO = cpupins["PE5"]
	PE6  gpio.PinIO = cpupins["PE6"]
	PE7  gpio.PinIO = cpupins["PE7"]
	PE8  gpio.PinIO = cpupins["PE8"]
	PE9  gpio.PinIO = cpupins["PE9"]
	PE10 gpio.PinIO = cpupins["PE10"]
	PE11 gpio.PinIO = cpupins["PE11"]
	PE12 gpio.PinIO = cpupins["PE12"]
	PE13 gpio.PinIO = cpupins["PE13"]
	PE14 gpio.PinIO = cpupins["PE14"]
	PE15 gpio.PinIO = cpupins["PE15"]
	PE16 gpio.PinIO = cpupins["PE16"]
	PE17 gpio.PinIO = cpupins["PE17"]
	PF0  gpio.PinIO = cpupins["PF0"]
	PF1  gpio.PinIO = cpupins["PF1"]
	PF2  gpio.PinIO = cpupins["PF2"]
	PF3  gpio.PinIO = cpupins["PF3"]
	PF4  gpio.PinIO = cpupins["PF4"]
	PF5  gpio.PinIO = cpupins["PF5"]
	PF6  gpio.PinIO = cpupins["PF6"]
	PG0  gpio.PinIO = cpupins["PG0"]
	PG1  gpio.PinIO = cpupins["PG1"]
	PG2  gpio.PinIO = cpupins["PG2"]
	PG3  gpio.PinIO = cpupins["PG3"]
	PG4  gpio.PinIO = cpupins["PG4"]
	PG5  gpio.PinIO = cpupins["PG5"]
	PG6  gpio.PinIO = cpupins["PG6"]
	PG7  gpio.PinIO = cpupins["PG7"]
	PG8  gpio.PinIO = cpupins["PG8"]
	PG9  gpio.PinIO = cpupins["PG9"]
	PG10 gpio.PinIO = cpupins["PG10"]
	PG11 gpio.PinIO = cpupins["PG11"]
	PG12 gpio.PinIO = cpupins["PG12"]
	PG13 gpio.PinIO = cpupins["PG13"]
	PH0  gpio.PinIO = cpupins["PH0"]
	PH1  gpio.PinIO = cpupins["PH1"]
	PH2  gpio.PinIO = cpupins["PH2"]
	PH3  gpio.PinIO = cpupins["PH3"]
	PH4  gpio.PinIO = cpupins["PH4"]
	PH5  gpio.PinIO = cpupins["PH5"]
	PH6  gpio.PinIO = cpupins["PH6"]
	PH7  gpio.PinIO = cpupins["PH7"]
	PH8  gpio.PinIO = cpupins["PH8"]
	PH9  gpio.PinIO = cpupins["PH9"]
	PH10 gpio.PinIO = cpupins["PH10"]
	PH11 gpio.PinIO = cpupins["PH11"]
)

// initPins initializes the mapping of pins by function, sets the alternate functions of each
// pin, and registers all the pins with gpio.
func initPins() error {
	for i := range cpupins {
		// register the pin with gpio
		if err := gpio.Register(cpupins[i]); err != nil {
			return err
		}
		// iterate through alternate functions and register function->pin mapping
		for _, f := range cpupins[i].altFunc {
			if f != "" {
				gpio.MapFunction(f, cpupins[i])
			}
		}
	}
	return nil
}

// function encodes the active functionality of a pin. The alternate functions
// are GPIO pin dependent.
type function uint8

// Page 23~24
// Each pin can have one of 7 functions.
const (
	in       function = 0
	out      function = 1
	alt1     function = 2
	alt2     function = 3
	alt3     function = 4
	alt4     function = 5
	alt5     function = 6 // often interrupt based edge detection as input
	disabled function = 7
)

// gpioGroup is a memory-mapped structure for the hardware registers that control a
// group of at most 32 pins. In practice the number of valid pins per group varies
// between 10 and 27.
//
// http://files.pine64.org/doc/datasheet/pine64/Allwinner_A64_User_Manual_V1.0.pdf
// Page 376 GPIO PB to PH.
type gpioGroup struct {
	// Pn_CFGx n*0x24+x*4       Port n Configure Register x (n from 1(B) to 7(H))
	cfg [4]uint32
	// Pn_DAT  n*0x24+0x10      Port n Data Register (n from 1(B) to 7(H))
	data uint32
	// Pn_DRVx n*0x24+0x14+x*4  Port n Multi-Driving Register x (n from 1 to 7)
	drv [2]uint32
	// Pn_PULL n*0x24+0x1C+x*4  Port n Pull Register (n from 1(B) to 7(H))
	pull [2]uint32
}

// gpioMap memory-maps all the gpio pin groups.
type gpioMap struct {
	// PB to PH. The first group is unused.
	groups [8]gpioGroup
}

var gpioMemory *gpioMap
