// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// i2c communicates to an I²C device.
package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	addr := flag.Int("a", -1, "I²C device address to query")
	busName := flag.String("b", "", "I²C bus to use")
	verbose := flag.Bool("v", false, "verbose mode")
	write := flag.Bool("w", false, "write instead of reading")
	reg := flag.Int("r", -1, "register to address")
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
			b, err := strconv.Atoi(a)
			if err != nil {
				return err
			}
			if b < 0 || b > 255 {
				return errors.New("invalid byte")
			}
			buf = append(buf, byte(b))
		}
	} else {
		if flag.NArg() != 0 {
			return errors.New("do not specify bytes when reading")
		}
		buf = make([]byte, *l)
	}

	bus, err := i2c.Open(*busName)
	if err != nil {
		return err
	}
	defer bus.Close()
	if p, ok := bus.(i2c.Pins); ok {
		log.Printf("Using pins SCL: %s  SDA: %s", p.SCL(), p.SDA())
	}
	d := i2c.Dev{Bus: bus, Addr: uint16(*addr)}
	if *write {
		_, err = d.Write(buf)
	} else {
		if err = d.Tx([]byte{byte(*reg)}, buf); err != nil {
			return err
		}
		fmt.Printf("%s\n", hex.EncodeToString(buf))
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "i2c: %s.\n", err)
		os.Exit(1)
	}
}
