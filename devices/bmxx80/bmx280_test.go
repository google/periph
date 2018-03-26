// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bmxx80

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spitest"
)

// Real data extracted from a device.
var calib280 = calibration280{
	t1: 28176,
	t2: 26220,
	t3: 350,
	p1: 38237,
	p2: -10824,
	p3: 3024,
	p4: 7799,
	p5: -99,
	p6: -7,
	p7: 9900,
	p8: -10230,
	p9: 4285,
	h2: 366, // Note they are inversed for bit packing.
	h1: 75,
	h3: 0,
	h4: 309,
	h5: 0,
	h6: 30,
}

func TestSPISenseBME280_success(t *testing.T) {
	s := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				// Chip ID detection.
				{
					W: []byte{0xD0, 0x00},
					R: []byte{0x00, 0x60},
				},
				// Calibration data.
				{
					W: []byte{0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					R: []byte{0x00, 0xC9, 0x6C, 0x63, 0x65, 0x32, 0x00, 0x77, 0x93, 0x98, 0xD5, 0xD0, 0x0B, 0x67, 0x23, 0xBA, 0x00, 0xF9, 0xFF, 0xAC, 0x26, 0x0A, 0xD8, 0xBD, 0x10, 0x00, 0x4B},
				},
				// Calibration data humidity.
				{
					W: []byte{0xE1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					R: []byte{0x00, 0x5C, 0x01, 0x00, 0x15, 0x0F, 0x00, 0x1E},
				},
				// Config.
				{W: []byte{0x74, 0xB4, 0x72, 0x05, 0x75, 0xA0, 0x74, 0xB4}},
				// Forced mode.
				{W: []byte{0x74, 0xB5}},
				// Check if idle.
				{W: []byte{0xF3, 0x00}, R: []byte{0, 0}},
				// Read measurement data.
				{
					W: []byte{0xF7, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					R: []byte{0x00, 0x51, 0x9F, 0xC0, 0x9E, 0x3A, 0x50, 0x5E, 0x5B},
				},
			},
		},
	}
	opts := Opts{
		Temperature: O16x,
		Pressure:    O16x,
		Humidity:    O16x,
	}
	dev, err := NewSPI(&s, &opts)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "BME280{playback}" {
		t.Fatal(s)
	}
	e := physic.Env{}
	if err := dev.Sense(&e); err != nil {
		t.Fatal(err)
	}
	// TODO(maruel): The values do not make sense but I think I burned my SPI
	// BME280 by misconnecting it in reverse for a few minutes. It still "work"
	// but fail to read data. It could also be a bug in the driver. :(
	if e.Temperature != 62680*physic.MilliCelsius+physic.ZeroCelsius {
		t.Fatalf("temp %d", e.Temperature)
	}
	if e.Pressure != 99575933593750*physic.NanoPascal {
		t.Fatalf("pressure %d", e.Pressure)
	}
	if e.Humidity != 9950*physic.MicroRH {
		t.Fatalf("humidity %d", e.Humidity)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewSPIBME280_fail_Connect(t *testing.T) {
	if dev, err := NewSPI(&spiFail{}, nil); dev != nil || err == nil {
		t.Fatal("read failed")
	}
}

func TestNewSPIBME280_fail_len(t *testing.T) {
	s := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				{
					// Chip ID detection.
					W: []byte{0xD0, 0x00},
					R: []byte{0x00},
				},
			},
			DontPanic: true,
		},
	}
	if dev, err := NewSPI(&s, nil); dev != nil || err == nil {
		t.Fatal("read failed")
	}
	// The I/O didn't occur.
	s.Count++
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewSPIBME280_fail_chipid(t *testing.T) {
	s := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				{
					// Chip ID detection.
					W: []byte{0xD0, 0x00},
					R: []byte{0x00, 0xFF},
				},
			},
		},
	}
	if dev, err := NewSPI(&s, nil); dev != nil || err == nil {
		t.Fatal("read failed")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2CBME280_fail_io(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}},
		},
		DontPanic: true,
	}
	if dev, err := NewI2C(&bus, 0x76, nil); dev != nil || err == nil {
		t.Fatal("read failed")
	}
	// The I/O didn't occur.
	bus.Count++
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2CBME280_fail_read_calib1(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
		},
		DontPanic: true,
	}
	if dev, err := NewI2C(&bus, 0x76, nil); dev != nil || err == nil {
		t.Fatal("invalid chip id")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2CBME280_read_calib2(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
		},
		DontPanic: true,
	}
	opts := Opts{Temperature: O1x}
	if dev, err := NewI2C(&bus, 0x76, &opts); dev != nil || err == nil {
		t.Fatal("2nd calib read failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2C280_write_config(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
		},
		DontPanic: true,
	}
	if dev, err := NewI2C(&bus, 0x76, nil); dev != nil || err == nil {
		t.Fatal("3rd calib read failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2C280Opts_temperature(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
		},
	}
	if dev, err := NewI2C(&bus, 0x76, &Opts{}); dev != nil || err == nil {
		t.Fatal("bad addr")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2C280_bad_addr(t *testing.T) {
	bus := i2ctest.Playback{}
	if dev, err := NewI2C(&bus, 1, nil); dev != nil || err == nil {
		t.Fatal("bad addr")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2CSenseBME280_fail(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Forced mode.
			{Addr: 0x76, W: []byte{0xF4, 0xB5}},
			// Check if idle fails.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	if dev.Sense(&physic.Env{}) == nil {
		t.Fatal("sense fail read")
	}
	// The I/O didn't occur.
	bus.Count++
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2CSenseBMP280_success(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x58}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Forced mode.
			{Addr: 0x76, W: []byte{0xF4, 0x6d}},
			// Check if idle; not idle.
			{Addr: 0x76, W: []byte{0xF3}, R: []byte{8}},
			// Check if idle.
			{Addr: 0x76, W: []byte{0xF3}, R: []byte{0}},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}, R: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0}},
		},
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "BMP280{playback(118)}" {
		t.Fatal(s)
	}
	e := physic.Env{}
	if err := dev.Sense(&e); err != nil {
		t.Fatal(err)
	}
	if e.Temperature != 23720*physic.MilliCelsius+physic.ZeroCelsius {
		t.Fatalf("temp %d", e.Temperature)
	}
	if e.Pressure != 100942695312500*physic.NanoPascal {
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

func TestI2CSenseBME280_success(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Forced mode.
			{Addr: 0x76, W: []byte{0xF4, 0x6d}},
			// Check if idle; not idle.
			{Addr: 0x76, W: []byte{0xF3}, R: []byte{8}},
			// Check if idle.
			{Addr: 0x76, W: []byte{0xF3}, R: []byte{0}},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}, R: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
		},
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "BME280{playback(118)}" {
		t.Fatal(s)
	}
	e := physic.Env{}
	if err := dev.Sense(&e); err != nil {
		t.Fatal(err)
	}
	if e.Temperature != 23720*physic.MilliCelsius+physic.ZeroCelsius {
		t.Fatalf("temp %d", e.Temperature)
	}
	if e.Pressure != 100942695312500*physic.NanoPascal {
		t.Fatalf("pressure %d", e.Pressure)
	}
	if e.Humidity != 65305*physic.MicroRH {
		t.Fatalf("humidity %d", e.Humidity)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2CSense280_idle_fail(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Forced mode.
			{Addr: 0x76, W: []byte{0xF4, 0x6d}},
			// Check if idle fails.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "BME280{playback(118)}" {
		t.Fatal(s)
	}
	e := physic.Env{}
	if dev.Sense(&e) == nil {
		t.Fatal("isIdle() should have failed")
	}
}

