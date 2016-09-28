// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/google/pio"
	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/host/distro"
	"github.com/google/pio/host/gpiomem"
	"github.com/google/pio/host/sysfs"
)

// Present returns true if running on a Broadcom bcm283x based CPU.
func Present() bool {
	if isArm {
		hardware, ok := distro.CPUInfo()["Hardware"]
		return ok && strings.HasPrefix(hardware, "BCM")
	}
	return false
}

var functional = map[string]gpio.PinIO{
	"GPCLK0":    gpio.INVALID,
	"GPCLK1":    gpio.INVALID,
	"GPCLK2":    gpio.INVALID,
	"I2C0_SCL":  gpio.INVALID,
	"I2C0_SDA":  gpio.INVALID,
	"I2C1_SCL":  gpio.INVALID,
	"I2C1_SDA":  gpio.INVALID,
	"PCM_CLK":   gpio.INVALID,
	"PCM_FS":    gpio.INVALID,
	"PCM_DIN":   gpio.INVALID,
	"PCM_DOUT":  gpio.INVALID,
	"PWM0_OUT":  gpio.INVALID,
	"PWM1_OUT":  gpio.INVALID,
	"SPI0_CS0":  gpio.INVALID,
	"SPI0_CS1":  gpio.INVALID,
	"SPI0_CLK":  gpio.INVALID,
	"SPI0_MISO": gpio.INVALID,
	"SPI0_MOSI": gpio.INVALID,
	"SPI1_CS0":  gpio.INVALID,
	"SPI1_CS1":  gpio.INVALID,
	"SPI1_CS2":  gpio.INVALID,
	"SPI1_CLK":  gpio.INVALID,
	"SPI1_MISO": gpio.INVALID,
	"SPI1_MOSI": gpio.INVALID,
	"SPI2_MISO": gpio.INVALID,
	"SPI2_MOSI": gpio.INVALID,
	"SPI2_CLK":  gpio.INVALID,
	"SPI2_CS0":  gpio.INVALID,
	"SPI2_CS1":  gpio.INVALID,
	"SPI2_CS2":  gpio.INVALID,
	"UART0_RXD": gpio.INVALID,
	"UART0_CTS": gpio.INVALID,
	"UART1_CTS": gpio.INVALID,
	"UART0_RTS": gpio.INVALID,
	"UART1_RTS": gpio.INVALID,
	"UART0_TXD": gpio.INVALID,
	"UART1_RXD": gpio.INVALID,
	"UART1_TXD": gpio.INVALID,
}

// Pin is a GPIO number (GPIOnn) on BCM238(5|6|7).
//
// Pin implements gpio.PinIO.
type Pin struct {
	number      int
	name        string
	defaultPull gpio.Pull
	edge        *sysfs.Pin // Mutable, set once, then never set back to nil
}

