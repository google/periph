// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bone implements headers J1, P8 and P9 found on many (but not all)
// BeagleBone micro-computer.
//
// In particular, the headers are found on the models using a TI AM335x
// processor: BeagleBone Black, Black Wireless, Green and Green Wireless.
//
// Reference
//
// http://beagleboard.org/Support/bone101/#hardware
package bone

import (
	"errors"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/beagle/black"
	"periph.io/x/periph/host/beagle/green"
	"periph.io/x/periph/host/sysfs"
)

// TODO(maruel): Use specialized am335x or pru implementation once available.

// Common pin types on BeagleBones.
var (
	PWR_BUT   = &pin.BasicPin{N: "PWR_BUT"}   //
	RESET_OUT = &pin.BasicPin{N: "RESET_OUT"} // SYS_RESETn
	VADC      = &pin.BasicPin{N: "VADC"}      // VDD_ADC
	AIN4      = &pin.BasicPin{N: "AIN4"}      // AIN4
	AGND      = &pin.BasicPin{N: "AGND"}      // GNDA_ADC
	AIN6      = &pin.BasicPin{N: "AIN6"}      // AIN6
	AIN5      = &pin.BasicPin{N: "AIN5"}      // AIN5
	AIN2      = &pin.BasicPin{N: "AIN2"}      // AIN2
	AIN3      = &pin.BasicPin{N: "AIN3"}      // AIN3
	AIN0      = &pin.BasicPin{N: "AIN0"}      // AIN0
	AIN1      = &pin.BasicPin{N: "AIN1"}      // AIN1
)

