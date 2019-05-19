// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ccs811

import (
	"fmt"
	"testing"

	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
)

func TestBasicInitialisationAndDataRead(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5A, W: []byte{0xf4}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg, 0x1C}, R: nil},
			{Addr: 0x5A, W: []byte{algoResultsReg}, R: []byte{0x1, 0x2, 0x2, 0x3, 0xF, 0x8, 0xF, 0xF}},
		},
		DontPanic: true,
	}

	opts := DefaultOpts
	opts.InterruptWhenReady = true
	opts.UseThreshold = true

	dev, err := New(&bus, &opts)
	if err != nil {
		t.Fatal(err)
	}

	data := &SensorValues{}
	if err := dev.Sense(data); err == nil {
		var vExpected physic.ElectricPotential
		vExpected.Set("1.65V") // 682 units
		var cExpected physic.ElectricCurrent
		cExpected.Set("63uA")
		if data.ECO2 != 0x102 &&
			data.VOC != 0x203 &&
			data.Status != 0xF &&
			data.Error != fmt.Errorf("sensor error: %s", "HEATER_FAULT: The Heater current in the CCS811 is not in range.") &&
			data.RawDataCurrent != cExpected &&
			data.RawDataVoltage != vExpected {
			t.Fatalf("Data parsed incorrectly, got %v", data)
		}
	} else {
		t.Fatal(err)
	}
}

func TestMeasurementModeRegisterRead(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5A, W: []byte{0xf4}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg, 0x4C}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg}, R: []byte{0x4C}},
		},
		DontPanic: true,
	}

	opts := DefaultOpts
	opts.MeasurementMode = MeasurementModeConstant250
	opts.InterruptWhenReady = true
	opts.UseThreshold = true

	dev, err := New(&bus, &opts)
	if err != nil {
		t.Fatal(err)
	}
	mode, err := dev.GetMeasurementModeRegister()
	if err != nil ||
		mode.GenerateInterrupt != true ||
		mode.UseThreshold != true ||
		mode.MeasurementMode != MeasurementModeConstant250 {
		t.Fatalf("Parsing of Measurement Mode register failed. Got: %+v", mode)
	}

}
func TestGetFirmwareData(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5A, W: []byte{0xf4}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg, 0x4C}, R: nil},
			{Addr: 0x5A, W: []byte{0x20}, R: []byte{0x81}},
			{Addr: 0x5A, W: []byte{0x21}, R: []byte{0x15}},
			{Addr: 0x5A, W: []byte{0x23}, R: []byte{0x12, 0x03}},
			{Addr: 0x5A, W: []byte{0x24}, R: []byte{0x89, 0x20}},
		},
		DontPanic: true,
	}

	opts := DefaultOpts
	opts.MeasurementMode = MeasurementModeConstant250
	opts.InterruptWhenReady = true
	opts.UseThreshold = true

	dev, err := New(&bus, &opts)
	if err != nil {
		t.Fatal(err)
	}
	versions, err := dev.GetFirmwareData()
	if err != nil ||
		versions.HWIdentifier != 0x81 ||
		versions.HWVersion != 0x15 ||
		versions.BootVersion != "1.2.3" ||
		versions.ApplicationVersion != "8.9.32" {
		t.Fatalf("Parsing of firmware version data failed. Got: %+v", versions)
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

func TestReadRawData(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5A, W: []byte{0xf4}, R: nil},
			{Addr: 0x5A, W: []byte{measurementModeReg, 0x10}, R: nil},
			{Addr: 0x5A, W: []byte{rawDataReg}, R: []byte{0x96, 0xAA}},
		},
	}
	dev, err := New(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	cur, vol, err := dev.ReadRawData()
	if err != nil {
		t.Fatal(err)
	}
	var vExpected physic.ElectricPotential
	vExpected.Set("1.1V") // 682 units
	var cExpected physic.ElectricCurrent
	cExpected.Set("37uA")

	if cur != cExpected || vol != vExpected {
		t.Fatalf("Raw data reading failed got values: %d, %d", cur, vol)
	}
}

func TestRawDataParsing(t *testing.T) {
	var vExpected physic.ElectricPotential
	vExpected.Set("0.825806451V") // 512 units
	var cExpected physic.ElectricCurrent
	cExpected.Set("62uA")
	c, v := valuesFromRawData([]byte{0xFA, 0x0})
	if c != cExpected && v != vExpected {
		t.Fatal("current and/or voltage data parsed incorrectly")
	}
}

func TestReset(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x5B, W: []byte{0xf4}, R: nil},
			{Addr: 0x5B, W: []byte{measurementModeReg, 0x10}, R: nil},
			{Addr: 0x5B, W: []byte{resetReg, 0x11, 0xE5, 0x72, 0x8A}, R: nil},
		},
	}
	opts := &DefaultOpts
	opts.Addr = 0x5B
	dev, err := New(&bus, opts)
	if err != nil {
		t.Fatal(err)
	}
	dev.Reset()
}
