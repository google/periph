// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/host/distro"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/sysfs"
	"periph.io/x/periph/host/videocore"
)

// All the pins supported by the CPU.
var (
	GPIO0  *Pin // I2C0_SDA
	GPIO1  *Pin // I2C0_SCL
	GPIO2  *Pin // I2C1_SDA
	GPIO3  *Pin // I2C1_SCL
	GPIO4  *Pin // CLK0
	GPIO5  *Pin // CLK1
	GPIO6  *Pin // CLK2
	GPIO7  *Pin // SPI0_CS1
	GPIO8  *Pin // SPI0_CS0
	GPIO9  *Pin // SPI0_MISO
	GPIO10 *Pin // SPI0_MOSI
	GPIO11 *Pin // SPI0_CLK
	GPIO12 *Pin // PWM0
	GPIO13 *Pin // PWM1
	GPIO14 *Pin // UART0_TX, UART1_TX
	GPIO15 *Pin // UART0_RX, UART1_RX
	GPIO16 *Pin // UART0_CTS, SPI1_CS2, UART1_CTS
	GPIO17 *Pin // UART0_RTS, SPI1_CS1, UART1_RTS
	GPIO18 *Pin // I2S_SCK, SPI1_CS0, PWM0
	GPIO19 *Pin // I2S_WS, SPI1_MISO, PWM1
	GPIO20 *Pin // I2S_DIN, SPI1_MOSI, CLK0
	GPIO21 *Pin // I2S_DOUT, SPI1_CLK, CLK1
	GPIO22 *Pin //
	GPIO23 *Pin //
	GPIO24 *Pin //
	GPIO25 *Pin //
	GPIO26 *Pin //
	GPIO27 *Pin //
	GPIO28 *Pin // I2C0_SDA, I2S_SCK
	GPIO29 *Pin // I2C0_SCL, I2S_WS
	GPIO30 *Pin // I2S_DIN, UART0_CTS, UART1_CTS
	GPIO31 *Pin // I2S_DOUT, UART0_RTS, UART1_RTS
	GPIO32 *Pin // CLK0, UART0_TX, UART1_TX
	GPIO33 *Pin // UART0_RX, UART1_RX
	GPIO34 *Pin // CLK0
	GPIO35 *Pin // SPI0_CS1
	GPIO36 *Pin // SPI0_CS0, UART0_TX
	GPIO37 *Pin // SPI0_MISO, UART0_RX
	GPIO38 *Pin // SPI0_MOSI, UART0_RTS
	GPIO39 *Pin // SPI0_CLK, UART0_CTS
	GPIO40 *Pin // PWM0, SPI2_MISO, UART1_TX
	GPIO41 *Pin // PWM1, SPI2_MOSI, UART1_RX
	GPIO42 *Pin // CLK1, SPI2_CLK, UART1_RTS
	GPIO43 *Pin // CLK2, SPI2_CS0, UART1_CTS
	GPIO44 *Pin // CLK1, I2C0_SDA, I2C1_SDA, SPI2_CS1
	GPIO45 *Pin // PWM1, I2C0_SCL, I2C1_SCL, SPI2_CS2
	GPIO46 *Pin //
	// Pins 47~53 are not exposed because using them would lead to immediate SD
	// Card corruption.
)

// Present returns true if running on a Broadcom bcm283x based CPU.
func Present() bool {
	if isArm {
		for _, line := range distro.DTCompatible() {
			if strings.HasPrefix(line, "brcm,bcm") {
				return true
			}
		}
	}
	return false
}

// PinsRead0To31 returns the value of all GPIO0 to GPIO31 at their corresponding
// bit as a single read operation.
//
// This function is extremely fast and does no error checking.
//
// The returned bits are valid for both inputs and outputs.
func PinsRead0To31() uint32 {
	return drvGPIO.gpioMemory.level[0]
}

// PinsClear0To31 clears the value of GPIO0 to GPIO31 pin for the bit set at
// their corresponding bit as a single write operation.
//
// This function is extremely fast and does no error checking.
func PinsClear0To31(mask uint32) {
	drvGPIO.gpioMemory.outputClear[0] = mask
}

// PinsSet0To31 sets the value of GPIO0 to GPIO31 pin for the bit set at their
// corresponding bit as a single write operation.
func PinsSet0To31(mask uint32) {
	drvGPIO.gpioMemory.outputSet[0] = mask
}

// PinsRead32To46 returns the value of all GPIO32 to GPIO46 at their
// corresponding 'bit minus 32' as a single read operation.
//
// This function is extremely fast and does no error checking.
//
// The returned bits are valid for both inputs and outputs.
//
// Bits above 15 are guaranteed to be 0.
//
// This function is not recommended on Raspberry Pis as these GPIOs are not
// easily accessible.
func PinsRead32To46() uint32 {
	return drvGPIO.gpioMemory.level[1] & 0x7fff
}

