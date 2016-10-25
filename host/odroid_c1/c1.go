// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package odroid_c1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/periph"
	"github.com/google/periph/conn/analog"
	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/pins"
	"github.com/google/periph/host/distro"
	"github.com/google/periph/host/headers"
	"github.com/google/periph/host/sysfs"
)

var (
	V1_8 pins.Pin = &pins.BasicPin{N: "V1_8"} // 1.8 volt output

	I2CA_SDA, I2CA_SCL, I2CB_SDA, I2CB_SCL    gpio.PinIO // pins 3, 5, 27, 28
	SPI0_MOSI, SPI0_MISO, SPI0_SCLK, SPI0_CS0 gpio.PinIO // pins 19, 21,23, 24
)

// The J2 header is rPi compatible, except for the two analog pins and the 1.8V output.
var (
	J2_1  pins.Pin     = pins.V3_3      // 3.3 volt; max 30mA
	J2_2  pins.Pin     = pins.V5        // 5 volt (after filtering)
	J2_3  gpio.PinIO   = gpio.INVALID   // I2C1_SDA
	J2_4  pins.Pin     = pins.V5        // 5 volt (after filtering)
	J2_5  gpio.PinIO   = gpio.INVALID   // I2C1_SCL
	J2_6  pins.Pin     = pins.GROUND    //
	J2_7  gpio.PinIO   = gpio.INVALID   // GPCLK0
	J2_8  gpio.PinIO   = gpio.INVALID   // UART0_TXD, UART1_TXD
	J2_9  pins.Pin     = pins.GROUND    //
	J2_10 gpio.PinIO   = gpio.INVALID   // UART0_RXD, UART1_RXD
	J2_11 gpio.PinIO   = gpio.INVALID   // UART0_RTS, SPI1_CS1, UART1_RTS
	J2_12 gpio.PinIO   = gpio.INVALID   // PCM_CLK, SPI1_CS0, PWM0_OUT
	J2_13 gpio.PinIO   = gpio.INVALID   // GPIO116
	J2_14 pins.Pin     = pins.GROUND    //
	J2_15 gpio.PinIO   = gpio.INVALID   // GPIO115
	J2_16 gpio.PinIO   = gpio.INVALID   // GPIO104
	J2_17 pins.Pin     = pins.V3_3      //
	J2_18 gpio.PinIO   = gpio.INVALID   // GPIO102
	J2_19 gpio.PinIO   = gpio.INVALID   // SPI0_MOSI
	J2_20 pins.Pin     = pins.GROUND    //
	J2_21 gpio.PinIO   = gpio.INVALID   // SPI0_MISO
	J2_22 gpio.PinIO   = gpio.INVALID   // GPIO103
	J2_23 gpio.PinIO   = gpio.INVALID   // SPI0_SCLK
	J2_24 gpio.PinIO   = gpio.INVALID   // SPI0_CS0
	J2_25 pins.Pin     = pins.GROUND    //
	J2_26 gpio.PinIO   = gpio.INVALID   // SPI0_CE1
	J2_27 gpio.PinIO   = gpio.INVALID   // I2C0_SDA
	J2_28 gpio.PinIO   = gpio.INVALID   // I2C0_SCL
	J2_29 gpio.PinIO   = gpio.INVALID   // GPCLK1
	J2_30 pins.Pin     = pins.GROUND    //
	J2_31 gpio.PinIO   = gpio.INVALID   // GPCLK2
	J2_32 gpio.PinIO   = gpio.INVALID   // PWM0_OUT
	J2_33 gpio.PinIO   = gpio.INVALID   // PWM1_OUT
	J2_34 pins.Pin     = pins.GROUND    //
	J2_35 gpio.PinIO   = gpio.INVALID   // PCM_FS, SPI1_MISO, PWM1_OUT
	J2_36 gpio.PinIO   = gpio.INVALID   // UART0_CTS, SPI1_CE2, UART1_CTS
	J2_37 analog.PinIO = analog.INVALID // TODO: make this functional
	J2_38 pins.Pin     = V1_8           //
	J2_39 pins.Pin     = pins.GROUND    //
	J2_40 analog.PinIO = analog.INVALID // TODO: make this functional
)

