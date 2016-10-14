// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// gpio-read reads a GPIO pin.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/host"
)

func printLevel(l gpio.Level) error {
	if l == gpio.Low {
		_, err := os.Stdout.Write([]byte{'0', '\n'})
		return err
	}
	_, err := os.Stdout.Write([]byte{'1', '\n'})
	return err
}

func mainImpl() error {
	pullUp := flag.Bool("u", false, "pull up")
	pullDown := flag.Bool("d", false, "pull down")
	edges := flag.Bool("e", false, "wait for edges")
	verbose := flag.Bool("v", false, "enable verbose logs")
	flag.Parse()

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	//pull := gpio.PullNoChange
	pull := gpio.Float
	if *pullUp {
		if *pullDown {
			return errors.New("use only one of -d or -u")
		}
		pull = gpio.Up
	}
	if *pullDown {
		pull = gpio.Down
	}
	if flag.NArg() != 1 {
		return errors.New("specify GPIO pin to read")
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	pin, err := strconv.Atoi(flag.Args()[0])
	if err != nil {
		return err
	}
	p := gpio.ByNumber(pin)
	if p == nil {
		return errors.New("specify a valid GPIO pin number")
	}
	edge := gpio.None
	if *edges {
		edge = gpio.Both
	}
	if err = p.In(pull, edge); err != nil {
		return err
	}
	if *edges {
		for {
			p.WaitForEdge(-1)
			printLevel(p.Read())
		}
	}
	return printLevel(p.Read())
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "gpio-read: %s.\n", err)
		os.Exit(1)
	}
}
