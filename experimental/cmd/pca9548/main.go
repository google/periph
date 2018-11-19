// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// pca9548 scans the 8 ports of a pca9548 i2c multiplexer for other i2c devices.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

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

	// Creates a multiplexer with 8 ports at address 0x70 if no other address
	// supplied with command line option.
	mux, err := pca9548.New(bus, &pca9548.Opts{Address: *address})
	if err != nil {
		return fmt.Errorf("failed to load new mux: %v", err)
	}

	if err := mux.RegisterPorts("mux"); err != nil {
		return fmt.Errorf("failed to register ports: %v", err.Error())
	}

	// Create a place to store the results from the scan.
	results := make(map[string][]uint16)
	fmt.Println("Starting Scan")

	// Loop through each bus scanning all addresses for a response.
	for i := 0; i < 8; i++ {
		// Open multiplexer port with alias mux0-mux7.
		m, err := i2creg.Open("mux" + strconv.Itoa(i))
		if err != nil {
			return err
		}
		defer m.Close()

		rx := []byte{0x00}
		for addr := uint16(1); addr < 0x77; addr++ {
			if err := m.Tx(addr, nil, rx); err == nil {
				results[m.String()] = append(results[m.String()], addr)
			}
		}
	}

	fmt.Println("Scan Results:")
	for bus, addrs := range results {
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