// Pins is all the pins as supported by the CPU. There is no guarantee that
// they are actually connected to anything on the board.
var Pins = [54]Pin{
	{number: 0, name: "GPIO0", defaultPull: gpio.Up},
	{number: 1, name: "GPIO1", defaultPull: gpio.Up},
	{number: 2, name: "GPIO2", defaultPull: gpio.Up},
	{number: 3, name: "GPIO3", defaultPull: gpio.Up},
	{number: 4, name: "GPIO4", defaultPull: gpio.Up},
	{number: 5, name: "GPIO5", defaultPull: gpio.Up},
	{number: 6, name: "GPIO6", defaultPull: gpio.Up},
	{number: 7, name: "GPIO7", defaultPull: gpio.Up},
	{number: 8, name: "GPIO8", defaultPull: gpio.Up},
	{number: 9, name: "GPIO9", defaultPull: gpio.Down},
	{number: 10, name: "GPIO10", defaultPull: gpio.Down},
	{number: 11, name: "GPIO11", defaultPull: gpio.Down},
	{number: 12, name: "GPIO12", defaultPull: gpio.Down},
	{number: 13, name: "GPIO13", defaultPull: gpio.Down},
	{number: 14, name: "GPIO14", defaultPull: gpio.Down},
	{number: 15, name: "GPIO15", defaultPull: gpio.Down},
	{number: 16, name: "GPIO16", defaultPull: gpio.Down},
	{number: 17, name: "GPIO17", defaultPull: gpio.Down},
	{number: 18, name: "GPIO18", defaultPull: gpio.Down},
	{number: 19, name: "GPIO19", defaultPull: gpio.Down},
	{number: 20, name: "GPIO20", defaultPull: gpio.Down},
	{number: 21, name: "GPIO21", defaultPull: gpio.Down},
	{number: 22, name: "GPIO22", defaultPull: gpio.Down},
	{number: 23, name: "GPIO23", defaultPull: gpio.Down},
	{number: 24, name: "GPIO24", defaultPull: gpio.Down},
	{number: 25, name: "GPIO25", defaultPull: gpio.Down},
	{number: 26, name: "GPIO26", defaultPull: gpio.Down},
	{number: 27, name: "GPIO27", defaultPull: gpio.Down},
	{number: 28, name: "GPIO28", defaultPull: gpio.Float},
	{number: 29, name: "GPIO29", defaultPull: gpio.Float},
	{number: 30, name: "GPIO30", defaultPull: gpio.Down},
	{number: 31, name: "GPIO31", defaultPull: gpio.Down},
	{number: 32, name: "GPIO32", defaultPull: gpio.Down},
	{number: 33, name: "GPIO33", defaultPull: gpio.Down},
	{number: 34, name: "GPIO34", defaultPull: gpio.Up},
	{number: 35, name: "GPIO35", defaultPull: gpio.Up},
	{number: 36, name: "GPIO36", defaultPull: gpio.Up},
	{number: 37, name: "GPIO37", defaultPull: gpio.Down},
	{number: 38, name: "GPIO38", defaultPull: gpio.Down},
	{number: 39, name: "GPIO39", defaultPull: gpio.Down},
	{number: 40, name: "GPIO40", defaultPull: gpio.Down},
	{number: 41, name: "GPIO41", defaultPull: gpio.Down},
	{number: 42, name: "GPIO42", defaultPull: gpio.Down},
	{number: 43, name: "GPIO43", defaultPull: gpio.Down},
	{number: 44, name: "GPIO44", defaultPull: gpio.Float},
	{number: 45, name: "GPIO45", defaultPull: gpio.Float},
	{number: 46, name: "GPIO46", defaultPull: gpio.Up},
	{number: 47, name: "GPIO47", defaultPull: gpio.Up},
	{number: 48, name: "GPIO48", defaultPull: gpio.Up},
	{number: 49, name: "GPIO49", defaultPull: gpio.Up},
	{number: 50, name: "GPIO50", defaultPull: gpio.Up},
	{number: 51, name: "GPIO51", defaultPull: gpio.Up},
	{number: 52, name: "GPIO52", defaultPull: gpio.Up},
	{number: 53, name: "GPIO53", defaultPull: gpio.Up},
}

