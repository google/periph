// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build usb

// lsusb prints out information about the USB devices.
package main

import (
	"fmt"
	"os"

	"github.com/google/pio/experimental/host/usbbus"
	"github.com/google/pio/host"
)

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return err
	}

	fmt.Printf("Addr  ID\n")
	for _, d := range usbbus.All() {
		fmt.Printf("%02x:%02x %s\n", d.Bus, d.Addr, d.ID)
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "lsusb: %s.\n", err)
		os.Exit(1)
	}
}
