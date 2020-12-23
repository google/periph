// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// tlv493d measures the magnetic flux sensed by a tlv493d component.
package main

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/tlv493d"
	"periph.io/x/periph/host"
)

func main() {
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

	// Create a new TLV493D hall effect sensor.
	tlv, err := tlv493d.New(bus, &tlv493d.DefaultOpts)
	if err != nil {
		log.Fatalln(err)
	}
	defer tlv.Halt()

	// Read a single value.
	if err := tlv.SetMode(tlv493d.LowPowerMode); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Single reading")
	reading, err := tlv.Read(tlv493d.HighPrecisionWithTemperature)

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(reading)

	// Read values continuously from the sensor.
	fmt.Println("Continuous reading")
	c, err := tlv.ReadContinuous(100*physic.Hertz, tlv493d.LowPrecision)
	if err != nil {
		log.Fatalln(err)
	}

	for reading := range c {
		fmt.Println(reading)
	}
}
