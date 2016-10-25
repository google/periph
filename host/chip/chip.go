// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package chip

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/periph"
	"github.com/google/periph/conn/analog"
	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/pins"
	"github.com/google/periph/host/allwinner"
	"github.com/google/periph/host/distro"
	"github.com/google/periph/host/headers"
	"github.com/google/periph/host/sysfs"
)

var (
	DC_IN    pins.Pin = &pins.BasicPin{N: "DC_IN"}
	BAT_PLUS pins.Pin = &pins.BasicPin{N: "BAT_PLUS"}
	V1_8     pins.Pin = &pins.BasicPin{N: "V1_8"} // 1.8 volt output

	TEMP_SENSOR gpio.PinIO = &gpio.BasicPin{N: "TEMP_SENSOR"}
	PWR_SWITCH  gpio.PinIO = &gpio.BasicPin{N: "PWR_SWITCH"}

	// XIO "gpio" pins attached to the pcf8574 I2c port extender, these get
	// initialized in the Init function
	XIO0, XIO1, XIO2, XIO3, XIO4, XIO5, XIO6, XIO7 gpio.PinIO
)

var (
	U13_1  pins.Pin   = pins.GROUND    //
	U13_2  pins.Pin   = DC_IN          // 5 volt input
	U13_3  pins.Pin   = pins.V5        // 5 volt output
	U13_4  pins.Pin   = pins.GROUND    //
	U13_5  pins.Pin   = pins.V3_3      // 3.3v output
	U13_6  gpio.PinIO = TEMP_SENSOR    // analog temp sensor input
	U13_7  pins.Pin   = V1_8           // 1.8v output
	U13_8  pins.Pin   = BAT_PLUS       // external LiPo battery
	U13_9  gpio.PinIO = allwinner.PB16 //
	U13_10 pins.Pin   = PWR_SWITCH     // power button
	U13_11 gpio.PinIO = allwinner.PB15 //
	U13_12 pins.Pin   = pins.GROUND    //
	U13_13 gpio.PinIO = allwinner.X1   // touch screen X1
	U13_14 gpio.PinIO = allwinner.X2   // touch screen X2
	U13_15 gpio.PinIO = allwinner.Y1   // touch screen Y1
	U13_16 gpio.PinIO = allwinner.Y2   // touch screen Y2
	U13_17 gpio.PinIO = allwinner.PD2  //
	U13_18 gpio.PinIO = allwinner.PB2  //
	U13_19 gpio.PinIO = allwinner.PD4  //
	U13_20 gpio.PinIO = allwinner.PD3  //
	U13_21 gpio.PinIO = allwinner.PD6  //
	U13_22 gpio.PinIO = allwinner.PD5  //
	U13_23 gpio.PinIO = allwinner.PD10 //
	U13_24 gpio.PinIO = allwinner.PD7  //
	U13_25 gpio.PinIO = allwinner.PD12 //
	U13_26 gpio.PinIO = allwinner.PD11 //
	U13_27 gpio.PinIO = allwinner.PD14 //
	U13_28 gpio.PinIO = allwinner.PD13 //
	U13_29 gpio.PinIO = allwinner.PD18 //
	U13_30 gpio.PinIO = allwinner.PD15 //
	U13_31 gpio.PinIO = allwinner.PD20 //
	U13_32 gpio.PinIO = allwinner.PD19 //
	U13_33 gpio.PinIO = allwinner.PD22 //
	U13_34 gpio.PinIO = allwinner.PD21 //
	U13_35 gpio.PinIO = allwinner.PD24 //
	U13_36 gpio.PinIO = allwinner.PD23 //
	U13_37 gpio.PinIO = allwinner.PD27 //
	U13_38 gpio.PinIO = allwinner.PD26 //
	U13_39 pins.Pin   = pins.GROUND    //
	U13_40 pins.Pin   = pins.GROUND    //

	U14_1  pins.Pin     = pins.GROUND //
	U14_2  pins.Pin     = pins.V5     // 5 volt output
	U14_3  gpio.PinIO   = allwinner.PG3
	U14_4  gpio.PinIO   = allwinner.HP_LEFT  // headphone left output
	U14_5  gpio.PinIO   = allwinner.PG4      //
	U14_6  pins.Pin     = allwinner.HP_COM   // headphone amp out
	U14_7  pins.Pin     = allwinner.FEL      // boot mode selection
	U14_8  gpio.PinIO   = allwinner.HP_RIGHT // headphone right output
	U14_9  pins.Pin     = pins.V3_3          // 3.3v output
	U14_10 pins.Pin     = allwinner.MIC_GND  // microphone ground
	U14_11 analog.PinIO = allwinner.KEY_ADC  // low res analog to digital
	U14_12 gpio.PinIO   = allwinner.MIC_IN   // microphone input
	U14_13 gpio.PinIO   = XIO0               // gpio via i2c controller
	U14_14 gpio.PinIO   = XIO1               // gpio via i2c controller
	U14_15 gpio.PinIO   = XIO2               // gpio via i2c controller
	U14_16 gpio.PinIO   = XIO3               // gpio via i2c controller
	U14_17 gpio.PinIO   = XIO4               // gpio via i2c controller
	U14_18 gpio.PinIO   = XIO5               // gpio via i2c controller
	U14_19 gpio.PinIO   = XIO6               // gpio via i2c controller
	U14_20 gpio.PinIO   = XIO7               // gpio via i2c controller
	U14_21 pins.Pin     = pins.GROUND        //
	U14_22 pins.Pin     = pins.GROUND        //
	U14_23 gpio.PinIO   = allwinner.PG1      //
	U14_24 gpio.PinIO   = allwinner.PB3      //
	U14_25 gpio.PinIO   = allwinner.PB18     //
	U14_26 gpio.PinIO   = allwinner.PB17     //
	U14_27 gpio.PinIO   = allwinner.PE0      //
	U14_28 gpio.PinIO   = allwinner.PE1      //
	U14_29 gpio.PinIO   = allwinner.PE2      //
	U14_30 gpio.PinIO   = allwinner.PE3      //
	U14_31 gpio.PinIO   = allwinner.PE4      //
	U14_32 gpio.PinIO   = allwinner.PE5      //
	U14_33 gpio.PinIO   = allwinner.PE6      //
	U14_34 gpio.PinIO   = allwinner.PE7      //
	U14_35 gpio.PinIO   = allwinner.PE8      //
	U14_36 gpio.PinIO   = allwinner.PE9      //
	U14_37 gpio.PinIO   = allwinner.PE10     //
	U14_38 gpio.PinIO   = allwinner.PE11     //
	U14_39 pins.Pin     = pins.GROUND        //
	U14_40 pins.Pin     = pins.GROUND        //
)

// Present returns true if running on a NextThing Co's C.H.I.P. board.
//
// https://www.getchip.com/
func Present() bool {
	return strings.Contains(distro.DTModel(), "C.H.I.P")
}

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

// driver implements drivers.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "chip"
}

func (d *driver) Type() periph.Type {
	return periph.Pins
}

func (d *driver) Prerequisites() []string {
	// has allwinner cpu, needs sysfs for XIO0-XIO7 "gpio" pins
	return []string{"allwinner", "sysfs-gpio"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("NextThing Co. CHIP board not detected")
	}

	// sysfsPin is a safe say to get a sysfs pin
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
	// Need to set header pins too 'cause XIOn are interfaces, i.e. pointers.
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
