// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains the definitions of all possible generic Allwinner pins and their
// implementation using a combination of sysfs and memory-mapped I/O.

package allwinner

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/sysfs"
)

// List of all known pins. These global variables can be used directly.
//
// The supported functionality of each gpio differs between CPUs. For example
// the R8 has the LCD-DE signal on gpio PD25 but the A64 has it on PD19.
//
// The availability of each gpio differs between CPUs. For example the R8 has
// 19 pins in the group PB but the A64 only has 10.
//
// So make sure to read the datasheet for the exact right CPU.
var (
	PB0, PB1, PB2, PB3, PB4, PB5, PB6, PB7, PB8, PB9, PB10, PB11, PB12, PB13, PB14, PB15, PB16, PB17, PB18                                                       *Pin
	PC0, PC1, PC2, PC3, PC4, PC5, PC6, PC7, PC8, PC9, PC10, PC11, PC12, PC13, PC14, PC15, PC16, PC17, PC18, PC19                                                 *Pin
	PD0, PD1, PD2, PD3, PD4, PD5, PD6, PD7, PD8, PD9, PD10, PD11, PD12, PD13, PD14, PD15, PD16, PD17, PD18, PD19, PD20, PD21, PD22, PD23, PD24, PD25, PD26, PD27 *Pin
	PE0, PE1, PE2, PE3, PE4, PE5, PE6, PE7, PE8, PE9, PE10, PE11, PE12, PE13, PE14, PE15, PE16, PE17                                                             *Pin
	PF0, PF1, PF2, PF3, PF4, PF5, PF6                                                                                                                            *Pin
	PG0, PG1, PG2, PG3, PG4, PG5, PG6, PG7, PG8, PG9, PG10, PG11, PG12, PG13                                                                                     *Pin
	PH0, PH1, PH2, PH3, PH4, PH5, PH6, PH7, PH8, PH9, PH10, PH11                                                                                                 *Pin
)

// Pin implements the gpio.PinIO interface for generic Allwinner CPU pins using
// memory mapping for gpio in/out functionality.
type Pin struct {
	// Immutable.
	group       uint8     // as per register offset calculation
	offset      uint8     // as per register offset calculation
	name        string    // name as per datasheet
	defaultPull gpio.Pull // default pull at startup
	altFunc     [5]string // alternate functions

	// Mutable.
	edge        *sysfs.Pin // Set once, then never set back to nil.
	usingEdge   bool       // Set when edge detection is enabled.
	available   bool       // Set when the pin is available on this CPU architecture.
	supportEdge bool       // Set when the pin supports interrupt based edge detection.
}

// String returns the name of the pin in the processor and the GPIO pin number.
func (p *Pin) String() string {
	return fmt.Sprintf("%s(%d)", p.name, p.Number())
}

// Name returns the pin name.
func (p *Pin) Name() string {
	return p.name
}

// Number returns the GPIO pin number as represented by gpio sysfs.
func (p *Pin) Number() int {
	return int(p.group)*32 + int(p.offset)
}

