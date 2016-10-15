// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spi

import (
	"fmt"
	"log"
	"testing"
)

func ExampleAll() {
	fmt.Print("SPI buses available:\n")
	for name := range All() {
		fmt.Printf("- %s\n", name)
	}
}

func Example() {
	// Find a specific device on all available SPI buses:
	for _, opener := range All() {
		bus, err := opener()
		if err != nil {
			log.Fatal(err)
		}
		w := []byte("command to device")
		r := make([]byte, 16)
		if err := bus.Tx(w, r); err != nil {
			log.Fatal(err)
		}
		// Handle 'r'.
		bus.Close()
	}
}

func TestInvalid(t *testing.T) {
	if _, err := New(-1, -1); err == nil {
		t.Fail()
	}
}

func TestAreInSPITest(t *testing.T) {
	// Real tests are in spitest due to cyclic dependency.
}
