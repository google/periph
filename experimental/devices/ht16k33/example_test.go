// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.
package ht16k33_test

import (
	"fmt"
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

	if _, err := display.WriteString("ABCD"); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	if _, err := display.WriteString("GO"); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	if _, err := display.WriteString(fmt.Sprintf("%d", 1234)); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	if _, err := display.WriteString(fmt.Sprintf("%d", 60)); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	if _, err := display.WriteString(fmt.Sprintf("%5f", 23.99)); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	if _, err := display.WriteString(fmt.Sprintf("%5f", 1.45)); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)
}
