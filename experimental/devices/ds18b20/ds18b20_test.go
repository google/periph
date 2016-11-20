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

func initOnewire(t *testing.T) onewire.Bus {
	bus, err := i2c.New(1)
	if err != nil {
		t.Fatal(err)
	}
	ow, err := ds248x.New(bus, nil)
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
