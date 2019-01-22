// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// spi-io writes to an SPI port data from stdin and outputs to stdout or writes
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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
)

// runTx does the I/O.
//
// If you find yourself with the need to do a one-off complicated transaction
// using TxPackets, temporarily override this function.
func runTx(s spi.Conn, args []string) error {
	hex := false
	var write []byte
	var err error
	if len(args) == 0 {
		write, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		hex = true
		for _, b := range args {
			i, err := strconv.ParseUint(b, 0, 8)
			if err != nil {
				return err
			}
			write = append(write, byte(i))
		}
	}

	read := make([]byte, len(write))
	if err = s.Tx(write, read); err != nil {
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
	// Sample custom testing:
	/*
		p := []spi.Packet{
			{
				W:      []byte{0x01},
				KeepCS: true,
			},
			{
				R: make([]byte, 16),
			},
		}
		return s.TxPackets(p)
	*/
}

func mainImpl() error {
	spiID := flag.String("b", "", "SPI port to use")
	hz := physic.MegaHertz
	flag.Var(&hz, "hz", "SPI port speed")

	nocs := flag.Bool("nocs", false, "do not assert the CS line")
	half := flag.Bool("half", false, "half duplex mode, sharing MOSI and MISO")
	lsbfirst := flag.Bool("lsb", false, "lsb first (default is msb)")
	mode := flag.Int("mode", 0, "CLK and data polarity, between 0 and 3")
	bits := flag.Int("bits", 8, "bits per word")

	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if *mode < 0 || *mode > 3 {
		return errors.New("invalid mode")
	}
	if *bits < 1 || *bits > 255 {
		return errors.New("invalid bits")
	}
	m := spi.Mode(*mode)
	if *half {
		m |= spi.HalfDuplex
	}
	if *nocs {
		m |= spi.NoCS
	}
	if *lsbfirst {
		m |= spi.LSBFirst
	}

	if _, err := hostInit(); err != nil {
		return err
	}
	s, err := spireg.Open(*spiID)
	if err != nil {
		return err
	}
	defer s.Close()
	c, err := s.Connect(hz, m, *bits)
	if err != nil {
		return err
	}
	if *verbose {
		if p, ok := c.(spi.Pins); ok {
			log.Printf("Using pins CLK: %s  MOSI: %s  MISO:  %s", p.CLK(), p.MOSI(), p.MISO())
		}
	}
	return runTx(c, flag.Args())
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "spi-io: %s.\n", err)
		os.Exit(1)
	}
}
