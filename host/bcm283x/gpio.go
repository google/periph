// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host/distro"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/sysfs"
)

// All the pins supported by the CPU.
var (
	GPIO0  *Pin // I2C0_SDA
	GPIO1  *Pin // I2C0_SCL
	GPIO2  *Pin // I2C1_SDA
	GPIO3  *Pin // I2C1_SCL
	GPIO4  *Pin // GPCLK0
	GPIO5  *Pin // GPCLK1
	GPIO6  *Pin // GPCLK2
	GPIO7  *Pin // SPI0_CS1
	GPIO8  *Pin // SPI0_CS0
	GPIO9  *Pin // SPI0_MISO
	GPIO10 *Pin // SPI0_MOSI
	GPIO11 *Pin // SPI0_CLK
	GPIO12 *Pin // PWM0_OUT
	GPIO13 *Pin // PWM1_OUT
	GPIO14 *Pin // UART0_TXD, UART1_TXD
	GPIO15 *Pin // UART0_RXD, UART1_RXD
	GPIO16 *Pin // UART0_CTS, SPI1_CS2, UART1_CTS
	GPIO17 *Pin // UART0_RTS, SPI1_CS1, UART1_RTS
	GPIO18 *Pin // PCM_CLK, SPI1_CS0, PWM0_OUT
	GPIO19 *Pin // PCM_FS, SPI1_MISO, PWM1_OUT
	GPIO20 *Pin // PCM_DIN, SPI1_MOSI, GPCLK0
	GPIO21 *Pin // PCM_DOUT, SPI1_CLK, GPCLK1
	GPIO22 *Pin //
	GPIO23 *Pin //
	GPIO24 *Pin //
	GPIO25 *Pin //
	GPIO26 *Pin //
	GPIO27 *Pin //
	GPIO28 *Pin // I2C0_SDA, PCM_CLK
	GPIO29 *Pin // I2C0_SCL, PCM_FS
	GPIO30 *Pin // PCM_DIN, UART0_CTS, UART1_CTS
	GPIO31 *Pin // PCM_DOUT, UART0_RTS, UART1_RTS
	GPIO32 *Pin // GPCLK0, UART0_TXD, UART1_TXD
	GPIO33 *Pin // UART0_RXD, UART1_RXD
	GPIO34 *Pin // GPCLK0
	GPIO35 *Pin // SPI0_CS1
	GPIO36 *Pin // SPI0_CS0, UART0_TXD
	GPIO37 *Pin // SPI0_MISO, UART0_RXD
	GPIO38 *Pin // SPI0_MOSI, UART0_RTS
	GPIO39 *Pin // SPI0_CLK, UART0_CTS
	GPIO40 *Pin // PWM0_OUT, SPI2_MISO, UART1_TXD
	GPIO41 *Pin // PWM1_OUT, SPI2_MOSI, UART1_RXD
	GPIO42 *Pin // GPCLK1, SPI2_CLK, UART1_RTS
	GPIO43 *Pin // GPCLK2, SPI2_CS0, UART1_CTS
	GPIO44 *Pin // GPCLK1, I2C0_SDA, I2C1_SDA, SPI2_CS1
	GPIO45 *Pin // PWM1_OUT, I2C0_SCL, I2C1_SCL, SPI2_CS2
	GPIO46 *Pin //
	// Pins 47~53 are not exposed because using them would lead to immediate SD
	// Card corruption.
)

// Present returns true if running on a Broadcom bcm283x based CPU.
func Present() bool {
	if isArm {
		hardware, ok := distro.CPUInfo()["Hardware"]
		return ok && strings.HasPrefix(hardware, "BCM")
	}
	return false
}

// Pin is a GPIO number (GPIOnn) on BCM238(5|6|7).
//
// Pin implements gpio.PinIO.
type Pin struct {
	// Immutable.
	number      int
	name        string
	defaultPull gpio.Pull

	// Mutable.
	edge       *sysfs.Pin // Set once, then never set back to nil.
	usingEdge  bool       // Set when edge detection is enabled.
	usingClock bool       // Set when a GPCLK or PWM clock is used.
}

// PinIO implementation.

// String returns the pin name, ex: "GPIO10".
func (p *Pin) String() string {
	return p.name
}

// Name returns the pin name, ex: "GPIO10".
func (p *Pin) Name() string {
	return p.name
}

// Number returns the pin number as assigned by gpio sysfs.
func (p *Pin) Number() int {
	return p.number
}

