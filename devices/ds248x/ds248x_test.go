// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds248x

import (
	"fmt"
	"testing"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
)

// TestInit tests the initialization of a ds2483 using a recording.
func TestInit(t *testing.T) {
	var ops = []i2ctest.IO{
		{Addr: 0x18, Write: []byte{0xf0}, Read: []byte(nil)},
		{Addr: 0x18, Write: []byte{0xe1, 0xf0}, Read: []byte{0x18}},
		{Addr: 0x18, Write: []byte{0xd2, 0xe1}, Read: []byte{0x1}},
		{Addr: 0x18, Write: []byte{0xe1, 0xb4}, Read: []byte(nil)},
		{Addr: 0x18, Write: []byte{0xc3, 0x6, 0x26, 0x46, 0x66, 0x86}, Read: []byte(nil)},
		{Addr: 0x18, Write: []byte{0x78, 0x0}, Read: []byte(nil)},
		{Addr: 0x18, Write: []byte{}, Read: []byte{0xe8}},
	}

	bus := &i2ctest.Playback{Ops: ops}
	if _, err := New(bus, nil); err != nil {
		t.Fatal(err)
	}
}

func Example() {
	// Open the I²C bus to which the DS248x is connected.
	i2cBus, err := i2c.New(-1)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer i2cBus.Close()

	// Open the DS248x to get a 1-wire bus.
	owBus, err := New(i2cBus, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Search devices on the bus
	devices, err := owBus.Search(false)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Found %d 1-wire devices: ", len(devices))
	for _, d := range devices {
		fmt.Printf(" %#16x", uint64(d))
	}
	fmt.Print('\n')
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

	i2cReal, err := i2c.New(-1)
	if err != nil {
		t.Fatal(err)
	}
	i2cBus := &i2ctest.Record{Bus: i2cReal}
	// Now init the ds248x.
	owBus, err := New(i2cBus, nil)
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
		t.Logf("  {Addr: %#x, Write: %#v, Read: %#v},\n", op.Addr, op.Write, op.Read)
	}
	t.Logf("}\n")
}

//

var record *bool

func init() {
	record = flag.Bool("record", false, "record real hardware accesses")
}
*/
