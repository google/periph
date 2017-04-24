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

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices"
	"periph.io/x/periph/devices/bme280"
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

func read(e devices.Environmental, interval time.Duration) error {
	var t *time.Ticker
	if interval != 0 {
		t = time.NewTicker(interval)
	}
	for {
		var env devices.Environment
		if err := e.Sense(&env); err != nil {
			return err
		}
		fmt.Printf("%8s %10s %9s\n", env.Temperature, env.Pressure, env.Humidity)
		if t == nil {
			break
		}
		<-t.C
	}
	return nil
}

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use")
	i2cADDR := flag.Uint("ia", 0, "I²C bus address to use")
	spiID := flag.String("spi", "", "SPI bus to use")
	hz := flag.Int("hz", 0, "I²C/SPI bus speed")
	sample1x := flag.Bool("s1", false, "sample at 1x")
	sample2x := flag.Bool("s2", false, "sample at 2x")
	sample4x := flag.Bool("s4", false, "sample at 4x")
	sample8x := flag.Bool("s8", false, "sample at 8x")
	sample16x := flag.Bool("s16", false, "sample at 16x")
	filter2x := flag.Bool("f2", false, "filter IIR at 2x")
	filter4x := flag.Bool("f4", false, "filter IIR at 4x")
	filter8x := flag.Bool("f8", false, "filter IIR at 8x")
	filter16x := flag.Bool("f16", false, "filter IIR at 16x")
	interval := flag.Duration("i", 0, "read data continuously with this interval")
	verbose := flag.Bool("v", false, "verbose mode")
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
	if *spiID != "" {
		bus, err := spireg.Open(*spiID)
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
		if *hz != 0 {
			if err := bus.LimitSpeed(int64(*hz)); err != nil {
				return err
			}
		}
		if dev, err = bme280.NewSPI(bus, &opts); err != nil {
			return err
		}
	} else {
		bus, err := i2creg.Open(*i2cID)
		if err != nil {
			return err
		}
		defer bus.Close()
		if p, ok := bus.(i2c.Pins); ok {
			printPin("SCL", p.SCL())
			printPin("SDA", p.SDA())
		}
		if *hz != 0 {
			if err := bus.SetSpeed(int64(*hz)); err != nil {
				return err
			}
		}
		if dev, err = bme280.NewI2C(bus, &opts); err != nil {
			return err
		}
	}

	err := read(dev, *interval)
	err2 := dev.Halt()
	if err != nil {
		return err
	}
	return err2
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "bme280: %s.\n", err)
		os.Exit(1)
	}
}
