// Copyright 2017 The Periph Authors. All rights reserved.
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
	"os/signal"
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

func printEnv(env *devices.Environment) {
	fmt.Printf("%8s %10s\n", env.Temperature, env.Pressure)
}

func run(dev devices.Environmental, interval time.Duration) error {
	if interval == 0 {
		e := devices.Environment{}
		if err := dev.Sense(&e); err != nil {
			return err
		}
		printEnv(&e)
		return nil
	}

	c, err := dev.SenseContinuous(interval)
	if err != nil {
		return err
	}
	chanSignal := make(chan os.Signal)
	signal.Notify(chanSignal, os.Interrupt)
	for {
		select {
		case <-chanSignal:
			return nil
		case e := <-c:
			printEnv(&e)
		}
	}
	return nil
}

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use")
	sample2x := flag.Bool("s2", false, "sample at 2x")
	sample4x := flag.Bool("s4", false, "sample at 4x")
	sample8x := flag.Bool("s8", false, "sample at 8x")
	interval := flag.Duration("i", 0, "read data continously with this interval")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	os := bmp180.O1x
	if *sample2x {
		os = bmp180.O2x
	} else if *sample4x {
		os = bmp180.O4x
	} else if *sample8x {
		os = bmp180.O8x
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	i, err := i2creg.Open(*i2cID)
	if err != nil {
		return err
	}
	defer i.Close()

	if p, ok := i.(i2c.Pins); ok {
		printPin("SCL", p.SCL())
		printPin("SDA", p.SDA())
	}

	dev, err := bmp180.New(i, os)
	if err != nil {
		return err
	}

	err = run(dev, *interval)
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
