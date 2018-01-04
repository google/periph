// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// onewire-list lists all onewire buses and devices.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/onewire/onewirereg"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/host"
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
	for _, ref := range onewirereg.All() {
		fmt.Printf("%s", ref.Name)
		if ref.Number != -1 {
			fmt.Printf(" #%d", ref.Number)
		}
		fmt.Print(":\n")
		if len(ref.Aliases) != 0 {
			fmt.Printf("  Aliases:\n")
			for _, a := range ref.Aliases {
				fmt.Printf("    %s\n", a)
			}
		}
		bus, err := ref.Open()
		if err != nil {
			fmt.Println(" Open error:", err)
			continue
		}
		if p, ok := bus.(onewire.Pins); ok {
			name, pos := pinreg.Position(p.Q())
			if name != "" {
				fmt.Printf("  Q: %-10s found on header %s, #%d\n", p, name, pos)
			} else {
				fmt.Printf("  Q: %-10s\n", p)
			}
		}
		addresses, err := bus.Search(false)
		if err != nil {
			fmt.Println("  Search error:", err)
			continue
		}
		for _, address := range addresses {
			fmt.Printf("  Device address: %#016X\n", address)
		}
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "onewire-list: %s.\n", err)
		os.Exit(1)
	}
}
