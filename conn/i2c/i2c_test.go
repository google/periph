// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2c

import (
	"fmt"
	"log"
	"testing"
)

func ExampleAll() {
	fmt.Print("I²C buses available:\n")
	for name := range All() {
		fmt.Printf("- %s\n", name)
	}
}

func ExampleDev() {
	// Find a specific device on all available I²C buses:
	for _, opener := range All() {
		bus, err := opener()
		if err != nil {
			log.Fatal(err)
		}
		dev := Dev{bus, 0x76}
		v, err := dev.ReadRegUint8(0xD0)
		if err != nil {
			log.Fatal(err)
		}
		if v == 0x60 {
			fmt.Printf("Found bme280 on bus %s\n", bus)
		}
		bus.Close()
	}
}

func TestInvalid(t *testing.T) {
	if _, err := New(-1); err == nil {
		t.Fail()
	}
}

func TestAreInI2CTest(t *testing.T) {
	// Real tests are in i2ctest due to cyclic dependency.
}