// All the pins supported by the CPU.
var (
	GPIO0  *Pin = &Pins[0]  // I2C0_SDA
	GPIO1  *Pin = &Pins[1]  // I2C0_SCL
	GPIO2  *Pin = &Pins[2]  // I2C1_SDA
	GPIO3  *Pin = &Pins[3]  // I2C1_SCL
	GPIO4  *Pin = &Pins[4]  // GPCLK0
	GPIO5  *Pin = &Pins[5]  // GPCLK1
	GPIO6  *Pin = &Pins[6]  // GPCLK2
	GPIO7  *Pin = &Pins[7]  // SPI0_CS1
	GPIO8  *Pin = &Pins[8]  // SPI0_CS0
	GPIO9  *Pin = &Pins[9]  // SPI0_MISO
	GPIO10 *Pin = &Pins[10] // SPI0_MOSI
	GPIO11 *Pin = &Pins[11] // SPI0_CLK
	GPIO12 *Pin = &Pins[12] // PWM0_OUT
	GPIO13 *Pin = &Pins[13] // PWM1_OUT
	GPIO14 *Pin = &Pins[14] // UART0_TXD, UART1_TXD
	GPIO15 *Pin = &Pins[15] // UART0_RXD, UART1_RXD
	GPIO16 *Pin = &Pins[16] // UART0_CTS, SPI1_CS2, UART1_CTS
	GPIO17 *Pin = &Pins[17] // UART0_RTS, SPI1_CS1, UART1_RTS
	GPIO18 *Pin = &Pins[18] // PCM_CLK, SPI1_CS0, PWM0_OUT
	GPIO19 *Pin = &Pins[19] // PCM_FS, SPI1_MISO, PWM1_OUT
	GPIO20 *Pin = &Pins[20] // PCM_DIN, SPI1_MOSI, GPCLK0
	GPIO21 *Pin = &Pins[21] // PCM_DOUT, SPI1_CLK, GPCLK1
	GPIO22 *Pin = &Pins[22] //
	GPIO23 *Pin = &Pins[23] //
	GPIO24 *Pin = &Pins[24] //
	GPIO25 *Pin = &Pins[25] //
	GPIO26 *Pin = &Pins[26] //
	GPIO27 *Pin = &Pins[27] //
	GPIO28 *Pin = &Pins[28] // I2C0_SDA, PCM_CLK
	GPIO29 *Pin = &Pins[29] // I2C0_SCL, PCM_FS
	GPIO30 *Pin = &Pins[30] // PCM_DIN, UART0_CTS, UART1_CTS
	GPIO31 *Pin = &Pins[31] // PCM_DOUT, UART0_RTS, UART1_RTS
	GPIO32 *Pin = &Pins[32] // GPCLK0, UART0_TXD, UART1_TXD
	GPIO33 *Pin = &Pins[33] // UART0_RXD, UART1_RXD
	GPIO34 *Pin = &Pins[34] // GPCLK0
	GPIO35 *Pin = &Pins[35] // SPI0_CS1
	GPIO36 *Pin = &Pins[36] // SPI0_CS0, UART0_TXD
	GPIO37 *Pin = &Pins[37] // SPI0_MISO, UART0_RXD
	GPIO38 *Pin = &Pins[38] // SPI0_MOSI, UART0_RTS
	GPIO39 *Pin = &Pins[39] // SPI0_CLK, UART0_CTS
	GPIO40 *Pin = &Pins[40] // PWM0_OUT, SPI2_MISO, UART1_TXD
	GPIO41 *Pin = &Pins[41] // PWM1_OUT, SPI2_MOSI, UART1_RXD
	GPIO42 *Pin = &Pins[42] // GPCLK1, SPI2_CLK, UART1_RTS
	GPIO43 *Pin = &Pins[43] // GPCLK2, SPI2_CS0, UART1_CTS
	GPIO44 *Pin = &Pins[44] // GPCLK1, I2C0_SDA, I2C1_SDA, SPI2_CS1
	GPIO45 *Pin = &Pins[45] // PWM1_OUT, I2C0_SCL, I2C1_SCL, SPI2_CS2
	GPIO46 *Pin = &Pins[46] //
	GPIO47 *Pin = &Pins[47] // SDCard
	GPIO48 *Pin = &Pins[48] // SDCard
	GPIO49 *Pin = &Pins[49] // SDCard
	GPIO50 *Pin = &Pins[50] // SDCard
	GPIO51 *Pin = &Pins[51] // SDCard
	GPIO52 *Pin = &Pins[52] // SDCard
	GPIO53 *Pin = &Pins[53] // SDCard
)