// Function returns the current pin function, ex: "In/PullUp".
func (p *Pin) Function() string {
	switch f := p.function(); f {
	case in:
		return "In/" + p.Read().String()
	case out:
		return "Out/" + p.Read().String()
	case alt0:
		if s := mapping[p.number][0]; len(s) != 0 {
			return s
		}
		return "<Alt0>"
	case alt1:
		if s := mapping[p.number][1]; len(s) != 0 {
			return s
		}
		return "<Alt1>"
	case alt2:
		if s := mapping[p.number][2]; len(s) != 0 {
			return s
		}
		return "<Alt2>"
	case alt3:
		if s := mapping[p.number][3]; len(s) != 0 {
			return s
		}
		return "<Alt3>"
	case alt4:
		if s := mapping[p.number][4]; len(s) != 0 {
			return s
		}
		return "<Alt4>"
	case alt5:
		if s := mapping[p.number][5]; len(s) != 0 {
			return s
		}
		return "<Alt5>"
	default:
		return "<Unknown>"
	}
}

// In setups a pin as an input and implements gpio.PinIn.
//
// Specifying a value for pull other than gpio.PullNoChange causes this
// function to be slightly slower (about 1ms).
//
// For pull down, the resistor is 50KOhm~60kOhm
// For pull up, the resistor is 50kOhm~65kOhm
//
// The pull resistor stays set even after the processor shuts down. It is not
// possible to 'read back' what value was specified for each pin.
//
// Will fail if requesting to change a pin that is set to special functionality.
//
// Using edge detection requires opening a gpio sysfs file handle. On Raspbian,
// make sure the user is member of group 'gpio'. The pin will be exported at
// /sys/class/gpio/gpio*/. Note that the pin will not be unexported at
// shutdown.
//
// For edge detection, the processor samples the input at its CPU clock rate
// and looks for '011' to rising and '100' for falling detection to avoid
// glitches. Because gpio sysfs is used, the latency is unpredictable.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	if gpioMemory == nil {
		return p.wrap(errors.New("subsystem not initialized"))
	}
	if err := p.haltClock(); err != nil {
		return err
	}
	p.setFunction(in)
	if pull != gpio.PullNoChange {
		// Changing pull resistor requires a specific dance as described at
		// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
		// page 101.

		// Set Pull
		switch pull {
		case gpio.PullDown:
			gpioMemory.pullEnable = 1
		case gpio.PullUp:
			gpioMemory.pullEnable = 2
		case gpio.Float:
			gpioMemory.pullEnable = 0
		}

		// Datasheet states caller needs to sleep 150 cycles.
		sleep150cycles()
		offset := p.number / 32
		gpioMemory.pullEnableClock[offset] = 1 << uint(p.number%32)

		sleep150cycles()
		gpioMemory.pullEnable = 0
		gpioMemory.pullEnableClock[offset] = 0
	}
	wasUsing := p.usingEdge
	p.usingEdge = edge != gpio.NoEdge
	if p.usingEdge {
		if p.edge == nil {
			ok := false
			n := p.Number()
			if p.edge, ok = sysfs.Pins[n]; !ok {
				return p.wrap(fmt.Errorf("pin %d is not exported by sysfs", n))
			}
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

// Read return the current pin level and implements gpio.PinIn.
//
// This function is very fast. It works even if the pin is set as output.
func (p *Pin) Read() gpio.Level {
	if gpioMemory == nil {
		return gpio.Low
	}
	return gpio.Level((gpioMemory.level[p.number/32] & (1 << uint(p.number&31))) != 0)
}

// WaitForEdge does edge detection and implements gpio.PinIn.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	if p.edge != nil {
		return p.edge.WaitForEdge(timeout)
	}
	return false
}

// Pull implemented gpio.PinIn.
//
// bcm283x doesn't support querying the pull resistor of any GPIO pin.
func (p *Pin) Pull() gpio.Pull {
	// TODO(maruel): The best that could be added is to cache the last set value
	// and return it.
	return gpio.PullNoChange
}

// Out sets a pin as output and implements gpio.PinOut.
//
// Fails if requesting to change a pin that is set to special functionality.
func (p *Pin) Out(l gpio.Level) error {
	if gpioMemory == nil {
		return p.wrap(errors.New("subsystem not initialized"))
	}
	if err := p.Halt(); err != nil {
		return err
	}
	// Change output before changing mode to not create any glitch.
	offset := p.number / 32
	if l == gpio.Low {
		gpioMemory.outputClear[offset] = 1 << uint(p.number&31)
	} else {
		gpioMemory.outputSet[offset] = 1 << uint(p.number&31)
	}
	p.setFunction(out)
	return nil
}

