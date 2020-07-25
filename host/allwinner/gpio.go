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
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
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
	PA0, PA1, PA2, PA3, PA4, PA5, PA6, PA7, PA8, PA9, PA10, PA11, PA12, PA13, PA14, PA15, PA16, PA17                                                             *Pin
	PB0, PB1, PB2, PB3, PB4, PB5, PB6, PB7, PB8, PB9, PB10, PB11, PB12, PB13, PB14, PB15, PB16, PB17, PB18, PB19, PB20, PB21, PB22, PB23                         *Pin
	PC0, PC1, PC2, PC3, PC4, PC5, PC6, PC7, PC8, PC9, PC10, PC11, PC12, PC13, PC14, PC15, PC16, PC17, PC18, PC19, PC20, PC21, PC22, PC23, PC24                   *Pin
	PD0, PD1, PD2, PD3, PD4, PD5, PD6, PD7, PD8, PD9, PD10, PD11, PD12, PD13, PD14, PD15, PD16, PD17, PD18, PD19, PD20, PD21, PD22, PD23, PD24, PD25, PD26, PD27 *Pin
	PE0, PE1, PE2, PE3, PE4, PE5, PE6, PE7, PE8, PE9, PE10, PE11, PE12, PE13, PE14, PE15, PE16, PE17                                                             *Pin
	PF0, PF1, PF2, PF3, PF4, PF5, PF6                                                                                                                            *Pin
	PG0, PG1, PG2, PG3, PG4, PG5, PG6, PG7, PG8, PG9, PG10, PG11, PG12, PG13                                                                                     *Pin
	PH0, PH1, PH2, PH3, PH4, PH5, PH6, PH7, PH8, PH9, PH10, PH11, PH12, PH13, PH14, PH15, PH16, PH17, PH18, PH19, PH20, PH21, PH22, PH23, PH24, PH25, PH26, PH27 *Pin
	PI0, PI1, PI2, PI3, PI4, PI5, PI6, PI7, PI8, PI9, PI10, PI11, PI12, PI13, PI14, PI15, PI16, PI17, PI18, PI19, PI20, PI21                                     *Pin
)

// Pin implements the gpio.PinIO interface for generic Allwinner CPU pins using
// memory mapping for gpio in/out functionality.
type Pin struct {
	// Immutable.
	group       uint8     // as per register offset calculation
	offset      uint8     // as per register offset calculation
	name        string    // name as per datasheet
	defaultPull gpio.Pull // default pull at startup

	// Immutable after driver initialization.
	altFunc     [5]pin.Func // alternate functions
	sysfsPin    *sysfs.Pin  // Set to the corresponding sysfs.Pin, if any.
	available   bool        // Set when the pin is available on this CPU architecture.
	supportEdge bool        // Set when the pin supports interrupt based edge detection.

	// Mutable.
	usingEdge bool // Set when edge detection is enabled.
}

// String implements conn.Resource.
//
// It returns the pin name and number, ex: "PB5(37)".
func (p *Pin) String() string {
	return fmt.Sprintf("%s(%d)", p.name, p.Number())
}

