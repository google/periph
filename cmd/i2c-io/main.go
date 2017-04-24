// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// i2c-io communicates to an I²C peripheral.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	addr := flag.Int("a", -1, "I²C peripheral address to query")
	busName := flag.String("b", "", "I²C bus to use")
	verbose := flag.Bool("v", false, "verbose mode")
	// TODO(maruel): This is not generic enough.
	write := flag.Bool("w", false, "write instead of reading")
	reg := flag.Int("r", -1, "register to address")
	hz := flag.Int("hz", 0, "I²C bus speed (may require root)")
	l := flag.Int("l", 1, "length of data to read; ignored if -w is specified")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	if *addr < 0 || *addr >= 1<<9 {
		return fmt.Errorf("-a is required and must be between 0 and %d", 1<<9-1)
	}
	if *reg < 0 || *reg > 255 {
		return errors.New("-r must be between 0 and 255")
	}
	if *l <= 0 || *l > 255 {
		return errors.New("-l must be between 1 and 255")
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	var buf []byte
	if *write {
		if flag.NArg() == 0 {
			return errors.New("specify data to write as a list of hex encoded bytes")
		}
		buf = make([]byte, 0, flag.NArg())
		for _, a := range flag.Args() {
			b, err := strconv.ParseUint(a, 0, 8)
			if err != nil {
				return err
			}
			buf = append(buf, byte(b))
		}
	} else {
		if flag.NArg() != 0 {
			return errors.New("do not specify bytes when reading")
		}
		buf = make([]byte, *l)
	}

	bus, err := i2creg.Open(*busName)
	if err != nil {
		return err
	}
	defer bus.Close()

	if *hz != 0 {
		if err := bus.SetSpeed(int64(*hz)); err != nil {
			return err
		}
	}
	if *verbose {
		if p, ok := bus.(i2c.Pins); ok {
			log.Printf("Using pins SCL: %s  SDA: %s", p.SCL(), p.SDA())
		}
	}
	d := i2c.Dev{Bus: bus, Addr: uint16(*addr)}
	if *write {
		_, err = d.Write(buf)
	} else {
		if err = d.Tx([]byte{byte(*reg)}, buf); err != nil {
			return err
		}
		for i, b := range buf {
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
		fmt.Fprintf(os.Stderr, "i2c-io: %s.\n", err)
		os.Exit(1)
	}
}
