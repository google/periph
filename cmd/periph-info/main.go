// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// periph-info prints out information about the loaded periph drivers.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/periph"
	"periph.io/x/periph/host"
)

// driverAfter is an optional function.
// TODO(maruel): Remove in v3.
type driverAfter interface {
	After() []string
}

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
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

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
			p := d.Prerequisites()
			var a []string
			if da, ok := d.(driverAfter); ok {
				a = da.After()
			}
			if len(p) == 0 && len(a) == 0 {
				fmt.Printf("- %s\n", d)
				continue
			}
			fmt.Printf("- %-*s:", max, d)
			if len(p) != 0 {
				fmt.Printf(" %s", p)
			}
			if len(a) != 0 {
				fmt.Printf(" optional: %s", a)
			}
			fmt.Printf("\n")
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
