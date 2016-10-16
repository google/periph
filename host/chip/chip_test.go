// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package chip

import (
	"sort"
	"testing"
	"time"

	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/host"
	"github.com/google/pio/host/allwinner"
	"github.com/google/pio/host/headers"
	"github.com/google/pio/host/sysfs"
)

// TestChipPresent verifies that CHIP and Allwinner are indeed detected
func TestChipPresent(t *testing.T) {
	if !Present() {
		t.Fatalf("Did not detect presence of CHIP")
	}
	if !allwinner.Present() {
		t.Fatalf("Did not detect presence of Allwinner CPU")
	}
}

// TestChipLoading versifies that no error occurs when loading all the drivers for chip
func TestChipLoading(t *testing.T) {
	state, err := host.Init()
	if err != nil {
		t.Fatalf("Error loading drivers: %s", err)
	}
	if len(state.Failed) > 0 {
		for _, failure := range state.Failed {
			t.Errorf("%s: %s", failure.D, failure.Err)
		}
	}
}

// TestChipHeaders verifies that the appropriate headers with the right pin count show up and point
// checks that a couple of pins are correct.
func TestChipHeaders(t *testing.T) {
	host.Init()
	h := headers.All()
	if len(h) != 2 {
		t.Fatalf("Expected to find 2 headers, not %d\n", len(h))
	}
	if len(h["U13"]) != 20 {
		t.Errorf("Expected U13 to have 20 rows, not %d\n", len(h["U13"]))
	}
	if len(h["U14"]) != 20 {
		t.Errorf("Expected U13 to have 20 rows, not %d\n", len(h["U13"]))
	}

	for r := range h["U13"] {
		if len(h["U13"][r]) != 2 {
			t.Errorf("Expected row %d of U13 to have 2 pins, not %d\n", len(h["U13"][r]))
		}
		if len(h["U14"][r]) != 2 {
			t.Errorf("Expected row %d of U14 to have 2 pins, not %d\n", len(h["U14"][r]))
		}
	}

	/* for debugging
	for i := range h["U13"] {
		for j := range h["U13"][i] {
			fmt.Printf("U13[%d][%d] is %s\n", i, j, h["U13"][i][j])
		})
	}*/

	u13_17 := h["U13"][8][0]
	if u13_17.String() != "PD2(98)" {
		t.Errorf("Expected U13_17 to be PD2(98), not %s\n", u13_17.String())
	}
	p := gpio.ByName("PD2(98)")
	if p.String() != u13_17.String() { // p is gpio.PinIO while u13_17 is pins.Pin
		t.Errorf(`Expected gpio.ByName("PD2(98)") to equal h["U13"][8][0], instead `+
			"got %s and %s", p, u13_17)
	}

	u14_24 := h["U14"][11][1]
	if u14_24.String() != "PB3(35)" {
		t.Errorf("Expected U14_24 to be PB3(35), not %s\n", u14_24.String())
	}

	u14_17 := h["U14"][8][0]
	if u14_17.String() != "GPIO1020" {
		t.Errorf("Expected U14_17 to be GPIO1020, not %s\n", u14_17.String())
	}
}

// TestChipGpioNames tests that the gpio pins get the right names
func TestChipGpioNames(t *testing.T) {
	host.Init()
	all := []string{}
	for _, p := range gpio.All() {
		all = append(all, p.String())
	}
	sort.Strings(all)

	//t.Log("Pins:", strings.Join(all, ","))

	// must verifies that a pin exists
	must := func(name string) {
		ix := sort.SearchStrings(all, name)
		if ix >= len(all) || all[ix] != name {
			t.Errorf("Expected to find gpio pin %s but it's missing", name)
		}
	}

	must("PB2(34)")
	must("PE11(139)")
	must("GPIO1022")
}

// TestChipGpioMem tests two connected pins using memory-mapped gpio
func TestChipGpioMem(t *testing.T) {
	host.Init()
	p1 := pinByName(t, "PB2(34)")
	p2 := pinByName(t, "PB3(35)")
	testGpioPair(t, p1, p2)
	testGpioPair(t, p2, p1)
}

// TestChipGpioSysfs tests two connected pins using sysfs gpio
func TestChipGpioSysfs(t *testing.T) {
	host.Init()
	p1 := pinByNumber(t, 34)
	p2 := pinByNumber(t, 35)
	testGpioPair(t, p1, p2)
	testGpioPair(t, p2, p1)
}

// TestChipGpioXIO tests two connected XIO pins using sysfs gpio
func TestChipGpioXIO(t *testing.T) {
	host.Init()
	p1 := pinByNumber(t, 1022)
	p2 := pinByNumber(t, 1023)
	testGpioPair(t, p1, p2)
	testGpioPair(t, p2, p1)
}

// pinByName gets a gpio pin by name and calls Fatal if it fails
func pinByName(t *testing.T, name string) gpio.PinIO {
	p := gpio.ByName(name)
	if p == nil {
		t.Fatalf("Failed to open %s", name)
	}
	return p
}

// pinByNumber gets a *sysfs* pin by number and calls Fatal if it fails
func pinByNumber(t *testing.T, n int) gpio.PinIO {
	p, err := sysfs.PinByNumber(n)
	if p == nil {
		t.Fatalf("Failed to open sysfs(%d): %s", n, err)
	}
	return p
}

// testGpioPair checks that output values on p1 are received on p2
func testGpioPair(t *testing.T, p1, p2 gpio.PinIO) {
	if err := p2.In(gpio.Float, gpio.None); err != nil {
		t.Fatalf("Cannot make %s an input: %s", p2, err)
	}
	if err := p1.Out(gpio.Low); err != nil {
		t.Fatalf("Cannot make %s an output: %s", p1, err)
	}
	defer p1.In(gpio.Float, gpio.None) // leave pin in a safe state

	// Test simple toggling the output and seeing it on the input.
	for i := 0; i < 10; i++ {
		o := gpio.Level(i&1 == 0)
		p1.Out(o)
		in := p2.Read()
		if in != o {
			t.Fatalf("output %v on %s but read %v on %s", o, p1, in, p2)
		}
	}

	// Test edge detection.
	p1.Out(gpio.Low)
	if err := p2.In(gpio.Float, gpio.Both); err != nil {
		t.Errorf("Cannot make %s an input with edge detection: %s", p2, err)
		return
	}
	defer p2.In(gpio.Float, gpio.None) // disable edges again
	if p2.Read() != gpio.Low {
		t.Errorf("whoops!")
	}
	// Make sure there's no pending edge.
	for p2.WaitForEdge(time.Millisecond) {
	}
	// Toggle a few times.
	for i := 0; i < 10; i++ {
		o := gpio.Level(i&1 == 0)
		p1.Out(o)
		if edge := p2.WaitForEdge(10 * time.Millisecond); !edge {
			t.Errorf("output %v on %s but no edge interrupt on %s (input %v, i=%d)",
				o, p1, p2, p2.Read(), i)
			return
		}
		if in := p2.Read(); in != o {
			t.Errorf("output %v but read %v", o, in)
			return
		}
	}
}