// Function returns the current function of the pin in printable form.
func (p *Pin) Function() string {
	if !p.available {
		return "N/A"
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

// Halt implements conn.Resource.
//
// It stops edge detection if enabled.
func (p *Pin) Halt() error {
	if p.usingEdge {
		if err := p.edge.Halt(); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	return nil
}

// In sets the pin direction to input and optionally enables a pull-up/down
// resistor as well as edge detection.
//
// Not all pins support edge detection on Allwinner processors!
//
// Edge detection requires opening a gpio sysfs file handle. The pin will be
// exported at /sys/class/gpio/gpio*/. Note that the pin will not be unexported
// at shutdown.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	if gpioMemory == nil {
		return p.wrap(errors.New("subsystem not initialized"))
	}
	if !p.available {
		return p.wrap(errors.New("not available on this CPU architecture"))
	}
	if edge != gpio.NoEdge && !p.supportEdge {
		return p.wrap(errors.New("edge detection is not supported on this pin"))
	}
	if p.usingEdge && edge == gpio.NoEdge {
		if err := p.edge.Halt(); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	p.setFunction(in)
	if pull != gpio.PullNoChange {
		off := p.offset / 16
		shift := 2 * (p.offset % 16)
		// Do it in a way that is concurrency safe.
		gpioMemory.groups[p.group].pull[off] &^= 3 << shift
		switch pull {
		case gpio.PullDown:
			gpioMemory.groups[p.group].pull[off] = 2 << shift
		case gpio.PullUp:
			gpioMemory.groups[p.group].pull[off] = 1 << shift
		default:
		}
	}
	if edge != gpio.NoEdge {
		if p.edge == nil {
			ok := false
			if p.edge, ok = sysfs.Pins[p.Number()]; !ok {
				return p.wrap(errors.New("pin is not exported by sysfs"))
			}
		}
		// This resets pending edges.
		if err := p.edge.In(gpio.PullNoChange, edge); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = true
	}
	return nil
}

// Read return the current pin level and implements gpio.PinIn.
//
// This function is very fast.
func (p *Pin) Read() gpio.Level {
	if gpioMemory == nil || !p.available {
		return gpio.Low
	}
	return gpio.Level(gpioMemory.groups[p.group].data&(1<<p.offset) != 0)
}

// WaitForEdge waits for an edge as previously set using In() or the expiration
// of a timeout.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	if p.edge != nil {
		return p.edge.WaitForEdge(timeout)
	}
	return false
}

// Pull returns the current pull-up/down registor setting.
func (p *Pin) Pull() gpio.Pull {
	if gpioMemory == nil || !p.available {
		return gpio.PullNoChange
	}
	v := gpioMemory.groups[p.group].pull[p.offset/16]
	switch (v >> (2 * (p.offset % 16))) & 3 {
	case 0:
		return gpio.Float
	case 1:
		return gpio.PullUp
	case 2:
		return gpio.PullDown
	default:
		// Confused.
		return gpio.PullNoChange
	}
}

// Out ensures that the pin is configured as an output and outputs the value.
func (p *Pin) Out(l gpio.Level) error {
	if gpioMemory == nil {
		return p.wrap(errors.New("subsystem not initialized"))
	}
	if !p.available {
		return p.wrap(errors.New("not available on this CPU architecture"))
	}
	// First disable edges.
	if err := p.Halt(); err != nil {
		return err
	}
	p.FastOut(l)
	p.setFunction(out)
	return nil
}

// FastOut sets a pin output level with Absolutely No error checking.
//
// Out() Must be called once first before calling FastOut(), otherwise the
// behavior is undefined. Then FastOut() can be used for minimal CPU overhead
// to reach Mhz scale bit banging.
func (p *Pin) FastOut(l gpio.Level) {
	bit := uint32(1 << p.offset)
	// Pn_DAT  n*0x24+0x10  Port n Data Register (n from 1(B) to 7(H))
	switch p.group {
	case 1:
		if l {
			gpioMemory.groups[1].data |= bit
		} else {
			gpioMemory.groups[1].data &^= bit
		}
	case 2:
		if l {
			gpioMemory.groups[2].data |= bit
		} else {
			gpioMemory.groups[2].data &^= bit
		}
	case 3:
		if l {
			gpioMemory.groups[3].data |= bit
		} else {
			gpioMemory.groups[3].data &^= bit
		}
	case 4:
		if l {
			gpioMemory.groups[4].data |= bit
		} else {
			gpioMemory.groups[4].data &^= bit
		}
	case 5:
		if l {
			gpioMemory.groups[5].data |= bit
		} else {
			gpioMemory.groups[5].data &^= bit
		}
	case 6:
		if l {
			gpioMemory.groups[6].data |= bit
		} else {
			gpioMemory.groups[6].data &^= bit
		}
	case 7:
		if l {
			gpioMemory.groups[7].data |= bit
		} else {
			gpioMemory.groups[7].data &^= bit
		}
	}
}

