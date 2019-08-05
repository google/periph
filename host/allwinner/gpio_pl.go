// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

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

// All the pins in the PL group.
var PL0, PL1, PL2, PL3, PL4, PL5, PL6, PL7, PL8, PL9, PL10, PL11, PL12 *PinPL

// PinPL defines one CPU supported pin in the PL group.
//
// PinPL implements gpio.PinIO.
type PinPL struct {
	// Immutable.
	offset      uint8     // as per register offset calculation
	name        string    // name as per datasheet
	defaultPull gpio.Pull // default pull at startup

	// Immutable after driver initialization.
	sysfsPin  *sysfs.Pin // Set to the corresponding sysfs.Pin, if any.
	available bool       // Set when the pin is available on this CPU architecture.

	// Mutable.
	usingEdge bool // Set when edge detection is enabled.
}

// String implements conn.Resource.
//
// It returns the pin name and number, ex: "PL5(352)".
func (p *PinPL) String() string {
	return fmt.Sprintf("%s(%d)", p.name, p.Number())
}

// Halt implements conn.Resource.
//
// It stops edge detection if enabled.
func (p *PinPL) Halt() error {
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
// It returns the pin name, ex: "PL5".
func (p *PinPL) Name() string {
	return p.name
}

// Number implements pin.Pin.
//
// It returns the GPIO pin number as represented by gpio sysfs.
func (p *PinPL) Number() int {
	return 11*32 + int(p.offset)
}

// Function implements pin.Pin.
func (p *PinPL) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *PinPL) Func() pin.Func {
	if !p.available {
		// We do not want the error message about uninitialized system.
		return pin.FuncNone
	}
	if drvGPIOPL.gpioMemoryPL == nil {
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
		if s := mappingPL[p.offset][0]; len(s) != 0 {
			return pin.Func(s)
		}
		return pin.Func("ALT1")
	case alt2:
		if s := mappingPL[p.offset][1]; len(s) != 0 {
			return pin.Func(s)
		}
		return pin.Func("ALT2")
	case alt3:
		if s := mappingPL[p.offset][2]; len(s) != 0 {
			return pin.Func(s)
		}
		return pin.Func("ALT3")
	case alt4:
		if s := mappingPL[p.offset][3]; len(s) != 0 {
			return pin.Func(s)
		}
		return pin.Func("ALT4")
	case alt5:
		if s := mappingPL[p.offset][4]; len(s) != 0 {
			if strings.Contains(string(s), "_EINT") {
				// It's an input supporting interrupts.
				if p.FastRead() {
					return gpio.IN_HIGH
				}
				return gpio.IN_LOW
			}
			return pin.Func(s)
		}
		return pin.Func("ALT5")
	case disabled:
		return pin.FuncNone
	default:
		return pin.FuncNone
	}
}

// SupportedFuncs implements pin.PinFunc.
func (p *PinPL) SupportedFuncs() []pin.Func {
	f := make([]pin.Func, 0, 2+2)
	f = append(f, gpio.IN, gpio.OUT)
	for _, m := range mappingPL[p.offset] {
		if m != pin.FuncNone && !strings.Contains(string(m), "_EINT") {
			f = append(f, m)
		}
	}
	return f
}

