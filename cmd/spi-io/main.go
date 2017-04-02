// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// spi-io writes to an SPI bus data from stdin and outputs to stdout or writes
// arguments and outputs hex encoded output.
//
// Usage:
//   echo -n -e '\x88\x00' | spi-io -b SPI0.0 | hexdump
//   spi-io -b SPI0.0 0x88 0
//
// For "read only" operation, writes zeros.
// For "write only" operation, ignore stdout.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	busName := flag.String("b", "", "SPI bus to use")
	speed := flag.Int("speed", 1000000, "SPI bus speed in Hz")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if _, err := host.Init(); err != nil {
		return err
	}
	bus, err := spireg.Open(*busName)
	if err != nil {
		return err
	}
	defer bus.Close()

	if *verbose {
		if p, ok := bus.(spi.Pins); ok {
			log.Printf("Using pins CLK: %s  MOSI: %s  MISO:  %s", p.CLK(), p.MOSI(), p.MISO())
		}
	}
	if err = bus.DevParams(int64(*speed), spi.Mode0, 8); err != nil {
		return err
	}

	hex := false
	var write []byte
	if flag.NArg() == 0 {
		write, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		hex = true
		for _, b := range flag.Args() {
			i, err := strconv.ParseUint(b, 0, 8)
			if err != nil {
				return err
			}
			write = append(write, byte(i))
		}
	}

	read := make([]byte, len(write))
	if err = bus.Tx(write, read); err != nil {
		return err
	}
	if !hex {
		_, err = os.Stdout.Write(read)
	} else {
		for i, b := range read {
			if i != 0 {
				if _, err = fmt.Print(", "); err != nil {
					break
				}
			}
			if _, err = fmt.Printf("0x%02X", b); err != nil {
				break
			}
		}
		_, err = fmt.Print("\n")
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "spi-io: %s.\n", err)
		os.Exit(1)
	}
}
