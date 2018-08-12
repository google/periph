// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// headers-list prints out the headers as found on the computer and print the
// functionality of each pin.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
)

func printFailures(state *periph.State) {
	max := 0
	for _, f := range state.Failed {
		if m := len(f.D.String()); m > max {
			max = m
		}
	}
	for _, f := range state.Failed {
		fmt.Fprintf(os.Stderr, "- %-*s: %v\n", max, f.D, f.Err)
	}
}

func altFuncs(p pin.Pin) string {
	alt, ok := p.(pin.PinFunc)
	if !ok {
		return ""
	}
	fn := alt.Func()
	out := ""
	for _, f := range alt.SupportedFuncs() {
		if f == gpio.IN || f == gpio.OUT || f == fn {
			continue
		}
		if out != "" {
			out += ", "
		}
		out += string(f)
	}
	return out
}

func printHardware(showFunctions bool, all map[string][][]pin.Pin) {
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)
	maxName := 0
	maxFn := 0
	maxAltFn := 0
	for _, header := range all {
		if len(header) == 0 || len(header[0]) != 2 {
			continue
		}
		for _, line := range header {
			for _, p := range line {
				if l := len(p.String()); l > maxName {
					maxName = l
				}
				if l := len(p.Function()); l > maxFn {
					maxFn = l
				}
				if showFunctions {
					if l := len(altFuncs(p)); l > maxAltFn {
						maxAltFn = l
					}
				}
			}
		}
	}
	for i, name := range names {
		if i != 0 {
			fmt.Print("\n")
		}
		header := all[name]
		if len(header) == 0 {
			fmt.Printf("%s: No pin connected\n", name)
			continue
		}
		sum := 0
		for _, line := range header {
			sum += len(line)
		}
		fmt.Printf("%s: %d pins\n", name, sum)
		if len(header[0]) == 2 {
			// Two lines.
			if showFunctions {
				// Super wide format.
				fmt.Printf("  %*s  %*s  %*s  Pos  Pos  %-*s  %-*s  %-*s\n",
					maxAltFn, "Alt", maxFn, "Func", maxName, "Name",
					maxName, "Name", maxFn, "Func", maxAltFn, "Alt")
				for i, line := range header {
					fmt.Printf("  %*s  %*s  %*s  %3d  %-3d  %-*s  %-*s  %-*s\n",
						maxAltFn, altFuncs(line[0]), maxFn, line[0].Function(), maxName, line[0], 2*i+1,
						2*i+2, maxName, line[1], maxFn, line[1].Function(), maxAltFn, altFuncs(line[1]))
				}
				continue
			}
			fmt.Printf("  %*s  %*s  Pos  Pos  %-*s  %-*s\n", maxFn, "Func", maxName, "Name", maxName, "Name", maxFn, "Func")
			for i, line := range header {
				fmt.Printf("  %*s  %*s  %3d  %-3d  %-*s  %-*s\n",
					maxFn, line[0].Function(), maxName, line[0], 2*i+1, 2*i+2, maxName, line[1], maxFn, line[1].Function())
			}
			continue
		}
		// One line.
		if showFunctions {
			// Super wide format.
			fmt.Printf("  Pos  %-*s  %-*s  %-*s\n", maxName, "Name", maxFn, "Func", maxAltFn, "Alt")
			pos := 1
			for _, line := range header {
				for _, item := range line {
					fmt.Printf("  %-3d  %-*s  %-*s  %-*s\n", pos, maxName, item, maxFn, item.Function(), maxAltFn, altFuncs(item))
					pos++
				}
			}
		}
		fmt.Printf("  Pos  %-*s  Func\n", maxName, "Name")
		pos := 1
		for _, line := range header {
			for _, item := range line {
				fmt.Printf("  %-3d  %-*s  %s\n", pos, maxName, item, item.Function())
				pos++
			}
		}
	}
}

func mainImpl() error {
	showFunctions := flag.Bool("f", false, "show all alternate functions")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

	state, err := hostInit()
	if err != nil {
		return err
	}
	all := pinreg.All()
	if len(all) == 0 && len(state.Failed) != 0 {
		fmt.Fprintf(os.Stderr, "Got the following driver failures:\n")
		printFailures(state)
		return errors.New("no header found")
	}
	if flag.NArg() == 0 {
		printHardware(*showFunctions, all)
	} else {
		for _, name := range flag.Args() {
			hdr, ok := all[name]
			if !ok {
				return fmt.Errorf("header %q is not registered", name)
			}
			printHardware(*showFunctions, map[string][][]pin.Pin{name: hdr})
		}
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "headers-list: %s.\n", err)
		os.Exit(1)
	}
}
