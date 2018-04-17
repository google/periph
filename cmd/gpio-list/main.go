// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// gpio-list prints out the function of each GPIO pin.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin/pinreg"
)

func printAliases(invalid bool) {
	max := 0
	aliases := gpioreg.Aliases()
	names := make([]string, 0, len(aliases))
	m := make(map[string]gpio.PinIO, len(aliases))
	for _, p := range aliases {
		n := p.Name()
		names = append(names, n)
		m[n] = p
		if l := len(n); l > max {
			max = l
		}
	}
	sort.Strings(names)
	for _, name := range names {
		p := m[name]
		if r, ok := p.(gpio.RealPin); ok {
			p = r.Real()
		}
		if invalid || p.String() != "INVALID" {
			fmt.Printf("%-*s: %s\n", max, name, p)
		}
	}
}

func printGPIO(invalid bool) {
	maxName := 0
	maxFn := 0
	all := gpioreg.All()
	for _, p := range all {
		if invalid || pinreg.IsConnected(p) {
			if l := len(p.String()); l > maxName {
				maxName = l
			}
			if l := len(p.Function()); l > maxFn {
				maxFn = l
			}
		}
	}
	for _, p := range all {
		if pinreg.IsConnected(p) {
			fmt.Printf("%-*s: %s\n", maxName, p, p.Function())
		} else if invalid {
			fmt.Printf("%-*s: %-*s (not connected)\n", maxName, p, maxFn, p.Function())
		}
	}
}

func mainImpl() error {
	all := flag.Bool("a", false, "print everything")
	aliases := flag.Bool("l", false, "print aliases pins (e.g. I2C1_SCL)")
	gpios := flag.Bool("g", false, "print GPIO pins (e.g. GPIO1) (default)")
	invalid := flag.Bool("n", false, "show not connected/INVALID pins")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

	if *all {
		*aliases = true
		*gpios = true
		*invalid = true
	} else if !*aliases && !*gpios {
		*gpios = true
	}

	if _, err := hostInit(); err != nil {
		return err
	}
	if *aliases {
		printAliases(*invalid)
	}
	if *gpios {
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
