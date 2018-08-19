// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package odroidc1

import (
	"errors"
	"strconv"
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/distro"
	"periph.io/x/periph/host/sysfs"
)

// The J2 header is rPi compatible, except for the two analog pins and the 1.8V
// output.
var (
	J2_1             = pin.V3_3     // 3.3V; max 30mA
	J2_2             = pin.V5       // 5V (after filtering)
	J2_3  gpio.PinIO = gpio.INVALID // I2C1_SDA
	J2_4             = pin.V5       // 5V (after filtering)
	J2_5  gpio.PinIO = gpio.INVALID // I2C1_SCL
	J2_6             = pin.GROUND   //
	J2_7  gpio.PinIO = gpio.INVALID // CLK0
	J2_8  gpio.PinIO = gpio.INVALID // UART0_TX, UART1_TX
	J2_9             = pin.GROUND   //
	J2_10 gpio.PinIO = gpio.INVALID // UART0_RX, UART1_RX
	J2_11 gpio.PinIO = gpio.INVALID // UART0_RTS, SPI1_CS1, UART1_RTS
	J2_12 gpio.PinIO = gpio.INVALID // I2S_SCK, SPI1_CS0, PWM0
	J2_13 gpio.PinIO = gpio.INVALID // GPIO116
	J2_14            = pin.GROUND   //
	J2_15 gpio.PinIO = gpio.INVALID // GPIO115
	J2_16 gpio.PinIO = gpio.INVALID // GPIO104
	J2_17            = pin.V3_3     //
	J2_18 gpio.PinIO = gpio.INVALID // GPIO102
	J2_19 gpio.PinIO = gpio.INVALID // SPI0_MOSI
	J2_20            = pin.GROUND   //
	J2_21 gpio.PinIO = gpio.INVALID // SPI0_MISO
	J2_22 gpio.PinIO = gpio.INVALID // GPIO103
	J2_23 gpio.PinIO = gpio.INVALID // SPI0_CLK
	J2_24 gpio.PinIO = gpio.INVALID // SPI0_CS0
	J2_25            = pin.GROUND   //
	J2_26 gpio.PinIO = gpio.INVALID // SPI0_CS1
	J2_27 gpio.PinIO = gpio.INVALID // I2C0_SDA
	J2_28 gpio.PinIO = gpio.INVALID // I2C0_SCL
	J2_29 gpio.PinIO = gpio.INVALID // CLK1
	J2_30            = pin.GROUND   //
	J2_31 gpio.PinIO = gpio.INVALID // CLK2
	J2_32 gpio.PinIO = gpio.INVALID // PWM0
	J2_33 gpio.PinIO = gpio.INVALID // PWM1
	J2_34            = pin.GROUND   //
	J2_35 gpio.PinIO = gpio.INVALID // I2S_WS, SPI1_MISO, PWM1
	J2_36 gpio.PinIO = gpio.INVALID // UART0_CTS, SPI1_CS2, UART1_CTS
	J2_37            = pin.INVALID  // BUG(tve): make pins J2_37 and J2_40 functional once analog support is implemented
	J2_38            = pin.V1_8     //
	J2_39            = pin.GROUND   //
	J2_40            = pin.INVALID  // See above.
)

// Present returns true if running on a Hardkernel ODROID-C0/C1/C1+ board.
//
// It looks for "8726_M8B" in the device tree or "ODROIDC" in cpuinfo. The
// following information is expected in the device dtree:
//   root@odroid:/proc/device-tree# od -c compatible
//   0000000   A   M   L   O   G   I   C   ,   8   7   2   6   _   M   8   B
func Present() bool {
	for _, c := range distro.DTCompatible() {
		if strings.Contains(c, "8726_M8B") {
			return true
		}
	}
	return distro.CPUInfo()["Hardware"] == "ODROIDC"
}

//

// aliases is a list of aliases for the various gpio pins, this allows users to
// refer to pins using the documented and labeled names instead of some GPIOnnn
// name. The map key is the alias and the value is the real pin name.
var aliases = map[string]int{
	"I2C0_SDA":  74,
	"I2C0_SCL":  75,
	"I2C1_SDA":  76,
	"I2C1_SCL":  77,
	"I2CA_SDA":  74,
	"I2CA_SCL":  75,
	"I2CB_SDA":  76,
	"I2CB_SCL":  77,
	"SPI0_MOSI": 107,
	"SPI0_MISO": 106,
	"SPI0_CLK":  105,
	"SPI0_CS0":  117,
}

// sysfsPin is a safe way to get a sysfs pin
func sysfsPin(n int) gpio.PinIO {
	if p, ok := sysfs.Pins[n]; ok {
		return p
	}
	return gpio.INVALID
}

// driver implements drivers.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "odroid-c1"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	return []string{"sysfs-gpio"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("Hardkernel ODROID-C0/C1/C1+ board not detected")
	}
	J2_3 = sysfsPin(74)
	J2_5 = sysfsPin(75)
	J2_7 = sysfsPin(83)   // usually taken by 1-wire driver
	J2_8 = sysfsPin(113)  // usually not available
	J2_10 = sysfsPin(114) // usually not available
	J2_11 = sysfsPin(88)
	J2_12 = sysfsPin(87)
	J2_13 = sysfsPin(116)
	J2_15 = sysfsPin(115)
	J2_16 = sysfsPin(104)
	J2_18 = sysfsPin(102)
	J2_19 = sysfsPin(107)
	J2_21 = sysfsPin(106)
	J2_22 = sysfsPin(103)
	J2_23 = sysfsPin(105)
	J2_24 = sysfsPin(117)
	J2_26 = sysfsPin(118)
	J2_27 = sysfsPin(76)
	J2_28 = sysfsPin(77)
	J2_29 = sysfsPin(101)
	J2_31 = sysfsPin(100)
	J2_32 = sysfsPin(99)
	J2_33 = sysfsPin(108)
	J2_35 = sysfsPin(97)
	J2_36 = sysfsPin(98)

	// J2 is the 40-pin rPi-compatible header
	J2 := [][]pin.Pin{
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
	if err := pinreg.Register("J2", J2); err != nil {
		return true, err
	}
	for alias, number := range aliases {
		if err := gpioreg.RegisterAlias(alias, strconv.Itoa(number)); err != nil {
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