// Halt implements conn.Resource.
//
// It stops edge detection if enabled.
func (p *Pin) Halt() error {
	if p.usingEdge {
		if err := p.sysfsPin.Halt(); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	return nil
}

// Name implements pin.Pin.
//
// It returns the pin name, ex: "PB5".
func (p *Pin) Name() string {
	return p.name
}

// Number implements pin.Pin.
//
// It returns the GPIO pin number as represented by gpio sysfs.
func (p *Pin) Number() int {
	return int(p.group)*32 + int(p.offset)
}

// Function implements pin.Pin.
func (p *Pin) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *Pin) Func() pin.Func {
	if !p.available {
		return pin.FuncNone
	}
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return pin.FuncNone
		}
		return p.sysfsPin.Func()
	}
	switch f := p.function(); f {
	case in:
		if p.FastRead() {
			return gpio.IN_HIGH
		}
		return gpio.IN_LOW
	case out:
		if p.FastRead() {
			return gpio.OUT_HIGH
		}
		return gpio.OUT_LOW
	case alt1:
		if p.altFunc[0] != "" {
			return pin.Func(p.altFunc[0])
		}
		return pin.Func("ALT1")
	case alt2:
		if p.altFunc[1] != "" {
			return pin.Func(p.altFunc[1])
		}
		return pin.Func("ALT2")
	case alt3:
		if p.altFunc[2] != "" {
			return pin.Func(p.altFunc[2])
		}
		return pin.Func("ALT3")
	case alt4:
		if p.altFunc[3] != "" {
			if strings.Contains(string(p.altFunc[3]), "EINT") {
				// It's an input supporting interrupts.
				if p.FastRead() {
					return gpio.IN_HIGH
				}
				return gpio.IN_LOW
			}
			return pin.Func(p.altFunc[3])
		}
		return pin.Func("ALT4")
	case alt5:
		if p.altFunc[4] != "" {
			if strings.Contains(string(p.altFunc[4]), "EINT") {
				// It's an input supporting interrupts.
				if p.FastRead() {
					return gpio.IN_HIGH
				}
				return gpio.IN_LOW
			}
			return pin.Func(p.altFunc[4])
		}
		return pin.Func("ALT5")
	case disabled:
		return pin.FuncNone
	default:
		return pin.FuncNone
	}
}

// SupportedFuncs implements pin.PinFunc.
func (p *Pin) SupportedFuncs() []pin.Func {
	f := make([]pin.Func, 0, 2+4)
	f = append(f, gpio.IN, gpio.OUT)
	for _, m := range p.altFunc {
		if m != pin.FuncNone && !strings.Contains(string(m), "EINT") {
			f = append(f, m)
		}
	}
	return f
}

// SetFunc implements pin.PinFunc.
func (p *Pin) SetFunc(f pin.Func) error {
	switch f {
	case gpio.FLOAT:
		return p.In(gpio.Float, gpio.NoEdge)
	case gpio.IN:
		return p.In(gpio.PullNoChange, gpio.NoEdge)
	case gpio.IN_LOW:
		return p.In(gpio.PullDown, gpio.NoEdge)
	case gpio.IN_HIGH:
		return p.In(gpio.PullUp, gpio.NoEdge)
	case gpio.OUT_HIGH:
		return p.Out(gpio.High)
	case gpio.OUT_LOW:
		return p.Out(gpio.Low)
	default:
		isGeneral := f == f.Generalize()
		for i, m := range p.altFunc {
			if m == f || (isGeneral && m.Generalize() == f) {
				if err := p.Halt(); err != nil {
					return err
				}
				switch i {
				case 0:
					p.setFunction(alt1)
				case 1:
					p.setFunction(alt2)
				case 2:
					p.setFunction(alt3)
				case 3:
					p.setFunction(alt4)
				case 4:
					p.setFunction(alt5)
				}
				return nil
			}
		}
		return p.wrap(errors.New("unsupported function"))
	}
}