// DefaultPull returns the default pull for the pin.
func (p *Pin) DefaultPull() gpio.Pull {
	return p.defaultPull
}

//

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	if gpioMemory == nil {
		return disabled
	}
	shift := 4 * (p.offset % 8)
	return function((gpioMemory.groups[p.group].cfg[p.offset/8] >> shift) & 7)
}

// setFunction changes the GPIO pin function.
func (p *Pin) setFunction(f function) {
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
}

func (p *Pin) wrap(err error) error {
	return fmt.Errorf("allwinner-gpio (%s): %v", p, err)
}

//

// A64: Page 23~24
// R8: Page 322-334.
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

var (
	// gpioMemory is the memory map of the CPU GPIO registers.
	gpioMemory *gpioMap
	// gpioBaseAddr is the physical base address of the GPIO registers.
	gpioBaseAddr uint32
)

// cpupins that may be implemented by a generic Allwinner CPU. Not all pins
// will be present on all models and even if the CPU model supports them they
// may not be connected to anything on the board. The net effect is that it may
// look like more pins are available than really are, but trying to get the pin
// list 100% correct on all platforms seems futile, hence periph errs on the
// side of caution.
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
	"PC3":  {group: 2, offset: 3, name: "PC3", defaultPull: gpio.PullUp},
	"PC4":  {group: 2, offset: 4, name: "PC4", defaultPull: gpio.PullUp},
	"PC5":  {group: 2, offset: 5, name: "PC5", defaultPull: gpio.Float},
	"PC6":  {group: 2, offset: 6, name: "PC6", defaultPull: gpio.PullUp},
	"PC7":  {group: 2, offset: 7, name: "PC7", defaultPull: gpio.PullUp},
	"PC8":  {group: 2, offset: 8, name: "PC8", defaultPull: gpio.Float},
	"PC9":  {group: 2, offset: 9, name: "PC9", defaultPull: gpio.Float},
	"PC10": {group: 2, offset: 10, name: "PC10", defaultPull: gpio.Float},
	"PC11": {group: 2, offset: 11, name: "PC11", defaultPull: gpio.Float},
	"PC12": {group: 2, offset: 12, name: "PC12", defaultPull: gpio.Float},
	"PC13": {group: 2, offset: 13, name: "PC13", defaultPull: gpio.Float},
	"PC14": {group: 2, offset: 14, name: "PC14", defaultPull: gpio.Float},
	"PC15": {group: 2, offset: 15, name: "PC15", defaultPull: gpio.Float},
	"PC16": {group: 2, offset: 16, name: "PC16", defaultPull: gpio.Float},
	"PC17": {group: 2, offset: 17, name: "PC17", defaultPull: gpio.Float},
	"PC18": {group: 2, offset: 18, name: "PC18", defaultPull: gpio.Float},
	"PC19": {group: 2, offset: 19, name: "PC19", defaultPull: gpio.Float},
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

