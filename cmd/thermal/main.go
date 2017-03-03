// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// thermal reads the state of thermal sensors exposed via sysfs.
package main

import (
	"flag"
	"fmt"
	"os"

	"periph.io/x/periph/devices"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/sysfs"
)

func mainImpl() error {
	flag.Parse()
	if _, err := host.Init(); err != nil {
		return err
	}
	for _, t := range sysfs.ThermalSensors {
		var env devices.Environment
		if err := t.Sense(&env); err != nil {
			return err
		}
		fmt.Printf("%s: %s: %s\n", t, t.Type(), env.Temperature)
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "thermal: %s.\n", err)
		os.Exit(1)
	}
}
