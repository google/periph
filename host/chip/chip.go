// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package chip

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/periph"
	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/pins"
	"github.com/google/periph/host/allwinner"
	"github.com/google/periph/host/distro"
	"github.com/google/periph/host/headers"
	"github.com/google/periph/host/sysfs"
)

// C.H.I.P. hardware pins.
var (
	TEMP_SENSOR = &gpio.BasicPin{N: "TEMP_SENSOR"}
	PWR_SWITCH  = &gpio.BasicPin{N: "PWR_SWITCH"}
	// XIO "gpio" pins attached to the pcf8574 I²C port extender.
	XIO0, XIO1, XIO2, XIO3, XIO4, XIO5, XIO6, XIO7 gpio.PinIO
)

// The U13 header is opposite the power LED.
//
// The alternate pin functionality is described at pages 322-323 of
// https://github.com/NextThingCo/CHIP-Hardware/raw/master/CHIP%5Bv1_0%5D/CHIPv1_0-BOM-Datasheets/Allwinner%20R8%20User%20Manual%20V1.1.pdf
var (
	U13_1  = pins.GROUND    //
	U13_2  = pins.DC_IN     //
	U13_3  = pins.V5        // (filtered)
	U13_4  = pins.GROUND    //
	U13_5  = pins.V3_3      //
	U13_6  = TEMP_SENSOR    // Analog temp sensor input
	U13_7  = pins.V1_8      //
	U13_8  = pins.BAT_PLUS  // External LiPo battery
	U13_9  = allwinner.PB16 // I2C1_SDA
	U13_10 = PWR_SWITCH     // Power button
	U13_11 = allwinner.PB15 // I2C1_SCL
	U13_12 = pins.GROUND    //
	U13_13 = allwinner.X1   // Touch screen X1
	U13_14 = allwinner.X2   // Touch screen X2
	U13_15 = allwinner.Y1   // Touch screen Y1
	U13_16 = allwinner.Y2   // Touch screen Y2
	U13_17 = allwinner.PD2  // LCD-D2; UART2_TX firmware probe for 1-wire to detect DIP at boot; http://docs.getchip.com/dip.html#dip-identification
	U13_18 = allwinner.PB2  // PWM0; EINT16
	U13_19 = allwinner.PD4  // LCD-D4; UART2_CTS
	U13_20 = allwinner.PD3  // LCD-D3; UART2_RX
	U13_21 = allwinner.PD6  // LCD-D6
	U13_22 = allwinner.PD5  // LCD-D5
	U13_23 = allwinner.PD10 // LCD-D10
	U13_24 = allwinner.PD7  // LCD-D7
	U13_25 = allwinner.PD12 // LCD-D12
	U13_26 = allwinner.PD11 // LCD-D11
	U13_27 = allwinner.PD14 // LCD-D14
	U13_28 = allwinner.PD13 // LCD-D13
	U13_29 = allwinner.PD18 // LCD-D18
	U13_30 = allwinner.PD15 // LCD-D15
	U13_31 = allwinner.PD20 // LCD-D20
	U13_32 = allwinner.PD19 // LCD-D19
	U13_33 = allwinner.PD22 // LCD-D22
	U13_34 = allwinner.PD21 // LCD-D21
	U13_35 = allwinner.PD24 // LCD-CLK
	U13_36 = allwinner.PD23 // LCD-D23
	U13_37 = allwinner.PD26 // LCD-VSYNC
	U13_38 = allwinner.PD27 // LCD-HSYNC
	U13_39 = pins.GROUND    //
	U13_40 = allwinner.PD25 // LCD-DE: RGB666 data
)

// The U14 header is right next to the power LED.
var (
	U14_1  = pins.GROUND        //
	U14_2  = pins.V5            // (filtered)
	U14_3  = allwinner.PG3      // UART1_TX; EINT3
	U14_4  = allwinner.HP_LEFT  // Headphone left output
	U14_5  = allwinner.PG4      // UART1_RX; EINT4
	U14_6  = allwinner.HP_COM   // Headphone amp out
	U14_7  = allwinner.FEL      // Boot mode selection
	U14_8  = allwinner.HP_RIGHT // Headphone right output
	U14_9  = pins.V3_3          //
	U14_10 = allwinner.MIC_GND  // Microphone ground
	U14_11 = allwinner.KEY_ADC  // LRADC Low res analog to digital
	U14_12 = allwinner.MIC_IN   // Microphone input
	U14_13 = XIO0               // gpio via I²C controller
	U14_14 = XIO1               // gpio via I²C controller
	U14_15 = XIO2               // gpio via I²C controller
	U14_16 = XIO3               // gpio via I²C controller
	U14_17 = XIO4               // gpio via I²C controller
	U14_18 = XIO5               // gpio via I²C controller
	U14_19 = XIO6               // gpio via I²C controller
	U14_20 = XIO7               // gpio via I²C controller
	U14_21 = pins.GROUND        //
	U14_22 = pins.GROUND        //
	U14_23 = allwinner.PG1      // GPS_CLK; AP-EINT1
	U14_24 = allwinner.PB3      // IR_TX; AP-EINT3 (EINT17)
	U14_25 = allwinner.PB18     // I2C2_SDA
	U14_26 = allwinner.PB17     // I2C2_SCL
	U14_27 = allwinner.PE0      // CSIPCK: CMOS serial interface; SPI2_CS0; EINT14
	U14_28 = allwinner.PE1      // CSICK: CMOS serial interface; SPI2_CLK; EINT15
	U14_29 = allwinner.PE2      // CSIHSYNC; SPI2_MOSI
	U14_30 = allwinner.PE3      // CSIVSYNC; SPI2_MISO
	U14_31 = allwinner.PE4      // CSID0
	U14_32 = allwinner.PE5      // CSID1
	U14_33 = allwinner.PE6      // CSID2
	U14_34 = allwinner.PE7      // CSID3
	U14_35 = allwinner.PE8      // CSID4
	U14_36 = allwinner.PE9      // CSID5
	U14_37 = allwinner.PE10     // CSID6; UART1_RX
	U14_38 = allwinner.PE11     // CSID7; UART1_TX
	U14_39 = pins.GROUND        //
	U14_40 = pins.GROUND        //
)

