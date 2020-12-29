// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pine64

import (
	"errors"
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/allwinner"
	"periph.io/x/periph/host/distro"
)

// Present returns true if running on a Pine64 board.
//
// https://www.pine64.org/
func Present() bool {
	if isArm {
		return strings.HasPrefix(distro.DTModel(), "Pine64")
	}
	return false
}

// Pine64 specific pins.
var (
	VCC         = &pin.BasicPin{N: "VCC"}         //
	IOVCC       = &pin.BasicPin{N: "IOVCC"}       // Power supply for port A
	TEMP_SENSOR = &pin.BasicPin{N: "TEMP_SENSOR"} //
	IR_RX       = &pin.BasicPin{N: "IR_RX"}       // IR Data Receive
	CHARGER_LED = &pin.BasicPin{N: "CHARGER_LED"} //
	RESET       = &pin.BasicPin{N: "RESET"}       //
	PWR_SWITCH  = &pin.BasicPin{N: "PWR_SWITCH "} //
)

// All the individual pins on the headers.
var (
	P1_1  = pin.V3_3       // max 40mA
	P1_2  = pin.V5         // (filtered)
	P1_3  = allwinner.PH3  //
	P1_4  = pin.V5         // (filtered)
	P1_5  = allwinner.PH2  //
	P1_6  = pin.GROUND     //
	P1_7  = allwinner.PL10 //
	P1_8  = allwinner.PB0  //
	P1_9  = pin.GROUND     //
	P1_10 = allwinner.PB1  //
	P1_11 = allwinner.PC7  //
	P1_12 = allwinner.PC8  //
	P1_13 = allwinner.PH9  //
	P1_14 = pin.GROUND     //
	P1_15 = allwinner.PC12 //
	P1_16 = allwinner.PC13 //
	P1_17 = pin.V3_3       //
	P1_18 = allwinner.PC14 //
	P1_19 = allwinner.PC0  //
	P1_20 = pin.GROUND     //
	P1_21 = allwinner.PC1  //
	P1_22 = allwinner.PC15 //
	P1_23 = allwinner.PC2  //
	P1_24 = allwinner.PC3  //
	P1_25 = pin.GROUND     //
	P1_26 = allwinner.PH7  //
	P1_27 = allwinner.PL9  //
	P1_28 = allwinner.PL8  //
	P1_29 = allwinner.PH5  //
	P1_30 = pin.GROUND     //
	P1_31 = allwinner.PH6  //
	P1_32 = allwinner.PC4  //
	P1_33 = allwinner.PC5  //
	P1_34 = pin.GROUND     //
	P1_35 = allwinner.PC9  //
	P1_36 = allwinner.PC6  //
	P1_37 = allwinner.PC16 //
	P1_38 = allwinner.PC10 //
	P1_39 = pin.GROUND     //
	P1_40 = allwinner.PC11 //

	EULER_1  = pin.V3_3          //
	EULER_2  = pin.DC_IN         //
	EULER_3  = pin.BAT_PLUS      //
	EULER_4  = pin.DC_IN         //
	EULER_5  = TEMP_SENSOR       //
	EULER_6  = pin.GROUND        //
	EULER_7  = IR_RX             //
	EULER_8  = pin.V5            //
	EULER_9  = pin.GROUND        //
	EULER_10 = allwinner.PH8     //
	EULER_11 = allwinner.PB3     //
	EULER_12 = allwinner.PB4     //
	EULER_13 = allwinner.PB5     //
	EULER_14 = pin.GROUND        //
	EULER_15 = allwinner.PB6     //
	EULER_16 = allwinner.PB7     //
	EULER_17 = pin.V3_3          //
	EULER_18 = allwinner.PD4     //
	EULER_19 = allwinner.PD2     //
	EULER_20 = pin.GROUND        //
	EULER_21 = allwinner.PD3     //
	EULER_22 = allwinner.PD5     //
	EULER_23 = allwinner.PD1     //
	EULER_24 = allwinner.PD0     //
	EULER_25 = pin.GROUND        //
	EULER_26 = allwinner.PD6     //
	EULER_27 = allwinner.PB2     //
	EULER_28 = allwinner.PD7     //
	EULER_29 = allwinner.PB8     //
	EULER_30 = allwinner.PB9     //
	EULER_31 = allwinner.EAROUTP //
	EULER_32 = allwinner.EAROUTN //
	EULER_33 = pin.INVALID       //
	EULER_34 = pin.GROUND        //

	EXP_1  = pin.V3_3          //
	EXP_2  = allwinner.PL7     //
	EXP_3  = CHARGER_LED       //
	EXP_4  = RESET             //
	EXP_5  = PWR_SWITCH        //
	EXP_6  = pin.GROUND        //
	EXP_7  = allwinner.PB8     //
	EXP_8  = allwinner.PB9     //
	EXP_9  = pin.GROUND        //
	EXP_10 = allwinner.KEY_ADC //

	WIFI_BT_1  = pin.GROUND         //
	WIFI_BT_2  = allwinner.PG6      //
	WIFI_BT_3  = allwinner.PG0      //
	WIFI_BT_4  = allwinner.PG7      //
	WIFI_BT_5  = pin.GROUND         //
	WIFI_BT_6  = allwinner.PG8      //
	WIFI_BT_7  = allwinner.PG1      //
	WIFI_BT_8  = allwinner.PG9      //
	WIFI_BT_9  = allwinner.PG2      //
	WIFI_BT_10 = allwinner.PG10     //
	WIFI_BT_11 = allwinner.PG3      //
	WIFI_BT_12 = allwinner.PG11     //
	WIFI_BT_13 = allwinner.PG4      //
	WIFI_BT_14 = allwinner.PG12     //
	WIFI_BT_15 = allwinner.PG5      //
	WIFI_BT_16 = allwinner.PG13     //
	WIFI_BT_17 = allwinner.PL2      //
	WIFI_BT_18 = pin.GROUND         //
	WIFI_BT_19 = allwinner.PL3      //
	WIFI_BT_20 = allwinner.PL5      //
	WIFI_BT_21 = allwinner.X32KFOUT //
	WIFI_BT_22 = allwinner.PL5      //
	WIFI_BT_23 = pin.GROUND         //
	WIFI_BT_24 = allwinner.PL6      //
	WIFI_BT_25 = VCC                //
	WIFI_BT_26 = IOVCC              //

	AUDIO_LEFT  = pin.INVALID // BUG(maruel): Fix once analog is implemented.
	AUDIO_RIGHT = pin.INVALID //
)