// PinsClear32To46 clears the value of GPIO31 to GPIO46 pin for the bit set at
// their corresponding 'bit minus 32' as a single write operation.
//
// This function is extremely fast and does no error checking.
//
// Bits above 15 are ignored.
//
// This function is not recommended on Raspberry Pis as these GPIOs are not
// easily accessible.
func PinsClear32To46(mask uint32) {
	drvGPIO.gpioMemory.outputClear[1] = (mask & 0x7fff)
}

// PinsSet32To46 sets the value of GPIO31 to GPIO46 pin for the bit set at
// their corresponding 'bit minus 32' as a single write operation.
//
// This function is extremely fast and does no error checking.
//
// Bits above 15 are ignored.
//
// This function is not recommended on Raspberry Pis as these GPIOs are not
// easily accessible.
func PinsSet32To46(mask uint32) {
	drvGPIO.gpioMemory.outputSet[1] = (mask & 0x7fff)
}

// PinsSetup0To27 sets the output current drive strength, output slew limiting
// and input hysteresis for GPIO 0 to 27.
//
// Default drive is 8mA, slew unlimited and hysteresis enabled.
//
// Can only be used if driver bcm283x-dma was loaded.
func PinsSetup0To27(drive physic.ElectricCurrent, slewLimit, hysteresis bool) error {
	if drvDMA.gpioPadMemory == nil {
		return errors.New("bcm283x-dma not initialized; try again as root?")
	}
	drvDMA.gpioPadMemory.pads0.set(toPad(drive, slewLimit, hysteresis))
	return nil
}

// PinsSetup28To45 sets the output current drive strength, output slew limiting
// and input hysteresis for GPIO 28 to 45.
//
// Default drive is 16mA, slew unlimited and hysteresis enabled.
//
// Can only be used if driver bcm283x-dma was loaded.
//
// This function is not recommended on Raspberry Pis as these GPIOs are not
// easily accessible.
func PinsSetup28To45(drive physic.ElectricCurrent, slewLimit, hysteresis bool) error {
	if drvDMA.gpioPadMemory == nil {
		return errors.New("bcm283x-dma not initialized; try again as root?")
	}
	drvDMA.gpioPadMemory.pads1.set(toPad(drive, slewLimit, hysteresis))
	return nil
}

// Pin is a GPIO number (GPIOnn) on BCM238(5|6|7).
//
// Pin implements gpio.PinIO.
type Pin struct {
	// Immutable.
	number      int
	name        string
	defaultPull gpio.Pull // Default pull at system boot, as per datasheet.

	// Immutable after driver initialization.
	sysfsPin *sysfs.Pin // Set to the corresponding sysfs.Pin, if any.

	// Mutable.
	usingEdge  bool           // Set when edge detection is enabled.
	usingClock bool           // Set when a CLK, PWM or I2S/PCM clock is used.
	dmaCh      *dmaChannel    // Set when DMA is used for PWM or I2S/PCM.
	dmaBuf     *videocore.Mem // Set when DMA is used for PWM or I2S/PCM.
}

// String implements conn.Resource.
func (p *Pin) String() string {
	return p.name
}

// Halt implements conn.Resource.
//
// If the pin is running a clock, PWM or waiting for edges, it is halted.
//
// In the case of clock or PWM, all pins with this clock source are also
// disabled.
func (p *Pin) Halt() error {
	if p.usingEdge {
		if err := p.sysfsPin.Halt(); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = false
	}
	return p.haltClock()
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
	case alt3:
		if s := mapping[p.number][3]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT3")
	case alt4:
		if s := mapping[p.number][4]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT4")
	case alt5:
		if s := mapping[p.number][5]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT5")
	default:
		return pin.FuncNone
	}
}

// SupportedFuncs implements pin.PinFunc.
func (p *Pin) SupportedFuncs() []pin.Func {
	f := make([]pin.Func, 0, 2+4)
	f = append(f, gpio.IN, gpio.OUT)
	for _, m := range mapping[p.number] {
		if m != "" {
			f = append(f, m)
		}
	}
	return f
}

