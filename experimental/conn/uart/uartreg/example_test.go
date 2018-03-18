// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package uartreg_test

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"periph.io/x/periph/experimental/conn/uart"
	"periph.io/x/periph/experimental/conn/uart/uartreg"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// How a command line tool may let the user choose an UART port, yet default
	// to the first bus known.
	name := flag.String("uart", "", "UART port to use")
	flag.Parse()
	p, err := uartreg.Open(*name)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	if err := p.Tx([]byte("cmd"), nil); err != nil {
		log.Fatal(err)
	}
}

func ExampleAll() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Enumerate all UART ports available and the corresponding pins.
	fmt.Print("UART ports available:\n")
	for _, ref := range uartreg.All() {
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
		if p, ok := b.(uart.Pins); ok {
			fmt.Printf("  RX : %s", p.RX())
			fmt.Printf("  TX : %s", p.TX())
			fmt.Printf("  RTS: %s", p.RTS())
			fmt.Printf("  CTS: %s", p.CTS())
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

	// On linux, the following calls will likely open the same bus.
	_, _ = uartreg.Open("/dev/ttyUSB0")
	_, _ = uartreg.Open("UART0")
	_, _ = uartreg.Open("0")
}