//

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "pine64"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	return []string{"allwinner-gpio", "allwinner-gpio-pl"}
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("pine64 board not detected")
	}
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
	if err := pinreg.Register("EULER", [][]pin.Pin{
		{EULER_1, EULER_2},
		{EULER_3, EULER_4},
		{EULER_5, EULER_6},
		{EULER_7, EULER_8},
		{EULER_9, EULER_10},
		{EULER_11, EULER_12},
		{EULER_13, EULER_14},
		{EULER_15, EULER_16},
		{EULER_17, EULER_18},
		{EULER_19, EULER_20},
		{EULER_21, EULER_22},
		{EULER_23, EULER_24},
		{EULER_25, EULER_26},
		{EULER_27, EULER_28},
		{EULER_29, EULER_30},
		{EULER_31, EULER_32},
		{EULER_33, EULER_34},
	}); err != nil {
		return true, err
	}

	if err := pinreg.Register("EXP", [][]pin.Pin{
		{EXP_1, EXP_2},
		{EXP_3, EXP_4},
		{EXP_5, EXP_6},
		{EXP_7, EXP_8},
		{EXP_9, EXP_10},
	}); err != nil {
		return true, err
	}

	if err := pinreg.Register("WIFI_BT", [][]pin.Pin{
		{WIFI_BT_1, WIFI_BT_2},
		{WIFI_BT_3, WIFI_BT_4},
		{WIFI_BT_5, WIFI_BT_6},
		{WIFI_BT_7, WIFI_BT_8},
		{WIFI_BT_9, WIFI_BT_10},
		{WIFI_BT_11, WIFI_BT_12},
		{WIFI_BT_13, WIFI_BT_14},
		{WIFI_BT_15, WIFI_BT_16},
		{WIFI_BT_17, WIFI_BT_18},
		{WIFI_BT_19, WIFI_BT_20},
		{WIFI_BT_21, WIFI_BT_22},
		{WIFI_BT_23, WIFI_BT_24},
		{WIFI_BT_25, WIFI_BT_26},
	}); err != nil {
		return true, err
	}

	if err := pinreg.Register("AUDIO", [][]pin.Pin{
		{AUDIO_LEFT},
		{AUDIO_RIGHT},
	}); err != nil {
		return true, err
	}

	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&drv)
	}
}

var drv driver