// SetFunc implements pin.PinFunc.
func (p *Pin) SetFunc(f pin.Func) error {
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return p.wrap(errors.New("subsystem gpiomem not initialized and sysfs not accessible"))
		}
		return p.sysfsPin.SetFunc(f)
	}
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
		for i, m := range mapping[p.number] {
			if m == f || (isGeneral && m.Generalize() == f) {
				if err := p.Halt(); err != nil {
					return err
				}
				switch i {
				case 0:
					p.setFunction(alt0)
				case 1:
					p.setFunction(alt1)
				case 2:
					p.setFunction(alt2)
				case 3:
					p.setFunction(alt3)
				case 4:
					p.setFunction(alt4)
				case 5:
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
	if p.usingEdge && edge == gpio.NoEdge {
		if err := p.sysfsPin.Halt(); err != nil {
			return p.wrap(err)
		}
	}
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return p.wrap(errors.New("subsystem gpiomem not initialized and sysfs not accessible"))
		}
		if pull != gpio.PullNoChange {
			return p.wrap(errors.New("pull cannot be used when subsystem gpiomem not initialized"))
		}
		if err := p.sysfsPin.In(pull, edge); err != nil {
			return p.wrap(err)
		}
		p.usingEdge = edge != gpio.NoEdge
		return nil
	}
	if err := p.haltClock(); err != nil {
		return err
	}
	p.setFunction(in)
	if pull != gpio.PullNoChange {
		// Changing pull resistor requires a specific dance as described at
		// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
		// page 101.
		// However, BCM2711 uses a simpler way of setting pull resistors, reference at
		// https://github.com/raspberrypi/documentation/blob/master/hardware/raspberrypi/bcm2711/rpi_DATA_2711_1p0.pdf
		// page 84 and 95 ~ 98.

		// If we are running on a newer chip such as BCM2711, set Pull directly.
		if !drvGPIO.useLegacyPull {
			// GPIO_PUP_PDN_CNTRL_REG0 for GPIO0-15
			// GPIO_PUP_PDN_CNTRL_REG1 for GPIO16-31
			// GPIO_PUP_PDN_CNTRL_REG2 for GPIO32-47
			// GPIO_PUP_PDN_CNTRL_REG3 for GPIO48-57
			offset := p.number / 16
			// Check page 94.
			// Resistor Select for GPIOXX
			// 00 = No resistor is selected
			// 01 = Pull up resistor is selected
			// 10 = Pull down resistor is selected
			// 11 = Reserved
			var pullState uint32
			switch pull {
			case gpio.PullDown:
				pullState = 2
			case gpio.PullUp:
				pullState = 1
			case gpio.Float:
				pullState = 0
			}
			drvGPIO.gpioMemory.pullRegister[offset] = pullState << uint((p.number%16)<<1)
		} else {
			// Set Pull
			switch pull {
			case gpio.PullDown:
				drvGPIO.gpioMemory.pullEnable = 1
			case gpio.PullUp:
				drvGPIO.gpioMemory.pullEnable = 2
			case gpio.Float:
				drvGPIO.gpioMemory.pullEnable = 0
			}

			// Datasheet states caller needs to sleep 150 cycles.
			sleep150cycles()
			offset := p.number / 32
			drvGPIO.gpioMemory.pullEnableClock[offset] = 1 << uint(p.number%32)

			sleep150cycles()
			drvGPIO.gpioMemory.pullEnable = 0
			drvGPIO.gpioMemory.pullEnableClock[offset] = 0
		}
	}
	if edge != gpio.NoEdge {
		if p.sysfsPin == nil {
			return p.wrap(fmt.Errorf("pin %d is not exported by sysfs", p.number))
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
// This function is fast. It works even if the pin is set as output.
func (p *Pin) Read() gpio.Level {
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return gpio.Low
		}
		return p.sysfsPin.Read()
	}
	if p.number < 32 {
		// Important: do not remove the &31 here even if not necessary. Testing
		// showed that it slows down the performance by several percents.
		return gpio.Level((drvGPIO.gpioMemory.level[0] & (1 << uint(p.number&31))) != 0)
	}
	return gpio.Level((drvGPIO.gpioMemory.level[1] & (1 << uint(p.number&31))) != 0)
}

// FastRead return the current pin level without any error checking.
//
// This function is very fast. It works even if the pin is set as output.
func (p *Pin) FastRead() gpio.Level {
	if p.number < 32 {
		// Important: do not remove the &31 here even if not necessary. Testing
		// showed that it slows down the performance by several percents.
		return gpio.Level((drvGPIO.gpioMemory.level[0] & (1 << uint(p.number&31))) != 0)
	}
	return gpio.Level((drvGPIO.gpioMemory.level[1] & (1 << uint(p.number&31))) != 0)
}

// WaitForEdge implements gpio.PinIn.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	if p.sysfsPin != nil {
		return p.sysfsPin.WaitForEdge(timeout)
	}
	return false
}

