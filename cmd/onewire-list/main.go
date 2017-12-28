// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// onewire-list lists all onewire buses and devices.
package main

import (
	"fmt"
	"os"

	"periph.io/x/periph/conn/onewire/onewirereg"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return err
	}
	for _, ref := range onewirereg.All() {
		fmt.Println("BUS name:", ref.Name)
		bus, err := ref.Open()
		if err != nil {
			fmt.Println(" Open error:", err)
			continue
		}
		addresses, err := bus.Search(false)
		if err != nil {
			fmt.Println(" Search error:", err)
			continue
		}
		for _, address := range addresses {
			fmt.Printf(" Device address: %#016X\n", address)
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