// Headers found on BeagleBones.
var (
	// Port J1 is the UART port where the default terminal is connected to.
	J1_1 pin.Pin    = pin.GROUND
	J1_2 pin.Pin    = pin.INVALID
	J1_3 pin.Pin    = pin.INVALID
	J1_4 gpio.PinIO = gpio.INVALID // GPIO42, UART0_RX
	J1_5 gpio.PinIO = gpio.INVALID // GPIO43, UART0_TX
	J1_6 pin.Pin    = pin.INVALID

	P8_1  pin.Pin    = pin.GROUND
	P8_2  pin.Pin    = pin.GROUND
	P8_3  gpio.PinIO = gpio.INVALID // GPIO38, MMC1_DAT6
	P8_4  gpio.PinIO = gpio.INVALID // GPIO39, MMC1_DAT7
	P8_5  gpio.PinIO = gpio.INVALID // GPIO34, MMC1_DAT2
	P8_6  gpio.PinIO = gpio.INVALID // GPIO35, MMC1_DAT3
	P8_7  gpio.PinIO = gpio.INVALID // GPIO66, Timer4
	P8_8  gpio.PinIO = gpio.INVALID // GPIO67, Timer7
	P8_9  gpio.PinIO = gpio.INVALID // GPIO69, Timer5
	P8_10 gpio.PinIO = gpio.INVALID // GPIO68, Timer6
	P8_11 gpio.PinIO = gpio.INVALID // GPIO45,
	P8_12 gpio.PinIO = gpio.INVALID // GPIO44,
	P8_13 gpio.PinIO = gpio.INVALID // GPIO23, EHRPWM2B
	P8_14 gpio.PinIO = gpio.INVALID // GPIO26,
	P8_15 gpio.PinIO = gpio.INVALID // GPIO47,
	P8_16 gpio.PinIO = gpio.INVALID // GPIO46,
	P8_17 gpio.PinIO = gpio.INVALID // GPIO27,
	P8_18 gpio.PinIO = gpio.INVALID // GPIO65,
	P8_19 gpio.PinIO = gpio.INVALID // GPIO22, EHRPWM2A
	P8_20 gpio.PinIO = gpio.INVALID // GPIO63, MMC1_CMD
	P8_21 gpio.PinIO = gpio.INVALID // GPIO62, MMC1_CLK
	P8_22 gpio.PinIO = gpio.INVALID // GPIO37, MMC1_DAT5
	P8_23 gpio.PinIO = gpio.INVALID // GPIO36, MMC1_DAT4
	P8_24 gpio.PinIO = gpio.INVALID // GPIO33, MMC1_DAT1
	P8_25 gpio.PinIO = gpio.INVALID // GPIO32, MMC1_DAT0
	P8_26 gpio.PinIO = gpio.INVALID // GPIO61,
	P8_27 gpio.PinIO = gpio.INVALID // GPIO86, LCD_VSYNC
	P8_28 gpio.PinIO = gpio.INVALID // GPIO88, LCD_PCLK
	P8_29 gpio.PinIO = gpio.INVALID // GPIO87, LCD_HSYNC
	P8_30 gpio.PinIO = gpio.INVALID // GPIO89, LCD_AC_BIAS_E
	P8_31 gpio.PinIO = gpio.INVALID // GPIO10, LCD_DATA14, UART4_CTS
	P8_32 gpio.PinIO = gpio.INVALID // GPIO11, LCD_DATA15, UART5_RTS
	P8_33 gpio.PinIO = gpio.INVALID // GPIO9, LCD_DATA13, UART4_RTS
	P8_34 gpio.PinIO = gpio.INVALID // GPIO81, LCD_DATA11, EHRPWM1B, UART3_RTS
	P8_35 gpio.PinIO = gpio.INVALID // GPIO8, LCD_DATA12, UART4_CTS
	P8_36 gpio.PinIO = gpio.INVALID // GPIO80, LCD_DATA10, EHRPWM1A, UART3_CTS
	P8_37 gpio.PinIO = gpio.INVALID // GPIO78, LCD_DATA8, UART5_TX
	P8_38 gpio.PinIO = gpio.INVALID // GPIO79, LCD_DATA9, UART5_RX
	P8_39 gpio.PinIO = gpio.INVALID // GPIO76, LCD_DATA6
	P8_40 gpio.PinIO = gpio.INVALID // GPIO77, LCD_DATA7
	P8_41 gpio.PinIO = gpio.INVALID // GPIO74, LCD_DATA4
	P8_42 gpio.PinIO = gpio.INVALID // GPIO75, LCD_DATA5
	P8_43 gpio.PinIO = gpio.INVALID // GPIO72, LCD_DATA2
	P8_44 gpio.PinIO = gpio.INVALID // GPIO73, LCD_DATA3
	P8_45 gpio.PinIO = gpio.INVALID // GPIO70, LCD_DATA0, EHRPWM2A
	P8_46 gpio.PinIO = gpio.INVALID // GPIO71, LCD_DATA1, EHRPWM2B

	P9_1  pin.Pin    = pin.GROUND
	P9_2  pin.Pin    = pin.GROUND
	P9_3  pin.Pin    = pin.V3_3
	P9_4  pin.Pin    = pin.V3_3
	P9_5  pin.Pin    = pin.V5
	P9_6  pin.Pin    = pin.V5
	P9_7  pin.Pin    = pin.V5
	P9_8  pin.Pin    = pin.V5
	P9_9  pin.Pin    = PWR_BUT      // PWR_BUT
	P9_10 pin.Pin    = RESET_OUT    // SYS_RESETn
	P9_11 gpio.PinIO = gpio.INVALID // GPIO30, UART4_RX
	P9_12 gpio.PinIO = gpio.INVALID // GPIO60
	P9_13 gpio.PinIO = gpio.INVALID // GPIO31, UART4_TX
	P9_14 gpio.PinIO = gpio.INVALID // GPIO50, EHRPWM1A
	P9_15 gpio.PinIO = gpio.INVALID // GPIO48
	P9_16 gpio.PinIO = gpio.INVALID // GPIO51, EHRPWM1B
	P9_17 gpio.PinIO = gpio.INVALID // GPIO5, I2C1_SCL, SPI0_CS0
	P9_18 gpio.PinIO = gpio.INVALID // GPIO4, I2C1_SDA, SPI0_MISO
	P9_19 gpio.PinIO = gpio.INVALID // GPIO13, I2C2_SCL, UART1_RTS, SPI1_CS1
	P9_20 gpio.PinIO = gpio.INVALID // GPIO12, I2C2_SDA, UART1_CTS, SPI1_CS0
	P9_21 gpio.PinIO = gpio.INVALID // GPIO3, EHRPWM0B, I2C2_SCL, UART2_TX, SPI0_MOSI
	P9_22 gpio.PinIO = gpio.INVALID // GPIO2, EHRPWM0A, I2C2_SDA, UART2_RX, SPI0_CLK
	P9_23 gpio.PinIO = gpio.INVALID // GPIO49
	P9_24 gpio.PinIO = gpio.INVALID // GPIO15, I2C1_SCL, UART1_TX
	P9_25 gpio.PinIO = gpio.INVALID // GPIO117
	P9_26 gpio.PinIO = gpio.INVALID // GPIO14, I2C1_SDA, UART1_RX
	P9_27 gpio.PinIO = gpio.INVALID // GPIO115
	P9_28 gpio.PinIO = gpio.INVALID // GPIO113, ECAPPWM2, SPI1_CS0
	P9_29 gpio.PinIO = gpio.INVALID // GPIO111, EHRPWM0B, SPI1_MOSI
	P9_30 gpio.PinIO = gpio.INVALID // GPIO112, SPI1_MISO
	P9_31 gpio.PinIO = gpio.INVALID // GPIO110, EHRPWM0A, SPI1_CLK
	P9_32 pin.Pin    = VADC         // VDD_ADC
	P9_33 pin.Pin    = AIN4         // AIN4
	P9_34 pin.Pin    = AGND         // GNDA_ADC
	P9_35 pin.Pin    = AIN6         // AIN6
	P9_36 pin.Pin    = AIN5         // AIN5
	P9_37 pin.Pin    = AIN2         // AIN2
	P9_38 pin.Pin    = AIN3         // AIN3
	P9_39 pin.Pin    = AIN0         // AIN0
	P9_40 pin.Pin    = AIN1         // AIN1
	P9_41 gpio.PinIO = gpio.INVALID // GPIO20
	P9_42 gpio.PinIO = gpio.INVALID // GPIO7, ECAPPWM0, UART3_TX, SPI1_CS1
	P9_43 pin.Pin    = pin.GROUND
	P9_44 pin.Pin    = pin.GROUND
	P9_45 pin.Pin    = pin.GROUND
	P9_46 pin.Pin    = pin.GROUND
)