// Pull implements gpio.PinIn.
//
// bcm2711/bcm2838 support querying the pull resistor of all GPIO pins. Prior
// to it, bcm283x doesn't support querying the pull resistor of any GPIO pin.
func (p *Pin) Pull() gpio.Pull {
	// sysfs does not have the capability to read pull resistor.
	if drvGPIO.gpioMemory != nil {
		if drvGPIO.useLegacyPull {
			// TODO(maruel): The best that could be added is to cache the last set value
			// and return it.
			return gpio.PullNoChange
		}
		offset := p.number / 16
		pullState := (drvGPIO.gpioMemory.pullRegister[offset] >> uint((p.number%16)<<1)) % 4
		switch pullState {
		case 0:
			return gpio.Float
		case 1:
			return gpio.PullUp
		case 2:
			return gpio.PullDown
		}
	}
	return gpio.PullNoChange
}

// DefaultPull implements gpio.PinIn.
//
// The CPU doesn't return the current pull.
func (p *Pin) DefaultPull() gpio.Pull {
	return p.defaultPull
}

// Out implements gpio.PinOut.
//
// Fails if requesting to change a pin that is set to special functionality.
func (p *Pin) Out(l gpio.Level) error {
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return p.wrap(errors.New("subsystem gpiomem not initialized and sysfs not accessible"))
		}
		return p.sysfsPin.Out(l)
	}
	// TODO(maruel): This function call is very costly.
	if err := p.Halt(); err != nil {
		return err
	}
	// Change output before changing mode to not create any glitch.
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
	mask := uint32(1) << uint(p.number&31)
	if l == gpio.Low {
		if p.number < 32 {
			drvGPIO.gpioMemory.outputClear[0] = mask
		} else {
			drvGPIO.gpioMemory.outputClear[1] = mask
		}
	} else {
		if p.number < 32 {
			drvGPIO.gpioMemory.outputSet[0] = mask
		} else {
			drvGPIO.gpioMemory.outputSet[1] = mask
		}
	}
}

// BUG(maruel): PWM(): There is no conflict verification when multiple pins are
// used simultaneously. The last call to PWM() will affect all pins of the same
// type (CLK0, CLK2, PWM0 or PWM1).

// PWM implements gpio.PinOut.
//
// It outputs a periodic signal on supported pins without CPU usage.
//
// PWM pins
//
// PWM0 is exposed on pins 12, 18 and 40. However, PWM0 is used for generating
// clock for DMA and unavailable for PWM.
//
// PWM1 is exposed on pins 13, 19, 41 and 45.
//
// PWM1 uses 25Mhz clock source. The frequency must be a divisor of 25Mhz.
//
// DMA driven PWM is available for all pins except PWM1 pins, its resolution is
// 200KHz which is down-sampled from 25MHz clock above. The number of DMA driven
// PWM is limited.
//
// Furthermore, these can only be used if the drive "bcm283x-dma" was loaded.
// It can only be loaded if the process has root level access.
//
// The user must call either Halt(), In(), Out(), PWM(0,..) or
// PWM(gpio.DutyMax,..) to stop the clock source and DMA engine before exiting
// the program.
func (p *Pin) PWM(duty gpio.Duty, freq physic.Frequency) error {
	if duty == 0 {
		return p.Out(gpio.Low)
	} else if duty == gpio.DutyMax {
		return p.Out(gpio.High)
	}
	f := out
	useDMA := false
	switch p.number {
	case 12, 40: // PWM0 alt0: disabled
		useDMA = true
	case 13, 41, 45: // PWM1
		f = alt0
	case 18: // PWM0 alt5: disabled
		useDMA = true
	case 19: // PWM1
		f = alt5
	default:
		useDMA = true
	}

	// Intentionally check later, so a more informative error is returned on
	// unsupported pins.
	if drvGPIO.gpioMemory == nil {
		return p.wrap(errors.New("subsystem gpiomem not initialized"))
	}
	if drvDMA.pwmMemory == nil || drvDMA.clockMemory == nil {
		return p.wrap(errors.New("bcm283x-dma not initialized; try again as root?"))
	}
	if useDMA {
		if m := drvDMA.pwmDMAFreq / 2; m < freq {
			return p.wrap(fmt.Errorf("frequency must be at most %s", m))
		}

		// Total cycles in the period
		rng := uint64(drvDMA.pwmDMAFreq / freq)
		// Pulse width cycles
		dat := uint32((rng*uint64(duty) + uint64(gpio.DutyHalf)) / uint64(gpio.DutyMax))
		var err error
		// TODO(simokawa): Reuse DMA buffer if possible.
		if err = p.haltDMA(); err != nil {
			return p.wrap(err)
		}
		// Start clock before DMA starts.
		if _, err = setPWMClockSource(); err != nil {
			return p.wrap(err)
		}
		if p.dmaCh, p.dmaBuf, err = startPWMbyDMA(p, uint32(rng), dat); err != nil {
			return p.wrap(err)
		}
	} else {
		if m := drvDMA.pwmBaseFreq / 2; m < freq {
			return p.wrap(fmt.Errorf("frequency must be at most %s", m))
		}
		// Total cycles in the period
		rng := uint64(drvDMA.pwmBaseFreq / freq)
		// Pulse width cycles
		dat := uint32((rng*uint64(duty) + uint64(gpio.DutyHalf)) / uint64(gpio.DutyMax))
		if _, err := setPWMClockSource(); err != nil {
			return p.wrap(err)
		}
		// Bit shift for PWM0 and PWM1
		shift := uint((p.number & 1) * 8)
		if shift == 0 {
			drvDMA.pwmMemory.rng1 = uint32(rng)
			Nanospin(10 * time.Nanosecond)
			drvDMA.pwmMemory.dat1 = uint32(dat)
		} else {
			drvDMA.pwmMemory.rng2 = uint32(rng)
			Nanospin(10 * time.Nanosecond)
			drvDMA.pwmMemory.dat2 = uint32(dat)
		}
		Nanospin(10 * time.Nanosecond)
		old := drvDMA.pwmMemory.ctl
		drvDMA.pwmMemory.ctl = (old & ^(0xff << shift)) | ((pwm1Enable | pwm1MS) << shift)
	}
	p.usingClock = true
	p.setFunction(f)
	return nil
}