// Halt implements conn.Resource.
//
// If the pin is running a clock, PWM or waiting for edges, it is halted.
//
// In the case of clock or PWM, all pins with this clock source are also
// disabled.
func (p *Pin) Halt() error {
	if p.usingEdge {
		if err := p.edge.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	return p.haltClock()
}

// BUG(maruel): PWM(): There is no conflict verification when multiple pins are
// used simultaneously. The last call to PWM() will affect all pins of the same
// type (GPCLK0, GPCLK2, PWM0 or PWM1).

// PWM outputs a periodic signal on supported pins.
//
// PWM pins
//
// PWM0 is exposed on pins 12, 18 and 40.
//
// PWM1 is exposed on pins 13, 19, 41 and 45.
//
// PWM0 and PWM1 share the same 25Mhz clock source. The period must be a
// divisor of 25Mhz.
//
// Clock pins
//
// Clock GPCLK0 is exposed on pins 4, 20 32 and 34.
//
// Clock GPCLK2 is exposed on pins 6 and 43.
//
// Clocks can only be used with duty 0, gpio.DutyHalf or gpio.DutyMax. They
// also have relatively limited frequency range from 4.8kHz to 1MHz with a
// specific set of values supported.
//
// Furthermore, these can only be used if the drive "bcm283x-dma" was loaded.
// It can only be loaded if the process has root level access.
//
// The user must call either Halt(), In(), Out(), PWM(0,..) or
// PWM(gpio.DutyMax,..) to stop the clock source before exiting the program.
func (p *Pin) PWM(duty gpio.Duty, period time.Duration) error {
	if duty == 0 {
		return p.Out(gpio.Low)
	} else if duty == gpio.DutyMax {
		return p.Out(gpio.High)
	}
	if period < 500*time.Nanosecond {
		// High clock rate tends to hang the RPi. Need to investigate more.
		return p.wrap(errors.New("period must be at least 500ns"))
	}
	f := alt0
	clkID := -1
	switch p.number {
	case 4, 32, 34:
		if duty != gpio.DutyHalf {
			return p.wrap(errors.New("pin can only be used with 50% duty cycle"))
		}
		clkID = 0
	case 5, 21, 42, 44:
		return p.wrap(errors.New("GPCLK1 cannot be safely used"))
	case 6, 43:
		if duty != gpio.DutyHalf {
			return p.wrap(errors.New("pin can only be used with 50% duty cycle"))
		}
		clkID = 2
	case 20:
		if duty != gpio.DutyHalf {
			return p.wrap(errors.New("pin can only be used with 50% duty cycle"))
		}
		clkID = 0
		f = alt5
	case 12, 13, 40, 41, 45: // PWM
	case 18, 19: // PWM
		f = alt5
	default:
		// Technically 52 and 53 could also support PWM as alt1 but they are assumed
		// to be used for the SD Card.
		return p.wrap(errors.New("pwm is not supported on this pin"))
	}

	// Intentionally check later, so a more informative error is returned on
	// unsupported pins.
	if gpioMemory == nil {
		return p.wrap(errors.New("subsystem not initialized"))
	}
	if pwmMemory == nil || clockMemory == nil {
		return p.wrap(errors.New("bcm283x-dma not initialized; try again as root?"))
	}

	// Convert period to frequency. This is lossy.
	hz := uint64(time.Second / period)

	if clkID != -1 {
		// Using a clock pin.
		var clk *clock
		switch clkID {
		case 0:
			clk = &clockMemory.gp0
		case 2:
			clk = &clockMemory.gp2
		}
		actual, waits, err := clk.set(hz, 1)
		if err != nil {
			return p.wrap(err)
		}
		p.usingClock = true
		if actual != hz {
			return p.wrap(fmt.Errorf("asked for %dHz, got %dHz", hz, actual))
		}
		if waits != 1 {
			return p.wrap(fmt.Errorf("got oversampled clock by %dx", waits))
		}
		p.setFunction(f)
		return nil
	}

	// TODO(maruel): Leverage oversampling.
	base_freq := uint64(25 * 1000 * 1000) // 25MHz
	// Total cycles in the period
	rng := base_freq * uint64(period) / uint64(time.Second)
	// Pulse width cycles
	dat := uint32(rng * uint64(duty) / uint64(gpio.DutyMax))
	if _, _, err := clockMemory.pwm.set(base_freq, 1); err != nil {
		return p.wrap(err)
	}
	p.usingClock = true
	// Bit shift for PWM0 and PWM1
	shift := uint((p.number & 1) * 8)
	if shift == 0 {
		pwmMemory.rng1 = uint32(rng)
		Nanospin(10 * time.Nanosecond)
		pwmMemory.dat1 = uint32(dat)
	} else {
		pwmMemory.rng2 = uint32(rng)
		Nanospin(10 * time.Nanosecond)
		pwmMemory.dat2 = uint32(dat)
	}
	Nanospin(10 * time.Nanosecond)
	old := pwmMemory.ctl
	pwmMemory.ctl = (old & ^(0xff << shift)) | ((pwm1Enable | pwm1MS) << shift)
	p.setFunction(f)
	return nil
}

// DefaultPull returns the default pull for the pin.
//
// Implements gpio.PinDefaultPull.
//
// The CPU doesn't return the current pull.
func (p *Pin) DefaultPull() gpio.Pull {
	return p.defaultPull
}

// Internal code.

// haltClock disables the GPCLK/PWM clock if used.
func (p *Pin) haltClock() error {
	if !p.usingClock {
		return nil
	}
	p.usingClock = false
	switch p.number {
	// GPCLKx
	case 4, 20, 32, 34:
		if _, _, err := clockMemory.gp0.set(0, 0); err != nil {
			return p.wrap(err)
		}
		return nil
	case 5, 21, 42, 44:
		return p.wrap(errors.New("GPCLK1 cannot be safely used"))
	case 6, 43:
		if _, _, err := clockMemory.gp2.set(0, 0); err != nil {
			return p.wrap(err)
		}
		return nil

	// PWMx
	case 12, 13, 18, 19, 40, 45:
		if _, _, err := clockMemory.pwm.set(0, 0); err != nil {
			return p.wrap(err)
		}
		// Bit shift for PWM0 and PWM1.
		shift := uint((p.number & 1) * 8)
		pwmMemory.ctl &= ^(0xff << shift)
		return nil

	default:
		// Technically 52 and 53 could also support PWM as alt1 but they are assumed
		// to be used for the SD Card.
		return p.wrap(errors.New("pwm is not supported on this pin"))
	}
}

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	if gpioMemory == nil {
		return alt5
	}
	return function((gpioMemory.functionSelect[p.number/10] >> uint((p.number%10)*3)) & 7)
}

