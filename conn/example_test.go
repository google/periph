// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package conn_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

func ExampleConn() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Using SPI as an example. See package ./spi/spireg for more details.
	p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	c, err := p.Connect(1000000, spi.Mode3, 8)
	if err != nil {
		log.Fatal(err)
	}

	// Write 0x10 to the device, and read a byte right after.
	write := []byte{0x10, 0x00}
	read := make([]byte, len(write))
	if err := c.Tx(write, read); err != nil {
		log.Fatal(err)
	}
	// Use read.
	fmt.Printf("%v\n", read[1:])
}
