// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds18b20

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/periph/conn/i2c"
	"github.com/google/periph/devices"
	"github.com/google/periph/experimental/conn/onewire"
	"github.com/google/periph/experimental/conn/onewire/onewiretest"
	"github.com/google/periph/experimental/devices/ds248x"
	"github.com/google/periph/host"
)

// TestMain lets periph load all drivers and then runs the tests.
func TestMain(m *testing.M) {
	host.Init()
	os.Exit(m.Run())
}

// TestTemperature tests a temperature conversion on a ds18b20 using
// recorded bus transactions.
func TestTemperature(t *testing.T) {
	// set-up playback using the recording output.
	var ops = []onewiretest.IO{
		// Match ROM + Read Scratchpad (init)
		{Write: []uint8{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 0xbe},
			Read: []uint8{0xe0, 0x1, 0x0, 0x0, 0x3f, 0xff, 0x10, 0x10, 0x3f}, Pull: false},
		// Match ROM + Convert
		{Write: []uint8{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 0x44},
			Read: []uint8(nil), Pull: true},
		// Match ROM + Read Scratchpad (read temp)
		{Write: []uint8{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 0xbe},
			Read: []uint8{0xe0, 0x1, 0x0, 0x0, 0x3f, 0xff, 0x10, 0x10, 0x3f}, Pull: false},
	}
	var addr onewire.Address = 0x740000070e41ac28
	var temp devices.Celsius = 30000 // 30.000°C
	owBus := &onewiretest.Playback{Ops: ops}
	// Init the ds18b20.
	ds18b20, err := New(owBus, addr, 10)
	if err != nil {
		t.Fatal(err)
	}
	// Read the temperature.
	t0 := time.Now()
	now, err := ds18b20.Temperature()
	dt := time.Since(t0)
	if err != nil {
		t.Fatal(err)
	}
	// Expect the correct value.
	if now != temp {
		t.Errorf("expected %s, got %s", temp.String(), now.String())
	}
	// Expect it to take >187ms
	if dt < 188*time.Millisecond {
		t.Errorf("expected conversion to take >187ms, took %dms", dt/time.Millisecond)
	}
}

// TestRecordTemp tests and records a temperature conversion. It outputs
// the recording if the tests are run with the verbose option.
//
// This test is skipped if no i2c bus with a ds248x and at least one ds18b20
// is found.
func TestRecordTemp(t *testing.T) {
	i2cBus, err := i2c.New(-1)
	if err != nil {
		t.Skip(err)
	}
	owBus, err := ds248x.New(i2cBus, nil)
	if err != nil {
		t.Skip(err)
	}
	devices, err := owBus.Search(false)
	if err != nil {
		t.Skip(err)
	}
	addrs := "1-wire devices found:"
	for _, a := range devices {
		addrs += fmt.Sprintf(" %#016x", a)
	}
	t.Log(addrs)
	// See whether there's a ds18b20 on the bus.
	var addr onewire.Address
	for _, a := range devices {
		if a&0xff == 0x28 {
			addr = a
			break
		}
	}
	if addr == 0 {
		t.Skip("no DS18B20 found")
	}
	t.Logf("var addr onewire.Address = %#016x", addr)
	// Start recording and perform a temperature conversion.
	rec := &onewiretest.Record{Bus: owBus}
	time.Sleep(50 * time.Millisecond)
	ds18b20, err := New(rec, addr, 10)
	if err != nil {
		t.Fatalf("ds18b20 init: %s", err)
	}
	temp, err := ds18b20.Temperature()
	if err != nil {
		t.Fatal(err)
	}
	// Output what got recorded.
	t.Log("var ops = []onewiretest.IO{")
	for _, op := range rec.Ops {
		t.Logf("  %#v,", op)
	}
	t.Log("}")
	t.Logf("var temp devices.Celsius = %d // %s", temp, temp.String())
}