// setFunction changes the GPIO pin function.
func (p *Pin) setFunction(f function) {
	off := p.number / 10
	shift := uint(p.number%10) * 3
	gpioMemory.functionSelect[off] = (gpioMemory.functionSelect[off] &^ (7 << shift)) | (uint32(f) << shift)
}

func (p *Pin) wrap(err error) error {
	return fmt.Errorf("bcm283x-gpio (%s): %v", p, err)
}

//

// Each pin can have one of 7 functions.
const (
	in   function = 0
	out  function = 1
	alt0 function = 4
	alt1 function = 5
	alt2 function = 6
	alt3 function = 7
	alt4 function = 3
	alt5 function = 2
)

var (
	// baseAddr is the base for all the CPU registers.
	//
	// It is initialized by driverGPIO.Init().
	baseAddr uint32
	// dramBus is high bits to address uncached memory. See virtToUncachedPhys()
	// in dma.go.
	dramBus uint32
	// gpioMemory is the memory map of the CPU GPIO registers.
	gpioMemory *gpioMap
	// gpioBaseAddr is needed for DMA transfers.
	gpioBaseAddr uint32
)

// cpuPins is all the pins as supported by the CPU. There is no guarantee that
// they are actually connected to anything on the board.
var cpuPins = []Pin{
	{number: 0, name: "GPIO0", defaultPull: gpio.PullUp},
	{number: 1, name: "GPIO1", defaultPull: gpio.PullUp},
	{number: 2, name: "GPIO2", defaultPull: gpio.PullUp},
	{number: 3, name: "GPIO3", defaultPull: gpio.PullUp},
	{number: 4, name: "GPIO4", defaultPull: gpio.PullUp},
	{number: 5, name: "GPIO5", defaultPull: gpio.PullUp},
	{number: 6, name: "GPIO6", defaultPull: gpio.PullUp},
	{number: 7, name: "GPIO7", defaultPull: gpio.PullUp},
	{number: 8, name: "GPIO8", defaultPull: gpio.PullUp},
	{number: 9, name: "GPIO9", defaultPull: gpio.PullDown},
	{number: 10, name: "GPIO10", defaultPull: gpio.PullDown},
	{number: 11, name: "GPIO11", defaultPull: gpio.PullDown},
	{number: 12, name: "GPIO12", defaultPull: gpio.PullDown},
	{number: 13, name: "GPIO13", defaultPull: gpio.PullDown},
	{number: 14, name: "GPIO14", defaultPull: gpio.PullDown},
	{number: 15, name: "GPIO15", defaultPull: gpio.PullDown},
	{number: 16, name: "GPIO16", defaultPull: gpio.PullDown},
	{number: 17, name: "GPIO17", defaultPull: gpio.PullDown},
	{number: 18, name: "GPIO18", defaultPull: gpio.PullDown},
	{number: 19, name: "GPIO19", defaultPull: gpio.PullDown},
	{number: 20, name: "GPIO20", defaultPull: gpio.PullDown},
	{number: 21, name: "GPIO21", defaultPull: gpio.PullDown},
	{number: 22, name: "GPIO22", defaultPull: gpio.PullDown},
	{number: 23, name: "GPIO23", defaultPull: gpio.PullDown},
	{number: 24, name: "GPIO24", defaultPull: gpio.PullDown},
	{number: 25, name: "GPIO25", defaultPull: gpio.PullDown},
	{number: 26, name: "GPIO26", defaultPull: gpio.PullDown},
	{number: 27, name: "GPIO27", defaultPull: gpio.PullDown},
	{number: 28, name: "GPIO28", defaultPull: gpio.Float},
	{number: 29, name: "GPIO29", defaultPull: gpio.Float},
	{number: 30, name: "GPIO30", defaultPull: gpio.PullDown},
	{number: 31, name: "GPIO31", defaultPull: gpio.PullDown},
	{number: 32, name: "GPIO32", defaultPull: gpio.PullDown},
	{number: 33, name: "GPIO33", defaultPull: gpio.PullDown},
	{number: 34, name: "GPIO34", defaultPull: gpio.PullUp},
	{number: 35, name: "GPIO35", defaultPull: gpio.PullUp},
	{number: 36, name: "GPIO36", defaultPull: gpio.PullUp},
	{number: 37, name: "GPIO37", defaultPull: gpio.PullDown},
	{number: 38, name: "GPIO38", defaultPull: gpio.PullDown},
	{number: 39, name: "GPIO39", defaultPull: gpio.PullDown},
	{number: 40, name: "GPIO40", defaultPull: gpio.PullDown},
	{number: 41, name: "GPIO41", defaultPull: gpio.PullDown},
	{number: 42, name: "GPIO42", defaultPull: gpio.PullDown},
	{number: 43, name: "GPIO43", defaultPull: gpio.PullDown},
	{number: 44, name: "GPIO44", defaultPull: gpio.Float},
	{number: 45, name: "GPIO45", defaultPull: gpio.Float},
	{number: 46, name: "GPIO46", defaultPull: gpio.PullUp},
}