// StreamIn implements gpiostream.PinIn.
//
// DMA driven StreamOut is available for GPIO0 to GPIO31 pin and the maximum
// resolution is 200kHz.
func (p *Pin) StreamIn(pull gpio.Pull, s gpiostream.Stream) error {
	b, ok := s.(*gpiostream.BitStream)
	if !ok {
		return errors.New("bcm283x: other Stream than BitStream are not implemented yet")
	}
	if !b.LSBF {
		return errors.New("bcm283x: MSBF BitStream is not implemented yet")
	}
	if b.Duration() == 0 {
		return errors.New("bcm283x: can't read to empty BitStream")
	}
	if drvGPIO.gpioMemory == nil {
		return p.wrap(errors.New("subsystem gpiomem not initialized"))
	}
	if err := p.In(pull, gpio.NoEdge); err != nil {
		return err
	}
	if err := dmaReadStream(p, b); err != nil {
		return p.wrap(err)
	}
	return nil
}

// StreamOut implements gpiostream.PinOut.
//
// I2S/PCM driven StreamOut is available for GPIO21 pin. The resolution is up to
// 250MHz.
//
// For GPIO0 to GPIO31 except GPIO21 pin, DMA driven StreamOut is available and
// the maximum resolution is 200kHz.
func (p *Pin) StreamOut(s gpiostream.Stream) error {
	if drvGPIO.gpioMemory == nil {
		return p.wrap(errors.New("subsystem gpiomem not initialized"))
	}
	if err := p.Out(gpio.Low); err != nil {
		return err
	}
	// If the pin is I2S_DOUT, use PCM for much nicer stream and lower memory
	// usage.
	if p.number == 21 || p.number == 31 {
		alt := alt0
		if p.number == 31 {
			alt = alt2
		}
		p.setFunction(alt)
		if err := dmaWriteStreamPCM(p, s); err != nil {
			return p.wrap(err)
		}
	} else if err := dmaWriteStreamEdges(p, s); err != nil {
		return p.wrap(err)
	}
	return nil
}

// Drive returns the configured output current drive strength for this GPIO.
//
// The current drive is configurable per GPIO groups: 0~27 and 28~45.
//
// The default value for GPIOs 0~27 is 8mA and for GPIOs 28~45 is 16mA.
//
// The value is a multiple 2mA between 2mA and 16mA.
//
// Can only be used if driver bcm283x-dma was loaded. Otherwise returns 0.
func (p *Pin) Drive() physic.ElectricCurrent {
	if drvDMA.gpioPadMemory == nil {
		return 0
	}
	var v pad
	if p.number < 28 {
		v = drvDMA.gpioPadMemory.pads0
	} else {
		// GPIO 46~53 are not exposed.
		v = drvDMA.gpioPadMemory.pads1
	}
	switch v & 7 {
	case padDrive2mA:
		return 2 * physic.MilliAmpere
	case padDrive4mA:
		return 4 * physic.MilliAmpere
	case padDrive6mA:
		return 6 * physic.MilliAmpere
	case padDrive8mA:
		return 8 * physic.MilliAmpere
	case padDrive10mA:
		return 10 * physic.MilliAmpere
	case padDrive12mA:
		return 12 * physic.MilliAmpere
	case padDrive14mA:
		return 14 * physic.MilliAmpere
	case padDrive16mA:
		return 16 * physic.MilliAmpere
	default:
		return 0
	}
}

