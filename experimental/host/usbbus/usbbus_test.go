// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build usb

package usbbus

import (
	"fmt"
	"log"
	"testing"

	"periph.io/x/periph/experimental/conn/usb"
	"periph.io/x/periph/host"
)

func Example() {
	usb.Register(usb.ID{0x1234, 0x5678}, func(dev usb.ConnCloser) error {
		fmt.Printf("Detected USB device: %s\n", dev)
		return dev.Close()
	})

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// TODO(maruel): Check if the device is there.
}

func TestUSBBus(t *testing.T) {
}