// This excludes the functions in and out.
var mapping = [][6]string{
	{"I2C0_SDA"}, // 0
	{"I2C0_SCL"},
	{"I2C1_SDA"},
	{"I2C1_SCL"},
	{"GPCLK0"},
	{"GPCLK1"}, // 5
	{"GPCLK2"},
	{"SPI0_CS1"},
	{"SPI0_CS0"},
	{"SPI0_MISO"},
	{"SPI0_MOSI"}, // 10
	{"SPI0_CLK"},
	{"PWM0_OUT"},
	{"PWM1_OUT"},
	{"UART0_TXD", "", "", "", "", "UART1_TXD"},
	{"UART0_RXD", "", "", "", "", "UART1_RXD"}, // 15
	{"", "", "", "UART0_CTS", "SPI1_CS2", "UART1_CTS"},
	{"", "", "", "UART0_RTS", "SPI1_CS1", "UART1_RTS"},
	// TODO(maruel): Alias the PCM_xxx to I2S0_xxx.
	{"PCM_CLK", "", "", "", "SPI1_CS0", "PWM0_OUT"},
	{"PCM_FS", "", "", "", "SPI1_MISO", "PWM1_OUT"},
	{"PCM_DIN", "", "", "", "SPI1_MOSI", "GPCLK0"}, // 20
	{"PCM_DOUT", "", "", "", "SPI1_CLK", "GPCLK1"},
	{""},
	{""},
	{""},
	{""}, // 25
	{""},
	{""},
	{"I2C0_SDA", "", "PCM_CLK", "", "", ""},
	{"I2C0_SCL", "", "PCM_FS", "", "", ""},
	{"", "", "PCM_DIN", "UART0_CTS", "", "UART1_CTS"}, // 30
	{"", "", "PCM_DOUT", "UART0_RTS", "", "UART1_RTS"},
	{"GPCLK0", "", "", "UART0_TXD", "", "UART1_TXD"},
	{"", "", "", "UART0_RXD", "", "UART1_RXD"},
	{"GPCLK0"},
	{"SPI0_CS1"}, // 35
	{"SPI0_CS0", "", "UART0_TXD", "", "", ""},
	{"SPI0_MISO", "", "UART0_RXD", "", "", ""},
	{"SPI0_MOSI", "", "UART0_RTS", "", "", ""},
	{"SPI0_CLK", "", "UART0_CTS", "", "", ""},
	{"PWM0_OUT", "", "", "", "SPI2_MISO", "UART1_TXD"}, // 40
	{"PWM1_OUT", "", "", "", "SPI2_MOSI", "UART1_RXD"},
	{"GPCLK1", "", "", "", "SPI2_CLK", "UART1_RTS"},
	{"GPCLK2", "", "", "", "SPI2_CS0", "UART1_CTS"},
	{"GPCLK1", "I2C0_SDA", "I2C1_SDA", "", "SPI2_CS1", ""},
	{"PWM1_OUT", "I2C0_SCL", "I2C1_SCL", "", "SPI2_CS2", ""}, // 45
	{""},
}

