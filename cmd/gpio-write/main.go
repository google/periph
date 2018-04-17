// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// gpio-write sets a GPIO pin to low or high.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

func mainImpl() error {
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 2 {
		return errors.New("specify GPIO pin to write to and its level (0 or 1)")
	}

	args := flag.Args()
	l := gpio.Low
	switch args[1] {
	case "0":
	case "1":
		l = gpio.High
	default:
		return errors.New("specify level as 0 or 1")
	}

	if _, err := hostInit(); err != nil {
		return err
	}

	p := gpioreg.ByName(args[0])
	if p == nil {
		return errors.New("invalid GPIO pin number")
	}

	return p.Out(l)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "gpio-write: %s.\n", err)
		os.Exit(1)
	}
}