// In implements gpio.PinIn.
//
// It sets the pin direction to input and optionally enables a pull-up/down
// resistor as well as edge detection.
//
// Not all pins support edge detection on Allwinner processors!
//
// Edge detection requires opening a gpio sysfs file handle. The pin will be
// exported at /sys/class/gpio/gpio*/. Note that the pin will not be unexported
// at shutdown.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	if !p.available {
		// We do not want the error message about uninitialized system.
		return p.wrap(errors.New("not available on this CPU architecture"))
	}
	if edge != gpio.NoEdge && !p.supportEdge {
		return p.wrap(errors.New("edge detection is not supported on this pin"))
	}
	if p.usingEdge && edge == gpio.NoEdge {
		if err := p.sysfsPin.Halt(); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return p.wrap(errors.New("subsystem gpiomem not initialized and sysfs not accessible; try running as root?"))
		}
		if pull != gpio.PullNoChange {
			return p.wrap(errors.New("pull cannot be used when subsystem gpiomem not initialized; try running as root?"))
		}
		if err := p.sysfsPin.In(pull, edge); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = edge != gpio.NoEdge
		return nil
	}
	p.setFunction(in)
	if pull != gpio.PullNoChange {
		off := p.offset / 16
		shift := 2 * (p.offset % 16)
		// Do it in a way that is concurrency safe.
		drvGPIO.gpioMemory.groups[p.group].pull[off] &^= 3 << shift
		switch pull {
		case gpio.PullDown:
			drvGPIO.gpioMemory.groups[p.group].pull[off] = 2 << shift
		case gpio.PullUp:
			drvGPIO.gpioMemory.groups[p.group].pull[off] = 1 << shift
		default:
		}
	}
	if edge != gpio.NoEdge {
		if p.sysfsPin == nil {
			return p.wrap(fmt.Errorf("pin %d is not exported by sysfs", p.Number()))
		}
		// This resets pending edges.
		if err := p.sysfsPin.In(gpio.PullNoChange, edge); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = true
	}
	return nil
}

// Read implements gpio.PinIn.
//
// It returns the current pin level. This function is fast.
func (p *Pin) Read() gpio.Level {
	if !p.available {
		return gpio.Low
	}
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return gpio.Low
		}
		return p.sysfsPin.Read()
	}
	return gpio.Level(drvGPIO.gpioMemory.groups[p.group].data&(1<<p.offset) != 0)
}

// FastRead return the current pin level without any error checking.
//
// This function is very fast.
func (p *Pin) FastRead() gpio.Level {
	return gpio.Level(drvGPIO.gpioMemory.groups[p.group].data&(1<<p.offset) != 0)
}

// WaitForEdge implements gpio.PinIn.
//
// It waits for an edge as previously set using In() or the expiration of a
// timeout.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	if p.sysfsPin != nil {
		return p.sysfsPin.WaitForEdge(timeout)
	}
	return false
}

// Pull implements gpio.PinIn.
func (p *Pin) Pull() gpio.Pull {
	if drvGPIO.gpioMemory == nil || !p.available {
		return gpio.PullNoChange
	}
	v := drvGPIO.gpioMemory.groups[p.group].pull[p.offset/16]
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

// DefaultPull implements gpio.PinIn.
func (p *Pin) DefaultPull() gpio.Pull {
	return p.defaultPull
}

// Out implements gpio.PinOut.
func (p *Pin) Out(l gpio.Level) error {
	if !p.available {
		// We do not want the error message about uninitialized system.
		return p.wrap(errors.New("not available on this CPU architecture"))
	}
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin != nil {
			return p.wrap(errors.New("subsystem gpiomem not initialized and sysfs not accessible; try running as root?"))
		}
		return p.sysfsPin.Out(l)
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
	// Pn_DAT  n*0x24+0x10  Port n Data Register (n from 0(A) to 8(I))
	// This is a switch on p.group rather than an index to the groups array for
	// performance reasons: to avoid Go's array bound checking code.
	// See https://periph.io/news/2017/gpio_perf/ for details.
	switch p.group {
	case 0:
		if l {
			drvGPIO.gpioMemory.groups[0].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[0].data &^= bit
		}
	case 1:
		if l {
			drvGPIO.gpioMemory.groups[1].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[1].data &^= bit
		}
	case 2:
		if l {
			drvGPIO.gpioMemory.groups[2].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[2].data &^= bit
		}
	case 3:
		if l {
			drvGPIO.gpioMemory.groups[3].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[3].data &^= bit
		}
	case 4:
		if l {
			drvGPIO.gpioMemory.groups[4].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[4].data &^= bit
		}
	case 5:
		if l {
			drvGPIO.gpioMemory.groups[5].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[5].data &^= bit
		}
	case 6:
		if l {
			drvGPIO.gpioMemory.groups[6].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[6].data &^= bit
		}
	case 7:
		if l {
			drvGPIO.gpioMemory.groups[7].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[7].data &^= bit
		}
	case 8:
		if l {
			drvGPIO.gpioMemory.groups[8].data |= bit
		} else {
			drvGPIO.gpioMemory.groups[8].data &^= bit
		}
	}
}

// PWM implements gpio.PinOut.
func (p *Pin) PWM(gpio.Duty, physic.Frequency) error {
	return p.wrap(errors.New("not available on this CPU architecture"))
}

//

// drive returns the configured output current drive strength for this GPIO.
//
// The value returned by this function is not yet verified to be correct. Use
// with suspicion.
func (p *Pin) drive() physic.ElectricCurrent {
	if drvGPIO.gpioMemory == nil {
		return 0
	}
	// Explanation of the buffer configuration, but doesn't state what's the
	// expected drive strength!
	// http://files.pine64.org/doc/datasheet/pine-h64/Allwinner_H6%20V200_User_Manual_V1.1.pdf
	// Section 3.21.3.4 page 381~382
	//
	// The A64 and H3 datasheets call for 20mA, so it could be reasonable to
	// think that the values are 5mA, 10mA, 15mA, 20mA but we don't know for
	// sure.
	v := (drvGPIO.gpioMemory.groups[p.group].drv[p.offset/16] >> (2 * (p.offset & 15))) & 3
	return physic.ElectricCurrent(v+1) * 5 * physic.MilliAmpere
}

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	if drvGPIO.gpioMemory == nil {
		return disabled
	}
	shift := 4 * (p.offset % 8)
	return function((drvGPIO.gpioMemory.groups[p.group].cfg[p.offset/8] >> shift) & 7)
}

