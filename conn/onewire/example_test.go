// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/onewire/onewirereg"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use onewirereg 1-wire bus registry to find the first available 1-wire bus.
	b, err := onewirereg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Dev is a valid conn.Conn.
	d := &onewire.Dev{Addr: 23, Bus: b}

	// Send a command and expect a 5 bytes reply.
	write := []byte{0x10}
	read := make([]byte, 5)
	if err := d.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", read)
}

func ExamplePins() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use onewirereg 1-wire bus registry to find the first available 1-wire bus.
	b, err := onewirereg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	// Prints out the gpio pin used.
	if p, ok := b.(onewire.Pins); ok {
		fmt.Printf("Q: %s", p.Q())
	}
}