// SlewLimit returns true if the output slew is limited to reduce interference.
//
// The slew is configurable per GPIO groups: 0~27 and 28~45.
//
// The default is true.
//
// Can only be used if driver bcm283x-dma was loaded. Otherwise returns false
// (the default value).
func (p *Pin) SlewLimit() bool {
	if drvDMA.gpioPadMemory == nil {
		return true
	}
	if p.number < 28 {
		return drvDMA.gpioPadMemory.pads0&padSlewUnlimited == 0
	}
	return drvDMA.gpioPadMemory.pads1&padSlewUnlimited == 0
}

// Hysteresis returns true if the input hysteresis via a Schmitt trigger is
// enabled.
//
// The hysteresis is configurable per GPIO groups: 0~27 and 28~45.
//
// The default is true.
//
// Can only be used if driver bcm283x-dma was loaded. Otherwise returns true
// (the default value).
func (p *Pin) Hysteresis() bool {
	if drvDMA.gpioPadMemory == nil {
		return true
	}
	if p.number < 28 {
		return drvDMA.gpioPadMemory.pads0&padHysteresisEnable != 0
	}
	return drvDMA.gpioPadMemory.pads1&padHysteresisEnable != 0
}

// Internal code.

func (p *Pin) haltDMA() error {
	if p.dmaCh != nil {
		p.dmaCh.reset()
		p.dmaCh = nil
	}
	if p.dmaBuf != nil {
		if err := p.dmaBuf.Close(); err != nil {
			return p.wrap(err)
		}
		p.dmaBuf = nil
	}
	return nil
}

// haltClock disables the CLK/PWM clock if used.
func (p *Pin) haltClock() error {
	if err := p.haltDMA(); err != nil {
		return err
	}
	if !p.usingClock {
		return nil
	}
	p.usingClock = false

	// Disable PWMx.
	switch p.number {
	// PWM0 is not used.
	case 12, 18, 40:
	// PWM1
	case 13, 19, 41, 45:
		for _, i := range []int{13, 19, 41, 45} {
			if cpuPins[i].usingClock {
				return nil
			}
		}
		shift := uint((p.number & 1) * 8)
		drvDMA.pwmMemory.ctl &= ^(0xff << shift)
	}

	// Disable PWM clock if nobody use.
	for _, pin := range cpuPins {
		if pin.usingClock {
			return nil
		}
	}
	err := resetPWMClockSource()
	return err
}

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	if drvGPIO.gpioMemory == nil {
		return alt5
	}
	return function((drvGPIO.gpioMemory.functionSelect[p.number/10] >> uint((p.number%10)*3)) & 7)
}

// setFunction changes the GPIO pin function.
func (p *Pin) setFunction(f function) {
	off := p.number / 10
	shift := uint(p.number%10) * 3
	drvGPIO.gpioMemory.functionSelect[off] = (drvGPIO.gpioMemory.functionSelect[off] &^ (7 << shift)) | (uint32(f) << shift)
	// If a pin switches from a specific functionality back to GPIO, the alias
	// should be updated. For example both GPIO13 and GPIO19 support PWM1. By
	// default, PWM1 will be associated to GPIO13, even if
	// GPIO19.SetFunc(gpio.PWM) is called.
	// TODO(maruel): pinreg.Unregister()
	// TODO(maruel): pinreg.Register()
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
var mapping = [][6]pin.Func{
	{"I2C0_SDA"}, // 0
	{"I2C0_SCL"},
	{"I2C1_SDA"},
	{"I2C1_SCL"},
	{"CLK0"},
	{"CLK1"}, // 5
	{"CLK2"},
	{"SPI0_CS1"},
	{"SPI0_CS0"},
	{"SPI0_MISO"},
	{"SPI0_MOSI"}, // 10
	{"SPI0_CLK"},
	{"PWM0"},
	{"PWM1"},
	{"UART0_TX", "", "", "", "", "UART1_TX"},
	{"UART0_RX", "", "", "", "", "UART1_RX"}, // 15
	{"", "", "", "UART0_CTS", "SPI1_CS2", "UART1_CTS"},
	{"", "", "", "UART0_RTS", "SPI1_CS1", "UART1_RTS"},
	{"I2S_SCK", "", "", "", "SPI1_CS0", "PWM0"},
	{"I2S_WS", "", "", "", "SPI1_MISO", "PWM1"},
	{"I2S_DIN", "", "", "", "SPI1_MOSI", "CLK0"}, // 20
	{"I2S_DOUT", "", "", "", "SPI1_CLK", "CLK1"},
	{""},
	{""},
	{""},
	{""}, // 25
	{""},
	{""},
	{"I2C0_SDA", "", "I2S_SCK", "", "", ""},
	{"I2C0_SCL", "", "I2S_WS", "", "", ""},
	{"", "", "I2S_DIN", "UART0_CTS", "", "UART1_CTS"}, // 30
	{"", "", "I2S_DOUT", "UART0_RTS", "", "UART1_RTS"},
	{"CLK0", "", "", "UART0_TX", "", "UART1_TX"},
	{"", "", "", "UART0_RX", "", "UART1_RX"},
	{"CLK0"},
	{"SPI0_CS1"}, // 35
	{"SPI0_CS0", "", "UART0_TX", "", "", ""},
	{"SPI0_MISO", "", "UART0_RX", "", "", ""},
	{"SPI0_MOSI", "", "UART0_RTS", "", "", ""},
	{"SPI0_CLK", "", "UART0_CTS", "", "", ""},
	{"PWM0", "", "", "", "SPI2_MISO", "UART1_TX"}, // 40
	{"PWM1", "", "", "", "SPI2_MOSI", "UART1_RX"},
	{"CLK1", "", "", "", "SPI2_CLK", "UART1_RTS"},
	{"CLK2", "", "", "", "SPI2_CS0", "UART1_CTS"},
	{"CLK1", "I2C0_SDA", "I2C1_SDA", "", "SPI2_CS1", ""},
	{"PWM1", "I2C0_SCL", "I2C1_SCL", "", "SPI2_CS2", ""}, // 45
	{""},
}

// function specifies the active functionality of a pin. The alternative
// function is GPIO pin dependent.
type function uint8

// Mapping as
// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
// pages 90-91.
// And
// https://github.com/raspberrypi/documentation/blob/master/hardware/raspberrypi/bcm2711/rpi_DATA_2711_1p0.pdf
// pages 83-84.
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
	// padding
	dummy11 [3]uint32
	// 0xB0    -    Test (byte)
	test uint32
	// padding
	dummy12 [12]uint32
	// New in BCM2711
	// 0xE4    RW   GPIO Pull-up / Pull-down Register 0 (GPIO0-15)
	// 0xE8    RW   GPIO Pull-up / Pull-down Register 1 (GPIO16-31)
	// 0xEC    RW   GPIO Pull-up / Pull-down Register 2 (GPIO32-47)
	// 0xF0    RW   GPIO Pull-up / Pull-down Register 3 (GPIO48-57)
	pullRegister [4]uint32 // GPIO_PUP_PDN_CNTRL_REG0-GPIO_PUP_PDN_CNTRL_REG3
}

