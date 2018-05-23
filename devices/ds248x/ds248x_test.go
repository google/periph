// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds248x

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/i2c/i2ctest"
)

func TestNew(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x18, W: []byte{0xf0}},
			{Addr: 0x18, W: []byte{0xe1, 0xf0}, R: []byte{0x18}},
			{Addr: 0x18, W: []byte{0xd2, 0xe1}, R: []byte{0x1}},
			{Addr: 0x18, W: []byte{0xe1, 0xb4}},
			{Addr: 0x18, W: []byte{0xc3, 0x6, 0x26, 0x46, 0x66, 0x86}},
		},
	}
	d, err := New(&bus, 0x18, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "DS2483{playback(24)}" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNew_opts(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x18, W: []byte{0xf0}},
			{Addr: 0x18, W: []byte{0xe1, 0xf0}, R: []byte{0x18}},
			{Addr: 0x18, W: []byte{0xd2, 0xe1}, R: []byte{0x1}},
			{Addr: 0x18, W: []byte{0xe1, 0xb4}},
			{Addr: 0x18, W: []byte{0xc3, 0x6, 0x26, 0x46, 0x66, 0x86}},
		},
	}
	if _, err := New(&bus, 0x18, &DefaultOpts); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func init() {
	sleep = func(time.Duration) {}
}

/* Commented out in order not to import periph/host, need to move to smoke test
// TestRecordInit tests and records the initialization of a ds248x by accessing
// real hardware and outputs the recording ready to use for playback in
// TestInit.
//
// This test is skipped unless the -record flag is passed to the test executable.
// Use either `go test -args -record` or `ds2483.test -test.v -record`.
func TestRecordInit(t *testing.T) {
	// Only proceed to init hardware and test if -record flag is passed
	if !*record {
		t.SkipNow()
	}
	host.Init()

	i2cReal, err := i2creg.Open("")
	if err != nil {
		t.Fatal(err)
	}
	i2cBus := &i2ctest.Record{Bus: i2cReal}
	// Now init the ds248x.
	owBus, err := New(i2cBus, 0x18, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	// Perform a search triplet operation to see whether anyone is on the bus
	// (we could do a full search but that would produce a very long recording).
	_, err = owBus.SearchTriplet(0)
	if err != nil {
		t.Fatal(err)
	}
	// Output the recording.
	t.Logf("var ops = i2ctest.IO{\n")
	for _, op := range i2cBus.Ops {
		t.Logf("  {Addr: %#x, W: %#v, R: %#v},\n", op.Addr, op.W, op.R)
	}
	t.Logf("}\n")
}

//

var record *bool

func init() {
	record = flag.Bool("record", false, "record real hardware accesses")
}
*/
