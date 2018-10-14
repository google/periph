// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// simple util to test if a ina219 sensor
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ina219"
	"periph.io/x/periph/host"
)

func main() {
	address := flag.Int("address", 0x40, "I²C address")
	i2cbus := flag.String("bus", "", "I²C bus (/dev/i2c-1)")

	flag.Parse()

	fmt.Println("Starting INA219 Current Sensor\nctrl+c to exit")
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// open default I²C bus
	bus, err := i2creg.Open(*i2cbus)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// create a new power sensor a sense resistor of 100 mΩ, 3.2A
	config := ina219.Config{
		Address:       *address,
		SenseResistor: 100 * physic.MilliOhm,
		MaxCurrent:    3200 * physic.MilliAmpere,
	}

	sensor, err := ina219.New(bus, config)
	if err != nil {
		log.Fatalln(err)
	}

	// read values from sensor every second
	everySecond := time.NewTicker(time.Second).C
	var halt = make(chan os.Signal)
	signal.Notify(halt, syscall.SIGTERM)
	signal.Notify(halt, syscall.SIGINT)

	for {
		select {
		case <-everySecond:
			p, err := sensor.Sense()
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(p)
		case <-halt:
			os.Exit(0)
		}
	}
}
