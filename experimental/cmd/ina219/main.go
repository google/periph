// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// ina219 communicates with an ina219 sensor reading voltage, current and power.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/ina219"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return err
	}
	address := flag.Int("address", 0x40, "I²C address")
	i2cbus := flag.String("bus", "", "I²C bus (/dev/i2c-1)")

	flag.Parse()

	fmt.Println("Starting INA219 Current Sensor")
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
	sensor, err := ina219.New(bus, &ina219.Opts{Address: *address})
	if err != nil {
		return fmt.Errorf("failed to open new sensor: %v", err)
	}

	// Read values from sensor every second.
	everySecond := time.NewTicker(time.Second).C
	var halt = make(chan os.Signal, 1)
	signal.Notify(halt, syscall.SIGTERM)
	signal.Notify(halt, syscall.SIGINT)

	fmt.Println("ctrl+c to exit")
	for {
		select {
		case <-everySecond:
			p, err := sensor.Sense()
			if err != nil {
				return fmt.Errorf("sensor reading error: %v", err)
			}
			fmt.Println(p)
		case <-halt:
			return nil
		}
	}
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "ina219: %s.\n", err)
		return
	}
}
