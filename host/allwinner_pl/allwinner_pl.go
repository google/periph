// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner_pl

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/periph"
	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/host/distro"
	"github.com/google/periph/host/pmem"
	"github.com/google/periph/host/sysfs"
)

// All the pins in the PL group.
var PL0, PL1, PL2, PL3, PL4, PL5, PL6, PL7, PL8, PL9, PL10, PL11, PL12 *PinPL

// Present returns true if running on an Allwinner CPU supporting the PL group.
func Present() bool {
	// BUG(maruel): Fix detection, need to specifically look for H3 and A64!
	if isArm {
		hardware, ok := distro.CPUInfo()["Hardware"]
		return ok && strings.HasPrefix(hardware, "sun")
		// /sys/class/sunxi_info/sys_info
	}
	return false
}

// PinPL defines one CPU supported pin in the PL group.
//
// PinPL implements gpio.PinIO.
type PinPL struct {
	// Immutable.
	offset      uint8     // as per register offset calculation
	name        string    // name as per datasheet
	defaultPull gpio.Pull // default pull at startup

	// Mutable.
	edge      *sysfs.Pin // Set once, then never set back to nil
	usingEdge bool       // Set when edge detection is enabled.
}

// PinIO implementation.

// String returns the pin name and number, ex: "PL5(352)".
func (p *PinPL) String() string {
	return fmt.Sprintf("%s(%d)", p.name, p.Number())
}

// Name returns the pin name, ex: "PL5".
func (p *PinPL) Name() string {
	return p.name
}

// Number returns the GPIO pin number as represented by gpio sysfs.
func (p *PinPL) Number() int {
	return 11*32 + int(p.offset)
}

// Function returns the current pin function, ex: "In/PullUp".
func (p *PinPL) Function() string {
	switch f := p.function(); f {
	case in:
		return "In/" + p.Read().String() + "/" + p.Pull().String()
	case out:
		return "Out/" + p.Read().String()
	case alt1:
		if s := mapping[p.offset][0]; len(s) != 0 {
			return s
		}
		return "<Alt1>"
	case alt2:
		if s := mapping[p.offset][1]; len(s) != 0 {
			return s
		}
		return "<Alt2>"
	case alt3:
		if s := mapping[p.offset][2]; len(s) != 0 {
			return s
		}
		return "<Alt3>"
	case alt4:
		if s := mapping[p.offset][3]; len(s) != 0 {
			return s
		}
		return "<Alt4>"
	case alt5:
		if s := mapping[p.offset][4]; len(s) != 0 {
			return s
		}
		return "<Alt5>"
	case disabled:
		return "<Disabled>"
	default:
		return "<Internal error>"
	}
}

// In implemented gpio.PinIn.
//
// This requires opening a gpio sysfs file handle. The pin will be exported at
// /sys/class/gpio/gpio*/. Note that the pin will not be unexported at
// shutdown.
//
// Not all pins support edge detection Allwinner processors!
func (p *PinPL) In(pull gpio.Pull, edge gpio.Edge) error {
	if gpioMemoryPL == nil {
		return p.wrap(errors.New("subsystem not initialized"))
	}
	if !p.setFunction(in) {
		return p.wrap(errors.New("failed to set pin as input"))
	}
	if pull != gpio.PullNoChange {
		off := p.offset / 16
		shift := 2 * (p.offset % 16)
		// Do it in a way that is concurrent safe.
		gpioMemoryPL.pull[off] &^= 3 << shift
		switch pull {
		case gpio.Down:
			gpioMemoryPL.pull[off] = 2 << shift
		case gpio.Up:
			gpioMemoryPL.pull[off] = 1 << shift
		default:
		}
	}
	wasUsing := p.usingEdge
	p.usingEdge = edge != gpio.None
	if p.usingEdge && p.edge == nil {
		ok := false
		n := p.Number()
		if p.edge, ok = sysfs.Pins[n]; !ok {
			return p.wrap(fmt.Errorf("pin %d is not exported by sysfs", n))
		}
	}
	if p.usingEdge || wasUsing {
		// This resets pending edges.
		if err := p.edge.In(gpio.PullNoChange, edge); err != nil {
			return p.wrap(err)
		}
	}
	return nil
}

// Read implements gpio.PinIn.
func (p *PinPL) Read() gpio.Level {
	return gpio.Level(gpioMemoryPL.data&(1<<p.offset) != 0)
}

// WaitForEdge does edge detection and implements gpio.PinIn.
func (p *PinPL) WaitForEdge(timeout time.Duration) bool {
	if p.edge != nil {
		return p.edge.WaitForEdge(timeout)
	}
	return false
}

