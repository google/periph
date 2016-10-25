// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/host"
	"github.com/google/periph/host/allwinner"
	"github.com/google/periph/host/chip"
	"github.com/google/periph/host/headers"
)

// testChipPresent verifies that CHIP and Allwinner are indeed detected.
func testChipPresent() error {
	if !chip.Present() {
		return fmt.Errorf("did not detect presence of CHIP")
	}
	if !allwinner.Present() {
		return fmt.Errorf("did not detect presence of Allwinner CPU")
	}
	return nil
}

// testChipLoading verifies that no error occurs when loading all the drivers for chip.
func testChipLoading() error {
	state, err := host.Init()
	if err != nil {
		return fmt.Errorf("error loading drivers: %s", err)
	}
	if len(state.Failed) > 0 {
		for _, failure := range state.Failed {
			return fmt.Errorf("%s: %s", failure.D, failure.Err)
		}
	}
	return nil
}

// testChipHeaders verifies that the appropriate headers with the right pin count show
// up and point checks that a couple of pins are correct.
func testChipHeaders() error {
	h := headers.All()
	if len(h) != 2 {
		return fmt.Errorf("expected to find 2 headers, not %d", len(h))
	}
	if len(h["U13"]) != 20 {
		return fmt.Errorf("expected U13 to have 20 rows, not %d", len(h["U13"]))
	}
	if len(h["U14"]) != 20 {
		return fmt.Errorf("expected U13 to have 20 rows, not %d", len(h["U13"]))
	}

	for r := range h["U13"] {
		if len(h["U13"][r]) != 2 {
			return fmt.Errorf("expected row %d of U13 to have 2 pins, not %d",
				r, len(h["U13"][r]))
		}
		if len(h["U14"][r]) != 2 {
			return fmt.Errorf("expected row %d of U14 to have 2 pins, not %d",
				r, len(h["U14"][r]))
		}
	}

	u13_17 := h["U13"][8][0]
	if u13_17.Name() != "PD2" {
		return fmt.Errorf("expected U13_17 to be PD2, not %s", u13_17.Name())
	}
	p := gpio.ByName("PD2")
	if p == nil || p.Name() != u13_17.Name() { // p is gpio.PinIO while u13_17 is pins.Pin
		return fmt.Errorf(`expected gpio.ByName("PD2") to equal h["U13"][8][0], instead `+
			"got %s and %s", p, u13_17)
	}

	u14_24 := h["U14"][11][1]
	if p == nil || u14_24.Name() != "PB3" {
		return fmt.Errorf("expected U14_24 to be PB3, not %s", u14_24.Name())
	}

	u14_17 := h["U14"][8][0]
	if p == nil || u14_17.Name() != "GPIO1020" {
		return fmt.Errorf("expected U14_17 to be GPIO1020, not %s", u14_17.Name())
	}
	return nil
}

// testChipGpioNumbers tests that the gpio pins get the right numbers.
func testChipGpioNumbers() error {
	must := map[int]string{34: "PB2", 108: "PD12", 139: "PE11", 1022: "GPIO1022"}
	for number, name := range must {
		pin := gpio.ByNumber(number)
		if pin == nil {
			return fmt.Errorf("could not get gpio pin %d (should be %s)", number, name)
		}
		if pin.Name() != name {
			return fmt.Errorf("expected gpio pin %s to be %s but it's %s",
				number, name, pin.Name())
		}
	}
	return nil
}

// testChipGpioNames tests that the gpio pins get the right names.
func testChipGpioNames() error {
	all := []string{}
	for _, p := range gpio.All() {
		all = append(all, p.Name())
	}
	sort.Strings(all)

	must := []string{"PB2", "PE11", "GPIO1022"}
	for _, name := range must {
		ix := sort.SearchStrings(all, name)
		if ix >= len(all) || all[ix] != name {
			return fmt.Errorf("expected to find gpio pin %s but it's missing", name)
		}
	}
	return nil
}

// testChipAliases tests that the various gpio pin aliases get set-up
func testChipAliases() error {
	tests := map[string]string{ // alias->real
		"XIO-P4": "GPIO1020", "LCD-D2": "PD2", "GPIO98": "PD2",
	}
	for a, r := range tests {
		p := gpio.ByName(a)
		if p == nil {
			return fmt.Errorf("failed to open %s", a)
		}
		pa, ok := p.(*gpio.PinAlias)
		if !ok {
			return fmt.Errorf("expected that pin %s is an alias, not %T", a, pa)
		}
		if pa.Name() != a {
			return fmt.Errorf("the name of alias %s is %s not %s", a, pa.Name(), a)
		}
		pr, ok := p.(gpio.RealPin)
		if !ok {
			return fmt.Errorf("expected that pin alias %s implement RealPin", a)
		}
		if pr.Real().Name() != r {
			return fmt.Errorf("expected that alias %s have real pin %s but it's %s",
				a, r, pr.Real().Name())
		}
	}
	return nil
}

func Test() error {
	tests := []func() error{
		testChipPresent, testChipLoading, testChipHeaders, testChipGpioNames,
		testChipAliases,
	}
	for _, t := range tests {
		err := t()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := Test()
	if err != nil {
		fmt.Printf("CHIP test failed: %s\n", err)
		os.Exit(1)
	}
}

/* The following gpio tests are commented out for now in favor of using gpio-test via a shell
 * script. Once the test startegy settles this can be deleted if it's not used.

// testChipGpioMem tests two connected pins using memory-mapped gpio
func testChipGpioMem() error {
	p1, err := pinByName(t, "PB2")
	if err != nil {
		return err
	}
	p2, err := pinByName(t, "PB3")
	if err != nil {
		return err
	}
	err = gpio.TestCycle(p1, p2, noPull, false)
	if err != nil {
		return err
	}
	err = gpio.TestCycle(p2, p1, noPull, false)
	if err != nil {
		return err
	}
}

// testChipGpioSysfs tests two connected pins using sysfs gpio
func testChipGpioSysfs() error {
	p1, err := pinByNumber(t, 34)
	if err != nil {
		return err
	}
	p2, err := pinByNumber(t, 35)
	if err != nil {
		return err
	}
	err = gpio.TestCycle(p1, p2, noPull, false)
	if err != nil {
		return err
	}
	err = gpio.TestCycle(p2, p1, noPull, false)
	if err != nil {
		return err
	}
}

// testChipGpioXIO tests two connected XIO pins using sysfs gpio
func testChipGpioXIO() error {
	p1, err := pinByNumber(t, 1022)
	if err != nil {
		return err
	}
	p2, err := pinByNumber(t, 1023)
	if err != nil {
		return err
	}
	err = gpio.TestCycle(p1, p2, noPull, false)
	if err != nil {
		return err
	}
	err = gpio.TestCycle(p2, p1, noPull, false)
	if err != nil {
		return err
	}
}

// pinByName gets a gpio pin by name and calls Fatal if it fails
func pinByName(name string) (gpio.PinIO, error) {
	p := gpio.ByName(name)
	if p == nil {
		return nil, fmt.Errorf("Failed to open %s", name)
	}
	return p, nil
}

// pinByNumber gets a *sysfs* pin by number and calls Fatal if it fails
func pinByNumber(n int) (gpio.PinIO, error) {
	p, err := sysfs.PinByNumber(n)
	if p == nil {
		return nil, fmt.Errorf("Failed to open sysfs(%d): %s", n, err)
	}
	return p, nil
}

*/
