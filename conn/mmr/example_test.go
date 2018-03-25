// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mmr_test

import (
	"encoding/binary"
	"fmt"
	"log"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/onewire/onewirereg"
	"periph.io/x/periph/host"
)

func ExampleDev8() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a connection, using I²C as an example:
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	c := &i2c.Dev{Bus: b, Addr: 0xD0}

	d := mmr.Dev8{Conn: c, Order: binary.BigEndian}
	v, err := d.ReadUint8(0xD0)
	if err != nil {
		log.Fatal(err)
	}
	if v == 0x60 {
		fmt.Printf("Found bme280 on bus %s\n", b)
	}
}

func ExampleDev8_ReadStruct() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a connection, using I²C as an example:
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	c := &i2c.Dev{Bus: b, Addr: 0xD0}

	d := mmr.Dev8{Conn: c, Order: binary.BigEndian}
	flags := struct {
		Flag16 uint16
		Flag8  [2]uint8
	}{}
	if err = d.ReadStruct(0xD0, &flags); err != nil {
		log.Fatal(err)
	}
	// Use flags.Flag16 and flags.Flag8.
}

func ExampleDev8_WriteStruct() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Open a connection, using 1-wire as an example:
	b, err := onewirereg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	c := &onewire.Dev{Bus: b, Addr: 0xD0}

	d := mmr.Dev8{Conn: c, Order: binary.LittleEndian}
	flags := struct {
		Flag16 uint16
		Flag8  [2]uint8
	}{
		0x1234,
		[2]uint8{1, 2},
	}
	if err = d.WriteStruct(0xD0, &flags); err != nil {
		log.Fatal(err)
	}
}