func TestI2CSense280_command_fail(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Forced mode.
			{Addr: 0x76, W: []byte{0xF4, 0x6d}},
			// Check if idle.
			{Addr: 0x76, W: []byte{0xF3}, R: []byte{0}},
			// Read fail.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "BME280{playback(118)}" {
		t.Fatal(s)
	}
	e := physic.Env{}
	if dev.Sense(&e) == nil {
		t.Fatal("isIdle() should have failed")
	}
}

func TestI2CSenseContinuous280_success(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Normal mode.
			{Addr: 0x76, W: []byte{0xF5, 0xa0, 0xf4, 0x6f}},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}, R: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
			// Normal mode.
			{Addr: 0x76, W: []byte{0xF5, 0, 0xf4, 0x6f}},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}, R: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}, R: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}, R: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
			// Forced mode.
			{Addr: 0x76, W: []byte{0xF5, 0xa0, 0xf4, 0x6c}},
		},
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	c, err := dev.SenseContinuous(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	e := physic.Env{}
	select {
	case e = <-c:
	case <-time.After(2 * time.Second):
		t.Fatal("failed")
	}
	if e.Temperature != 23720*physic.MilliCelsius+physic.ZeroCelsius {
		t.Fatalf("temp %d", e.Temperature)
	}
	if e.Pressure != 100942695312500*physic.NanoPascal {
		t.Fatalf("pressure %d", e.Pressure)
	}
	if e.Humidity != 65305*physic.MicroRH {
		t.Fatalf("humidity %d", e.Humidity)
	}

	// This cancels the previous channel and resets the interval.
	c2, err := dev.SenseContinuous(time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := <-c; ok {
		t.Fatal("c should be closed")
	}
	select {
	case e = <-c2:
	case <-time.After(2 * time.Second):
		t.Fatal("failed")
	}
	select {
	case e = <-c2:
	case <-time.After(2 * time.Second):
		t.Fatal("failed")
	}
	if e.Temperature != 23720*physic.MilliCelsius+physic.ZeroCelsius {
		t.Fatalf("temp %d", e.Temperature)
	}
	if e.Pressure != 100942695312500*physic.NanoPascal {
		t.Fatalf("pressure %d", e.Pressure)
	}
	if e.Humidity != 65305*physic.MicroRH {
		t.Fatalf("humidity %d", e.Humidity)
	}

	if dev.Sense(&e) == nil {
		t.Fatal("can't Sense() during SenseContinously")
	}

	// Inspect to make sure senseContinuous() had the time to do its things. This
	// is a bit sad but it is the simplest way to make this test deterministic.
	for {
		bus.Lock()
		count := bus.Count
		bus.Unlock()
		if count == 10 {
			break
		}
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

func TestI2CSenseContinuous280_command_fail(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Normal mode fails.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dev.SenseContinuous(time.Minute); err == nil {
		t.Fatal("send command should have failed")
	}
}

func TestI2CSenseContinuous280_sense_fail(t *testing.T) {
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
		defer log.SetOutput(os.Stderr)
	}
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chip ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data humidity.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xa0, 0xf4, 0x6c}, R: nil},
			// Normal mode.
			{Addr: 0x76, W: []byte{0xF5, 0xa0, 0xf4, 0x6f}},
			// Read fail.
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, 0x76, nil)
	if err != nil {
		t.Fatal(err)
	}
	c, err := dev.SenseContinuous(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case _, ok := <-c:
		if ok {
			t.Fatal("expecting channel to be closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("failed")
	}
}

func TestCalibration280Float(t *testing.T) {
	// Real data extracted from measurements from this device.
	tRaw := int32(524112)
	pRaw := int32(309104)
	hRaw := int32(30987)

	// Compare the values with the 3 algorithms.
	tempf, tFine := calib280.compensateTempFloat(tRaw)
	presf := calib280.compensatePressureFloat(pRaw, tFine)
	humif := calib280.compensateHumidityFloat(hRaw, tFine)
	if tFine != 117494 {
		t.Fatalf("tFine %d", tFine)
	}
	if !floatEqual(tempf, 22.94811987876892) {
		// 22.95°C
		t.Fatalf("temp %g != %g", tempf, 22.94811987876892)
	}
	if temp := physic.Temperature((tempf+273.15)*1000000000.) * physic.NanoKelvin; temp != 22948119878*physic.NanoKelvin+physic.ZeroCelsius {
		t.Fatalf("temp %d (%s) != %d", temp, temp, 22948119878*physic.NanoKelvin+physic.ZeroCelsius)
	}
	if !floatEqual(presf, 100.04606913901831) {
		// 100.046kPa
		t.Fatalf("pressure %g != %g", presf, 100.04606913901831)
	}
	if pres := physic.Pressure(presf*1000000000000.) * physic.NanoPascal; pres != 100046069139018*physic.NanoPascal {
		t.Fatalf("pressure %d (%s) != %d", pres, pres, 100046069139018*physic.NanoPascal)
	}
	if !floatEqual(humif, 63.16788804728185) {
		// 63.17%
		t.Fatalf("humidity %g != %g", humif, 63.16788804728185)
	}
	if humi := physic.RelativeHumidity(humif*10000.) * physic.MicroRH; humi != 631678*physic.MicroRH {
		t.Fatalf("humidity %d (%s) != %d", humi, humi, 631678*physic.MicroRH)
	}
}

func TestCalibration280Int(t *testing.T) {
	// Real data extracted from measurements from this device.
	tRaw := int32(524112)
	pRaw := int32(309104)
	hRaw := int32(30987)

	temp, tFine := calib280.compensateTempInt(tRaw)
	pres64 := calib280.compensatePressureInt64(pRaw, tFine)
	pres32 := calib280.compensatePressureInt32(pRaw, tFine)
	humi := calib280.compensateHumidityInt(hRaw, tFine)
	if tFine != 117407 {
		t.Fatalf("tFine %d", tFine)
	}
	if temp != 2293 {
		// 2293/100 = 22.93°C
		// Delta is <0.02°C which is pretty good.
		t.Fatalf("temp %d", temp)
	}
	if pres64 != 25611063 {
		// 25611063/256/1000 = 100.043214844
		// Delta is 3Pa which is ok.
		t.Fatalf("pressure64 %d", pres64)
	}
	if pres32 != 100045 {
		// 100045/1000 = 100.045kPa
		// Delta is 1Pa which is pretty good.
		t.Fatalf("pressure32 %d", pres32)
	}
	if humi != 64686 {
		// 64686/1024 = 63.17%
		// Delta is <0.01% which is pretty good.
		t.Fatalf("humidity %d", humi)
	}
}

func TestCalibration280_limits_0(t *testing.T) {
	c := calibration280{h1: 0xFF, h2: 1, h3: 1, h6: 1}
	if v := c.compensateHumidityInt(0x7FFFFFFF>>14, 0xFFFFFFF); v != 0 {
		t.Fatal(v)
	}
}

func TestCalibration280_limits_419430400(t *testing.T) {
	// TODO(maruel): Reverse the equation to overflow  419430400
}

//

func TestOversampling(t *testing.T) {
	data := []struct {
		o Oversampling
		v int
		s string
	}{
		{Off, 0, "Off"},
		{O1x, 1, "1x"},
		{O2x, 2, "2x"},
		{O4x, 4, "4x"},
		{O8x, 8, "8x"},
		{O16x, 16, "16x"},
		{Oversampling(100), 0, "Oversampling(100)"},
	}
	for i, line := range data {
		if v := line.o.asValue(); v != line.v {
			t.Fatalf("#%d %d != %d", i, v, line.v)
		}
		if s := line.o.String(); s != line.s {
			t.Fatalf("#%d %s != %s", i, s, line.s)
		}
	}
}

func TestStandby(t *testing.T) {
	data := []struct {
		isBME bool
		d     time.Duration
		s     standby
	}{
		{true, 0, s500us},
		{true, time.Millisecond, s500us},
		{true, 10 * time.Millisecond, s10msBME},
		{true, 20 * time.Millisecond, s20msBME},
		{true, 62500 * time.Microsecond, s62ms},
		{true, 125 * time.Millisecond, s125ms},
		{true, 250 * time.Millisecond, s250ms},
		{true, 500 * time.Millisecond, s500ms},
		{true, time.Second, s1s},
		{true, 2 * time.Second, s1s},
		{true, 4 * time.Second, s1s},
		{true, time.Minute, s1s},
		{false, 0, s500us},
		{false, time.Millisecond, s500us},
		{false, 10 * time.Millisecond, s62ms},
		{false, 20 * time.Millisecond, s62ms},
		{false, 62500 * time.Microsecond, s62ms},
		{false, 125 * time.Millisecond, s125ms},
		{false, 250 * time.Millisecond, s250ms},
		{false, 500 * time.Millisecond, s500ms},
		{false, time.Second, s1s},
		{false, 2 * time.Second, s2sBMP},
		{false, 4 * time.Second, s4sBMP},
		{false, time.Minute, s4sBMP},
	}
	for i, line := range data {
		if s := chooseStandby(line.isBME, line.d); s != line.s {
			t.Fatalf("#%d chooseStandby(%s) = %d != %d", i, line.d, s, line.s)
		}
	}
}

func TestCalibration280_compensatePressureInt64(t *testing.T) {
	c := calibration280{}
	if x := c.compensatePressureInt64(0, 0); x != 0 {
		t.Fatal(x)
	}
}

func TestCalibration280_compensateHumidityInt(t *testing.T) {
	c := calibration280{
		h1: 0xFF,
	}
	if x := c.compensateHumidityInt(0, 0); x != 0 {
		t.Fatal(x)
	}
}

func BenchmarkCalibration280Int32(b *testing.B) {
	tRaw := int32(524112)
	pRaw := int32(309104)
	hRaw := int32(30987)
	for i := 0; i < b.N; i++ {
		temp, tFine := calib280.compensateTempInt(tRaw)
		pres := calib280.compensatePressureInt32(pRaw, tFine)
		humi := calib280.compensateHumidityInt(hRaw, tFine)
		if tFine != 117407 {
			b.Fatal(tFine)
		}
		if temp != 2293 || pres != 100045 || humi != 64686 {
			b.Fatal(temp, pres, humi)
		}
		// Include the conversion back to int64 as part of the cost.
		if t := physic.Temperature((temp*10+273150)*1000) * physic.MicroKelvin; t != 22930*physic.MilliCelsius+physic.ZeroCelsius {
			b.Fatalf("temp %d (%s) != %d", t, t, 22930*physic.MilliCelsius+physic.ZeroCelsius)
		}
		if p := physic.Pressure(pres) * physic.Pascal; p != 100045*physic.Pascal {
			b.Fatalf("pressure %d (%s) != %d", p, p, 100045*physic.Pascal)
		}
		if h := physic.RelativeHumidity(humi*10) * physic.MicroRH; h != 646860*physic.MicroRH {
			b.Fatalf("humidity %d (%s) != %d", h, h, 646860*physic.MicroRH)
		}
	}
}

func BenchmarkCalibration280Int64(b *testing.B) {
	tRaw := int32(524112)
	pRaw := int32(309104)
	hRaw := int32(30987)
	for i := 0; i < b.N; i++ {
		temp, tFine := calib280.compensateTempInt(tRaw)
		pres := calib280.compensatePressureInt64(pRaw, tFine)
		humi := calib280.compensateHumidityInt(hRaw, tFine)
		if tFine != 117407 {
			b.Fatal(tFine)
		}
		if temp != 2293 || pres != 25611063 || humi != 64686 {
			b.Fatal(temp, pres, humi)
		}
		// Include the conversion back to int64 as part of the cost.
		if t := physic.Temperature((temp*10+273150)*1000) * physic.MicroKelvin; t != 22930*physic.MilliCelsius+physic.ZeroCelsius {
			b.Fatalf("temp %d (%s) != %d", t, t, 22930*physic.MilliCelsius+physic.ZeroCelsius)
		}
		if p := physic.Pressure(pres) * 15625 * physic.MicroPascal / 4; p != 100043214843750*physic.NanoPascal {
			b.Fatalf("pressure %d (%s) != %d", p, p, 100043214843750*physic.NanoPascal)
		}
		if h := physic.RelativeHumidity(humi*10) * physic.MicroRH; h != 646860*physic.MicroRH {
			b.Fatalf("humidity %d (%s) != %d", h, h, 646860*physic.MicroRH)
		}
	}
}

func BenchmarkCalibration280Float64(b *testing.B) {
	tRaw := int32(524112)
	pRaw := int32(309104)
	hRaw := int32(30987)
	for i := 0; i < b.N; i++ {
		tempf, tFine := calib280.compensateTempFloat(tRaw)
		presf := calib280.compensatePressureFloat(pRaw, tFine)
		humif := calib280.compensateHumidityFloat(hRaw, tFine)
		if tFine != 117494 {
			b.Fatal(tFine)
		}
		// Include the conversion back to int64 as part of the cost.
		if temp := physic.Temperature((tempf+273.15)*1000000000.) * physic.NanoKelvin; temp != 22948119878*physic.NanoKelvin+physic.ZeroCelsius {
			b.Fatalf("temp %d (%s) != %d", temp, temp, 22948119878*physic.NanoKelvin+physic.ZeroCelsius)
		}
		if pres := physic.Pressure(presf*1000000000000.) * physic.NanoPascal; pres != 100046069139018*physic.NanoPascal {
			b.Fatalf("pressure %d (%s) != %d", pres, pres, 100046069139018*physic.NanoPascal)
		}
		if humi := physic.RelativeHumidity(humif*10000.) * physic.MicroRH; humi != 631678*physic.MicroRH {
			b.Fatalf("humidity %d (%s) != %d", humi, humi, 631678*physic.MicroRH)
		}
	}
}

//

var epsilon float64 = 0.000000001

func floatEqual(a, b float64) bool {
	return (a-b) < epsilon && (b-a) < epsilon
}

// Page 50

// compensatePressureInt32 returns pressure in Pa. Output value of "96386"
// equals 96386 Pa = 963.86 hPa
//
// "Compensating the pressure value with 32 bit integer has an accuracy of
// typically 1 Pa"
//
// raw has 20 bits of resolution.
//
// BUG(maruel): Output is incorrect.
func (c *calibration280) compensatePressureInt32(raw, tFine int32) uint32 {
	x := tFine>>1 - 64000
	y := (((x >> 2) * (x >> 2)) >> 11) * int32(c.p6)
	y += (x * int32(c.p5)) << 1
	y = y>>2 + int32(c.p4)<<16
	x = (((int32(c.p3) * (((x >> 2) * (x >> 2)) >> 13)) >> 3) + ((int32(c.p2) * x) >> 1)) >> 18
	x = ((32768 + x) * int32(c.p1)) >> 15
	if x == 0 {
		return 0
	}
	p := ((uint32(int32(1048576)-raw) - uint32(y>>12)) * 3125)
	if p < 0x80000000 {
		p = (p << 1) / uint32(x)
	} else {
		p = (p / uint32(x)) * 2
	}
	x = (int32(c.p9) * int32(((p>>3)*(p>>3))>>13)) >> 12
	y = (int32(p>>2) * int32(c.p8)) >> 13
	return uint32(int32(p) + ((x + y + int32(c.p7)) >> 4))
}

// Page 49

// compensateTempFloat returns temperature in °C. Output value of "51.23"
// equals 51.23 °C.
//
// raw has 20 bits of resolution.
func (c *calibration280) compensateTempFloat(raw int32) (float64, int32) {
	x := (float64(raw)/16384. - float64(c.t1)/1024.) * float64(c.t2)
	y := (float64(raw)/131072. - float64(c.t1)/8192.) * float64(c.t3)
	tFine := int32(x + y)
	return float64((x + y) / 5120.), tFine
}

// compensateHumidityFloat returns pressure in Pa. Output value of "96386.2"
// equals 96386.2 Pa = 963.862 hPa.
//
// raw has 20 bits of resolution.
func (c *calibration280) compensatePressureFloat(raw, tFine int32) float64 {
	x := float64(tFine)*0.5 - 64000.
	y := x * x * float64(c.p6) / 32768.
	y += x * float64(c.p5) * 2.
	y = y*0.25 + float64(c.p4)*65536.
	x = (float64(c.p3)*x*x/524288. + float64(c.p2)*x) / 524288.
	x = (1. + x/32768.) * float64(c.p1)
	if x <= 0 {
		return 0
	}
	p := float64(1048576 - raw)
	p = (p - y/4096.) * 6250. / x
	x = float64(c.p9) * p * p / 2147483648.
	y = p * float64(c.p8) / 32768.
	return float64(p+(x+y+float64(c.p7))/16.) / 1000.
}

// compensateHumidityFloat returns humidity in %rH. Output value of "46.332"
// represents 46.332 %rH.
//
// raw has 16 bits of resolution.
func (c *calibration280) compensateHumidityFloat(raw, tFine int32) float64 {
	h := float64(tFine - 76800)
	h = (float64(raw) - float64(c.h4)*64. + float64(c.h5)/16384.*h) * float64(c.h2) / 65536. * (1. + float64(c.h6)/67108864.*h*(1.+float64(c.h3)/67108864.*h))
	h *= 1. - float64(c.h1)*h/524288.
	if h > 100. {
		return 100.
	}
	if h < 0. {
		return 0.
	}
	return float64(h)
}

type spiFail struct {
	spitest.Playback
}

func (s *spiFail) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	return nil, errors.New("failing")
}
