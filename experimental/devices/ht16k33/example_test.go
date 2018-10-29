// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.
package ht16k33_test

import (
	"log"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/ht16k33"

	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	display, err := ht16k33.NewAlphaNumericDisplay(bus, ht16k33.I2CAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer display.Halt()

	display.DisplayString("ABCD", true)
	time.Sleep(1 * time.Second)

	display.DisplayString("GO", true)
	time.Sleep(1 * time.Second)

	display.DisplayInt(1234, true)
	time.Sleep(1 * time.Second)

	display.DisplayInt(60, true)
	time.Sleep(1 * time.Second)

	display.DisplayFloat(23.99, true)
	time.Sleep(1 * time.Second)

	display.DisplayFloat(1.45, true)
	time.Sleep(1 * time.Second)
}
