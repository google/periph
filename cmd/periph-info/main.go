// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// periph-info prints out information about the loaded periph drivers.
package main

import (
	"fmt"
	"os"

	"periph.io/x/periph"
	"periph.io/x/periph/host"
)

func printDrivers(drivers []periph.DriverFailure) {
	if len(drivers) == 0 {
		fmt.Print("  <none>\n")
		return
	}
	max := 0
	for _, f := range drivers {
		if m := len(f.D.String()); m > max {
			max = m
		}
	}
	for _, f := range drivers {
		fmt.Printf("- %-*s: %v\n", max, f.D, f.Err)
	}
}

func mainImpl() error {
	state, err := host.Init()
	if err != nil {
		return err
	}

	fmt.Printf("Drivers loaded and their dependencies, if any:\n")
	if len(state.Loaded) == 0 {
		fmt.Print("  <none>\n")
	} else {
		max := 0
		for _, d := range state.Loaded {
			if m := len(d.String()); m > max {
				max = m
			}
		}
		for _, d := range state.Loaded {
			if p := d.Prerequisites(); len(p) != 0 {
				fmt.Printf("- %-*s: %s\n", max, d, p)
			} else {
				fmt.Printf("- %s\n", d)
			}
		}
	}

	fmt.Printf("Drivers skipped and the reason why:\n")
	printDrivers(state.Skipped)
	fmt.Printf("Drivers failed to load and the error:\n")
	printDrivers(state.Failed)
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "periph-info: %s.\n", err)
		os.Exit(1)
	}
}
