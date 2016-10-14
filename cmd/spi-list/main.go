// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// spi-list lists all SPI buses.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/google/pio/conn/pins"
	"github.com/google/pio/conn/spi"
	"github.com/google/pio/host"
	"github.com/google/pio/host/headers"
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
	all := spi.All()
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("%s:\n", name)
		bus, err := all[name]()
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
		bus.Close()
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "spi-list: %s.\n", err)
		os.Exit(1)
	}
}
