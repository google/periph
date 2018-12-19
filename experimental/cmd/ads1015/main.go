// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ads1x15"
	"periph.io/x/periph/host"
)

// Resistor values for voltage divider. ADC measures between r2 and ground.
// In this example a 24 tolerant cirquit is used with a voltage divider of
// r1=820kΩ and r2=120kΩ.
const (
	r1 = 820
	r2 = 120
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()
	adc, err := ads1x15.NewADS1015(bus, &ads1x15.DefaultOpts)
	if err != nil {
		log.Fatalln(err)
	}

	// Obtain an analog pin from the ADC.
	pin, err := adc.PinForChannel(ads1x15.Channel0, 1*physic.Volt, 1*physic.Hertz, ads1x15.SaveEnergy)
	if err != nil {
		log.Fatalln(err)
	}
	defer pin.Halt()

	// Read values from ADC.
	fmt.Println("Single reading")
	reading, err := pin.Read()

	if err != nil {
		log.Fatalln(err)
	}

	actualV := (reading.V * (r1 + r2) / r2)
	fmt.Println(actualV)

	// Read values continously from ADC.
	fmt.Println("Continuous reading")
	c := pin.ReadContinuous()

	for reading := range c {
		actualV := (reading.V * (r1 + r2) / r2)
		fmt.Println(actualV)
	}
}
