// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cap1188_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/cap1188"
)

func Example() {
	// Open the I²C bus to which the cap1188 is connected.
	i2cBus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer i2cBus.Close()

	// We need to set an alert ping that will let us know when a touch event
	// occurs. The alert pin is the pin connected to the IRQ/interrupt pin.
	alertPin := gpioreg.ByName("GPIO25")
	if alertPin == nil {
		log.Fatal("invalid alert GPIO pin number")
	}
	// We set the alert pin to monitor for interrupts.
	if err := alertPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Fatalf("Can't monitor the alert pin")
	}

	// Optionally but highly recommended, we can also set a reset pin to
	// start/leave things in a clean state.
	resetPin := gpioreg.ByName("GPIO21")
	if resetPin == nil {
		log.Fatal("invalid reset GPIO pin number")
	}

	// We will configure the cap1188 by setting some options, we can start by the
	// defaults.
	opts := cap1188.DefaultOpts
	opts.AlertPin = alertPin
	opts.ResetPin = resetPin

	// Open the device so we can detect touch events.
	dev, err := cap1188.NewI2C(i2cBus, &opts)
	if err != nil {
		log.Fatalf("couldn't open cap1188: %v", err)
	}

	fmt.Println("Monitoring for touch events")
	maxTouches := 42 // Stop the program after 42 touches.
	for maxTouches > 0 {
		if alertPin.WaitForEdge(-1) {
			maxTouches--
			var statuses [8]cap1188.TouchStatus
			if err := dev.InputStatus(statuses[:]); err != nil {
				fmt.Printf("Error reading inputs: %v\n", err)
				continue
			}
			// print the status of each sensor
			for i, st := range statuses {
				fmt.Printf("#%d: %s\t", i, st)
			}
			fmt.Println()
			// we need to clear the interrupt so it can be triggered again
			if err := dev.ClearInterrupt(); err != nil {
				fmt.Println(err, "while clearing the interrupt")
			}
		}
	}
	fmt.Print("\n")
}