// Special functions that can be assigned to a GPIO. The values are probed and
// set at runtime. Changing the value of the variables has no effect.
var (
	GPCLK0    gpio.PinIO = gpio.INVALID // GPIO4, GPIO20, GPIO32, GPIO34 (also named GPIO_GCLK)
	GPCLK1    gpio.PinIO = gpio.INVALID // GPIO5, GPIO21, GPIO42, GPIO44
	GPCLK2    gpio.PinIO = gpio.INVALID // GPIO6, GPIO43
	I2C0_SCL  gpio.PinIO = gpio.INVALID // GPIO1, GPIO29, GPIO45
	I2C0_SDA  gpio.PinIO = gpio.INVALID // GPIO0, GPIO28, GPIO44
	I2C1_SCL  gpio.PinIO = gpio.INVALID // GPIO3, GPIO45
	I2C1_SDA  gpio.PinIO = gpio.INVALID // GPIO2, GPIO44
	PCM_CLK   gpio.PinIO = gpio.INVALID // GPIO18, GPIO28 (I2S)
	PCM_FS    gpio.PinIO = gpio.INVALID // GPIO19, GPIO29
	PCM_DIN   gpio.PinIO = gpio.INVALID // GPIO20, GPIO30
	PCM_DOUT  gpio.PinIO = gpio.INVALID // GPIO21, GPIO31
	PWM0_OUT  gpio.PinIO = gpio.INVALID // GPIO12, GPIO18, GPIO40
	PWM1_OUT  gpio.PinIO = gpio.INVALID // GPIO13, GPIO19, GPIO41, GPIO45
	SPI0_CS0  gpio.PinIO = gpio.INVALID // GPIO8,  GPIO36
	SPI0_CS1  gpio.PinIO = gpio.INVALID // GPIO7,  GPIO35
	SPI0_CLK  gpio.PinIO = gpio.INVALID // GPIO11, GPIO39
	SPI0_MISO gpio.PinIO = gpio.INVALID // GPIO9,  GPIO37
	SPI0_MOSI gpio.PinIO = gpio.INVALID // GPIO10, GPIO38
	SPI1_CS0  gpio.PinIO = gpio.INVALID // GPIO18
	SPI1_CS1  gpio.PinIO = gpio.INVALID // GPIO17
	SPI1_CS2  gpio.PinIO = gpio.INVALID // GPIO16
	SPI1_CLK  gpio.PinIO = gpio.INVALID // GPIO21
	SPI1_MISO gpio.PinIO = gpio.INVALID // GPIO19
	SPI1_MOSI gpio.PinIO = gpio.INVALID // GPIO20
	SPI2_MISO gpio.PinIO = gpio.INVALID // GPIO40
	SPI2_MOSI gpio.PinIO = gpio.INVALID // GPIO41
	SPI2_CLK  gpio.PinIO = gpio.INVALID // GPIO42
	SPI2_CS0  gpio.PinIO = gpio.INVALID // GPIO43
	SPI2_CS1  gpio.PinIO = gpio.INVALID // GPIO44
	SPI2_CS2  gpio.PinIO = gpio.INVALID // GPIO45
	UART0_CTS gpio.PinIO = gpio.INVALID // GPIO16, GPIO30, GPIO39
	UART0_RTS gpio.PinIO = gpio.INVALID // GPIO17, GPIO31, GPIO38
	UART0_RXD gpio.PinIO = gpio.INVALID // GPIO15, GPIO33, GPIO37
	UART0_TXD gpio.PinIO = gpio.INVALID // GPIO14, GPIO32, GPIO36
	UART1_CTS gpio.PinIO = gpio.INVALID // GPIO16, GPIO30
	UART1_RTS gpio.PinIO = gpio.INVALID // GPIO17, GPIO31
	UART1_RXD gpio.PinIO = gpio.INVALID // GPIO15, GPIO33, GPIO41
	UART1_TXD gpio.PinIO = gpio.INVALID // GPIO14, GPIO32, GPIO40
)

// PinIO implementation.

func (p *Pin) String() string {
	return p.name
}

// Number implements pins.Pin.
func (p *Pin) Number() int {
	return p.number
}