// Present returns true if running on a Hardkernel ODROID-C0/C1/C1+ board.
func Present() bool {
	for _, c := range distro.DTCompatible() {
		if strings.Contains(c, "8726_M8B") {
			return true
		}
	}
	return distro.CPUInfo()["Hardware"] == "ODROIDC"
}

// aliases is a list of aliases for the various gpio pins, this allows users to refer to pins
// using the documented and labeled names instead of some GPIOnnn name. The map key is the
// alias and the value is the real pin name.
var aliases = map[string]string{
	"I2CA_SDA":  "GPIO74",
	"I2CA_SCL":  "GPIO75",
	"I2CB_SDA":  "GPIO76",
	"I2CB_SCL":  "GPIO77",
	"SPI0_MOSI": "GPIO107",
	"SPI0_MISO": "GPIO106",
	"SPI0_SCLK": "GPIO105",
	"SPI0_CS0":  "GPIO117",
}

// driver implements drivers.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "odroid_c1"
}

func (d *driver) Type() periph.Type {
	return periph.Pins
}

func (d *driver) Prerequisites() []string {
	return []string{"sysfs-gpio"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("Hardkernel ODROID-C0/C1/C1+ board not detected")
	}

	// sysfsPin is a safe way to get a sysfs pin
	sysfsPin := func(n int) gpio.PinIO {
		if pin, present := sysfs.Pins[n]; present {
			return pin
		} else {
			return gpio.INVALID
		}
	}

	// At this point the sysfs driver has initialized and discovered its pins,
	// we can now hook-up the appropriate named pins to sysfs gpio pins.
	I2CA_SDA = sysfsPin(74)
	I2CA_SCL = sysfsPin(75)
	I2CB_SDA = sysfsPin(76)
	I2CB_SCL = sysfsPin(77)
	SPI0_MOSI = sysfsPin(107)
	SPI0_MISO = sysfsPin(106)
	SPI0_SCLK = sysfsPin(105)
	SPI0_CS0 = sysfsPin(117)

	J2_3 = I2CA_SDA
	J2_5 = I2CA_SCL
	J2_7 = sysfsPin(83)   // usually taken by 1-wire driver
	J2_8 = sysfsPin(113)  // usually not available
	J2_10 = sysfsPin(114) // usually not available
	J2_11 = sysfsPin(88)
	J2_12 = sysfsPin(87)
	J2_13 = sysfsPin(116)
	J2_15 = sysfsPin(115)
	J2_16 = sysfsPin(104)
	J2_18 = sysfsPin(102)
	J2_19 = SPI0_MOSI
	J2_21 = SPI0_MISO
	J2_22 = sysfsPin(103)
	J2_23 = SPI0_SCLK
	J2_24 = SPI0_CS0
	J2_26 = sysfsPin(118)
	J2_27 = I2CB_SDA
	J2_28 = I2CB_SCL
	J2_29 = sysfsPin(101)
	J2_31 = sysfsPin(100)
	J2_32 = sysfsPin(99)
	J2_33 = sysfsPin(108)
	J2_35 = sysfsPin(97)
	J2_36 = sysfsPin(98)

	// J2 is the 40-pin rPi-compatible header
	J2 := [][]pins.Pin{
		{J2_1, J2_2},
		{J2_3, J2_4},
		{J2_5, J2_6},
		{J2_7, J2_8},
		{J2_9, J2_10},
		{J2_11, J2_12},
		{J2_13, J2_14},
		{J2_15, J2_16},
		{J2_17, J2_18},
		{J2_19, J2_20},
		{J2_21, J2_22},
		{J2_23, J2_24},
		{J2_25, J2_26},
		{J2_27, J2_28},
		{J2_29, J2_30},
		{J2_31, J2_32},
		{J2_33, J2_34},
		{J2_35, J2_36},
		{J2_37, J2_38},
		{J2_39, J2_40},
	}
	if err := headers.Register("J2", J2); err != nil {
		return true, err
	}

	// Register explicit pin aliases.
	for alias, real := range aliases {
		r := gpio.ByName(real)
		if r == nil {
			return true, fmt.Errorf("Cannot create alias for %s: it doesn't exist",
				real)
		}
		a := &gpio.PinAlias{N: alias, PinIO: r}
		if err := gpio.RegisterAlias(a); err != nil {
			return true, fmt.Errorf("Cannot create alias %s for %s: %s",
				alias, real, err)
		}
	}

	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&driver{})
	}
}