func init() {
	PB0 = cpupins["PB0"]
	PB1 = cpupins["PB1"]
	PB2 = cpupins["PB2"]
	PB3 = cpupins["PB3"]
	PB4 = cpupins["PB4"]
	PB5 = cpupins["PB5"]
	PB6 = cpupins["PB6"]
	PB7 = cpupins["PB7"]
	PB8 = cpupins["PB8"]
	PB9 = cpupins["PB9"]
	PB10 = cpupins["PB10"]
	PB11 = cpupins["PB11"]
	PB12 = cpupins["PB12"]
	PB13 = cpupins["PB13"]
	PB14 = cpupins["PB14"]
	PB15 = cpupins["PB15"]
	PB16 = cpupins["PB16"]
	PB17 = cpupins["PB17"]
	PB18 = cpupins["PB18"]
	PC0 = cpupins["PC0"]
	PC1 = cpupins["PC1"]
	PC2 = cpupins["PC2"]
	PC3 = cpupins["PC3"]
	PC4 = cpupins["PC4"]
	PC5 = cpupins["PC5"]
	PC6 = cpupins["PC6"]
	PC7 = cpupins["PC7"]
	PC8 = cpupins["PC8"]
	PC9 = cpupins["PC9"]
	PC10 = cpupins["PC10"]
	PC11 = cpupins["PC11"]
	PC12 = cpupins["PC12"]
	PC13 = cpupins["PC13"]
	PC14 = cpupins["PC14"]
	PC15 = cpupins["PC15"]
	PC16 = cpupins["PC16"]
	PD0 = cpupins["PD0"]
	PD1 = cpupins["PD1"]
	PD2 = cpupins["PD2"]
	PD3 = cpupins["PD3"]
	PD4 = cpupins["PD4"]
	PD5 = cpupins["PD5"]
	PD6 = cpupins["PD6"]
	PD7 = cpupins["PD7"]
	PD8 = cpupins["PD8"]
	PD9 = cpupins["PD9"]
	PD10 = cpupins["PD10"]
	PD11 = cpupins["PD11"]
	PD12 = cpupins["PD12"]
	PD13 = cpupins["PD13"]
	PD14 = cpupins["PD14"]
	PD15 = cpupins["PD15"]
	PD16 = cpupins["PD16"]
	PD17 = cpupins["PD17"]
	PD18 = cpupins["PD18"]
	PD19 = cpupins["PD19"]
	PD20 = cpupins["PD20"]
	PD21 = cpupins["PD21"]
	PD22 = cpupins["PD22"]
	PD23 = cpupins["PD23"]
	PD24 = cpupins["PD24"]
	PD25 = cpupins["PD25"]
	PD26 = cpupins["PD26"]
	PD27 = cpupins["PD27"]
	PE0 = cpupins["PE0"]
	PE1 = cpupins["PE1"]
	PE2 = cpupins["PE2"]
	PE3 = cpupins["PE3"]
	PE4 = cpupins["PE4"]
	PE5 = cpupins["PE5"]
	PE6 = cpupins["PE6"]
	PE7 = cpupins["PE7"]
	PE8 = cpupins["PE8"]
	PE9 = cpupins["PE9"]
	PE10 = cpupins["PE10"]
	PE11 = cpupins["PE11"]
	PE12 = cpupins["PE12"]
	PE13 = cpupins["PE13"]
	PE14 = cpupins["PE14"]
	PE15 = cpupins["PE15"]
	PE16 = cpupins["PE16"]
	PE17 = cpupins["PE17"]
	PF0 = cpupins["PF0"]
	PF1 = cpupins["PF1"]
	PF2 = cpupins["PF2"]
	PF3 = cpupins["PF3"]
	PF4 = cpupins["PF4"]
	PF5 = cpupins["PF5"]
	PF6 = cpupins["PF6"]
	PG0 = cpupins["PG0"]
	PG1 = cpupins["PG1"]
	PG2 = cpupins["PG2"]
	PG3 = cpupins["PG3"]
	PG4 = cpupins["PG4"]
	PG5 = cpupins["PG5"]
	PG6 = cpupins["PG6"]
	PG7 = cpupins["PG7"]
	PG8 = cpupins["PG8"]
	PG9 = cpupins["PG9"]
	PG10 = cpupins["PG10"]
	PG11 = cpupins["PG11"]
	PG12 = cpupins["PG12"]
	PG13 = cpupins["PG13"]
	PH0 = cpupins["PH0"]
	PH1 = cpupins["PH1"]
	PH2 = cpupins["PH2"]
	PH3 = cpupins["PH3"]
	PH4 = cpupins["PH4"]
	PH5 = cpupins["PH5"]
	PH6 = cpupins["PH6"]
	PH7 = cpupins["PH7"]
	PH8 = cpupins["PH8"]
	PH9 = cpupins["PH9"]
	PH10 = cpupins["PH10"]
	PH11 = cpupins["PH11"]
}

