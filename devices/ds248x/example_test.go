// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds248x_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/devices/ds248x"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Open the DS248x to get a 1-wire bus.
	ob, err := ds248x.New(b, 0x18, &ds248x.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}
	// Search devices on the bus
	devices, err := ob.Search(false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d 1-wire devices: ", len(devices))
	for _, d := range devices {
		fmt.Printf(" %#16x", uint64(d))
	}
	fmt.Print("\n")
}