// Function implements pins.Pin.
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
		return errors.New("subsystem not initialized")
	}
	if !p.setFunction(in) {
		return errors.New("failed to change pin mode")
	}
	if pull != gpio.PullNoChange {
		// Changing pull resistor requires a specific dance as described at
		// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
		// page 101.

		// Set Pull
		switch pull {
		case gpio.Down:
			gpioMemory.pullEnable = 1
		case gpio.Up:
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
	if edge != gpio.None {
		// This is a race condition but this is fine; at worst PinByNumber() is
		// called twice but it is guaranteed to return the same value. p.edge is
		// never set to nil.
		if p.edge == nil {
			var err error
			if p.edge, err = sysfs.PinByNumber(p.Number()); err != nil {
				return err
			}
		}
		if err := p.edge.In(gpio.PullNoChange, edge); err != nil {
			return err
		}
	} else if p.edge != nil {
		if err := p.edge.In(gpio.PullNoChange, edge); err != nil {
			return err
		}
	}
	return nil
}

// Read return the current pin level and implements gpio.PinIn.
//
// This function is very fast. It works even if the pin is set as output.
func (p *Pin) Read() gpio.Level {
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
		return errors.New("subsystem not initialized")
	}
	// Change output before changing mode to not create any glitch.
	offset := p.number / 32
	if l == gpio.Low {
		gpioMemory.outputClear[offset] = 1 << uint(p.number&31)
	} else {
		gpioMemory.outputSet[offset] = 1 << uint(p.number&31)
	}
	if !p.setFunction(out) {
		return errors.New("failed to change pin mode")
	}
	return nil
}

// PWM implements gpio.PinOut.
func (p *Pin) PWM(duty int) error {
	return errors.New("pwm is not supported")
}

// Special functionality.

// DefaultPull returns the default pull for the function.
//
// The CPU doesn't return the current pull.
func (p *Pin) DefaultPull() gpio.Pull {
	return p.defaultPull
}

// Internal code.

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	if gpioMemory == nil {
		return alt5
	}
	return function((gpioMemory.functionSelect[p.number/10] >> uint((p.number%10)*3)) & 7)
}

// setFunction changes the GPIO pin function.
//
// Returns false if the pin was in AltN. Only accepts in and out
func (p *Pin) setFunction(f function) bool {
	if f != in && f != out {
		return false
	}
	if actual := p.function(); actual != in && actual != out {
		return false
	}
	off := p.number / 10
	shift := uint(p.number%10) * 3
	gpioMemory.functionSelect[off] = (gpioMemory.functionSelect[off] &^ (7 << shift)) | (uint32(f) << shift)
	return true
}

//

// function specifies the active functionality of a pin. The alternative
// function is GPIO pin dependent.
type function uint8

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
	functionSelect [6]uint32
	// 0x18    -    Reserved
	dummy0 uint32
	// 0x1C    W    GPIO Pin Output Set 0 (GPIO0-31)
	// 0x20    W    GPIO Pin Output Set 1 (GPIO32-53)
	outputSet [2]uint32
	// 0x24    -    Reserved
	dummy1 uint32
	// 0x28    W    GPIO Pin Output Clear 0 (GPIO0-31)
	// 0x2C    W    GPIO Pin Output Clear 1 (GPIO32-53)
	outputClear [2]uint32
	// 0x30    -    Reserved
	dummy2 uint32
	// 0x34    R    GPIO Pin Level 0 (GPIO0-31)
	// 0x38    R    GPIO Pin Level 1 (GPIO32-53)
	level [2]uint32
	// 0x3C    -    Reserved
	dummy3 uint32
	// 0x40    RW   GPIO Pin Event Detect Status 0 (GPIO0-31)
	// 0x44    RW   GPIO Pin Event Detect Status 1 (GPIO32-53)
	eventDetectStatus [2]uint32
	// 0x48    -    Reserved
	dummy4 uint32
	// 0x4C    RW   GPIO Pin Rising Edge Detect Enable 0 (GPIO0-31)
	// 0x50    RW   GPIO Pin Rising Edge Detect Enable 1 (GPIO32-53)
	risingEdgeDetectEnable [2]uint32
	// 0x54    -    Reserved
	dummy5 uint32
	// 0x58    RW   GPIO Pin Falling Edge Detect Enable 0 (GPIO0-31)
	// 0x5C    RW   GPIO Pin Falling Edge Detect Enable 1 (GPIO32-53)
	fallingEdgeDetectEnable [2]uint32
	// 0x60    -    Reserved
	dummy6 uint32
	// 0x64    RW   GPIO Pin High Detect Enable 0 (GPIO0-31)
	// 0x68    RW   GPIO Pin High Detect Enable 1 (GPIO32-53)
	highDetectEnable [2]uint32
	// 0x6C    -    Reserved
	dummy7 uint32
	// 0x70    RW   GPIO Pin Low Detect Enable 0 (GPIO0-31)
	// 0x74    RW   GPIO Pin Low Detect Enable 1 (GPIO32-53)
	lowDetectEnable [2]uint32
	// 0x78    -    Reserved
	dummy8 uint32
	// 0x7C    RW   GPIO Pin Async Rising Edge Detect 0 (GPIO0-31)
	// 0x80    RW   GPIO Pin Async Rising Edge Detect 1 (GPIO32-53)
	asyncRisingEdgeDetectEnable [2]uint32
	// 0x84    -    Reserved
	dummy9 uint32
	// 0x88    RW   GPIO Pin Async Falling Edge Detect 0 (GPIO0-31)
	// 0x8C    RW   GPIO Pin Async Falling Edge Detect 1 (GPIO32-53)
	asyncFallingEdgeDetectEnable [2]uint32
	// 0x90    -    Reserved
	dummy10 uint32
	// 0x94    RW   GPIO Pin Pull-up/down Enable (00=Float, 01=Down, 10=Up)
	pullEnable uint32
	// 0x98    RW   GPIO Pin Pull-up/down Enable Clock 0 (GPIO0-31)
	// 0x9C    RW   GPIO Pin Pull-up/down Enable Clock 1 (GPIO32-53)
	pullEnableClock [2]uint32
	// 0xA0    -    Reserved
	dummy uint32
	// 0xB0    -    Test (byte)
}