// function specifies the active functionality of a pin. The alternative
// function is GPIO pin dependent.
type function uint8

// Mapping as
// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
// pages 90-91.
type gpioMap struct {
	// 0x00    RW   GPIO Function Select 0 (GPIO0-9)
	// 0x04    RW   GPIO Function Select 1 (GPIO10-19)
	// 0x08    RW   GPIO Function Select 2 (GPIO20-29)
	// 0x0C    RW   GPIO Function Select 3 (GPIO30-39)
	// 0x10    RW   GPIO Function Select 4 (GPIO40-49)
	// 0x14    RW   GPIO Function Select 5 (GPIO50-53)
	functionSelect [6]uint32 // GPFSEL0~GPFSEL5
	// 0x18    -    Reserved
	dummy0 uint32
	// 0x1C    W    GPIO Pin Output Set 0 (GPIO0-31)
	// 0x20    W    GPIO Pin Output Set 1 (GPIO32-53)
	outputSet [2]uint32 // GPSET0-GPSET1
	// 0x24    -    Reserved
	dummy1 uint32
	// 0x28    W    GPIO Pin Output Clear 0 (GPIO0-31)
	// 0x2C    W    GPIO Pin Output Clear 1 (GPIO32-53)
	outputClear [2]uint32 // GPCLR0-GPCLR1
	// 0x30    -    Reserved
	dummy2 uint32
	// 0x34    R    GPIO Pin Level 0 (GPIO0-31)
	// 0x38    R    GPIO Pin Level 1 (GPIO32-53)
	level [2]uint32 // GPLEV0-GPLEV1
	// 0x3C    -    Reserved
	dummy3 uint32
	// 0x40    RW   GPIO Pin Event Detect Status 0 (GPIO0-31)
	// 0x44    RW   GPIO Pin Event Detect Status 1 (GPIO32-53)
	eventDetectStatus [2]uint32 // GPEDS0-GPEDS1
	// 0x48    -    Reserved
	dummy4 uint32
	// 0x4C    RW   GPIO Pin Rising Edge Detect Enable 0 (GPIO0-31)
	// 0x50    RW   GPIO Pin Rising Edge Detect Enable 1 (GPIO32-53)
	risingEdgeDetectEnable [2]uint32 // GPREN0-GPREN1
	// 0x54    -    Reserved
	dummy5 uint32
	// 0x58    RW   GPIO Pin Falling Edge Detect Enable 0 (GPIO0-31)
	// 0x5C    RW   GPIO Pin Falling Edge Detect Enable 1 (GPIO32-53)
	fallingEdgeDetectEnable [2]uint32 // GPFEN0-GPFEN1
	// 0x60    -    Reserved
	dummy6 uint32
	// 0x64    RW   GPIO Pin High Detect Enable 0 (GPIO0-31)
	// 0x68    RW   GPIO Pin High Detect Enable 1 (GPIO32-53)
	highDetectEnable [2]uint32 // GPHEN0-GPHEN1
	// 0x6C    -    Reserved
	dummy7 uint32
	// 0x70    RW   GPIO Pin Low Detect Enable 0 (GPIO0-31)
	// 0x74    RW   GPIO Pin Low Detect Enable 1 (GPIO32-53)
	lowDetectEnable [2]uint32 // GPLEN0-GPLEN1
	// 0x78    -    Reserved
	dummy8 uint32
	// 0x7C    RW   GPIO Pin Async Rising Edge Detect 0 (GPIO0-31)
	// 0x80    RW   GPIO Pin Async Rising Edge Detect 1 (GPIO32-53)
	asyncRisingEdgeDetectEnable [2]uint32 // GPAREN0-GPAREN1
	// 0x84    -    Reserved
	dummy9 uint32
	// 0x88    RW   GPIO Pin Async Falling Edge Detect 0 (GPIO0-31)
	// 0x8C    RW   GPIO Pin Async Falling Edge Detect 1 (GPIO32-53)
	asyncFallingEdgeDetectEnable [2]uint32 // GPAFEN0-GPAFEN1
	// 0x90    -    Reserved
	dummy10 uint32
	// 0x94    RW   GPIO Pin Pull-up/down Enable (00=Float, 01=Down, 10=Up)
	pullEnable uint32 // GPPUD
	// 0x98    RW   GPIO Pin Pull-up/down Enable Clock 0 (GPIO0-31)
	// 0x9C    RW   GPIO Pin Pull-up/down Enable Clock 1 (GPIO32-53)
	pullEnableClock [2]uint32 // GPPUDCLK0-GPPUDCLK1
	// 0xA0    -    Reserved
	dummy uint32
	// 0xB0    -    Test (byte)
}

