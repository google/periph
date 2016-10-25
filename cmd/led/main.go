// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// led reads the state of a LED or change it.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/google/periph/host"
	"github.com/google/periph/host/sysfs"
)

func mainImpl() error {
	flag.Parse()
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