// Present returns true if the host is a BeagleBone Black/Green or their
// Wireless version.
func Present() bool {
	return black.Present() || green.Present()
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "beaglebone"
}

func (d *driver) Prerequisites() []string {
	return []string{"am335x"}
}

func (d *driver) After() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("BeagleBone board not detected")
	}

	J1_4 = sysfs.Pins[42]
	J1_5 = sysfs.Pins[43]

	P8_3 = sysfs.Pins[38]
	P8_4 = sysfs.Pins[39]
	P8_5 = sysfs.Pins[34]
	P8_6 = sysfs.Pins[35]
	P8_7 = sysfs.Pins[66]
	P8_8 = sysfs.Pins[67]
	P8_9 = sysfs.Pins[69]
	P8_10 = sysfs.Pins[68]
	P8_11 = sysfs.Pins[45]
	P8_12 = sysfs.Pins[44]
	P8_13 = sysfs.Pins[23]
	P8_14 = sysfs.Pins[26]
	P8_15 = sysfs.Pins[47]
	P8_16 = sysfs.Pins[46]
	P8_17 = sysfs.Pins[27]
	P8_18 = sysfs.Pins[65]
	P8_19 = sysfs.Pins[22]
	P8_20 = sysfs.Pins[63]
	P8_21 = sysfs.Pins[62]
	P8_22 = sysfs.Pins[37]
	P8_23 = sysfs.Pins[36]
	P8_24 = sysfs.Pins[33]
	P8_25 = sysfs.Pins[32]
	P8_26 = sysfs.Pins[61]
	P8_27 = sysfs.Pins[86]
	P8_28 = sysfs.Pins[88]
	P8_29 = sysfs.Pins[87]
	P8_30 = sysfs.Pins[89]
	P8_31 = sysfs.Pins[10]
	P8_32 = sysfs.Pins[11]
	P8_33 = sysfs.Pins[9]
	P8_34 = sysfs.Pins[81]
	P8_35 = sysfs.Pins[8]
	P8_36 = sysfs.Pins[80]
	P8_37 = sysfs.Pins[78]
	P8_38 = sysfs.Pins[79]
	P8_39 = sysfs.Pins[76]
	P8_40 = sysfs.Pins[77]
	P8_41 = sysfs.Pins[74]
	P8_42 = sysfs.Pins[75]
	P8_43 = sysfs.Pins[72]
	P8_44 = sysfs.Pins[73]
	P8_45 = sysfs.Pins[70]
	P8_46 = sysfs.Pins[71]

	P9_11 = sysfs.Pins[30]
	P9_12 = sysfs.Pins[60]
	P9_13 = sysfs.Pins[31]
	P9_14 = sysfs.Pins[50]
	P9_15 = sysfs.Pins[48]
	P9_16 = sysfs.Pins[51]
	P9_17 = sysfs.Pins[5]
	P9_18 = sysfs.Pins[4]
	P9_19 = sysfs.Pins[13]
	P9_20 = sysfs.Pins[12]
	P9_21 = sysfs.Pins[3]
	P9_22 = sysfs.Pins[2]
	P9_23 = sysfs.Pins[49]
	P9_24 = sysfs.Pins[15]
	P9_25 = sysfs.Pins[117]
	P9_26 = sysfs.Pins[14]
	P9_27 = sysfs.Pins[115]
	P9_28 = sysfs.Pins[113]
	P9_29 = sysfs.Pins[111]
	P9_30 = sysfs.Pins[112]
	P9_31 = sysfs.Pins[110]
	P9_41 = sysfs.Pins[20]
	P9_42 = sysfs.Pins[7]

	hdr := [][]pin.Pin{{J1_1}, {J1_2}, {J1_3}, {J1_4}, {J1_5}, {J1_6}}
	if err := pinreg.Register("J1", hdr); err != nil {
		return true, err
	}

	hdr = [][]pin.Pin{
		{P8_1, P8_2},
		{P8_3, P8_4},
		{P8_5, P8_6},
		{P8_7, P8_8},
		{P8_9, P8_10},
		{P8_11, P8_12},
		{P8_13, P8_14},
		{P8_15, P8_16},
		{P8_17, P8_18},
		{P8_19, P8_20},
		{P8_21, P8_22},
		{P8_23, P8_24},
		{P8_25, P8_26},
		{P8_27, P8_28},
		{P8_29, P8_30},
		{P8_31, P8_32},
		{P8_33, P8_34},
		{P8_35, P8_36},
		{P8_37, P8_38},
		{P8_39, P8_40},
		{P8_41, P8_42},
		{P8_43, P8_44},
		{P8_45, P8_46},
	}
	if err := pinreg.Register("P8", hdr); err != nil {
		return true, err
	}

	hdr = [][]pin.Pin{
		{P9_1, P9_2},
		{P9_3, P9_4},
		{P9_5, P9_6},
		{P9_7, P9_8},
		{P9_9, P9_10},
		{P9_11, P9_12},
		{P9_13, P9_14},
		{P9_15, P9_16},
		{P9_17, P9_18},
		{P9_19, P9_20},
		{P9_21, P9_22},
		{P9_23, P9_24},
		{P9_25, P9_26},
		{P9_27, P9_28},
		{P9_29, P9_30},
		{P9_31, P9_32},
		{P9_33, P9_34},
		{P9_35, P9_36},
		{P9_37, P9_38},
		{P9_39, P9_40},
		{P9_41, P9_42},
		{P9_43, P9_44},
		{P9_45, P9_46},
	}
	err := pinreg.Register("P9", hdr)
	return true, err
}

func init() {
	if isArm {
		periph.MustRegister(&drv)
	}
}

var drv driver
