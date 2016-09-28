// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// gpio-list prints out the function of each GPIO pin.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/conn/pins"
	"github.com/google/pio/host"
	"github.com/google/pio/host/headers"
)

func printFunc(invalid bool) {
	max := 0
	functional := gpio.Functional()
	funcs := make([]string, 0, len(functional))
	for f := range functional {
		if l := len(f); l > 0 && f[0] != '<' {
			funcs = append(funcs, f)
			if l > max {
				max = l
			}
		}
	}
	sort.Strings(funcs)
	for _, name := range funcs {
		pin := functional[name]
		if invalid || pin != pins.INVALID {
			if pin == nil {
				fmt.Printf("%-*s: INVALID\n", max, name)
			} else {
				fmt.Printf("%-*s: %s\n", max, name, pin)
			}
		}
	}
}

func printGPIO(invalid bool) {
	maxName := 0
	maxFn := 0
	all := gpio.All()
	for _, p := range all {
		if invalid || headers.IsConnected(p) {
			if l := len(p.String()); l > maxName {
				maxName = l
			}
			if l := len(p.Function()); l > maxFn {
				maxFn = l
			}
		}
	}
	for _, p := range all {
		if headers.IsConnected(p) {
			fmt.Printf("%-*s: %s\n", maxName, p, p.Function())
		} else if invalid {
			fmt.Printf("%-*s: %-*s (not connected)\n", maxName, p, maxFn, p.Function())
		}
	}
}

func mainImpl() error {
	all := flag.Bool("a", false, "print everything")
	fun := flag.Bool("f", false, "print functional pins (e.g. I2C1_SCL)")
	gpio := flag.Bool("g", false, "print GPIO pins (e.g. GPIO1) (default)")
	invalid := flag.Bool("n", false, "show not connected/INVALID pins")
	verbose := flag.Bool("v", false, "enable verbose logs")
	flag.Parse()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(0)

	if *all {
		*fun = true
		*gpio = true
		*invalid = true
	} else if !*fun && !*gpio {
		*gpio = true
	}

	if _, err := host.Init(); err != nil {
		return err
	}
	if *fun {
		printFunc(*invalid)
	}
	if *gpio {
		printGPIO(*invalid)
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "gpio-list: %s.\n", err)
		os.Exit(1)
	}
}
