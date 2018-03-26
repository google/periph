// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// thermal reads the state of thermal sensors exposed via sysfs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/periph/conn/physic"
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
	for _, t := range sysfs.ThermalSensors {
		e := physic.Env{}
		if err := t.Sense(&e); err != nil {
			return err
		}
		fmt.Printf("%s: %s: %s\n", t, t.Type(), e.Temperature)
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "thermal: %s.\n", err)
		os.Exit(1)
	}
}
