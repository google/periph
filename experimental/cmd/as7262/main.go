// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// as7262 communicates with an as7262 continually reading the spectrum.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/as7262"
	"periph.io/x/periph/host"
)

func mainImpl() error {

	i2cbus := flag.String("bus", "", "I²C bus (/dev/i2c-1)")

	flag.Parse()

	fmt.Println("Starting AS7262 Visible Spectrum Sensor")
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open I²C bus.
	bus, err := i2creg.Open(*i2cbus)
	if err != nil {
		return fmt.Errorf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// Create a spectrum sensor.
	sensor, err := as7262.New(bus, &as7262.Opts{Gain: as7262.G16x})
	if err != nil {
		return fmt.Errorf("failed to open new sensor: %v", err)
	}

	// Create a ticker to read values from sensor every second.
	everySecond := time.NewTicker(time.Second).C
	defer sensor.Halt()

	fmt.Println("ctrl+c to quit")

	senseTime := time.Millisecond * 300

	for {
		select {
		case <-everySecond:
			spectrum, err := sensor.Sense(12500*physic.MicroAmpere, senseTime)
			if err != nil {
				return fmt.Errorf("sensor reading error: %v", err)
			}
			fmt.Println(spectrum)
		}
	}
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "as7262: %s.\n", err)
		return
	}
}
