// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// as7262 communicates with an as7262 continually reading the spectrum..

// +build go1.7

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/as7262"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return err
	}
	i2cbus := flag.String("bus", "", "I²C bus (/dev/i2c-1)")

	flag.Parse()

	fmt.Println("Starting AS7262 Visible Spectrum Sensor")
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open default I²C bus.
	bus, err := i2creg.Open(*i2cbus)
	if err != nil {
		return fmt.Errorf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// Create a new power sensor a sense with default options of 100 mΩ, 3.2A at
	// address of 0x40 if no other address supplied with command line option.
	sensor, err := as7262.New(bus, &as7262.Opts{Gain: as7262.G16x})
	if err != nil {
		return fmt.Errorf("failed to open new sensor: %v", err)
	}

	// Read values from sensor every second.
	everySecond := time.NewTicker(time.Second).C
	var halt = make(chan os.Signal)
	signal.Notify(halt, syscall.SIGTERM)
	signal.Notify(halt, syscall.SIGINT)

	fmt.Println("ctrl+c to exit")

	senseTime := time.Millisecond * 300

	for {
		select {
		case <-halt:
			sensor.Halt()
			return fmt.Errorf("got exit signal")
		case <-everySecond:
			spectrum, err := sensor.Sense(100*physic.MilliAmpere, senseTime)
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
