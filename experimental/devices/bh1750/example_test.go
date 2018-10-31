// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bh1750_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/bh1750"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open default I²C bus.
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// Create a new light sensor.
	sensor, err := bh1750.NewI2C(bus, bh1750.I2CAddr)
	if err != nil {
		log.Fatalln(err)
	}

	// Read value from sensor.
	measurement, err := sensor.Sense()

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(measurement)
}
