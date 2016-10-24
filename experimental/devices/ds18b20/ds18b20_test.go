// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds18b20

import (
	"os"
	"testing"

	"github.com/google/periph/conn/i2c"
	"github.com/google/periph/experimental/conn/onewire"
	"github.com/google/periph/experimental/devices/ds248x"
	"github.com/google/periph/host"
)

func TestMain(m *testing.M) {
	host.Init()
	os.Exit(m.Run())
}

func initOnewire(t *testing.T) onewire.Conn {
	bus, err := i2c.New(1)
	if err != nil {
		t.Fatal(err)
	}
	ow, err := ds248x.NewI2C(bus, nil)
	if err != nil {
		t.Fatal(err)
	}
	return ow
}

func TestInit(t *testing.T) {
	ow := initOnewire(t)
	// Search on the bus and expect two temperature sensors.
	addrs, err := ow.Search(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) != 2 {
		t.Fatalf("found %d 1-wire devices, expected 2", len(addrs))
	}
	for _, a := range addrs {
		if a&0xff != 0x28 {
			t.Fatalf("found family %#x, expected 0x28 (DS18B20)", a&0xff)
		}
	}
	// Init each sensor.
	sens := make([]*Dev, len(addrs))
	for i := range addrs {
		sens[i], err = New(ow, addrs[i], 10)
		if err != nil {
			t.Fatal(err)
		}
	}
	// Run a conversion on all sensors.
	err = ConvertAll(ow, 10)
	if err != nil {
		t.Fatal(err)
	}
	// Read the temperatures and expect them to be within 5 degrees of one-another.
	temp0 := 0.0
	for _, s := range sens {
		temp, err := s.LastTempFloat()
		if err != nil {
			t.Fatal(err)
		}
		if temp < 10 || temp > 40 {
			t.Errorf("Invalid temperature: %f", temp)
		}
		if temp0 > 1 && (temp-temp0 > 5 || temp-temp0 < -5) {
			t.Errorf("Large temperature discrepancy: %f vs %f", temp0, temp)
		}
		temp0 = temp
	}
}

/* foo
func TestRead(t *testing.T) {
	// This data was generated with "bme280 -r"
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, Write: []byte{0xd0}, Read: []byte{0x60}},
			// Calibration data.
			{
				Addr:  0x76,
				Write: []byte{0x88},
				Read:  []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data.
			{Addr: 0x76, Write: []byte{0xe1}, Read: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, Write: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xe0, 0xf4, 0x6f}, Read: nil},
			// Read.
			{Addr: 0x76, Write: []byte{0xf7}, Read: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
		},
	}
	dev, err := NewI2C(&bus, nil)
	if err != nil {
		t.Fatal(err)
	}
	env := devices.Environment{}
	if err := dev.Sense(&env); err != nil {
		t.Fatalf("Sense(): %v", err)
	}
	if env.Temperature != 23720 {
		t.Fatalf("temp %d", env.Temperature)
	}
	if env.Pressure != 100943 {
		t.Fatalf("pressure %d", env.Pressure)
	}
	if env.Humidity != 6531 {
		t.Fatalf("humidity %d", env.Humidity)
	}
}

func Example() {
	bus, err := i2c.New(-1)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()
	dev, err := NewI2C(bus, nil)
	if err != nil {
		log.Fatalf("failed to initialize ds2483: %v", err)
	}
}
*/