// setFunction changes the GPIO pin function.
func (p *Pin) setFunction(f function) {
	off := p.offset / 8
	shift := 4 * (p.offset % 8)
	mask := uint32(disabled) << shift
	v := (uint32(f) << shift) ^ mask
	// First disable, then setup. This is concurrent safe.
	drvGPIO.gpioMemory.groups[p.group].cfg[off] |= mask
	drvGPIO.gpioMemory.groups[p.group].cfg[off] &^= v
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

// cpupins that may be implemented by a generic Allwinner CPU. Not all pins
// will be present on all models and even if the CPU model supports them they
// may not be connected to anything on the board. The net effect is that it may
// look like more pins are available than really are, but trying to get the pin
// list 100% correct on all platforms seems futile, hence periph errs on the
// side of caution.
var cpupins = map[string]*Pin{
	"PA0":  {group: 0, offset: 0, name: "PA0", defaultPull: gpio.Float},
	"PA1":  {group: 0, offset: 1, name: "PA1", defaultPull: gpio.Float},
	"PA2":  {group: 0, offset: 2, name: "PA2", defaultPull: gpio.Float},
	"PA3":  {group: 0, offset: 3, name: "PA3", defaultPull: gpio.Float},
	"PA4":  {group: 0, offset: 4, name: "PA4", defaultPull: gpio.Float},
	"PA5":  {group: 0, offset: 5, name: "PA5", defaultPull: gpio.Float},
	"PA6":  {group: 0, offset: 6, name: "PA6", defaultPull: gpio.Float},
	"PA7":  {group: 0, offset: 7, name: "PA7", defaultPull: gpio.Float},
	"PA8":  {group: 0, offset: 8, name: "PA8", defaultPull: gpio.Float},
	"PA9":  {group: 0, offset: 9, name: "PA9", defaultPull: gpio.Float},
	"PA10": {group: 0, offset: 10, name: "PA10", defaultPull: gpio.Float},
	"PA11": {group: 0, offset: 11, name: "PA11", defaultPull: gpio.Float},
	"PA12": {group: 0, offset: 12, name: "PA12", defaultPull: gpio.Float},
	"PA13": {group: 0, offset: 13, name: "PA13", defaultPull: gpio.Float},
	"PA14": {group: 0, offset: 14, name: "PA14", defaultPull: gpio.Float},
	"PA15": {group: 0, offset: 15, name: "PA15", defaultPull: gpio.Float},
	"PA16": {group: 0, offset: 16, name: "PA16", defaultPull: gpio.Float},
	"PA17": {group: 0, offset: 17, name: "PA17", defaultPull: gpio.Float},
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
	"PB19": {group: 1, offset: 19, name: "PB19", defaultPull: gpio.Float},
	"PB20": {group: 1, offset: 20, name: "PB20", defaultPull: gpio.Float},
	"PB21": {group: 1, offset: 21, name: "PB21", defaultPull: gpio.Float},
	"PB22": {group: 1, offset: 22, name: "PB22", defaultPull: gpio.Float},
	"PB23": {group: 1, offset: 23, name: "PB23", defaultPull: gpio.Float},
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
	"PC20": {group: 2, offset: 20, name: "PC20", defaultPull: gpio.Float},
	"PC21": {group: 2, offset: 21, name: "PC21", defaultPull: gpio.Float},
	"PC22": {group: 2, offset: 22, name: "PC22", defaultPull: gpio.Float},
	"PC23": {group: 2, offset: 23, name: "PC23", defaultPull: gpio.Float},
	"PC24": {group: 2, offset: 24, name: "PC24", defaultPull: gpio.Float},
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
	"PH12": {group: 7, offset: 12, name: "PH12", defaultPull: gpio.Float},
	"PH13": {group: 7, offset: 13, name: "PH13", defaultPull: gpio.Float},
	"PH14": {group: 7, offset: 14, name: "PH14", defaultPull: gpio.Float},
	"PH15": {group: 7, offset: 15, name: "PH15", defaultPull: gpio.Float},
	"PH16": {group: 7, offset: 16, name: "PH16", defaultPull: gpio.Float},
	"PH17": {group: 7, offset: 17, name: "PH17", defaultPull: gpio.Float},
	"PH18": {group: 7, offset: 18, name: "PH18", defaultPull: gpio.Float},
	"PH19": {group: 7, offset: 19, name: "PH19", defaultPull: gpio.Float},
	"PH20": {group: 7, offset: 20, name: "PH20", defaultPull: gpio.Float},
	"PH21": {group: 7, offset: 21, name: "PH21", defaultPull: gpio.Float},
	"PH22": {group: 7, offset: 22, name: "PH22", defaultPull: gpio.Float},
	"PH23": {group: 7, offset: 23, name: "PH23", defaultPull: gpio.Float},
	"PH24": {group: 7, offset: 24, name: "PH24", defaultPull: gpio.Float},
	"PH25": {group: 7, offset: 25, name: "PH25", defaultPull: gpio.Float},
	"PH26": {group: 7, offset: 26, name: "PH26", defaultPull: gpio.Float},
	"PH27": {group: 7, offset: 27, name: "PH27", defaultPull: gpio.Float},
	"PI0":  {group: 8, offset: 0, name: "PI0", defaultPull: gpio.Float},
	"PI1":  {group: 8, offset: 1, name: "PI1", defaultPull: gpio.Float},
	"PI2":  {group: 8, offset: 2, name: "PI2", defaultPull: gpio.Float},
	"PI3":  {group: 8, offset: 3, name: "PI3", defaultPull: gpio.Float},
	"PI4":  {group: 8, offset: 4, name: "PI4", defaultPull: gpio.Float},
	"PI5":  {group: 8, offset: 5, name: "PI5", defaultPull: gpio.Float},
	"PI6":  {group: 8, offset: 6, name: "PI6", defaultPull: gpio.Float},
	"PI7":  {group: 8, offset: 7, name: "PI7", defaultPull: gpio.Float},
	"PI8":  {group: 8, offset: 8, name: "PI8", defaultPull: gpio.Float},
	"PI9":  {group: 8, offset: 9, name: "PI9", defaultPull: gpio.Float},
	"PI10": {group: 8, offset: 10, name: "PI10", defaultPull: gpio.Float},
	"PI11": {group: 8, offset: 11, name: "PI11", defaultPull: gpio.Float},
	"PI12": {group: 8, offset: 12, name: "PI12", defaultPull: gpio.Float},
	"PI13": {group: 8, offset: 13, name: "PI13", defaultPull: gpio.Float},
	"PI14": {group: 8, offset: 14, name: "PI14", defaultPull: gpio.Float},
	"PI15": {group: 8, offset: 15, name: "PI15", defaultPull: gpio.Float},
	"PI16": {group: 8, offset: 16, name: "PI16", defaultPull: gpio.Float},
	"PI17": {group: 8, offset: 17, name: "PI17", defaultPull: gpio.Float},
	"PI18": {group: 8, offset: 18, name: "PI18", defaultPull: gpio.Float},
	"PI19": {group: 8, offset: 19, name: "PI19", defaultPull: gpio.Float},
	"PI20": {group: 8, offset: 20, name: "PI20", defaultPull: gpio.Float},
	"PI21": {group: 8, offset: 21, name: "PI21", defaultPull: gpio.Float},
}

func init() {
	PA0 = cpupins["PA0"]
	PA1 = cpupins["PA1"]
	PA2 = cpupins["PA2"]
	PA3 = cpupins["PA3"]
	PA4 = cpupins["PA4"]
	PA5 = cpupins["PA5"]
	PA6 = cpupins["PA6"]
	PA7 = cpupins["PA7"]
	PA8 = cpupins["PA8"]
	PA9 = cpupins["PA9"]
	PA10 = cpupins["PA10"]
	PA11 = cpupins["PA11"]
	PA12 = cpupins["PA12"]
	PA13 = cpupins["PA13"]
	PA14 = cpupins["PA14"]
	PA15 = cpupins["PA15"]
	PA16 = cpupins["PA16"]
	PA17 = cpupins["PA17"]
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
	PB19 = cpupins["PB19"]
	PB20 = cpupins["PB20"]
	PB21 = cpupins["PB21"]
	PB22 = cpupins["PB22"]
	PB23 = cpupins["PB23"]
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
	PC17 = cpupins["PC17"]
	PC18 = cpupins["PC18"]
	PC19 = cpupins["PC19"]
	PC20 = cpupins["PC20"]
	PC21 = cpupins["PC21"]
	PC22 = cpupins["PC22"]
	PC23 = cpupins["PC23"]
	PC24 = cpupins["PC24"]
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
	PH12 = cpupins["PH12"]
	PH13 = cpupins["PH13"]
	PH14 = cpupins["PH14"]
	PH15 = cpupins["PH15"]
	PH16 = cpupins["PH16"]
	PH17 = cpupins["PH17"]
	PH18 = cpupins["PH18"]
	PH19 = cpupins["PH19"]
	PH20 = cpupins["PH20"]
	PH21 = cpupins["PH21"]
	PH22 = cpupins["PH22"]
	PH23 = cpupins["PH23"]
	PH24 = cpupins["PH24"]
	PH25 = cpupins["PH25"]
	PH26 = cpupins["PH26"]
	PH27 = cpupins["PH27"]
	PI0 = cpupins["PI0"]
	PI1 = cpupins["PI1"]
	PI2 = cpupins["PI2"]
	PI3 = cpupins["PI3"]
	PI4 = cpupins["PI4"]
	PI5 = cpupins["PI5"]
	PI6 = cpupins["PI6"]
	PI7 = cpupins["PI7"]
	PI8 = cpupins["PI8"]
	PI9 = cpupins["PI9"]
	PI10 = cpupins["PI10"]
	PI11 = cpupins["PI11"]
	PI12 = cpupins["PI12"]
	PI13 = cpupins["PI13"]
	PI14 = cpupins["PI14"]
	PI15 = cpupins["PI15"]
	PI16 = cpupins["PI16"]
	PI17 = cpupins["PI17"]
	PI18 = cpupins["PI18"]
	PI19 = cpupins["PI19"]
	PI20 = cpupins["PI20"]
	PI21 = cpupins["PI21"]
}

// initPins initializes the mapping of pins by function, sets the alternate
// functions of each pin, and registers all the pins with gpio.
func initPins() error {
	functions := map[pin.Func]struct{}{}
	for name, p := range cpupins {
		num := strconv.Itoa(p.Number())
		gpion := "GPIO" + num

		// Unregister the pin if already registered. This happens with sysfs-gpio.
		// Do not error on it, since sysfs-gpio may have failed to load.
		_ = gpioreg.Unregister(gpion)
		_ = gpioreg.Unregister(num)

		// Register the pin with gpio.
		if err := gpioreg.Register(p); err != nil {
			return err
		}
		if err := gpioreg.RegisterAlias(gpion, name); err != nil {
			return err
		}
		if err := gpioreg.RegisterAlias(num, name); err != nil {
			return err
		}
		switch f := p.Func(); f {
		case gpio.IN, gpio.OUT, pin.FuncNone:
		default:
			// Registering the same alias twice fails. This can happen if two pins
			// are configured with the same function.
			if _, ok := functions[f]; !ok {
				functions[f] = struct{}{}
				if err := gpioreg.RegisterAlias(string(f), name); err != nil {
					return err
				}
			}
		}
	}

	// Now do a second loop but do the alternate functions.
	for name, p := range cpupins {
		for _, f := range p.SupportedFuncs() {
			switch f {
			case gpio.IN, gpio.OUT:
			default:
				if _, ok := functions[f]; !ok {
					functions[f] = struct{}{}
					if err := gpioreg.RegisterAlias(string(f), name); err != nil {
						return err
					}
				}
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
	drv [2]uint32
	// Pn_PULL n*0x24+0x1C+x*4  Port n Pull Register (n from 1(B) to 7(H))
	pull [2]uint32
}

// gpioMap memory-maps all the gpio pin groups.
type gpioMap struct {
	// PA to PI.
	groups [9]gpioGroup
}

// driverGPIO implements periph.Driver.
type driverGPIO struct {
	// gpioMemory is the memory map of the CPU GPIO registers.
	gpioMemory *gpioMap
}

func (d *driverGPIO) String() string {
	return "allwinner-gpio"
}

func (d *driverGPIO) Prerequisites() []string {
	return nil
}

func (d *driverGPIO) After() []string {
	return []string{"sysfs-gpio"}
}

// Init does nothing if an allwinner processor is not detected. If one is
// detected, it memory maps gpio CPU registers and then sets up the pin mapping
// for the exact processor model detected.
func (d *driverGPIO) Init() (bool, error) {
	if !Present() {
		return false, errors.New("no Allwinner CPU detected")
	}

	// Mark the right pins as available even if the memory map fails so they can
	// callback to sysfs.Pins.
	switch {
	case IsA64():
		if err := mapA64Pins(); err != nil {
			return true, err
		}
	case IsR8():
		if err := mapR8Pins(); err != nil {
			return true, err
		}
	case IsA20():
		if err := mapA20Pins(); err != nil {
			return true, err
		}
	default:
		return false, errors.New("unknown Allwinner CPU model")
	}

	// gpioBaseAddr is the physical base address of the GPIO registers.
	gpioBaseAddr := uint32(getBaseAddress())
	if err := pmem.MapAsPOD(uint64(gpioBaseAddr), &d.gpioMemory); err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}

	return true, initPins()
}

func init() {
	if isArm {
		periph.MustRegister(&drvGPIO)
	}
}

// getBaseAddress queries the virtual file system to retrieve the base address
// of the GPIO registers for GPIO pins in groups PA to PI.
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

var drvGPIO driverGPIO

// Ensure that the various structs implement the interfaces they're supposed to.
var _ gpio.PinIO = &Pin{}
var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ pin.PinFunc = &Pin{}
