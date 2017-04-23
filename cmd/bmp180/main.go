// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// bmp180 reads the current temperature and pressure from a BMP180.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/devices"
	"periph.io/x/periph/devices/bmp180"
	"periph.io/x/periph/host"
)

func printPin(fn string, p pin.Pin) {
	name, pos := pinreg.Position(p)
	if name != "" {
		log.Printf("  %-4s: %-10s found on header %s, #%d\n", fn, p, name, pos)
	} else {
		log.Printf("  %-4s: %-10s\n", fn, p)
	}
}

func read(e devices.Environmental, loop bool) error {
	for {
		var env devices.Environment
		if err := e.Sense(&env); err != nil {
			return err
		}
		fmt.Printf("%8s %10s %9s\n", env.Temperature, env.Pressure, env.Humidity)
		if !loop {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use")
	hz := flag.Int("hz", 0, "I²C bus speed")
	sample1x := flag.Bool("s1", false, "sample at 1x")
	sample2x := flag.Bool("s2", false, "sample at 2x")
	sample4x := flag.Bool("s4", false, "sample at 4x")
	sample8x := flag.Bool("s8", false, "sample at 8x")
	loop := flag.Bool("l", false, "loop every 100ms")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	opts := &bmp180.Opts{}
	if *sample1x {
		opts.Pressure = bmp180.No
	} else if *sample2x {
		opts.Pressure = bmp180.O2x
	} else if *sample4x {
		opts.Pressure = bmp180.O4x
	} else if *sample8x {
		opts.Pressure = bmp180.O8x
	}

	fmt.Printf("initialize host\n")
	if _, err := host.Init(); err != nil {
		return err
	}

	fmt.Printf("open i2c bus\n")
	bus, err := i2creg.Open(*i2cID)
	if err != nil {
		return err
	}
	defer bus.Close()

	fmt.Printf("i2c bus is open\n")
	if p, ok := bus.(i2c.Pins); ok {
		printPin("SCL", p.SCL())
		printPin("SDA", p.SDA())
	}
	if *hz != 0 {
		if err := bus.SetSpeed(int64(*hz)); err != nil {
			return err
		}
	}

	var dev *bmp180.Dev

	fmt.Printf("open bmp180\n")
	if dev, err = bmp180.New(bus, opts); err != nil {
		return err
	}

	fmt.Printf("read measurement\n")
	err = read(dev, *loop)
	err2 := dev.Halt()
	if err != nil {
		return err
	}
	return err2
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "bmp180: %s.\n", err)
		os.Exit(1)
	}
}