var gpioMemory *gpioMap

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

func setIfAlt0(p *Pin, special *gpio.PinIO) {
	if p.function() == alt0 {
		if (*special).String() != "INVALID" {
			//fmt.Printf("%s and %s have same functionality\n", p, *special)
		}
		*special = p
	}
}

func setIfAlt(p *Pin, special0 *gpio.PinIO, special1 *gpio.PinIO, special2 *gpio.PinIO, special3 *gpio.PinIO, special4 *gpio.PinIO, special5 *gpio.PinIO) {
	switch p.function() {
	case alt0:
		if special0 != nil {
			if (*special0).String() != "INVALID" {
				//fmt.Printf("%s and %s have same functionality\n", p, *special0)
			}
			*special0 = p
		}
	case alt1:
		if special1 != nil {
			if (*special1).String() != "INVALID" {
				//fmt.Printf("%s and %s have same functionality\n", p, *special1)
			}
			*special1 = p
		}
	case alt2:
		if special2 != nil {
			if (*special2).String() != "INVALID" {
				//log.Printf("%s and %s have same functionality\n", p, *special2)
			}
			*special2 = p
		}
	case alt3:
		if special3 != nil {
			if (*special3).String() != "INVALID" {
				//log.Printf("%s and %s have same functionality\n", p, *special3)
			}
			*special3 = p
		}
	case alt4:
		if special4 != nil {
			if (*special4).String() != "INVALID" {
				//log.Printf("%s and %s have same functionality\n", p, *special4)
			}
			*special4 = p
		}
	case alt5:
		if special5 != nil {
			if (*special5).String() != "INVALID" {
				//log.Printf("%s and %s have same functionality\n", p, *special5)
			}
			*special5 = p
		}
	}
}