// pad defines the settings for a GPIO pad group.
type pad uint32

const (
	padPasswd           pad = 0x5A << 24 // Write protection
	padSlewUnlimited    pad = 1 << 4     // Output bandwidth limit to reduce bounce.
	padHysteresisEnable pad = 1 << 3     // Schmitt trigger
	padDrive2mA         pad = 0
	padDrive4mA         pad = 1
	padDrive6mA         pad = 2
	padDrive8mA         pad = 3
	padDrive10mA        pad = 4
	padDrive12mA        pad = 5
	padDrive14mA        pad = 6
	padDrive16mA        pad = 7
)

// set changes the current drive strength for the GPIO pad group.
//
// We could disable the schmitt trigger or the slew limit.
func (p *pad) set(settings pad) {
	*p = padPasswd | settings
}

func toPad(drive physic.ElectricCurrent, slewLimit, hysteresis bool) pad {
	var p pad
	d := int(drive / physic.MilliAmpere)
	switch {
	case d <= 2:
		p = padDrive2mA
	case d <= 4:
		p = padDrive4mA
	case d <= 6:
		p = padDrive6mA
	case d <= 8:
		p = padDrive8mA
	case d <= 10:
		p = padDrive10mA
	case d <= 12:
		p = padDrive12mA
	case d <= 14:
		p = padDrive14mA
	default:
		p = padDrive16mA
	}
	if !slewLimit {
		p |= padSlewUnlimited
	}
	if hysteresis {
		p |= padHysteresisEnable
	}
	return p
}

// Mapping as https://scribd.com/doc/101830961/GPIO-Pads-Control2
type gpioPadMap struct {
	dummy [11]uint32 // 0x00~0x28
	pads0 pad        // 0x2c GPIO 0~27
	pads1 pad        // 0x30 GPIO 28~45
	pads2 pad        // 0x34 GPIO 46~53
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
		out += drvGPIO.gpioMemory.functionSelect[0]
	}
	return out
}

// driverGPIO implements periph.Driver.
type driverGPIO struct {
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
	// useLegacyPull is set when the old slow pull resistor setup method before
	// bcm2711 must be used.
	useLegacyPull bool
}

func (d *driverGPIO) Close() {
	d.baseAddr = 0
	d.dramBus = 0
	d.gpioMemory = nil
	d.gpioBaseAddr = 0
}

func (d *driverGPIO) String() string {
	return "bcm283x-gpio"
}

func (d *driverGPIO) Prerequisites() []string {
	return nil
}