func init() {
	GPIO0 = &cpuPins[0]
	GPIO1 = &cpuPins[1]
	GPIO2 = &cpuPins[2]
	GPIO3 = &cpuPins[3]
	GPIO4 = &cpuPins[4]
	GPIO5 = &cpuPins[5]
	GPIO6 = &cpuPins[6]
	GPIO7 = &cpuPins[7]
	GPIO8 = &cpuPins[8]
	GPIO9 = &cpuPins[9]
	GPIO10 = &cpuPins[10]
	GPIO11 = &cpuPins[11]
	GPIO12 = &cpuPins[12]
	GPIO13 = &cpuPins[13]
	GPIO14 = &cpuPins[14]
	GPIO15 = &cpuPins[15]
	GPIO16 = &cpuPins[16]
	GPIO17 = &cpuPins[17]
	GPIO18 = &cpuPins[18]
	GPIO19 = &cpuPins[19]
	GPIO20 = &cpuPins[20]
	GPIO21 = &cpuPins[21]
	GPIO22 = &cpuPins[22]
	GPIO23 = &cpuPins[23]
	GPIO24 = &cpuPins[24]
	GPIO25 = &cpuPins[25]
	GPIO26 = &cpuPins[26]
	GPIO27 = &cpuPins[27]
	GPIO28 = &cpuPins[28]
	GPIO29 = &cpuPins[29]
	GPIO30 = &cpuPins[30]
	GPIO31 = &cpuPins[31]
	GPIO32 = &cpuPins[32]
	GPIO33 = &cpuPins[33]
	GPIO34 = &cpuPins[34]
	GPIO35 = &cpuPins[35]
	GPIO36 = &cpuPins[36]
	GPIO37 = &cpuPins[37]
	GPIO38 = &cpuPins[38]
	GPIO39 = &cpuPins[39]
	GPIO40 = &cpuPins[40]
	GPIO41 = &cpuPins[41]
	GPIO42 = &cpuPins[42]
	GPIO43 = &cpuPins[43]
	GPIO44 = &cpuPins[44]
	GPIO45 = &cpuPins[45]
	GPIO46 = &cpuPins[46]
}

// Changing pull resistor require a 150 cycles sleep.
//
// Do not inline so the temporary value is not optimized out.
//
//go:noinline
func sleep150cycles() uint32 {
	// Do not call into any kernel function, since this causes a high chance of
	// being preempted.
	// Abuse the fact that gpioMemory is uncached memory.
	// TODO(maruel): No idea if this is too much or enough.
	var out uint32
	for i := 0; i < 150; i++ {
		out += gpioMemory.functionSelect[0]
	}
	return out
}

// driverGPIO implements periph.Driver.
type driverGPIO struct {
}

func (d *driverGPIO) String() string {
	return "bcm283x-gpio"
}

func (d *driverGPIO) Prerequisites() []string {
	return nil
}

