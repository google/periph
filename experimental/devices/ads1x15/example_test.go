// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ads1x15_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ads1x15"
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

	// Create a new ADS1115 ADC.
	adc, err := ads1x15.NewADS1115(bus, &ads1x15.DefaultOpts)
	if err != nil {
		log.Fatalln(err)
	}

	// Obtain an analog pin from the ADC
	pin, err := adc.PinForDifferenceOfChannels(ads1x15.Channel0, ads1x15.Channel3, 5*physic.Volt, 1*physic.Hertz, ads1x15.SaveEnergy)
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

	fmt.Println(reading)

	// Read values continously from ADC.
	fmt.Println("Continuous reading")
	c := pin.ReadContinuous()

	for reading := range c {
		fmt.Println(reading)
	}
}
