// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// led reads the state of a LED or change it.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/periph/host"
	"periph.io/x/periph/host/sysfs"
)

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

	if _, err := host.Init(); err != nil {
		return err
	}
	for _, led := range sysfs.LEDs {
		fmt.Printf("%s: %s\n", led, led.Function())
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "led: %s.\n", err)
		os.Exit(1)
	}
}