// Present returns true if running on a NextThing Co's C.H.I.P. board.
//
// It looks for "C.H.I.P" in the device tree. The following information is
// expected in the device dtree:
//   root@chip2:/proc/device-tree# od -c compatible
//   0000000   n   e   x   t   t   h   i   n   g   ,   c   h   i   p  \0   a
//   0000020   l   l   w   i   n   n   e   r   ,   s   u   n   5   i   -   r
//   0000040   8  \0
//   root@chip2:/proc/device-tree# od -c model
//   0000000   N   e   x   t   T   h   i   n   g       C   .   H   .   I   .
//   0000020   P   .  \0
func Present() bool {
	return strings.Contains(distro.DTModel(), "C.H.I.P")
}

//

// aliases is a list of aliases for the various gpio pins, this allows users to refer to pins
// using the documented and labeled names instead of some GPIOnnn name. The map key is the
// alias and the value is the real pin name.
var aliases = map[string]string{
	"XIO-P0": "GPIO1016",
	"XIO-P1": "GPIO1017",
	"XIO-P2": "GPIO1018",
	"XIO-P3": "GPIO1019",
	"XIO-P4": "GPIO1020",
	"XIO-P5": "GPIO1021",
	"XIO-P6": "GPIO1022",
	"XIO-P7": "GPIO1023",
	"LCD-D2": "PD2",
}

func init() {
	// These are initialized later by the driver.
	XIO0 = gpio.INVALID
	XIO1 = gpio.INVALID
	XIO2 = gpio.INVALID
	XIO3 = gpio.INVALID
	XIO4 = gpio.INVALID
	XIO5 = gpio.INVALID
	XIO6 = gpio.INVALID
	XIO7 = gpio.INVALID
	// These must be reinitialized.
	U14_13 = XIO0
	U14_14 = XIO1
	U14_15 = XIO2
	U14_16 = XIO3
	U14_17 = XIO4
	U14_18 = XIO5
	U14_19 = XIO6
	U14_20 = XIO7
}

// driver implements drivers.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "chip"
}

func (d *driver) Prerequisites() []string {
	// has allwinner cpu, needs sysfs for XIO0-XIO7 "gpio" pins
	return []string{"allwinner-gpio", "sysfs-gpio"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("NextThing Co. CHIP board not detected")
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
	// we can now hook-up the appropriate CHIP pins to sysfs gpio pins.
	XIO0 = sysfsPin(1016)
	XIO1 = sysfsPin(1017)
	XIO2 = sysfsPin(1018)
	XIO3 = sysfsPin(1019)
	XIO4 = sysfsPin(1020)
	XIO5 = sysfsPin(1021)
	XIO6 = sysfsPin(1022)
	XIO7 = sysfsPin(1023)
	// These must be reinitialized.
	U14_13 = XIO0
	U14_14 = XIO1
	U14_15 = XIO2
	U14_16 = XIO3
	U14_17 = XIO4
	U14_18 = XIO5
	U14_19 = XIO6
	U14_20 = XIO7

	// U13 is one of the 20x2 connectors.
	U13 := [][]pins.Pin{
		{U13_1, U13_2},
		{U13_3, U13_4},
		{U13_5, U13_6},
		{U13_7, U13_8},
		{U13_9, U13_10},
		{U13_11, U13_12},
		{U13_13, U13_14},
		{U13_15, U13_16},
		{U13_17, U13_18},
		{U13_19, U13_20},
		{U13_21, U13_22},
		{U13_23, U13_24},
		{U13_25, U13_26},
		{U13_27, U13_28},
		{U13_29, U13_30},
		{U13_31, U13_32},
		{U13_33, U13_34},
		{U13_35, U13_36},
		{U13_37, U13_38},
		{U13_39, U13_40},
	}
	if err := headers.Register("U13", U13); err != nil {
		return true, err
	}

	// U14 is one of the 20x2 connectors.
	U14 := [][]pins.Pin{
		{U14_1, U14_2},
		{U14_3, U14_4},
		{U14_5, U14_6},
		{U14_7, U14_8},
		{U14_9, U14_10},
		{U14_11, U14_12},
		{U14_13, U14_14},
		{U14_15, U14_16},
		{U14_17, U14_18},
		{U14_19, U14_20},
		{U14_21, U14_22},
		{U14_23, U14_24},
		{U14_25, U14_26},
		{U14_27, U14_28},
		{U14_29, U14_30},
		{U14_31, U14_32},
		{U14_33, U14_34},
		{U14_35, U14_36},
		{U14_37, U14_38},
		{U14_39, U14_40},
	}
	if err := headers.Register("U14", U14); err != nil {
		return true, err
	}

	// Register explicit pin aliases.
	for alias, real := range aliases {
		r := gpio.ByName(real)
		if r == nil {
			return true, fmt.Errorf("cannot create alias for %s: it doesn't exist", real)
		}
		if err := gpio.RegisterAlias(alias, r.Number()); err != nil {
			return true, err
		}
	}

	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&driver{})
	}
}
