// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package chipsmoketest is leveraged by periph-smoketest to verify that basic
// CHIP specific functionality works.
package chipsmoketest

import (
	"errors"
	"flag"
	"fmt"
	"sort"
	"strconv"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/allwinner"
	"periph.io/x/periph/host/chip"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

func (s *SmokeTest) String() string {
	return s.Name()
}

// Name implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Name() string {
	return "chip"
}

// Description implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Description() string {
	return "Single CPU low cost board available at getchip.com"
}

// Run implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) error {
	f.Parse(args)
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}
	if !chip.Present() {
		f.Usage()
		return errors.New("this smoke test can only be run on a C.H.I.P. based host")
	}
	if !allwinner.Present() {
		f.Usage()
		return errors.New("this smoke test can only be run on an allwinner based host")
	}
	tests := []func() error{
		testChipHeaders, testChipGpioNumbers, testChipGpioNames, testChipAliases,
	}
	for _, t := range tests {
		if err := t(); err != nil {
			return err
		}
	}
	return nil
}

// testChipHeaders verifies that the appropriate headers with the right pin count show
// up and point checks that a couple of pins are correct.
func testChipHeaders() error {
	h := pinreg.All()
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
	if u13_17.Name() != "LCD-D2" {
		return fmt.Errorf("expected U13_17 to be LCD-D2, not %s", u13_17.Name())
	}
	if u13_17.String() != "LCD-D2(PD2)" {
		return fmt.Errorf("expected U13_17.String() to be \"LCD-D2(PD2)\", instead got %s", u13_17.String())
	}

	u14_24 := h["U14"][11][1]
	if u14_24.Name() != "AP-EINT3" {
		return fmt.Errorf("expected U14_24 to be AP-EINT3, not %s", u14_24.Name())
	}

	u14_17 := h["U14"][8][0]
	if u14_17.Name() != "XIO-P4" {
		return fmt.Errorf("expected U14_17 to be XIO-P4, not %s", u14_17.Name())
	}
	return nil
}

// testChipGpioNumbers tests that the gpio pins get the right numbers.
func testChipGpioNumbers() error {
	must := map[int]string{34: "PB2", 108: "PD12", 139: "PE11", 1022: "GPIO1022"}
	for number, name := range must {
		pin := gpioreg.ByName(strconv.Itoa(number))
		if pin == nil {
			return fmt.Errorf("could not get gpio pin %d (should be %s)", number, name)
		}
		if pin.Name() != name {
			return fmt.Errorf("expected gpio pin %d to be %s but it's %s",
				number, name, pin.Name())
		}
	}
	return nil
}

// testChipGpioNames tests that the gpio pins get the right names.
func testChipGpioNames() error {
	all := []string{}
	for _, p := range gpioreg.All() {
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
		"XIO-P4":   "GPIO1017",
		"LCD-D2":   "PD2",
		"AP-EINT3": "PB3",
	}
	for a, r := range tests {
		p := gpioreg.ByName(a)
		if p == nil {
			return fmt.Errorf("failed to open %s", a)
		}
		pa, ok := p.(gpio.RealPin)
		if !ok {
			return fmt.Errorf("expected that pin %s is an alias, not %T", a, p)
		}
		if pr := pa.Real(); pr.Name() != r {
			return fmt.Errorf("expected that alias %s have real pin %s but it's %s",
				a, r, pr.Name())
		}
	}
	return nil
}

/* The following gpio tests are commented out for now in favor of using gpio-test via a shell
 * script. Once the test strategy settles this can be deleted if it's not used.

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
	p := gpioreg.ByName(name)
	if p == nil {
		return nil, fmt.Errorf("Failed to open %s", name)
	}
	return p, nil
}

// pinByNumber gets a *sysfs* pin by number and calls Fatal if it fails
func pinByNumber(n int) (gpio.PinIO, error) {
	p := sysfs.Pins[n]
	if p == nil {
		return nil, fmt.Errorf("Failed to open sysfs(%d): %s", n, err)
	}
	return p, nil
}

*/
