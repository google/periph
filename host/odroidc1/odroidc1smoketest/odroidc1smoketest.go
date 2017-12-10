// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package odroidc1smoketest is leveraged by periph-smoketest to verify that
// basic ODROID-C1 specific functionality works.
package odroidc1smoketest

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host/odroidc1"
)

// testOdroidC1Present verifies that odroidc1 is indeed detected.
func testOdroidC1Present() error {
	if !odroidc1.Present() {
		return fmt.Errorf("did not detect presence of ODROID-C1")
	}
	// TODO: add amlogic s805 detection check once that is implemented
	return nil
}

// testOdroidC1Headers verifies that the appropriate headers with the right pin
// count show up and point checks that a couple of pins are correct.
func testOdroidC1Headers() error {
	h := pinreg.All()
	if len(h) != 1 {
		return fmt.Errorf("expected to find 1 header, not %d", len(h))
	}
	if len(h["J2"]) != 20 {
		return fmt.Errorf("expected J2 to have 20 rows, not %d", len(h["J2"]))
	}

	for r := range h["J2"] {
		if len(h["J2"][r]) != 2 {
			return fmt.Errorf("expected row %d of J2 to have 2 pins, not %d",
				r, len(h["J2"][r]))
		}
	}

	j2_3 := h["J2"][1][0]
	if j2_3.Name() != "GPIO74" {
		return fmt.Errorf("expected J2_3 to be GPIO74, not %s", j2_3.Name())
	}
	p := gpioreg.ByName("GPIO74")
	if p == nil || p.Name() != j2_3.Name() { // p is gpio.PinIO while j2_3 is pins.Pin
		return fmt.Errorf(`expected gpioreg.ByName("GPIO74") to equal h["J2"][1][0], instead `+
			"got %s and %s", p, j2_3)
	}

	return nil
}

// testOdroidC1GpioNumbers tests that the gpio pins get the right numbers.
func testOdroidC1GpioNumbers() error {
	must := map[int]string{74: "I2CA_SDA", 75: "I2CA_SCL", 76: "I2CB_SDA", 77: "I2C_SCL",
		107: "SPI0_MOSI", 106: "SPI0_MISO", 105: "SPI0_SCLK", 117: "SPI0_CS0"}
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

// testOdroidC1GpioNames tests that the gpio pins get the right names.
func testOdroidC1GpioNames() error {
	all := []string{}
	for _, p := range gpioreg.All() {
		all = append(all, p.Name())
	}
	sort.Strings(all)

	must := []string{"GPIO74", "GPIO118"}
	for _, name := range must {
		ix := sort.SearchStrings(all, name)
		if ix >= len(all) || all[ix] != name {
			return fmt.Errorf("expected to find gpio pin %s but it's missing", name)
		}
	}
	return nil
}

// testOdroidC1Aliases tests that the various gpio pin aliases get set-up
func testOdroidC1Aliases() error {
	tests := map[string]string{ // alias->real
		"I2CA_SDA":  "GPIO74",
		"I2CA_SCL":  "GPIO75",
		"I2CB_SDA":  "GPIO76",
		"I2CB_SCL":  "GPIO77",
		"SPI0_MOSI": "GPIO107", // Amlogic S805: "GPIO107": "X10",
		"SPI0_MISO": "GPIO106", // Amlogic S805: "GPIO106": "X9",
		"SPI0_SCLK": "GPIO105", // Amlogic S805: "GPIO105": "X8",
		"SPI0_CS0":  "GPIO117", // Amlogic S805: "GPIO117": "X20",
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

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

func (s *SmokeTest) String() string {
	return s.Name()
}

// Name implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Name() string {
	return "odroid-c1"
}

// Description implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Description() string {
	return "Quad core low cost board made by hardkernel.com"
}

// Run implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Run(args []string) error {
	if len(args) != 0 {
		return errors.New("unrecognized arguments")
	}
	tests := []func() error{
		testOdroidC1Present, testOdroidC1Headers,
		testOdroidC1GpioNames, testOdroidC1Aliases,
	}
	for _, t := range tests {
		if err := t(); err != nil {
			return err
		}
	}
	return nil
}
