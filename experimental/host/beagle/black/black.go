// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package black

import (
	"errors"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/distro"
	"periph.io/x/periph/host/sysfs"
)

var (
	P8_1  pin.Pin = pin.GROUND
	P8_2  pin.Pin = pin.GROUND
	P8_3  pin.Pin = sysfs.Pins[38]
	P8_4  pin.Pin = sysfs.Pins[39]
	P8_5  pin.Pin = sysfs.Pins[34]
	P8_6  pin.Pin = sysfs.Pins[35]
	P8_7  pin.Pin = sysfs.Pins[66]
	P8_8  pin.Pin = sysfs.Pins[67]
	P8_9  pin.Pin = sysfs.Pins[69]
	P8_10 pin.Pin = sysfs.Pins[68]
	P8_11 pin.Pin = sysfs.Pins[45]
	P8_12 pin.Pin = sysfs.Pins[44]
	P8_13 pin.Pin = sysfs.Pins[23]
	P8_14 pin.Pin = sysfs.Pins[26]
	P8_15 pin.Pin = sysfs.Pins[47]
	P8_16 pin.Pin = sysfs.Pins[46]
	P8_17 pin.Pin = sysfs.Pins[27]
	P8_18 pin.Pin = sysfs.Pins[65]
	P8_19 pin.Pin = sysfs.Pins[22]
	P8_20 pin.Pin = sysfs.Pins[63]
	P8_21 pin.Pin = sysfs.Pins[62]
	P8_22 pin.Pin = sysfs.Pins[37]
	P8_23 pin.Pin = sysfs.Pins[36]
	P8_24 pin.Pin = sysfs.Pins[33]
	P8_25 pin.Pin = sysfs.Pins[32]
	P8_26 pin.Pin = sysfs.Pins[61]
	P8_27 pin.Pin = sysfs.Pins[86]
	P8_28 pin.Pin = sysfs.Pins[88]
	P8_29 pin.Pin = sysfs.Pins[87]
	P8_30 pin.Pin = sysfs.Pins[89]
	P8_31 pin.Pin = sysfs.Pins[10]
	P8_32 pin.Pin = sysfs.Pins[11]
	P8_33 pin.Pin = sysfs.Pins[9]
	P8_34 pin.Pin = sysfs.Pins[81]
	P8_35 pin.Pin = sysfs.Pins[8]
	P8_36 pin.Pin = sysfs.Pins[80]
	P8_37 pin.Pin = sysfs.Pins[78]
	P8_38 pin.Pin = sysfs.Pins[79]
	P8_39 pin.Pin = sysfs.Pins[76]
	P8_40 pin.Pin = sysfs.Pins[77]
	P8_41 pin.Pin = sysfs.Pins[74]
	P8_42 pin.Pin = sysfs.Pins[75]
	P8_43 pin.Pin = sysfs.Pins[72]
	P8_44 pin.Pin = sysfs.Pins[73]
	P8_45 pin.Pin = sysfs.Pins[70]
	P8_46 pin.Pin = sysfs.Pins[71]

	P9_1  pin.Pin = pin.GROUND
	P9_2  pin.Pin = pin.GROUND
	P9_3  pin.Pin = pin.V3_3
	P9_4  pin.Pin = pin.V3_3
	P9_5  pin.Pin = pin.V5
	P9_6  pin.Pin = pin.V5
	P9_7  pin.Pin = pin.V5
	P9_8  pin.Pin = pin.V5
	P9_11 pin.Pin = sysfs.Pins[30]
	P9_12 pin.Pin = sysfs.Pins[60]
	P9_13 pin.Pin = sysfs.Pins[31]
	P9_14 pin.Pin = sysfs.Pins[50]
	P9_15 pin.Pin = sysfs.Pins[48]
	P9_16 pin.Pin = sysfs.Pins[51]
	P9_17 pin.Pin = sysfs.Pins[5]
	P9_18 pin.Pin = sysfs.Pins[4]
	P9_21 pin.Pin = sysfs.Pins[3]
	P9_22 pin.Pin = sysfs.Pins[2]
	P9_23 pin.Pin = sysfs.Pins[49]
	P9_24 pin.Pin = sysfs.Pins[15]
	P9_25 pin.Pin = sysfs.Pins[117]
	P9_26 pin.Pin = sysfs.Pins[14]
	P9_27 pin.Pin = sysfs.Pins[115]
	P9_28 pin.Pin = sysfs.Pins[113]
	P9_29 pin.Pin = sysfs.Pins[111]
	P9_30 pin.Pin = sysfs.Pins[112]
	P9_31 pin.Pin = sysfs.Pins[110]
	P9_41 pin.Pin = sysfs.Pins[20]
	P9_42 pin.Pin = sysfs.Pins[7]
)

// Present returns true if there is evidence that we have a
// Beagleboard present.
func Present() bool {
	if isArm {
		return distro.DTModel() == "TI AM335x BeagleBone Black"
	}
	return false
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "black"
}

func (d *driver) Prerequisites() []string {
	return []string{"beagle"}
}

func (d *driver) After() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	if !Present() {
		return false, errors.New("BeagleBone Black board not detected")
	}

	err := pinreg.Register("P8", [][]pin.Pin{
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
	})
	if err != nil {
		return true, err
	}

	err = pinreg.Register("P9", [][]pin.Pin{
		{P9_11, P9_12},
		{P9_13, P9_14},
		{P9_15, P9_16},
		{P9_17, P9_18},
		{P9_21, P9_22},
		{P9_23, P9_24},
		{P9_25, P9_26},
		{P9_27, P9_28},
		{P9_29, P9_30},
		{P9_41, P9_42},
	})
	if err != nil {
		return true, err
	}

	return true, nil
}

func init() {
	if isArm {
		periph.MustRegister(&driver{})
	}
}

var _ periph.Driver = &driver{}
