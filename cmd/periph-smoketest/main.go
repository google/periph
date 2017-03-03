// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// periph-smoketest runs all known smoke tests.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"periph.io/x/periph/conn/gpio/gpiosmoketest"
	"periph.io/x/periph/conn/i2c/i2csmoketest"
	"periph.io/x/periph/conn/spi/spismoketest"
	"periph.io/x/periph/experimental/conn/onewire/onewiresmoketest"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/chip/chipsmoketest"
	"periph.io/x/periph/host/odroid_c1/odroidc1smoketest"
)

// SmokeTest must be implemented by a smoke test. It will be run by this
// executable.
type SmokeTest interface {
	// Name is the name of the smoke test, it is the identifier used on the
	// command line.
	Name() string
	// Description returns a short description to be printed to the user in the
	// help page, to explain what this test does and any requirement to make it
	// work.
	Description() string
	// Run runs the test and return an error in case of failure.
	Run(args []string) error
}

// tests is the list of registered smoke tests.
var tests = []SmokeTest{
	&chipsmoketest.SmokeTest{},
	&gpiosmoketest.SmokeTest{},
	&i2csmoketest.SmokeTest{},
	&odroidc1smoketest.SmokeTest{},
	&onewiresmoketest.SmokeTest{},
	&spismoketest.SmokeTest{},
}

func usage() {
	io.WriteString(os.Stderr, "Usage: periph-smoketest <args> <name> ...\n\n")
	flag.PrintDefaults()
	io.WriteString(os.Stderr, "\nTests available:\n")
	names := make([]string, len(tests))
	desc := make(map[string]string, len(tests))
	for i := range tests {
		n := tests[i].Name()
		names[i] = n
		desc[n] = tests[i].Description()
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", name, desc[name])
	}
}

func mainImpl() error {
	state, err := host.Init()
	if err != nil {
		return fmt.Errorf("error loading drivers: %v", err)
	}
	verbose := flag.Bool("v", false, "verbose mode")
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.Usage = usage
	if err := flag.CommandLine.Parse(os.Args[1:]); err == flag.ErrHelp {
		return nil
	} else if err != nil {
		return err
	}
	if flag.NArg() == 0 {
		return errors.New("please specify a test to run or use -help")
	}
	cmd := flag.Arg(0)
	if cmd == "help" {
		usage()
		return nil
	}

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	if *verbose {
		if len(state.Failed) > 0 {
			log.Print("Failed to load some drivers:")
			for _, failure := range state.Failed {
				log.Printf("- %s: %v", failure.D, failure.Err)
			}
		}
		log.Printf("Using drivers:")
		for _, driver := range state.Loaded {
			log.Printf("- %s", driver)
		}
		if len(state.Skipped) > 0 {
			log.Printf("Drivers skipped:")
			for _, failure := range state.Skipped {
				log.Printf("- %s: %v", failure.D, failure.Err)
			}
		}
	}

	for i := range tests {
		if tests[i].Name() == cmd {
			if err = tests[i].Run(flag.Args()[1:]); err == nil {
				log.Printf("Test %s successful", cmd)
			}
			return err
		}
	}
	return fmt.Errorf("test case %q was not found", cmd)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "periph-smoketest: %s.\n", err)
		os.Exit(1)
	}
}
