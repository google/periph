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
	"github.com/google/periph/host/headers"
	"github.com/google/periph/host/odroid_c1"
)

// testOdroidC1Present verifies that odroid_c1 is indeed detected.
func testOdroidC1Present() error {
	if !odroid_c1.Present() {
		return fmt.Errorf("did not detect presence of ODROID-C1")
	}
	// TODO: add amlogic s805 detection check once that is implemented
	return nil
}

// testOdroidC1Loading verifies that no error occurs when loading all the drivers.
func testOdroidC1Loading() error {
	state, err := host.Init()
	if err != nil {
		return fmt.Errorf("error loading drivers: %s", err)
	}
	if len(state.Failed) > 0 {
		for _, failure := range state.Failed {
			return fmt.Errorf("%s: %s", failure.D, failure.Err)
		}
	}

	// Print some info.
	fmt.Printf("Using drivers:\n")
	for _, driver := range state.Loaded {
		fmt.Printf("- %s\n", driver)
	}
	if len(state.Skipped) > 0 {
		fmt.Printf("Drivers skipped:\n")
		for _, failure := range state.Skipped {
			fmt.Printf("- %s: %s\n", failure.D, failure.Err)
		}
	}

	return nil
}

// testOdroidC1Headers verifies that the appropriate headers with the right pin count show
// up and point checks that a couple of pins are correct.
func testOdroidC1Headers() error {
	h := headers.All()
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
	p := gpio.ByName("GPIO74")
	if p == nil || p.Name() != j2_3.Name() { // p is gpio.PinIO while j2_3 is pins.Pin
		return fmt.Errorf(`expected gpio.ByName("GPIO74") to equal h["J2"][1][0], instead `+
			"got %s and %s", p, j2_3)
	}

	return nil
}

// testOdroidC1GpioNumbers tests that the gpio pins get the right numbers.
func testOdroidC1GpioNumbers() error {
	must := map[int]string{74: "I2CA_SDA", 75: "I2CA_SCL", 76: "I2CB_SDA", 77: "I2C_SCL",
		107: "SPI0_MOSI", 106: "SPI0_MISO", 105: "SPI0_SCLK", 117: "SPI0_CS0"}
	for number, name := range must {
		pin := gpio.ByNumber(number)
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
	for _, p := range gpio.All() {
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
		testOdroidC1Present, testOdroidC1Loading, testOdroidC1Headers,
		testOdroidC1GpioNames, testOdroidC1Aliases,
	}
	for _, t := range tests {
		if err := t(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := Test(); err != nil {
		fmt.Printf("ODROID-C1 test failed: %s\n", err)
		os.Exit(1)
	}
}
