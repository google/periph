// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// i2c-list lists all I²C buses.
package main

import (
	"fmt"
	"os"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host"
)

func printPin(fn string, p pin.Pin) {
	name, pos := pinreg.Position(p)
	if name != "" {
		fmt.Printf("  %-3s: %-10s found on header %s, #%d\n", fn, p, name, pos)
	} else {
		fmt.Printf("  %-3s: %-10s\n", fn, p)
	}
}

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return err
	}
	for _, ref := range i2creg.All() {
		fmt.Printf("%s:\n", ref.Name)
		bus, err := ref.Open()
		if err != nil {
			fmt.Printf("  Failed to open: %v\n", err)
			continue
		}
		if p, ok := bus.(i2c.Pins); ok {
			printPin("SCL", p.SCL())
			printPin("SDA", p.SDA())
		}
		if err := bus.Close(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "i2c-list: %s.\n", err)
		os.Exit(1)
	}
}