// Pull implements gpio.PinIn.
func (p *PinPL) Pull() gpio.Pull {
	if gpioMemoryPL == nil {
		return gpio.PullNoChange
	}
	switch (gpioMemoryPL.pull[p.offset/16] >> (2 * (p.offset % 16))) & 3 {
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

// Out implements gpio.PinOut.
func (p *PinPL) Out(l gpio.Level) error {
	if gpioMemoryPL == nil {
		return p.wrap(errors.New("subsystem not initialized"))
	}
	if p.usingEdge {
		// First disable edges.
		if err := p.edge.In(gpio.PullNoChange, gpio.None); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	if !p.setFunction(out) {
		return p.wrap(errors.New("failed to set pin as output"))
	}
	// TODO(maruel): Set the value *before* changing the pin to be an output, so
	// there is no glitch.
	bit := uint32(1 << p.offset)
	if l {
		gpioMemoryPL.data |= bit
	} else {
		gpioMemoryPL.data &^= bit
	}
	return nil
}

// PWM implements gpio.PinOut.
func (p *PinPL) PWM(duty int) error {
	return p.wrap(errors.New("pwm is not supported"))
}

//

// function returns the current GPIO pin function.
func (p *PinPL) function() function {
	if gpioMemoryPL == nil {
		return disabled
	}
	shift := 4 * (p.offset % 8)
	return function((gpioMemoryPL.cfg[p.offset/8] >> shift) & 7)
}

// setFunction changes the GPIO pin function.
//
// Returns false if the pin was in AltN. Only accepts in and out
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
	gpioMemoryPL.cfg[off] |= mask
	gpioMemoryPL.cfg[off] &^= v
	if p.function() != f {
		panic(f)
	}
	return true
}

func (p *PinPL) wrap(err error) error {
	return fmt.Errorf("allwinner-gpio-pl (%s): %v", p, err)
}

//

// Page 23~24
// Each pin can have one of 7 functions.
const (
	in       function = 0
	out      function = 1
	alt1     function = 2
	alt2     function = 3
	alt3     function = 4
	alt4     function = 5
	alt5     function = 6
	disabled function = 7
)

// cpuPinsPL is all the pins as supported by the CPU. There is no guarantee that
// they are actually connected to anything on the board.
var cpuPinsPL = []PinPL{
	{offset: 0, name: "PL0", defaultPull: gpio.Up},
	{offset: 1, name: "PL1", defaultPull: gpio.Up},
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

// gpioMemoryPL is only the PL group in that case. Note that groups PI, PJ, PK
// do not exist.
var gpioMemoryPL *gpioGroup

// See ../allwinner/allwinner.go for details.
// TODO(maruel): Figure out what the S_ prefix means.
var mapping = [13][5]string{
	{"S_RSB_SCK", "S_I2C_SCL", "", "", "S_PL_EINT0"}, // PL0
	{"S_RSB_SDA", "S_I2C_SDA", "", "", "S_PL_EINT1"}, // PL1
	{"S_UART_TX", "", "", "", "S_PL_EINT2"},          // PL2
	{"S_UART_RX", "", "", "", "S_PL_EINT3"},          // PL3
	{"S_JTAG_MS", "", "", "", "S_PL_EINT4"},          // PL4
	{"S_JTAG_CK", "", "", "", "S_PL_EINT5"},          // PL5
	{"S_JTAG_DO", "", "", "", "S_PL_EINT6"},          // PL6
	{"S_JTAG_DI", "", "", "", "S_PL_EINT7"},          // PL7
	{"S_I2C_CSK", "", "", "", "S_PL_EINT8"},          // PL8
	{"S_I2C_SDA", "", "", "", "S_PL_EINT9"},          // PL9
	{"S_PWM", "", "", "", "S_PL_EINT10"},             // PL10
	{"S_CIR_RX", "", "", "", "S_PL_EINT11"},          // PL11
	{"", "", "", "", "S_PL_EINT12"},                  // PL12
}

// function specifies the active functionality of a pin. The alternative
// function is GPIO pin dependent.
type function uint8

// http://files.pine64.org/doc/datasheet/pine64/Allwinner_A64_User_Manual_V1.0.pdf
// Page 410 GPIO PL.
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

// getBaseAddress queries the virtual file system to retrieve the base address
// of the GPIO registers for GPIO pins in group PL.
//
// Defaults to 0x01F02C00 as per datasheet if could query the file system.
func getBaseAddress() uint64 {
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
}

func (d *driverGPIOPL) String() string {
	return "allwinner-gpio-pl"
}

func (d *driverGPIOPL) Prerequisites() []string {
	return []string{"allwinner-gpio"}
}

func (d *driverGPIOPL) Init() (bool, error) {
	if !Present() {
		return false, errors.New("A64 CPU not detected")
	}
	m, err := pmem.Map(getBaseAddress(), 4096)
	if err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}
	if err := m.Struct(reflect.ValueOf(&gpioMemoryPL)); err != nil {
		return true, err
	}

	for i := range cpuPinsPL {
		p := &cpuPinsPL[i]
		if err := gpio.Register(p, true); err != nil {
			return true, err
		}
		// TODO(maruel): There's a problem where multiple pins may be set to the
		// same function. Need investigation. For now just ignore errors.
		if f := p.Function(); f[0] != '<' && f[:2] != "In" && f[:3] != "Out" {
			// TODO(maruel): Stop ignoring errors by not registering the same
			// function multiple times.
			gpio.RegisterAlias(f, p.Number())
			/*
				if err := gpio.RegisterAlias(f, p.Number()); err != nil {
					return true, err
				}
			*/
		}
	}
	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&driverGPIOPL{})
	}
}

var _ gpio.PinIn = &PinPL{}
var _ gpio.PinOut = &PinPL{}
var _ gpio.PinIO = &PinPL{}