// SetFunc implements pin.PinFunc.
func (p *PinPL) SetFunc(f pin.Func) error {
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
		for i, m := range mappingPL[p.offset] {
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
func (p *PinPL) In(pull gpio.Pull, edge gpio.Edge) error {
	if !p.available {
		// We do not want the error message about uninitialized system.
		return p.wrap(errors.New("not available on this CPU architecture"))
	}
	if p.usingEdge && edge == gpio.NoEdge {
		if err := p.sysfsPin.Halt(); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	if drvGPIOPL.gpioMemoryPL == nil {
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
	if !p.setFunction(in) {
		return p.wrap(errors.New("failed to set pin as input"))
	}
	if pull != gpio.PullNoChange {
		off := p.offset / 16
		shift := 2 * (p.offset % 16)
		// Do it in a way that is concurrent safe.
		drvGPIOPL.gpioMemoryPL.pull[off] &^= 3 << shift
		switch pull {
		case gpio.PullDown:
			drvGPIOPL.gpioMemoryPL.pull[off] = 2 << shift
		case gpio.PullUp:
			drvGPIOPL.gpioMemoryPL.pull[off] = 1 << shift
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
func (p *PinPL) Read() gpio.Level {
	if drvGPIOPL.gpioMemoryPL == nil {
		if p.sysfsPin == nil {
			return gpio.Low
		}
		return p.sysfsPin.Read()
	}
	return gpio.Level(drvGPIOPL.gpioMemoryPL.data&(1<<p.offset) != 0)
}

// FastRead reads without verification.
func (p *PinPL) FastRead() gpio.Level {
	return gpio.Level(drvGPIOPL.gpioMemoryPL.data&(1<<p.offset) != 0)
}

// WaitForEdge implements gpio.PinIn.
func (p *PinPL) WaitForEdge(timeout time.Duration) bool {
	if p.sysfsPin != nil {
		return p.sysfsPin.WaitForEdge(timeout)
	}
	return false
}

// Pull implements gpio.PinIn.
func (p *PinPL) Pull() gpio.Pull {
	if drvGPIOPL.gpioMemoryPL == nil {
		// If gpioMemoryPL is set, p.available is true.
		return gpio.PullNoChange
	}
	switch (drvGPIOPL.gpioMemoryPL.pull[p.offset/16] >> (2 * (p.offset % 16))) & 3 {
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
func (p *PinPL) DefaultPull() gpio.Pull {
	return p.defaultPull
}

// Out implements gpio.PinOut.
func (p *PinPL) Out(l gpio.Level) error {
	if !p.available {
		// We do not want the error message about uninitialized system.
		return p.wrap(errors.New("not available on this CPU architecture"))
	}
	if drvGPIOPL.gpioMemoryPL == nil {
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
	if !p.setFunction(out) {
		return p.wrap(errors.New("failed to set pin as output"))
	}
	return nil
}

// FastOut sets a pin output level with Absolutely No error checking.
//
// See Pin.FastOut for more information.
func (p *PinPL) FastOut(l gpio.Level) {
	bit := uint32(1 << p.offset)
	if l {
		drvGPIOPL.gpioMemoryPL.data |= bit
	} else {
		drvGPIOPL.gpioMemoryPL.data &^= bit
	}
}

// PWM implements gpio.PinOut.
func (p *PinPL) PWM(gpio.Duty, physic.Frequency) error {
	// TODO(maruel): PWM support for PL10.
	return p.wrap(errors.New("not available on this CPU architecture"))
}

//

// function returns the current GPIO pin function.
//
// It must not be called if drvGPIOPL.gpioMemoryPL is nil.
func (p *PinPL) function() function {
	shift := 4 * (p.offset % 8)
	return function((drvGPIOPL.gpioMemoryPL.cfg[p.offset/8] >> shift) & 7)
}

// setFunction changes the GPIO pin function.
//
// Returns false if the pin was in AltN. Only accepts in and out
//
// It must not be called if drvGPIOPL.gpioMemoryPL is nil.
func (p *PinPL) setFunction(f function) bool {
	if f != in && f != out {
		return false
	}
	// Interrupt based edge triggering is Alt5 but this is only supported on some
	// pins.
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
	drvGPIOPL.gpioMemoryPL.cfg[off] |= mask
	drvGPIOPL.gpioMemoryPL.cfg[off] &^= v
	if p.function() != f {
		panic(f)
	}
	return true
}

func (p *PinPL) wrap(err error) error {
	return fmt.Errorf("allwinner-gpio-pl (%s): %v", p, err)
}

//

// cpuPinsPL is all the pins as supported by the CPU. There is no guarantee that
// they are actually connected to anything on the board.
var cpuPinsPL = []PinPL{
	{offset: 0, name: "PL0", defaultPull: gpio.PullUp},
	{offset: 1, name: "PL1", defaultPull: gpio.PullUp},
	{offset: 2, name: "PL2", defaultPull: gpio.Float},
	{offset: 3, name: "PL3", defaultPull: gpio.Float},
	{offset: 4, name: "PL4", defaultPull: gpio.Float},
	{offset: 5, name: "PL5", defaultPull: gpio.Float},
	{offset: 6, name: "PL6", defaultPull: gpio.Float},
	{offset: 7, name: "PL7", defaultPull: gpio.Float},
	{offset: 8, name: "PL8", defaultPull: gpio.Float},
	{offset: 9, name: "PL9", defaultPull: gpio.Float},
	{offset: 10, name: "PL10", defaultPull: gpio.Float},
	{offset: 11, name: "PL11", defaultPull: gpio.Float},
	{offset: 12, name: "PL12", defaultPull: gpio.Float},
}

// See gpio.go for details.
var mappingPL = [13][5]pin.Func{
	{"RSB_SCK", "I2C_SCL", "", "", "PL_EINT0"}, // PL0
	{"RSB_SDA", "I2C_SDA", "", "", "PL_EINT1"}, // PL1
	{"UART_TX", "", "", "", "PL_EINT2"},        // PL2
	{"UART_RX", "", "", "", "PL_EINT3"},        // PL3
	{"JTAG_TMS", "", "", "", "PL_EINT4"},       // PL4
	{"JTAG_TCK", "", "", "", "PL_EINT5"},       // PL5
	{"JTAG_TDO", "", "", "", "PL_EINT6"},       // PL6
	{"JTAG_TDI", "", "", "", "PL_EINT7"},       // PL7
	{"I2C_SCL", "", "", "", "PL_EINT8"},        // PL8
	{"I2C_SDA", "", "", "", "PL_EINT9"},        // PL9
	{"PWM0", "", "", "", "PL_EINT10"},          // PL10
	{"CIR_RX", "", "", "", "PL_EINT11"},        // PL11
	{"", "", "", "", "PL_EINT12"},              // PL12
}

func init() {
	PL0 = &cpuPinsPL[0]
	PL1 = &cpuPinsPL[1]
	PL2 = &cpuPinsPL[2]
	PL3 = &cpuPinsPL[3]
	PL4 = &cpuPinsPL[4]
	PL5 = &cpuPinsPL[5]
	PL6 = &cpuPinsPL[6]
	PL7 = &cpuPinsPL[7]
	PL8 = &cpuPinsPL[8]
	PL9 = &cpuPinsPL[9]
	PL10 = &cpuPinsPL[10]
	PL11 = &cpuPinsPL[11]
	PL12 = &cpuPinsPL[12]
}

// getBaseAddressPL queries the virtual file system to retrieve the base address
// of the GPIO registers for GPIO pins in group PL.
//
// Defaults to 0x01F02C00 as per datasheet if could query the file system.
func getBaseAddressPL() uint64 {
	base := uint64(0x01F02C00)
	link, err := os.Readlink("/sys/bus/platform/drivers/sun50i-r-pinctrl/driver")
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

// driverGPIOPL implements periph.Driver.
type driverGPIOPL struct {
	// gpioMemoryPL is only the PL group in that case. Note that groups PI, PJ, PK
	// do not exist.
	gpioMemoryPL *gpioGroup
}

func (d *driverGPIOPL) String() string {
	return "allwinner-gpio-pl"
}

func (d *driverGPIOPL) Prerequisites() []string {
	return nil
}

func (d *driverGPIOPL) After() []string {
	return []string{"sysfs-gpio"}
}

func (d *driverGPIOPL) Init() (bool, error) {
	// BUG(maruel): H3 supports group PL too.
	if !IsA64() {
		return false, errors.New("A64 CPU not detected")
	}

	// Mark the right pins as available even if the memory map fails so they can
	// callback to sysfs.Pins.
	functions := map[pin.Func]struct{}{}
	for i := range cpuPinsPL {
		name := cpuPinsPL[i].Name()
		num := strconv.Itoa(cpuPinsPL[i].Number())
		cpuPinsPL[i].available = true
		gpion := "GPIO" + num

		// Unregister the pin if already registered. This happens with sysfs-gpio.
		// Do not error on it, since sysfs-gpio may have failed to load.
		_ = gpioreg.Unregister(gpion)
		_ = gpioreg.Unregister(num)

		// Register the pin with gpio.
		if err := gpioreg.Register(&cpuPinsPL[i]); err != nil {
			return true, err
		}
		if err := gpioreg.RegisterAlias(gpion, name); err != nil {
			return true, err
		}
		if err := gpioreg.RegisterAlias(num, name); err != nil {
			return true, err
		}
		switch f := cpuPinsPL[i].Func(); f {
		case gpio.IN, gpio.OUT, pin.FuncNone:
		default:
			// Registering the same alias twice fails. This can happen if two pins
			// are configured with the same function.
			if _, ok := functions[f]; !ok {
				// TODO(maruel): We'd have to clear out the ones from allwinner-gpio
				// too.
				functions[f] = struct{}{}
				_ = gpioreg.RegisterAlias(string(f), name)
			}
		}
	}

	// Now do a second loop but do the alternate functions.
	for i := range cpuPinsPL {
		for _, f := range cpuPinsPL[i].SupportedFuncs() {
			switch f {
			case gpio.IN, gpio.OUT:
			default:
				if _, ok := functions[f]; !ok {
					// TODO(maruel): We'd have to clear out the ones from allwinner-gpio
					// too.
					functions[f] = struct{}{}
					_ = gpioreg.RegisterAlias(string(f), cpuPinsPL[i].name)
				}
			}
		}
	}

	m, err := pmem.Map(getBaseAddressPL(), 4096)
	if err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}
	if err := m.AsPOD(&d.gpioMemoryPL); err != nil {
		return true, err
	}

	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&drvGPIOPL)
	}
}

var drvGPIOPL driverGPIOPL

var _ gpio.PinIO = &PinPL{}
var _ gpio.PinIn = &PinPL{}
var _ gpio.PinOut = &PinPL{}
var _ pin.PinFunc = &PinPL{}
