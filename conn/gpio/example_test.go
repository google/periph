// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpio_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use gpioreg GPIO pin registry to find a GPIO pin by name.
	p := gpioreg.ByName("GPIO6")
	if p == nil {
		log.Fatal("Failed to find GPIO6")
	}

	// A pin can be read, independent of its state; it doesn't matter if it is
	// set as input or output.
	fmt.Printf("%s is %s\n", p, p.Read())
}

func ExampleParseDuty() {
	d, err := gpio.ParseDuty("33%")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", d)
	// Output:
	// 33%
}

func ExamplePinIn() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use gpioreg GPIO pin registry to find a GPIO pin by name.
	p := gpioreg.ByName("GPIO6")
	if p == nil {
		log.Fatal("Failed to find GPIO6")
	}

	// Set it as input, with a pull down (defaults to Low when unconnected) and
	// enable rising edge triggering.
	if err := p.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is %s\n", p, p.Read())

	// Wait for rising edges (Low -> High) and print when one occur.
	for p.WaitForEdge(-1) {
		fmt.Printf("%s went %s\n", p, gpio.High)
	}
}

func ExamplePinOut() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use gpioreg GPIO pin registry to find a GPIO pin by name.
	p := gpioreg.ByName("GPIO6")
	if p == nil {
		log.Fatal("Failed to find GPIO6")
	}

	// Set the pin as output High.
	if err := p.Out(gpio.High); err != nil {
		log.Fatal(err)
	}
}

func ExamplePinOut_PWM() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use gpioreg GPIO pin registry to find a GPIO pin by name.
	p := gpioreg.ByName("GPIO6")
	if p == nil {
		log.Fatal("Failed to find GPIO6")
	}

	// Generate a 33% duty cycle 10KHz signal.
	if err := p.PWM(gpio.DutyMax/3, 10*physic.KiloHertz); err != nil {
		log.Fatal(err)
	}
}

func ExampleRealPin() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use gpioreg GPIO pin registry to find a GPIO pin by name.
	p := gpioreg.ByName("P1_3")
	if p == nil {
		log.Fatal("Failed to find P1_3")
	}
	fmt.Printf("P1_3: %s", p)

	// Resolve the real underlying pin.
	if r, ok := p.(gpio.RealPin); ok {
		// On Raspberry Pis, pin #3 on header P1 is an alias for GPIO2.
		fmt.Printf("%s is in fact %s", p, r.Real())
	} else {
		log.Printf("%s is not an alias", p)
	}
}
