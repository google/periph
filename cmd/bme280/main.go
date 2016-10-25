// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// bme280 reads environmental data from a BME280.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/google/periph/conn/i2c"
	"github.com/google/periph/conn/i2c/i2ctest"
	"github.com/google/periph/conn/pins"
	"github.com/google/periph/conn/spi"
	"github.com/google/periph/devices"
	"github.com/google/periph/devices/bme280"
	"github.com/google/periph/host"
	"github.com/google/periph/host/headers"
)

func printPin(fn string, p pins.Pin) {
	name, pos := headers.Position(p)
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
	i2cID := flag.Int("i", -1, "I²C bus to use")
	i2cADDR := flag.Uint("ia", 0, "I²C bus address to use")
	spiID := flag.Int("s", -1, "SPI bus to use")
	cs := flag.Int("cs", -1, "SPI chip select (CS) line to use")
	sample1x := flag.Bool("s1", false, "sample at 1x")
	sample2x := flag.Bool("s2", false, "sample at 2x")
	sample4x := flag.Bool("s4", false, "sample at 4x")
	sample8x := flag.Bool("s8", false, "sample at 8x")
	sample16x := flag.Bool("s16", false, "sample at 16x")
	filter2x := flag.Bool("f2", false, "filter IIR at 2x")
	filter4x := flag.Bool("f4", false, "filter IIR at 4x")
	filter8x := flag.Bool("f8", false, "filter IIR at 8x")
	filter16x := flag.Bool("f16", false, "filter IIR at 16x")
	loop := flag.Bool("l", false, "loop every 100ms")
	verbose := flag.Bool("v", false, "verbose mode")
	record := flag.Bool("r", false, "record operation (for playback unit testing, only works with I²C)")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	opts := bme280.Opts{Standby: bme280.S20ms}
	s := bme280.O4x
	if *sample1x {
		s = bme280.O1x
	} else if *sample2x {
		s = bme280.O2x
	} else if *sample4x {
		s = bme280.O4x
	} else if *sample8x {
		s = bme280.O8x
	} else if *sample16x {
		s = bme280.O16x
	}
	opts.Temperature = s
	opts.Pressure = s
	opts.Humidity = s
	if *filter2x {
		opts.Filter = bme280.F2
	} else if *filter4x {
		opts.Filter = bme280.F4
	} else if *filter8x {
		opts.Filter = bme280.F8
	} else if *filter16x {
		opts.Filter = bme280.F16
	}
	if *i2cADDR != 0 {
		opts.Address = uint16(*i2cADDR)
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	var dev *bme280.Dev
	var recorder i2ctest.Record
	if *spiID != -1 && *cs != -1 {
		// Spec calls for max 10Mhz. In practice so little data is used.
		bus, err := spi.New(*spiID, *cs)
		if err != nil {
			return err
		}
		defer bus.Close()
		if p, ok := bus.(spi.Pins); ok {
			printPin("CLK", p.CLK())
			printPin("MOSI", p.MOSI())
			printPin("MISO", p.MISO())
			printPin("CS", p.CS())
		}
		if dev, err = bme280.NewSPI(bus, &opts); err != nil {
			return err
		}
	} else {
		bus, err := i2c.New(*i2cID)
		if err != nil {
			return err
		}
		defer bus.Close()
		if p, ok := bus.(i2c.Pins); ok {
			printPin("SCL", p.SCL())
			printPin("SDA", p.SDA())
		}
		var base i2c.Conn = bus
		if *record {
			recorder.Conn = bus
			base = &recorder
		}
		if dev, err = bme280.NewI2C(base, &opts); err != nil {
			return err
		}
	}

	defer dev.Stop()
	err := read(dev, *loop)
	if *record {
		for _, op := range recorder.Ops {
			fmt.Printf("%# v\n", op)
		}
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "bme280: %s.\n", err)
		os.Exit(1)
	}
}