// This excludes the functions in and out.
var mapping = [54][6]string{
	{"I2C0_SDA"}, // 0
	{"I2C0_SCL"},
	{"I2C1_SDA"},
	{"I2C1_SCL"},
	{"GPCLK0"},
	{"GPCLK1"},
	{"GPCLK2"},
	{"SPI0_CS1"},
	{"SPI0_CS0"},
	{"SPI0_MISO"},
	{"SPI0_MOSI"}, // 10
	{"SPI0_CLK"},
	{"PWM0_OUT"},
	{"PWM1_OUT"},
	{"UART0_TXD", "", "", "", "", "UART1_TXD"},
	{"UART0_RXD", "", "", "", "", "UART1_RXD"},
	{"", "", "", "UART0_CTS", "SPI1_CS2", "UART1_CTS"},
	{"", "", "", "UART0_RTS", "SPI1_CS1", "UART1_RTS"},
	{"PCM_CLK", "", "", "", "SPI1_CS0", "PWM0_OUT"},
	{"PCM_FS", "", "", "", "SPI1_MISO", "PWM1_OUT"},
	{"PCM_DIN", "", "", "", "SPI1_MOSI", "GPCLK0"}, // 20
	{"PCM_DOUT", "", "", "", "SPI1_CLK", "GPCLK1"},
	{},
	{},
	{},
	{},
	{},
	{},
	{"I2C0_SDA", "", "PCM_CLK", "", "", ""},
	{"I2C0_SCL", "", "PCM_FS", "", "", ""},
	{"", "", "PCM_DIN", "UART0_CTS", "", "UART1_CTS"}, // 30
	{"", "", "PCM_DOUT", "UART0_RTS", "", "UART1_RTS"},
	{"GPCLK0", "", "", "UART0_TXD", "", "UART1_TXD"},
	{"", "", "", "UART0_RXD", "", "UART1_RXD"},
	{"GPCLK0"},
	{"SPI0_CS1"},
	{"SPI0_CS0", "", "UART0_TXD", "", "", ""},
	{"SPI0_MISO", "", "UART0_RXD", "", "", ""},
	{"SPI0_MOSI", "", "UART0_RTS", "", "", ""},
	{"SPI0_CLK", "", "UART0_CTS", "", "", ""},
	{"PWM0_OUT", "", "", "", "SPI2_MISO", "UART1_TXD"}, // 40
	{"PWM1_OUT", "", "", "", "SPI2_MOSI", "UART1_RXD"},
	{"GPCLK1", "", "", "", "SPI2_CLK", "UART1_RTS"},
	{"GPCLK2", "", "", "", "SPI2_CS0", "UART1_CTS"},
	{"GPCLK1", "I2C0_SDA", "I2C1_SDA", "", "SPI2_CS1", ""},
	{"PWM1_OUT", "I2C0_SCL", "I2C1_SCL", "", "SPI2_CS2", ""},
}

// getBaseAddress queries the virtual file system to retrieve the base address
// of the GPIO registers.
//
// Defaults to 0x3F200000 as per datasheet if could query the file system.
func getBaseAddress() uint64 {
	items, _ := ioutil.ReadDir("/sys/bus/platform/drivers/pinctrl-bcm2835/")
	for _, item := range items {
		if item.Mode()&os.ModeSymlink != 0 {
			parts := strings.SplitN(path.Base(item.Name()), ".", 2)
			if len(parts) != 2 {
				continue
			}
			base, err := strconv.ParseUint(parts[0], 16, 64)
			if err != nil {
				continue
			}
			return base
		}
	}
	return 0x3F200000
}

// driver implements pio.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "bcm283x"
}

