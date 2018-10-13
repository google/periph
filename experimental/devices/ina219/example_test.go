// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ina219_test

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ina219"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// open default I²C bus
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// create a new power sensor
	sensor, err := ina219.New(bus)
	if err != nil {
		log.Fatalln(err)
	}

	// read values from sensor
	p, err := sensor.Sense()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(p)
}

func ExampleNew() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// open default I²C bus
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// create a new power sensor
	options := []ina219.Option{
		ina219.Address(0x40),
		ina219.SenseResistor(100 * physic.MilliOhm),
		ina219.MaxCurrent(3200 * physic.MilliAmpere),
	}
	sensor, err := ina219.New(bus, options...)
	if err != nil {
		log.Fatalln(err)
	}

	// read values from sensor
	p, err := sensor.Sense()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(p)
}

func ExampleAddress() {
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()
	// create a new power sensor with an address of 0x4a (74)
	sensor, err := ina219.New(bus, ina219.Address(0x4a))
	if err != nil {
		log.Fatalln(err)
	}
	p, err := sensor.Sense()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(p)
}

func ExampleSenseResistor() {
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()
	// create a new power sensor a sense resistor of 10 mΩ
	sensor, err := ina219.New(bus, ina219.SenseResistor(10*physic.MilliOhm))
	if err != nil {
		log.Fatalln(err)
	}
	p, err := sensor.Sense()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(p)
}

func ExampleMaxCurrent() {
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()
	// create a new power sensor a maximum current of 1.2A
	sensor, err := ina219.New(bus, ina219.MaxCurrent(1200*physic.MilliAmpere))
	if err != nil {
		log.Fatalln(err)
	}
	p, err := sensor.Sense()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(p)
}

func ExampleSense() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// open default I²C bus
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// create a new power sensor a sense resistor of 10 mΩ
	options := []ina219.Option{
		ina219.Address(0x40),
		ina219.SenseResistor(100 * physic.MilliOhm),
		ina219.MaxCurrent(3200 * physic.MilliAmpere),
	}

	sensor, err := ina219.New(bus, options...)
	if err != nil {
		log.Fatalln(err)
	}

	// read values from sensor every second
	everySecond := time.NewTicker(time.Second).C
	for i := 5; i < 10; i++ {
		select {
		case <-everySecond:
			p, err := sensor.Sense()
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("Bus Voltage: %v\n", p.Voltage)
		}
	}

}