func (d *driverGPIO) After() []string {
	return []string{"sysfs-gpio"}
}

func (d *driverGPIO) Init() (bool, error) {
	if !Present() {
		return false, errors.New("bcm283x CPU not detected")
	}
	// It's kind of messy, some report bcm283x while others show bcm27xx.
	// Let's play safe here.
	dTCompatible := strings.Join(distro.DTCompatible(), " ")
	// Reference: https://www.raspberrypi.org/documentation/hardware/raspberrypi/peripheral_addresses.md
	if strings.Contains(dTCompatible, "bcm2708") ||
		strings.Contains(dTCompatible, "bcm2835") {
		// RPi0/1.
		d.baseAddr = 0x20000000
		d.dramBus = 0x40000000
		d.useLegacyPull = true
	} else if strings.Contains(dTCompatible, "bcm2709") ||
		strings.Contains(dTCompatible, "bcm2836") ||
		strings.Contains(dTCompatible, "bcm2710") ||
		strings.Contains(dTCompatible, "bcm2837") {
		// RPi2+
		d.baseAddr = 0x3F000000
		d.dramBus = 0xC0000000
		d.useLegacyPull = true
	} else {
		// RPi4B+
		d.baseAddr = 0xFE000000
		d.dramBus = 0xC0000000
		// BCM2711 (and perhaps future versions?) uses a simpler way to
		// setup internal pull resistors.
		d.useLegacyPull = false
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
	d.gpioBaseAddr = d.baseAddr + 0x200000

	// Mark the right pins as available even if the memory map fails so they can
	// callback to sysfs.Pins.
	functions := map[pin.Func]struct{}{}
	for i := range cpuPins {
		name := cpuPins[i].name
		num := strconv.Itoa(cpuPins[i].number)

		// Initializes the sysfs corresponding pin right away.
		cpuPins[i].sysfsPin = sysfs.Pins[cpuPins[i].number]

		// Unregister the pin if already registered. This happens with sysfs-gpio.
		// Do not error on it, since sysfs-gpio may have failed to load.
		_ = gpioreg.Unregister(name)
		_ = gpioreg.Unregister(num)

		if err := gpioreg.Register(&cpuPins[i]); err != nil {
			return true, err
		}
		if err := gpioreg.RegisterAlias(num, name); err != nil {
			return true, err
		}
		switch f := cpuPins[i].Func(); f {
		case gpio.IN, gpio.OUT, gpio.IN_LOW, gpio.IN_HIGH, gpio.OUT_LOW, gpio.OUT_HIGH, pin.FuncNone:
		default:
			// Registering the same alias twice fails. This can happen if two pins
			// are configured with the same function. For example both pin #12, #18
			// and #40 could be configured to work as PWM0.
			if _, ok := functions[f]; !ok {
				functions[f] = struct{}{}
				if err := gpioreg.RegisterAlias(string(f), name); err != nil {
					return true, err
				}
			}
		}
	}

	// Now do a second loop but do the alternate functions.
	for i := range cpuPins {
		for _, f := range cpuPins[i].SupportedFuncs() {
			switch f {
			case gpio.IN, gpio.OUT:
			default:
				if _, ok := functions[f]; !ok {
					functions[f] = struct{}{}
					if err := gpioreg.RegisterAlias(string(f), cpuPins[i].name); err != nil {
						return true, err
					}
				}
			}
		}
	}

	// Register some BCM-documentation specific names.
	// Do not do UARTx_TXD/RXD nor the PCM_xxx ones.
	aliases := [][2]string{
		{"GPCLK0", "CLK0"},
		{"GPCLK1", "CLK1"},
		{"GPCLK2", "CLK2"},
		{"PWM0_OUT", "PWM0"},
		{"PWM1_OUT", "PWM1"},
	}
	for _, a := range aliases {
		if err := gpioreg.RegisterAlias(a[0], a[1]); err != nil {
			return true, err
		}
	}

	m, err := pmem.MapGPIO()
	if err != nil {
		// Try without /dev/gpiomem. This is the case of not running on Raspbian or
		// raspbian before Jessie. This requires running as root.
		var err2 error
		m, err2 = pmem.Map(uint64(d.gpioBaseAddr), 4096)
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
	if err := m.AsPOD(&d.gpioMemory); err != nil {
		return true, err
	}

	return true, sysfs.I2CSetSpeedHook(setSpeed)
}

func setSpeed(f physic.Frequency) error {
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
		periph.MustRegister(&drvGPIO)
	}
}

var drvGPIO driverGPIO

var _ gpio.PinIO = &Pin{}
var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ gpiostream.PinIn = &Pin{}
var _ gpiostream.PinOut = &Pin{}
var _ pin.PinFunc = &Pin{}
