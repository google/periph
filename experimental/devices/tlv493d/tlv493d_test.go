// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package tlv493d

import (
	"testing"

	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
)

func TestDev_String(t *testing.T) {
	b := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Recovery
			{
				Addr: 0x5e,
				W:    []byte{0xff},
				R:    []byte{},
			},
			// Reset
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{},
			},
			// Read configuration
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{0xfd, 0x2d, 0x79, 0x14, 0xab, 0x22, 0x51, 0x81, 0x4, 0x60},
			},
			// Configure
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x80, 0x4, 0x20},
				R:    []byte{},
			},
			// Halt: power down
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x80, 0x4, 0x20},
				R:    []byte{},
			},
		},
	}
	defer b.Close()

	d, err := New(&b, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "TLV493D" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

func TestTLV493D_Read(t *testing.T) {
	b := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Recovery
			{
				Addr: 0x5e,
				W:    []byte{0xff},
				R:    []byte{},
			},
			// Reset
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{},
			},
			// Read configuration
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{0xfd, 0x2d, 0x79, 0x14, 0xab, 0x22, 0x51, 0x81, 0x4, 0x60},
			},
			// Configure
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x81, 0x4, 0x60},
				R:    []byte{},
			},
			// Read measurements
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{0xfd, 0x2d, 0x79, 0x18, 0xbb, 0x31, 0x51},
			},
			// Halt: power down
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x80, 0x4, 0x20},
				R:    []byte{},
			},
		},
	}
	defer b.Close()

	opts := DefaultOpts
	opts.Mode = LowPowerMode

	d, err := New(&b, &opts)
	if err != nil {
		t.Fatal(err)
	}

	// Read values from ADC.
	reading, err := d.Read(HighPrecisionWithTemperature)
	if err != nil {
		t.Fatal(err)
	}

	assertSample(t, Sample{
		Bx:          -3626 * physic.MicroTesla,
		By:          71638 * physic.MicroTesla,
		Bz:          189826 * physic.MicroTesla,
		Temperature: 294850 * physic.MilliKelvin,
	}, reading)

	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

func TestTLV493D_ReadContinous(t *testing.T) {
	b := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Recovery
			{
				Addr: 0x5e,
				W:    []byte{0xff},
				R:    []byte{},
			},
			// Reset
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{},
			},
			// Read configuration
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{0xfd, 0x2d, 0x79, 0x14, 0xab, 0x22, 0x51, 0x81, 0x4, 0x60},
			},
			// Configure
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x80, 0x4, 0x20},
				R:    []byte{},
			},
			// Configure for continuous mode
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x81, 0x4, 0x60},
				R:    []byte{},
			},
			// Read measurements
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{0xfd, 0x2d, 0x79},
			},
			// Read measurements
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{0xf3, 0xd5, 0xa},
			},
			// End of continuous reading, restore previous mode
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x80, 0x4, 0x20},
				R:    []byte{},
			},
			// Halt: power down
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x80, 0x4, 0x20},
				R:    []byte{},
			},
		},
	}
	defer b.Close()

	samples := []Sample{
		{
			Bx: -4704 * physic.MicroTesla,
			By: 70560 * physic.MicroTesla,
			Bz: 189728 * physic.MicroTesla,
		},
		{
			Bx: -20384 * physic.MicroTesla,
			By: -67424 * physic.MicroTesla,
			Bz: 15680 * physic.MicroTesla,
		},
	}

	d, err := New(&b, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}

	// Read values from sensor.
	c, err := d.ReadContinuous(100*physic.Hertz, LowPrecision)
	if err != nil {
		t.Fatal(err)
	}

	var i = 0
	for reading := range c {
		assertSample(t, samples[i], reading)

		i++
		if i >= len(samples) {
			break
		}
	}

	d.StopContinousRead()

	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

func TestTLV493D_Configuration(t *testing.T) {
	b := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Recovery
			{
				Addr: 0x5e,
				W:    []byte{0xff},
				R:    []byte{},
			},
			// Reset
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{},
			},
			// Read configuration
			{
				Addr: 0x5e,
				W:    []byte{0x0},
				R:    []byte{0xfd, 0x2d, 0x79, 0x14, 0xab, 0x22, 0x51, 0x81, 0x4, 0x60},
			},
			// Configure
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x80, 0x4, 0x20},
				R:    []byte{},
			},
			// Configure for UltraLowPowerMode
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x1, 0x4, 0x20},
				R:    []byte{},
			},
			// Disable temperature measurement
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x81, 0x4, 0xa0},
				R:    []byte{},
			},
			// Disable parity test
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x1, 0x4, 0x80},
				R:    []byte{},
			},
			// Disable interruptions
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x85, 0x4, 0x80},
				R:    []byte{},
			},
			// Halt: power down
			{
				Addr: 0x5e,
				W:    []byte{0x0, 0x4, 0x4, 0x80},
				R:    []byte{},
			},
		},
	}
	defer b.Close()

	opts := DefaultOpts

	d, err := New(&b, &opts)
	if err != nil {
		t.Fatal(err)
	}

	// Change configuration items
	err = d.SetMode(UltraLowPowerMode)
	if err != nil {
		t.Fatal(err)
	}

	err = d.EnableTemperatureMeasurement(false)
	if err != nil {
		t.Fatal(err)
	}

	err = d.EnableParityTest(false)
	if err != nil {
		t.Fatal(err)
	}

	err = d.EnableInterruptions(true)
	if err != nil {
		t.Fatal(err)
	}

	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

func assertSample(t *testing.T, expected Sample, actual Sample) {
	if actual.Bx != expected.Bx {
		t.Fatalf("Bx: Found %d, expected %d", actual.Bx, expected.Bx)
	}

	if actual.By != expected.By {
		t.Fatalf("By: Found %d, expected %d", actual.By, expected.By)
	}

	if actual.Bz != expected.Bz {
		t.Fatalf("Bz: Found %d, expected %d", actual.Bz, expected.Bz)
	}

	if actual.Temperature != expected.Temperature {
		t.Fatalf("Temperature: Found %d, expected %d", actual.Temperature, expected.Temperature)
	}

}