// initPins initializes the mapping of pins by function, sets the alternate
// functions of each pin, and registers all the pins with gpio.
func initPins() error {
	for i := range cpupins {
		// Register the pin with gpio.
		if err := gpioreg.Register(cpupins[i], true); err != nil {
			return err
		}
		// Iterate through alternate functions and register function->pin mapping.
		// TODO(maruel): There's a problem where multiple pins may be set to the
		// same function. Need investigation. For now just ignore errors.
		for _, f := range cpupins[i].altFunc {
			if f != "" && f[0] != '<' && f[:2] != "In" && f[:3] != "Out" {
				// TODO(maruel): Stop ignoring errors by not registering the same
				// function multiple times.
				gpioreg.RegisterAlias(f, cpupins[i].Name())
				/*
					if err := gpioreg.RegisterAlias(f, cpupins[i].Number()); err != nil {
						return true, err
					}
				*/
			}
		}
	}
	return nil
}

// function encodes the active functionality of a pin. The alternate functions
// are GPIO pin dependent.
type function uint8

// gpioGroup is a memory-mapped structure for the hardware registers that
// control a group of at most 32 pins. In practice the number of valid pins per
// group varies between 10 and 27.
//
// http://files.pine64.org/doc/datasheet/pine64/Allwinner_A64_User_Manual_V1.0.pdf
// Page 376 GPIO PB to PH.
// Page 410 GPIO PL.
// Size is 36 bytes.
type gpioGroup struct {
	// Pn_CFGx n*0x24+x*4       Port n Configure Register x (n from 1(B) to 7(H))
	cfg [4]uint32
	// Pn_DAT  n*0x24+0x10      Port n Data Register (n from 1(B) to 7(H))
	data uint32
	// Pn_DRVx n*0x24+0x14+x*4  Port n Multi-Driving Register x (n from 1 to 7)
	drv [2]uint32 // TODO(maruel): Figure out how to use this.
	// Pn_PULL n*0x24+0x1C+x*4  Port n Pull Register (n from 1(B) to 7(H))
	pull [2]uint32
}

// gpioMap memory-maps all the gpio pin groups.
type gpioMap struct {
	// PB to PH. The first group is unused.
	groups [8]gpioGroup
}

// driverGPIO implements periph.Driver.
type driverGPIO struct {
}

func (d *driverGPIO) String() string {
	return "allwinner-gpio"
}

func (d *driverGPIO) Prerequisites() []string {
	return nil
}

// Init does nothing if an allwinner processor is not detected. If one is
// detected, it memory maps gpio CPU registers and then sets up the pin mapping
// for the exact processor model detected.
func (d *driverGPIO) Init() (bool, error) {
	if !Present() {
		return false, errors.New("Allwinner CPU not detected")
	}
	gpioBaseAddr = uint32(getBaseAddress())
	if err := pmem.MapAsPOD(uint64(gpioBaseAddr), &gpioMemory); err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}

	switch {
	case IsA64():
		if err := mapA64Pins(); err != nil {
			return true, err
		}
	case IsR8():
		if err := mapR8Pins(); err != nil {
			return true, err
		}
	default:
		return false, errors.New("unknown Allwinner CPU model")
	}

	return true, initPins()
}

func init() {
	if isArm {
		periph.MustRegister(&driverGPIO{})
	}
}

// getBaseAddress queries the virtual file system to retrieve the base address
// of the GPIO registers for GPIO pins in groups PB to PH.
//
// Defaults to 0x01C20800 as per datasheet if it could not query the file
// system.
func getBaseAddress() uint64 {
	base := uint64(0x01C20800)
	link, err := os.Readlink("/sys/bus/platform/drivers/sun50i-pinctrl/driver")
	if err != nil {
		return base
	}
	parts := strings.SplitN(path.Base(link), ".", 2)
	if len(parts) != 2 {
		return base
	}
	base2, err := strconv.ParseUint(parts[0], 16, 64)
	if err != nil {
		return base
	}
	return base2
}

// Ensure that the various structs implement the interfaces they're supposed to.

var _ gpio.PinDefaultPull = &Pin{}
var _ gpio.PinIO = &Pin{}
var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
