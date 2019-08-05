// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ads1x15

import (
	"reflect"
	"testing"

	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/experimental/conn/analog"
)

func TestChannel_String(t *testing.T) {
	// Mainly to increase test coverage...
	data := []struct {
		c        Channel
		expected string
	}{
		{Channel0, "0"},
		{Channel1, "1"},
		{Channel2, "2"},
		{Channel3, "3"},
		{Channel0Minus1, "0-1"},
		{Channel0Minus3, "0-3"},
		{Channel1Minus3, "1-3"},
		{Channel2Minus3, "2-3"},
		{Channel(-1), "Invalid"},
	}
	for _, line := range data {
		if actual := line.c.String(); actual != line.expected {
			t.Fatalf("%s != %s", line.expected, actual)
		}
	}
}

func TestChannel_number(t *testing.T) {
	// Mainly to increase test coverage...
	data := []struct {
		c        Channel
		expected int
	}{
		{Channel0, 0},
		{Channel1, 1},
		{Channel2, 2},
		{Channel3, 3},
		{Channel0Minus1, 4},
		{Channel0Minus3, 5},
		{Channel1Minus3, 6},
		{Channel2Minus3, 7},
		{Channel(-1), -1},
	}
	for _, line := range data {
		if actual := line.c.number(); actual != line.expected {
			t.Fatalf("%d != %d", line.expected, actual)
		}
	}
}

func TestDev_String(t *testing.T) {
	b := i2ctest.Playback{}
	defer b.Close()
	d, err := NewADS1115(&b, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "ADS1115" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

func TestPinADC(t *testing.T) {
	b := i2ctest.Playback{}
	defer b.Close()
	d, err := NewADS1015(&b, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	// Obtain an analog pin from the ADC
	p, err := d.PinForChannel(Channel0Minus3, 5*physic.Volt, 1*physic.Hertz, BestQuality)
	if err != nil {
		t.Fatal(err)
	}
	if v := p.String(); v != "ADS1015(0-3)" {
		t.Fatal(v)
	}
	if v := p.Name(); v != "ADS1015(0-3)" {
		t.Fatal(v)
	}
	if v := p.Number(); v != 5 {
		t.Fatal(v)
	}
	if v := p.Function(); v != "ADC" {
		t.Fatal(v)
	}
	if v := p.(pin.PinFunc).Func(); v != analog.ADC {
		t.Fatal(v)
	}
	if v := p.(pin.PinFunc).SupportedFuncs(); !reflect.DeepEqual(v, []pin.Func{analog.ADC}) {
		t.Fatal(v)
	}
	if err := p.(pin.PinFunc).SetFunc(analog.ADC); err != nil {
		t.Fatal(err)
	}
	if err := p.(pin.PinFunc).SetFunc(pin.FuncNone); err == nil {
		t.Fatal("expected failure")
	}
}

func TestPinADC_Read(t *testing.T) {
	b := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x48,
				W:    []byte{0x1, 0x91, 0x3},
				R:    []byte{},
			},
			{
				Addr: 0x48,
				W:    []byte{0x0},
				R:    []byte{0xff, 0x50},
			},
		},
	}
	defer b.Close()

	d, err := NewADS1015(&b, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	// Obtain an analog pin from the ADC
	p, err := d.PinForChannel(Channel0Minus3, 5*physic.Volt, 1*physic.Hertz, BestQuality)
	if err != nil {
		t.Fatal(err)
	}

	// Read values from ADC.
	reading, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}

	if reading.Raw != -176 {
		t.Fatalf("Found %d, expected %d", reading.Raw, -176)
	}

	if reading.V != -33*physic.MilliVolt {
		t.Fatalf("Found %s, expected %s", reading.V, -33*physic.MilliVolt)
	}

	if err := p.Halt(); err != nil {
		t.Fatal(err)
	}

	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

func TestPinADC_ReadContinous(t *testing.T) {
	b := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{
				Addr: 0x48,
				W:    []byte{0x1, 0x91, 0xc3},
				R:    []byte{},
			},
			{
				Addr: 0x48,
				W:    []byte{0x0},
				R:    []byte{0x52, 0xd0},
			},
			{
				Addr: 0x48,
				W:    []byte{0x1, 0x91, 0xc3},
				R:    []byte{},
			},
			{
				Addr: 0x48,
				W:    []byte{0x0},
				R:    []byte{0x52, 0xc0},
			},
			// We add 2 extra exchanges, as halting the polling is not instant
			{
				Addr: 0x48,
				W:    []byte{0x1, 0x91, 0xc3},
				R:    []byte{},
			},
			{
				Addr: 0x48,
				W:    []byte{0x0},
				R:    []byte{0x52, 0xc0},
			},
			{
				Addr: 0x48,
				W:    []byte{0x1, 0x91, 0xc3},
				R:    []byte{},
			},
			{
				Addr: 0x48,
				W:    []byte{0x0},
				R:    []byte{0x52, 0xc0},
			},
		},
		DontPanic: true,
	}
	defer b.Close()

	rawValues := []int32{21200, 21184}
	voltValues := []physic.ElectricPotential{3975 * physic.MilliVolt, 3972 * physic.MilliVolt}

	d, err := NewADS1015(&b, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	// Obtain an analog pin from the ADC
	p, err := d.PinForChannel(Channel0Minus3, 5*physic.Volt, 100*physic.Hertz, SaveEnergy)
	if err != nil {
		t.Fatal(err)
	}

	// Read values from ADC.
	c := p.ReadContinuous()

	var i = 0
	for reading := range c {
		if reading.Raw != rawValues[i] {
			t.Fatalf("Found %d, expected %d", reading.Raw, rawValues[i])
		}

		if reading.V != voltValues[i] {
			t.Fatalf("Found %s, expected %s", reading.V, voltValues[i])
		}

		i++
		if i >= len(rawValues) {
			break
		}
	}

	if err := p.Halt(); err != nil {
		t.Fatal(err)
	}

	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}
