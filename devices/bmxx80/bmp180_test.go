// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bmxx80

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
)

var opts180 = &Opts{Temperature: O1x, Pressure: O1x}

func TestNew180_fail_read_chipid(t *testing.T) {
	bus := i2ctest.Playback{
		// Chip ID detection read fail.
		Ops:       []i2ctest.IO{},
		DontPanic: true,
	}
	if _, err := NewI2C(&bus, 0x77, opts180); err == nil {
		t.Fatal("can't read chip ID")
	}
}

func TestNew180_bad_chipid(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Bad Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x61}},
		},
	}
	if _, err := NewI2C(&bus, 0x77, opts180); err == nil {
		t.Fatal("bad chip ID")
	}
}

func TestNew180_fail_calib(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
			// Fail to read calibration data.
		},
		DontPanic: true,
	}
	if _, err := NewI2C(&bus, 0x77, opts180); err == nil {
		t.Fatal("can't read calibration")
	}
}

func TestNew180_bad_calib(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
			// Calibration data.
			{
				Addr: 0x77,
				W:    []byte{0xaa},
				R:    []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			},
		},
		DontPanic: true,
	}
	if _, err := NewI2C(&bus, 0x77, opts180); err == nil {
		t.Fatal("bad calibration")
	}
}

func TestSense180_success(t *testing.T) {
	values := []struct {
		o Oversampling
		c byte
		p physic.Pressure
	}{
		{Oversampling(42), 0x34, 100567 * physic.Pascal},
		{O1x, 0x34, 100567 * physic.Pascal},
		{O2x, 0x74, 100567 * physic.Pascal},
		{O4x, 0xB4, 100568 * physic.Pascal},
		{O8x, 0xF4, 100568 * physic.Pascal},
	}
	for _, line := range values {
		bus := i2ctest.Playback{
			Ops: []i2ctest.IO{
				// Chip ID detection.
				{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
				// Calibration data.
				{
					Addr: 0x77,
					W:    []byte{0xaa},
					R:    []byte{35, 136, 251, 103, 199, 169, 135, 91, 98, 137, 80, 22, 25, 115, 0, 46, 128, 0, 209, 246, 10, 123},
				},
				// Request temperature.
				{Addr: 0x77, W: []byte{0xF4, 0x2E}},
				// Read temperature.
				{Addr: 0x77, W: []byte{0xF6}, R: []byte{0x71, 0xBf}},
				// Request pressure.
				{Addr: 0x77, W: []byte{0xF4, line.c}},
				// Read pressure.
				{Addr: 0x77, W: []byte{0xF6}, R: []byte{0xAb, 0x96, 0}},
			},
		}
		dev, err := NewI2C(&bus, 0x77, &Opts{Pressure: line.o})
		if err != nil {
			t.Fatal(err)
		}
		if s := dev.String(); s != "BMP180{playback(119)}" {
			t.Fatal(s)
		}
		e := physic.Env{}
		if err := dev.Sense(&e); err != nil {
			t.Fatal(err)
		}
		if e.Temperature != 25300*physic.MilliCelsius+physic.ZeroCelsius {
			t.Fatalf("temp %d", e.Temperature)
		}
		if e.Pressure != line.p {
			t.Fatalf("pressure %d", e.Pressure)
		}
		if e.Humidity != 0 {
			t.Fatalf("humidity %d", e.Humidity)
		}
		if err := dev.Halt(); err != nil {
			t.Fatal(err)
		}
		if err := bus.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSense180_fail_1(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
			// Calibration data.
			{
				Addr: 0x77,
				W:    []byte{0xaa},
				R:    []byte{35, 136, 251, 103, 199, 169, 135, 91, 98, 137, 80, 22, 25, 115, 0, 46, 128, 0, 209, 246, 10, 123},
			},
			// Request temperature fail.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x77, opts180)
	if err != nil {
		t.Fatal(err)
	}
	e := physic.Env{}
	if dev.Sense(&e) == nil {
		t.Fatal("sensing should have failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSense180_fail_2(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
			// Calibration data.
			{
				Addr: 0x77,
				W:    []byte{0xaa},
				R:    []byte{35, 136, 251, 103, 199, 169, 135, 91, 98, 137, 80, 22, 25, 115, 0, 46, 128, 0, 209, 246, 10, 123},
			},
			// Request temperature.
			{Addr: 0x77, W: []byte{0xF4, 0x2E}},
			// Read temperature fail.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x77, opts180)
	if err != nil {
		t.Fatal(err)
	}
	e := physic.Env{}
	if dev.Sense(&e) == nil {
		t.Fatal("sensing should have failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSense180_fail_3(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
			// Calibration data.
			{
				Addr: 0x77,
				W:    []byte{0xaa},
				R:    []byte{35, 136, 251, 103, 199, 169, 135, 91, 98, 137, 80, 22, 25, 115, 0, 46, 128, 0, 209, 246, 10, 123},
			},
			// Request temperature.
			{Addr: 0x77, W: []byte{0xF4, 0x2E}},
			// Read temperature fail.
			{Addr: 0x77, W: []byte{0xF6}, R: []byte{0x71, 0xBf}},
			// Request pressure fail.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x77, opts180)
	if err != nil {
		t.Fatal(err)
	}
	e := physic.Env{}
	if dev.Sense(&e) == nil {
		t.Fatal("sensing should have failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSense180_fail_4(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
			// Calibration data.
			{
				Addr: 0x77,
				W:    []byte{0xaa},
				R:    []byte{35, 136, 251, 103, 199, 169, 135, 91, 98, 137, 80, 22, 25, 115, 0, 46, 128, 0, 209, 246, 10, 123},
			},
			// Request temperature.
			{Addr: 0x77, W: []byte{0xF4, 0x2E}},
			// Read temperature fail.
			{Addr: 0x77, W: []byte{0xF6}, R: []byte{0x71, 0xBf}},
			// Request pressure.
			{Addr: 0x77, W: []byte{0xF4, 0x34}},
			// Read pressure fail.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x77, opts180)
	if err != nil {
		t.Fatal(err)
	}
	e := physic.Env{}
	if dev.Sense(&e) == nil {
		t.Fatal("sensing should have failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSenseContinuous180_success(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x77, W: []byte{0xd0}, R: []byte{0x55}},
			// Calibration data.
			{
				Addr: 0x77,
				W:    []byte{0xaa},
				R:    []byte{35, 136, 251, 103, 199, 169, 135, 91, 98, 137, 80, 22, 25, 115, 0, 46, 128, 0, 209, 246, 10, 123},
			},
			// Request temperature.
			{Addr: 0x77, W: []byte{0xF4, 0x2E}},
			// Read temperature.
			{Addr: 0x77, W: []byte{0xF6}, R: []byte{0x71, 0xBf}},
			// Request pressure.
			{Addr: 0x77, W: []byte{0xF4, 0x34}},
			// Read pressure.
			{Addr: 0x77, W: []byte{0xF6}, R: []byte{0xAb, 0x96, 0}},
			// Request temperature.
			{Addr: 0x77, W: []byte{0xF4, 0x2E}},
			// Read temperature.
			{Addr: 0x77, W: []byte{0xF6}, R: []byte{0x71, 0xBf}},
			// Request pressure.
			{Addr: 0x77, W: []byte{0xF4, 0x34}},
			// Read pressure.
			{Addr: 0x77, W: []byte{0xF6}, R: []byte{0xAb, 0x96, 0}},
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x77, opts180)
	if err != nil {
		t.Fatal(err)
	}
	c, err := dev.SenseContinuous(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	e := <-c
	if e.Temperature != 25300*physic.MilliCelsius+physic.ZeroCelsius {
		t.Fatalf("temp %d", e.Temperature)
	}
	if e.Pressure != 100567*physic.Pascal {
		t.Fatalf("pressure %d", e.Pressure)
	}
	if e.Humidity != 0 {
		t.Fatalf("humidity %d", e.Humidity)
	}

	if dev.Sense(&e) == nil {
		t.Fatal("Sense() should have failed")
	}

	c2, err := dev.SenseContinuous(time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	e = <-c2

	if _, ok := <-c; ok {
		t.Fatal("c should be closed")
	}

	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	if _, ok := <-c2; ok {
		t.Fatal("c2 should be closed")
	}

	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

/*
func TestOversampling(t *testing.T) {
	data := []struct {
		o Oversampling
		v int
		s string
	}{
		{O1x, 1, "1x"},
		{O2x, 2, "2x"},
		{O4x, 4, "4x"},
		{O8x, 8, "8x"},
		{Oversampling(100), 0, "Oversampling(100)"},
	}
	for i, line := range data {
		if s := line.o.String(); s != line.s {
			t.Fatalf("#%d %s != %s", i, s, line.s)
		}
	}
}
*/

func TestCompensate180(t *testing.T) {
	c := calibration180{
		AC1: 408,
		AC2: -72,
		AC3: -14383,
		AC4: 32741,
		AC5: 32757,
		AC6: 23153,
		B1:  6190,
		B2:  4,
		MB:  -32768,
		MC:  -8711,
		MD:  2868,
	}

	if temp := c.compensateTemp(27898); temp != 150 {
		t.Errorf("temperature is wrong, want %v, got %v", 150, temp)
	}

	if pressure := c.compensatePressure(23843, 27898, 0); pressure != 69964 {
		t.Errorf("pressure is wrong, want %v, got %v", 69964, pressure)
	}
}

func init() {
	doSleep = func(time.Duration) {}
}
