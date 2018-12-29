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
	dev := flag.String("dev", "", "Read only the device with this name. If not specified, read all devices found")
	interval := flag.Duration("interval", 0, "Poll continuously with the given interval")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}
	if *interval != 0 && *dev == "" {
		return errors.New("-dev is required when -interval is used")
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	// Find the named device if provided, otherwise use all devices.
	var sensors []*sysfs.ThermalSensor
	if *dev == "" {
		sensors = sysfs.ThermalSensors
	} else {
		t, err := sysfs.ThermalSensorByName(*dev)
		if err != nil {
			return err
		}
		sensors = []*sysfs.ThermalSensor{t}
	}

	// Read continuously if an interval was provided.
	if *interval != 0 {
		t := sensors[0] // There is exactly 1 device, enforced above.
		ch, err := t.SenseContinuous(*interval)
		if err != nil {
			return err
		}
		for {
			e := <-ch
			fmt.Printf("%s: %s: %s\n", t, t.Type(), e.Temperature)
		}
	}

	for _, t := range sensors {
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
