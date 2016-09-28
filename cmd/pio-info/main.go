// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// pio-info prints out information about the loaded pio drivers.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/google/pio"
	"github.com/google/pio/host"
)

type failures []pio.DriverFailure

func (f failures) Len() int           { return len(f) }
func (f failures) Less(i, j int) bool { return f[i].D.String() < f[j].D.String() }
func (f failures) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func printOrdered(drivers []pio.DriverFailure) {
	if len(drivers) == 0 {
		fmt.Print("  <none>\n")
	} else {
		list := failures(drivers)
		sort.Sort(list)
		max := 0
		for _, f := range list {
			if m := len(f.D.String()); m > max {
				max = m
			}
		}
		for _, f := range list {
			fmt.Printf("- %-*s: %v\n", max, f.D, f.Err)
		}
	}
}

func mainImpl() error {
	state, err := host.Init()
	if err != nil {
		return err
	}

	fmt.Printf("Drivers loaded and their dependencies, if any:\n")
	if len(state.Loaded) == 0 {
		fmt.Print("  <none>\n")
	} else {
		names := make([]string, 0, len(state.Loaded))
		m := make(map[string]pio.Driver, len(state.Loaded))
		max := 0
		for _, d := range state.Loaded {
			n := d.String()
			if m := len(n); m > max {
				max = m
			}
			names = append(names, n)
			m[n] = d
		}
		sort.Strings(names)
		for _, d := range names {
			p := m[d].Prerequisites()
			if len(p) != 0 {
				fmt.Printf("- %-*s: %s\n", max, d, p)
			} else {
				fmt.Printf("- %s\n", d)
			}
		}
	}

	fmt.Printf("Drivers skipped and the reason why:\n")
	printOrdered(state.Skipped)
	fmt.Printf("Drivers failed to load and the error:\n")
	printOrdered(state.Failed)
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "pio-info: %s.\n", err)
		os.Exit(1)
	}
}
