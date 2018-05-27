// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Raspberry Pi pin out.

package rpi

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/bcm283x"
	"periph.io/x/periph/host/distro"
)

// Present returns true if running on a Raspberry Pi board.
//
// https://www.raspberrypi.org/
func Present() bool {
	if isArm {
		// This is iffy at best.
		_, err := os.Stat("/sys/bus/platform/drivers/raspberrypi-firmware")
		return err == nil
	}
	return false
}

// Pin as connect on the 40 pins extension header.
//
// Schematics are useful to know what is connected to what:
// https://www.raspberrypi.org/documentation/hardware/raspberrypi/schematics/README.md
//
// The actual pin mapping depends on the board revision! The default values are
// set as the 40 pins header on Raspberry Pi 2 and Raspberry Pi 3.
//
// Some header info here: http://elinux.org/RPi_Low-level_peripherals
//
// P1 is also known as J8 on A+, B+, 2 and later.
var (
	// Raspberry Pi A and B, 26 pin header:
	P1_1  pin.Pin    = pin.V3_3       // max 30mA
	P1_2  pin.Pin    = pin.V5         // (filtered)
	P1_3  gpio.PinIO = bcm283x.GPIO2  // High, I2C1_SDA
	P1_4  pin.Pin    = pin.V5         //
	P1_5  gpio.PinIO = bcm283x.GPIO3  // High, I2C1_SCL
	P1_6  pin.Pin    = pin.GROUND     //
	P1_7  gpio.PinIO = bcm283x.GPIO4  // High, GPCLK0
	P1_8  gpio.PinIO = bcm283x.GPIO14 // Low,  UART0_TXD, UART1_TXD
	P1_9  pin.Pin    = pin.GROUND     //
	P1_10 gpio.PinIO = bcm283x.GPIO15 // Low,  UART0_RXD, UART1_RXD
	P1_11 gpio.PinIO = bcm283x.GPIO17 // Low,  UART0_RTS, SPI1_CE1, UART1_RTS
	P1_12 gpio.PinIO = bcm283x.GPIO18 // Low,  PCM_CLK, SPI1_CE0, PWM0_OUT
	P1_13 gpio.PinIO = bcm283x.GPIO27 // Low,
	P1_14 pin.Pin    = pin.GROUND     //
	P1_15 gpio.PinIO = bcm283x.GPIO22 // Low,
	P1_16 gpio.PinIO = bcm283x.GPIO23 // Low,
	P1_17 pin.Pin    = pin.V3_3       //
	P1_18 gpio.PinIO = bcm283x.GPIO24 // Low,
	P1_19 gpio.PinIO = bcm283x.GPIO10 // Low, SPI0_MOSI
	P1_20 pin.Pin    = pin.GROUND     //
	P1_21 gpio.PinIO = bcm283x.GPIO9  // Low, SPI0_MISO
	P1_22 gpio.PinIO = bcm283x.GPIO25 // Low,
	P1_23 gpio.PinIO = bcm283x.GPIO11 // Low, SPI0_CLK
	P1_24 gpio.PinIO = bcm283x.GPIO8  // High, SPI0_CE0
	P1_25 pin.Pin    = pin.GROUND     //
	P1_26 gpio.PinIO = bcm283x.GPIO7  // High, SPI0_CE1

	// Raspberry Pi A+, B+, 2 and later, 40 pin header (also named J8):
	P1_27 gpio.PinIO = bcm283x.GPIO0  // High, I2C0_SDA used to probe for HAT EEPROM, see https://github.com/raspberrypi/hats
	P1_28 gpio.PinIO = bcm283x.GPIO1  // High, I2C0_SCL
	P1_29 gpio.PinIO = bcm283x.GPIO5  // High, GPCLK1
	P1_30 pin.Pin    = pin.GROUND     //
	P1_31 gpio.PinIO = bcm283x.GPIO6  // High, GPCLK2
	P1_32 gpio.PinIO = bcm283x.GPIO12 // Low,  PWM0_OUT
	P1_33 gpio.PinIO = bcm283x.GPIO13 // Low,  PWM1_OUT
	P1_34 pin.Pin    = pin.GROUND     //
	P1_35 gpio.PinIO = bcm283x.GPIO19 // Low,  PCM_FS, SPI1_MISO, PWM1_OUT
	P1_36 gpio.PinIO = bcm283x.GPIO16 // Low,  UART0_CTS, SPI1_CE2, UART1_CTS
	P1_37 gpio.PinIO = bcm283x.GPIO26 //
	P1_38 gpio.PinIO = bcm283x.GPIO20 // Low,  PCM_DIN, SPI1_MOSI, GPCLK0
	P1_39 pin.Pin    = pin.GROUND     //
	P1_40 gpio.PinIO = bcm283x.GPIO21 // Low,  PCM_DOUT, SPI1_CLK, GPCLK1

	// P5 header on Raspberry Pi A and B, PCB v2:
	P5_1 pin.Pin    = pin.V5
	P5_2 pin.Pin    = pin.V3_3
	P5_3 gpio.PinIO = bcm283x.GPIO28 // Float, I2C0_SDA, PCM_CLK
	P5_4 gpio.PinIO = bcm283x.GPIO29 // Float, I2C0_SCL, PCM_FS
	P5_5 gpio.PinIO = bcm283x.GPIO30 // Low,   PCM_DIN, UART0_CTS, UART1_CTS
	P5_6 gpio.PinIO = bcm283x.GPIO31 // Low,   PCM_DOUT, UART0_RTS, UART1_RTS
	P5_7 pin.Pin    = pin.GROUND
	P5_8 pin.Pin    = pin.GROUND

	AUDIO_RIGHT         = bcm283x.GPIO40 // Low,   PWM0_OUT, SPI2_MISO, UART1_TXD
	AUDIO_LEFT          = bcm283x.GPIO41 // Low,   PWM1_OUT, SPI2_MOSI, UART1_RXD
	HDMI_HOTPLUG_DETECT = bcm283x.GPIO46 // High,
)

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "rpi"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	return []string{"bcm283x-gpio"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("Raspberry Pi board not detected")
	}

	// Setup headers based on board revision.
	//
	// This code is not futureproof, it will error out on a Raspberry Pi 4
	// whenever it comes out.
	// Revision codes from: http://elinux.org/RPi_HardwareHistory
	has26PinP1Header := false
	has40PinP1Header := false
	hasP5Header := false
	hasAudio := false
	hasNewAudio := false
	hasHDMI := false
	rev := distro.CPUInfo()["Revision"]
	if i, err := strconv.ParseInt(rev, 16, 32); err == nil {
		// Ignore the overclock bit.
		i &= 0xFFFFFF
		switch i {
		case 0x0002, 0x0003: // B v1.0
			has26PinP1Header = true
			hasAudio = true
		case 0x0004, 0x0005, 0x0006, // B v2.0
			0x0007, 0x0008, 0x0009, // A v2.0
			0x000d, 0x000e, 0x000f: // B v2.0
			has26PinP1Header = true
			// Only the v2 PCB has the P5 header.
			hasP5Header = true
			hasAudio = true
			hasHDMI = true
		case 0x0010, // B+ v1.0
			0x0012,             // A+ v1.1
			0x0013,             // B+ v1.2
			0x0015,             // A+ v1.1
			0x90021,            // A+ v1.1
			0x90032,            // B+ v1.2
			0xa01040,           // 2 Model B v1.0
			0xa01041, 0xa21041, // 2 Model B v1.1
			0xa22042: // 2 Model B v1.2
			has40PinP1Header = true
			hasAudio = true
			hasHDMI = true
		case 0x900092, // Zero v1.2
			0x900093, // Zero v1.3
			0x920093, // Zero v1.3
			0x9000c1: // Zero W v1.1
			has40PinP1Header = true
			hasHDMI = true
		case 0x0011, // Compute Module 1
			0x0014,   // Compute Module 1
			0xa020a0: // Compute Module 3 v1.0
			// NOTE: Could define the use of a SODIMM header here.
		case 0xa02082, 0xa22082, 0xa32082, 0xa020d3: // 3 Model B v1.2, B+
			has40PinP1Header = true
			hasAudio = true
			hasNewAudio = true
			hasHDMI = true
		default:
			return true, fmt.Errorf("rpi: unknown hardware version: 0x%x", i)
		}
	} else {
		return true, fmt.Errorf("rpi: failed to read cpu_info: %v", err)
	}

	if has26PinP1Header {
		if err := pinreg.Register("P1", [][]pin.Pin{
			{P1_1, P1_2},
			{P1_3, P1_4},
			{P1_5, P1_6},
			{P1_7, P1_8},
			{P1_9, P1_10},
			{P1_11, P1_12},
			{P1_13, P1_14},
			{P1_15, P1_16},
			{P1_17, P1_18},
			{P1_19, P1_20},
			{P1_21, P1_22},
			{P1_23, P1_24},
			{P1_25, P1_26},
		}); err != nil {
			return true, err
		}

		// TODO(maruel): Models from 2012 and earlier have P1_3=GPIO0, P1_5=GPIO1 and P1_13=GPIO21.
		// P2 and P3 are not useful.
		// P6 has a RUN pin for reset but it's not available after Pi version 1.
		P1_27 = gpio.INVALID
		P1_28 = gpio.INVALID
		P1_29 = gpio.INVALID
		P1_30 = pin.INVALID
		P1_31 = gpio.INVALID
		P1_32 = gpio.INVALID
		P1_33 = gpio.INVALID
		P1_34 = pin.INVALID
		P1_35 = gpio.INVALID
		P1_36 = gpio.INVALID
		P1_37 = gpio.INVALID
		P1_38 = gpio.INVALID
		P1_39 = pin.INVALID
		P1_40 = gpio.INVALID
	} else if has40PinP1Header {
		if err := pinreg.Register("P1", [][]pin.Pin{
			{P1_1, P1_2},
			{P1_3, P1_4},
			{P1_5, P1_6},
			{P1_7, P1_8},
			{P1_9, P1_10},
			{P1_11, P1_12},
			{P1_13, P1_14},
			{P1_15, P1_16},
			{P1_17, P1_18},
			{P1_19, P1_20},
			{P1_21, P1_22},
			{P1_23, P1_24},
			{P1_25, P1_26},
			{P1_27, P1_28},
			{P1_29, P1_30},
			{P1_31, P1_32},
			{P1_33, P1_34},
			{P1_35, P1_36},
			{P1_37, P1_38},
			{P1_39, P1_40},
		}); err != nil {
			return true, err
		}
	} else {
		P1_1 = pin.INVALID
		P1_2 = pin.INVALID
		P1_3 = gpio.INVALID
		P1_4 = pin.INVALID
		P1_5 = gpio.INVALID
		P1_6 = pin.INVALID
		P1_7 = gpio.INVALID
		P1_8 = gpio.INVALID
		P1_9 = pin.INVALID
		P1_10 = gpio.INVALID
		P1_11 = gpio.INVALID
		P1_12 = gpio.INVALID
		P1_13 = gpio.INVALID
		P1_14 = pin.INVALID
		P1_15 = gpio.INVALID
		P1_16 = gpio.INVALID
		P1_17 = pin.INVALID
		P1_18 = gpio.INVALID
		P1_19 = gpio.INVALID
		P1_20 = pin.INVALID
		P1_21 = gpio.INVALID
		P1_22 = gpio.INVALID
		P1_23 = gpio.INVALID
		P1_24 = gpio.INVALID
		P1_25 = pin.INVALID
		P1_26 = gpio.INVALID
		P1_27 = gpio.INVALID
		P1_28 = gpio.INVALID
		P1_29 = gpio.INVALID
		P1_30 = pin.INVALID
		P1_31 = gpio.INVALID
		P1_32 = gpio.INVALID
		P1_33 = gpio.INVALID
		P1_34 = pin.INVALID
		P1_35 = gpio.INVALID
		P1_36 = gpio.INVALID
		P1_37 = gpio.INVALID
		P1_38 = gpio.INVALID
		P1_39 = pin.INVALID
		P1_40 = gpio.INVALID
	}

	// Only the A and B v2 PCB has the P5 header.
	if hasP5Header {
		if err := pinreg.Register("P5", [][]pin.Pin{
			{P5_1, P5_2},
			{P5_3, P5_4},
			{P5_5, P5_6},
			{P5_7, P5_8},
		}); err != nil {
			return true, err
		}
	} else {
		P5_1 = pin.INVALID
		P5_2 = pin.INVALID
		P5_3 = gpio.INVALID
		P5_4 = gpio.INVALID
		P5_5 = gpio.INVALID
		P5_6 = gpio.INVALID
		P5_7 = pin.INVALID
		P5_8 = pin.INVALID
	}

	if hasAudio {
		if !hasNewAudio {
			AUDIO_LEFT = bcm283x.GPIO45 // PWM1_OUT
		}
		if err := pinreg.Register("AUDIO", [][]pin.Pin{
			{AUDIO_LEFT},
			{AUDIO_RIGHT},
		}); err != nil {
			return true, err
		}
	}

	if hasHDMI {
		if err := pinreg.Register("HDMI", [][]pin.Pin{{HDMI_HOTPLUG_DETECT}}); err != nil {
			return true, err
		}
	}
	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&drv)
	}
}

var drv driver
