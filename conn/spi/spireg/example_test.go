// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spireg_test

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// A command line tool may let the user choose a SPI port, yet default to the
	// first port known.
	name := flag.String("spi", "", "SPI port to use")
	flag.Parse()
	p, err := spireg.Open(*name)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// Convert the spi.Port into a spi.Conn so it can be used for communication.
	c, err := p.Connect(physic.MegaHertz, spi.Mode3, 8)
	if err != nil {
		log.Fatal(err)
	}

	// Write 0x10 to the device, and read a byte right after.
	write := []byte{0x10, 0x00}
	read := make([]byte, len(write))
	if err := c.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	// Use read.
	fmt.Printf("%v\n", read[1:])
}

func ExampleAll() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Enumerate all SPI ports available and the corresponding pins.
	fmt.Print("SPI ports available:\n")
	for _, ref := range spireg.All() {
		fmt.Printf("- %s\n", ref.Name)
		if ref.Number != -1 {
			fmt.Printf("  %d\n", ref.Number)
		}
		if len(ref.Aliases) != 0 {
			fmt.Printf("  %s\n", strings.Join(ref.Aliases, " "))
		}

		p, err := ref.Open()
		if err != nil {
			fmt.Printf("  Failed to open: %v", err)
		}
		if p, ok := p.(spi.Pins); ok {
			fmt.Printf("  CLK : %s", p.CLK())
			fmt.Printf("  MOSI: %s", p.MOSI())
			fmt.Printf("  MISO: %s", p.MISO())
			fmt.Printf("  CS  : %s", p.CS())
		}
		if err := p.Close(); err != nil {
			fmt.Printf("  Failed to close: %v", err)
		}
	}
}

func ExampleOpen() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// On Linux, the following calls will likely open the same port.
	_, _ = spireg.Open("/dev/spidev1.0")
	_, _ = spireg.Open("SPI1.0")
	_, _ = spireg.Open("1")

	// Opens the first default SPI bus found:
	_, _ = spireg.Open("")

	// Wondering what to do with the opened spi.PortCloser? Look at the package's
	// example above.
}

//
