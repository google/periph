// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds18b20

import (
	"reflect"
	"testing"
	"time"

	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/onewire/onewiretest"
	"periph.io/x/periph/conn/physic"
)

func TestNew_fail_resolution(t *testing.T) {
	bus := &onewiretest.Playback{}
	var addr onewire.Address = 0x740000070e41ac28
	if d, err := New(bus, addr, 1); d != nil || err == nil {
		t.Fatal("invalid resolution")
	}
}

func TestNew_fail_read(t *testing.T) {
	bus := &onewiretest.Playback{DontPanic: true}
	var addr onewire.Address = 0x740000070e41ac28
	if d, err := New(bus, addr, 9); d != nil || err == nil {
		t.Fatal("invalid resolution")
	}
}

// TestTemperature tests a temperature conversion on a ds18b20 using
// recorded bus transactions.
func TestTemperature(t *testing.T) {
	// set-up playback using the recording output.
	ops := []onewiretest.IO{
		// Match ROM + Read Scratchpad (init)
		{
			W: []uint8{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 0xbe},
			R: []uint8{0xe0, 0x1, 0x0, 0x0, 0x3f, 0xff, 0x10, 0x10, 0x3f},
		},
		// Match ROM + Convert
		{
			W:    []uint8{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 0x44},
			Pull: true,
		},
		// Match ROM + Read Scratchpad (read temp)
		{
			W: []uint8{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 0xbe},
			R: []uint8{0xe0, 0x1, 0x0, 0x0, 0x3f, 0xff, 0x10, 0x10, 0x3f},
		},
	}
	var addr onewire.Address = 0x740000070e41ac28
	temp := 30*physic.Celsius + physic.ZeroCelsius
	bus := onewiretest.Playback{Ops: ops}
	dev, err := New(&bus, addr, 10)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "DS18B20{{playback 8358680938703596584}}" {
		t.Fatal(s)
	}
	// Read the temperature.
	var sleeps []time.Duration
	sleep = func(d time.Duration) { sleeps = append(sleeps, d) }
	defer func() { sleep = func(time.Duration) {} }()
	now, err := dev.Temperature()
	if err != nil {
		t.Fatal(err)
	}
	// Expect the correct value.
	if now != temp {
		t.Errorf("expected %s, got %s", temp.String(), now.String())
	}
	// Expect it to take >187ms
	if !reflect.DeepEqual(sleeps, []time.Duration{188 * time.Millisecond}) {
		t.Errorf("expected conversion to sleep: %v", sleeps)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestConvertAll tests a temperature conversion on all ds18b20 using
// recorded bus transactions.
func TestConvertAll(t *testing.T) {
	// set-up playback using the recording output.
	ops := []onewiretest.IO{
		// Skip ROM + Convert
		{W: []uint8{0xcc, 0x44}, R: []uint8(nil), Pull: true},
	}
	bus := onewiretest.Playback{Ops: ops}
	// Perform the conversion
	var sleeps []time.Duration
	sleep = func(d time.Duration) { sleeps = append(sleeps, d) }
	defer func() { sleep = func(time.Duration) {} }()
	if err := ConvertAll(&bus, 9); err != nil {
		t.Fatal(err)
	}
	// Expect it to take >93ms
	if !reflect.DeepEqual(sleeps, []time.Duration{94 * time.Millisecond}) {
		t.Errorf("expected conversion to take >93ms, took %s", sleeps)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestConvertAll_fail_resolution(t *testing.T) {
	bus := &onewiretest.Playback{}
	if err := ConvertAll(bus, 1); err == nil {
		t.Fatal("invalid resolution")
	}
}

func TestConvertAll_fail_io(t *testing.T) {
	bus := &onewiretest.Playback{DontPanic: true}
	if err := ConvertAll(bus, 9); err == nil {
		t.Fatal("invalid io")
	}
}

func init() {
	sleep = func(time.Duration) {}
}

/* Commented out in order not to import periph/host, need to move to smoke test
// TestRecordTemp tests and records a temperature conversion. It outputs
// the recording if the tests are run with the verbose option.
//
// This test is skipped unless the -record flag is passed to the test executable.
// Use either `go test -args -record` or `ds18b20.test -test.v -record`.
func TestRecordTemp(t *testing.T) {
	// Only proceed to init hardware and test if -record flag is passed
	if !*record {
		t.SkipNow()
	}
	host.Init()

	i2cBus, err := i2c.New(-1)
	if err != nil {
		t.Fatal(err)
	}
	owBus, err := ds248x.New(i2cBus, nil)
	if err != nil {
		t.Fatal(err)
	}
	devices, err := owBus.Search(false)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal("no DS18B20 found")
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
	t.Logf("var temp physic.Temperature = %d  // %s", temp, temp.String())
}

//

var record *bool

func init() {
	record = flag.Bool("record", false, "record real hardware accesses")
}
*/