func (d *driver) Type() pio.Type {
	return pio.Processor
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("bcm283x CPU not detected")
	}
	mem, err := gpiomem.OpenGPIO()
	if err != nil {
		// Try without /dev/gpiomem. This is the case of not running on Raspbian or
		// raspbian before Jessie. This requires running as root.
		var err2 error
		mem, err2 = gpiomem.OpenMem(getBaseAddress())
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
	mem.Struct(unsafe.Pointer(&gpioMemory))

	// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
	// Page 102.
	setIfAlt0(GPIO0, &I2C0_SDA)
	setIfAlt0(GPIO1, &I2C0_SCL)
	setIfAlt0(GPIO2, &I2C1_SDA)
	setIfAlt0(GPIO3, &I2C1_SCL)
	setIfAlt0(GPIO4, &GPCLK0)
	setIfAlt0(GPIO5, &GPCLK1)
	setIfAlt0(GPIO6, &GPCLK2)
	setIfAlt0(GPIO7, &SPI0_CS1)
	setIfAlt0(GPIO8, &SPI0_CS0)
	setIfAlt0(GPIO9, &SPI0_MISO)
	setIfAlt0(GPIO10, &SPI0_MOSI)
	setIfAlt0(GPIO11, &SPI0_CLK)
	setIfAlt0(GPIO12, &PWM0_OUT)
	setIfAlt0(GPIO13, &PWM1_OUT)
	setIfAlt(GPIO14, &UART0_TXD, nil, nil, nil, nil, &UART1_TXD)
	setIfAlt(GPIO15, &UART0_RXD, nil, nil, nil, nil, &UART1_RXD)
	setIfAlt(GPIO16, nil, nil, nil, &UART0_CTS, &SPI1_CS2, &UART1_CTS)
	setIfAlt(GPIO17, nil, nil, nil, &UART0_RTS, &SPI1_CS1, &UART1_RTS)
	setIfAlt(GPIO18, &PCM_CLK, nil, nil, nil, &SPI1_CS0, &PWM0_OUT)
	setIfAlt(GPIO19, &PCM_FS, nil, nil, nil, &SPI1_MISO, &PWM1_OUT)
	setIfAlt(GPIO20, &PCM_DIN, nil, nil, nil, &SPI1_MOSI, &GPCLK0)
	setIfAlt(GPIO21, &PCM_DOUT, nil, nil, nil, &SPI1_CLK, &GPCLK1)
	// GPIO22-GPIO27 do not have interesting alternate function.
	setIfAlt(GPIO28, &I2C0_SDA, nil, &PCM_CLK, nil, nil, nil)
	setIfAlt(GPIO29, &I2C0_SCL, nil, &PCM_FS, nil, nil, nil)
	setIfAlt(GPIO30, nil, nil, &PCM_DIN, &UART0_CTS, nil, &UART1_CTS)
	setIfAlt(GPIO31, nil, nil, &PCM_DOUT, &UART0_RTS, nil, &UART1_RTS)
	setIfAlt(GPIO32, &GPCLK0, nil, nil, &UART0_TXD, nil, &UART1_TXD)
	setIfAlt(GPIO33, nil, nil, nil, &UART0_RXD, nil, &UART1_RXD)
	setIfAlt0(GPIO34, &GPCLK0)
	setIfAlt0(GPIO35, &SPI0_CS1)
	setIfAlt(GPIO36, &SPI0_CS0, nil, &UART0_TXD, nil, nil, nil)
	setIfAlt(GPIO37, &SPI0_MISO, nil, &UART0_RXD, nil, nil, nil)
	setIfAlt(GPIO38, &SPI0_MOSI, nil, &UART0_RTS, nil, nil, nil)
	setIfAlt(GPIO39, &SPI0_CLK, nil, &UART0_CTS, nil, nil, nil)
	setIfAlt(GPIO40, &PWM0_OUT, nil, nil, nil, &SPI2_MISO, &UART1_TXD)
	setIfAlt(GPIO41, &PWM1_OUT, nil, nil, nil, &SPI2_MOSI, &UART1_RXD)
	setIfAlt(GPIO42, &GPCLK1, nil, nil, nil, &SPI2_CLK, &UART1_RTS)
	setIfAlt(GPIO43, &GPCLK2, nil, nil, nil, &SPI2_CS0, &UART1_CTS)
	setIfAlt(GPIO44, &GPCLK1, &I2C0_SDA, &I2C1_SDA, nil, &SPI2_CS1, nil)
	setIfAlt(GPIO45, &PWM1_OUT, &I2C0_SCL, &I2C1_SCL, nil, &SPI2_CS2, nil)
	// GPIO46 doesn't have interesting alternate function.
	// GPIO47-GPIO53 are connected to the SDCard.

	// TODO(maruel): Remove all the functional variables?
	for i := range Pins {
		if err := gpio.Register(&Pins[i]); err != nil {
			return true, err
		}
		if i < 46 {
			if f := Pins[i].Function(); len(f) < 3 || (f[:2] != "In" && f[:3] != "Out") {
				functional[f] = &Pins[i]
			}
		}
	}
	for k, v := range functional {
		gpio.MapFunction(k, v)
	}
	return true, nil
}

func init() {
	if isArm {
		pio.MustRegister(&driver{})
	}
}

var _ pio.Driver = &driver{}
var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ gpio.PinIO = &Pin{}
