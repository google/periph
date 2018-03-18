// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2creg_test

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// A command line tool may let the user choose a I²C port, yet default to the
	// first port known.
	name := flag.String("i2c", "", "I²C bus to use")
	flag.Parse()
	b, err := i2creg.Open(*name)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Dev is a valid conn.Conn.
	d := &i2c.Dev{Addr: 23, Bus: b}

	// Send a command 0x10 and expect a 5 bytes reply.
	write := []byte{0x10}
	read := make([]byte, 5)
	if err := d.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", read)
}

func ExampleAll() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Enumerate all I²C buses available and the corresponding pins.
	fmt.Print("I²C buses available:\n")
	for _, ref := range i2creg.All() {
		fmt.Printf("- %s\n", ref.Name)
		if ref.Number != -1 {
			fmt.Printf("  %d\n", ref.Number)
		}
		if len(ref.Aliases) != 0 {
			fmt.Printf("  %s\n", strings.Join(ref.Aliases, " "))
		}

		b, err := ref.Open()
		if err != nil {
			fmt.Printf("  Failed to open: %v", err)
		}
		if p, ok := b.(i2c.Pins); ok {
			fmt.Printf("  SDA: %s", p.SDA())
			fmt.Printf("  SCL: %s", p.SCL())
		}
		if err := b.Close(); err != nil {
			fmt.Printf("  Failed to close: %v", err)
		}
	}
}

func ExampleOpen() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// On Linux, the following calls will likely open the same bus.
	_, _ = i2creg.Open("/dev/i2c-1")
	_, _ = i2creg.Open("I2C1")
	_, _ = i2creg.Open("1")

	// Opens the first default I²C bus found:
	_, _ = i2creg.Open("")

	// Wondering what to do with the opened i2c.BusCloser? Look at the package's
	// example above.
}
