// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// spi-list lists all SPI buses.
package main

import (
	"fmt"
	"os"

	"periph.io/x/periph/conn/pins"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/headers"
)

func printPin(fn string, p pins.Pin) {
	name, pos := headers.Position(p)
	if name != "" {
		fmt.Printf("  %-4s: %-10s found on header %s, #%d\n", fn, p, name, pos)
	} else {
		fmt.Printf("  %-4s: %-10s\n", fn, p)
	}
}

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return err
	}
	for _, ref := range spi.All() {
		fmt.Printf("%s:\n", ref)
		bus, err := ref.Open()
		if err != nil {
			fmt.Printf("  Failed to open: %v\n", err)
			continue
		}
		if pins, ok := bus.(spi.Pins); ok {
			printPin("CLK", pins.CLK())
			printPin("MOSI", pins.MOSI())
			printPin("MISO", pins.MISO())
			printPin("CS", pins.CS())
		}
		if err := bus.Close(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "spi-list: %s.\n", err)
		os.Exit(1)
	}
}
