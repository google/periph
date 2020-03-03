// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mcp23xxx_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/mcp23xxx"
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

	// Create a new I2C IO extender
	extender, err := mcp23xxx.NewI2C(bus, mcp23xxx.MCP23017, 0x20)
	if err != nil {
		log.Fatalln(err)
	}

	for _, port := range extender.Pins {
		for _, pin := range port {
			err = pin.In(gpio.Float, gpio.NoEdge)
			if err != nil {
				log.Fatalln(err)
			}
			level := pin.Read()
			fmt.Printf("%s\t%s\n", pin.Name(), level.String())
		}
	}

	if err != nil {
		log.Fatalln(err)
	}
}
