// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioreg_test

import (
	"flag"
	"fmt"
	"log"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// A command line tool may let the user choose a GPIO pin.
	name := flag.String("p", "", "GPIO pin to use")
	flag.Parse()
	if *name == "" {
		log.Fatal("-p is required")
	}
	p := gpioreg.ByName(*name)
	if p == nil {
		log.Fatalf("Failed to find %s", *name)
	}

	// Set the pin as output High.
	if err := p.Out(gpio.High); err != nil {
		log.Fatal(err)
	}
}

func ExampleAll() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	fmt.Print("GPIO pins available:\n")
	for _, p := range gpioreg.All() {
		fmt.Printf("- %s: %s\n", p, p.Function())
	}
}

func ExampleByName_alias() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// LCD-D2 is a pin found on the C.H.I.P.
	p := gpioreg.ByName("LCD-D2")
	if p == nil {
		log.Fatal("Failed to find LCD-D2")
	}
	if rp, ok := p.(gpio.RealPin); ok {
		fmt.Printf("%s is an alias for %s\n", p, rp.Real())
	} else {
		fmt.Printf("%s is not an alias!\n", p)
	}
}

func ExampleByName_number() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// The string representation of a number works too.
	p := gpioreg.ByName("6")
	if p == nil {
		log.Fatal("Failed to find GPIO6")
	}
	fmt.Printf("%s: %s\n", p, p.Function())
}