func (d *driverGPIO) Init() (bool, error) {
	if !Present() {
		return false, errors.New("bcm283x CPU not detected")
	}
	model := distro.CPUInfo()["model name"]
	if strings.Contains(model, "ARMv6") {
		baseAddr = 0x20000000
		dramBus = 0x40000000
	} else {
		// RPi2+
		baseAddr = 0x3F000000
		dramBus = 0xC0000000
	}
	// Page 6.
	// Virtual addresses in kernel mode will range between 0xC0000000 and
	// 0xEFFFFFFF.
	// Virtual addresses in user mode (i.e. seen by processes running in ARM
	// Linux) will range between 0x00000000 and 0xBFFFFFFF.
	// Peripherals (at physical address 0x20000000 on) are mapped into the kernel
	// virtual address space starting at address 0xF2000000. Thus a peripheral
	// advertised here at bus address 0x7Ennnnnn is available in the ARM kenel at
	// virtual address 0xF2nnnnnn.

	gpioBaseAddr = baseAddr + 0x200000
	m, err := pmem.MapGPIO()
	if err != nil {
		// Try without /dev/gpiomem. This is the case of not running on Raspbian or
		// raspbian before Jessie. This requires running as root.
		var err2 error
		m, err2 = pmem.Map(uint64(gpioBaseAddr), 4096)
		var err error
		if err2 != nil {
			if distro.IsRaspbian() {
				// Raspbian specific error code to help guide the user to troubleshoot
				// the problems.
				if os.IsNotExist(err) && os.IsPermission(err2) {
					return true, fmt.Errorf("/dev/gpiomem wasn't found; please upgrade to Raspbian Jessie or run as root")
				}
			}
			if os.IsPermission(err2) {
				return true, fmt.Errorf("need more access, try as root: %v", err)
			}
			return true, err
		}
	}
	if err := m.AsPOD(&gpioMemory); err != nil {
		return true, err
	}

	functions := map[string]struct{}{}
	for i := range cpuPins {
		if err := gpioreg.Register(&cpuPins[i], true); err != nil {
			return true, err
		}
		// A pin set in alternate function but not described in `mapping` will
		// show up as "<AltX>". We don't want them to be registered as aliases.
		if f := cpuPins[i].Function(); len(f) < 3 || (f[:2] != "In" && f[:3] != "Out" && f[0] != '<') {
			// Registering the same alias twice fails. This can happen if two pins
			// are configured with the same function. For example both pin #12, #18
			// and #40 could be configured to work as PWM0_OUT.
			// TODO(maruel): Dynamically register and unregister the pins as their
			// functionality is changed.
			if _, ok := functions[f]; !ok {
				functions[f] = struct{}{}
				if err := gpioreg.RegisterAlias(f, fmt.Sprintf("GPIO%d", i)); err != nil {
					return true, err
				}
			}
		}
	}
	// The CS pins are never switched to their alternate mode, so manually map
	// them. They are dependent on their parent pins.
	csMap := map[string]int{
		"SPI1_CS0": 18,
		"SPI1_CS1": 17,
		"SPI1_CS2": 16,
		"SPI2_CS0": 43,
		"SPI2_CS1": 44,
		"SPI2_CS2": 45,
	}
	if GPIO11.Function() == "SPI0_CLK" {
		csMap["SPI0_CS1"] = 7
		csMap["SPI0_CS0"] = 8
	} else if GPIO39.Function() == "SPI0_CLK" {
		csMap["SPI0_CS1"] = 35
		csMap["SPI0_CS0"] = 36
	}
	for a, n := range csMap {
		if err := gpioreg.RegisterAlias(a, fmt.Sprintf("GPIO%d", n)); err != nil {
			return true, err
		}
	}
	return true, sysfs.SetSpeedHook(setSpeed)
}

func setSpeed(hz int64) error {
	// Writing to "/sys/module/i2c_bcm2708/parameters/baudrate" was confirmed to
	// not work.
	// modprobe hangs when a bus is opened, so this must be called *before* the
	// bus is opened.
	// TL;DR: we can't do anything here.
	/*
		if err := exec.Command("modprobe", "-r", "i2c_bcm2708").Run(); err != nil {
			return fmt.Errorf("bcm283x: failed to unload driver i2c_bcm2708: %v", err)
		}
		if err := exec.Command("modprobe", "i2c_bcm2708", "baudrate=600000"); err != nil {
			return fmt.Errorf("bcm283x: failed to reload driver i2c_bcm2708: %v", err)
		}
	*/
	return errors.New("bcm283x: to change the I²C bus speed, please refer to https://periph.io/platform/raspberrypi/#i²c")
}

func init() {
	if isArm {
		periph.MustRegister(&driverGPIO{})
	}
}

var _ gpio.PinDefaultPull = &Pin{}
var _ gpio.PinIO = &Pin{}
var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ gpio.PinPWM = &Pin{}
