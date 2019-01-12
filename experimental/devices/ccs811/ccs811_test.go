// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ccs811

import (
	"testing"

	"periph.io/x/periph/conn/i2c/i2ctest"
)

func TestBasicInitialisationAndDataRead(t *testing.T) {
	// FIXME: Try to include basic code coverage. You can use "replay" tests by
	// leveraging i2ctest and spitest.
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5A, W: []byte{0xf4}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg, 0x10}, R: nil},
			{Addr: 0x5A, W: []byte{algoResultsReg}, R: []byte{0x1, 0x2, 0x2, 0x3}},
		},
		DontPanic: true,
	}
	dev, err := New(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	if data, err := dev.Sense(ReadCO2VOC); err == nil {
		if data.ECO2 != 0x102 && data.VOC != 0x203 {
			t.Fatal("eCO2 and/or VOC data parsed incorrectly")
		}
	} else {
		t.Fatal(err)
	}
}

func TestInvalidSensorAddress(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x0, W: nil, R: nil},
		},
	}
	if dev, err := New(&bus, &Opts{Addr: 0xFF, MeasurementMode: MeasurementModeConstant1000}); dev != nil || err == nil {
		t.Fatal("New should have failed")
	}
}

func TestSetEnvironmentData(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5A, W: []byte{0xf4}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg, 0x10}, R: nil},
			{Addr: 0x5A, W: []byte{environmentReg, 0x61, 0x00, 0x64, 0x00}, R: nil},
			{Addr: 0x5A, W: []byte{environmentReg, 0x64, 0x00, 0x61, 0x00}, R: nil},
		},
	}
	dev, err := New(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	dev.SetEnvironmentData(25, 48.5)
	dev.SetEnvironmentData(23.5, 50)
}

func TestBaseline(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5A, W: []byte{0xf4}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg, 0x10}, R: nil},
			{Addr: 0x5A, W: []byte{baselineReg}, R: []byte{0xAA, 0xDD}},
			{Addr: 0x5A, W: []byte{baselineReg, 0xAA, 0xDD}, R: nil},
		},
	}
	dev, err := New(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	base, err := dev.GetBaseline()
	if err != nil {
		t.Fatal(err)
	}
	dev.SetBaseline(base)
}

func TestRawDataParsing(t *testing.T) {
	c, v := valuesFromRawData([]byte{0xF9, 0x0})
	if c != 62 && v != 512 {
		t.Fatal("current and/or voltage data parsed incorrectly")
	}
}
