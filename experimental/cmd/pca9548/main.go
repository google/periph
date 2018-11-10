// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// pca9548 scans the 8 ports of a pca9548 i2c multiplexer for other i2c devices.
package main

import (
	"flag"
	"fmt"
	"os"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/pca9548"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	address := flag.Int("address", 0x70, "I²C address")
	i2cbus := flag.String("bus", "", "I²C bus (/dev/i2c-1)")

	flag.Parse()

	if _, err := host.Init(); err != nil {
		return err
	}

	// Open default I²C bus.
	bus, err := i2creg.Open(*i2cbus)
	if err != nil {
		return fmt.Errorf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// Registers a multiplexer with 8 ports at address 0x70 if no other address
	// supplied with command line option.
	_, err = pca9548.Register(bus, &pca9548.Opts{Address: *address})
	if err != nil {
		return fmt.Errorf("failed to load new mux: %v", err)
	}

	devices := make(map[string][]uint16)
	fmt.Println("Starting Scan")
	for i := 0; i < 8; i++ {
		busname := fmt.Sprintf("mux-%x-%d", *address, i)
		mux, err := i2creg.Open(busname)
		defer mux.Close()
		rx := []byte{0x00}
		if err != nil {
			return err
		}
		for addr := uint16(0); addr < 0x77; addr++ {
			if err := mux.Tx(addr, nil, rx); err == nil {
				devices[busname] = append(devices[busname], addr)
			}
		}
	}

	fmt.Println("Scan Results:")
	for bus, addrs := range devices {
		fmt.Printf("Bus[%s] %d found\n", bus, len(addrs))
		for _, addr := range addrs {
			fmt.Printf("\tDevice at %#2x\n", addr)
		}
	}

	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "pca9548: %s.\n", err)
		return
	}
}
