// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// bme280 reads environmental data from a BME280.
package main

import (
	"errors"
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

func printEnv(env *devices.Environment) {
	fmt.Printf("%8s %10s %9s\n", env.Temperature, env.Pressure, env.Humidity)
}

func run(dev devices.Environmental, interval time.Duration) (err error) {
	defer func() {
		if err2 := dev.Halt(); err == nil {
			err = err2
		}
	}()

	if interval == 0 {
		e := devices.Environment{}
		if err = dev.Sense(&e); err != nil {
			return err
		}
		printEnv(&e)
		return nil
	}

	c, err2 := dev.SenseContinuous(interval)
	if err2 != nil {
		return err2
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
}

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use (default)")
	i2cADDR := flag.Uint("ia", 0x76, "I²C bus address to use; either 0x76 or 0x77")
	spiID := flag.String("spi", "", "SPI port to use")
	hz := flag.Int("hz", 0, "I²C bus/SPI port speed")
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
	opts := bme280.Opts{Temperature: s, Pressure: s, Humidity: s}
	if *filter2x {
		if *interval == 0 {
			return errors.New("-f2 only makes sense with -i")
		}
		opts.Filter = bme280.F2
	} else if *filter4x {
		if *interval == 0 {
			return errors.New("-f4 only makes sense with -i")
		}
		opts.Filter = bme280.F4
	} else if *filter8x {
		if *interval == 0 {
			return errors.New("-f8 only makes sense with -i")
		}
		opts.Filter = bme280.F8
	} else if *filter16x {
		if *interval == 0 {
			return errors.New("-f16 only makes sense with -i")
		}
		opts.Filter = bme280.F16
	}
	if *i2cADDR != 0x76 && *i2cADDR > 0x77 {
		return errors.New("-ia must be either 0x76 or 0x77")
	}
	opts.Address = uint16(*i2cADDR)

	if _, err := host.Init(); err != nil {
		return err
	}

	var dev *bme280.Dev
	if *spiID != "" {
		s, err := spireg.Open(*spiID)
		if err != nil {
			return err
		}
		defer s.Close()
		if p, ok := s.(spi.Pins); ok {
			printPin("CLK", p.CLK())
			printPin("MOSI", p.MOSI())
			printPin("MISO", p.MISO())
			printPin("CS", p.CS())
		}
		if *hz != 0 {
			if err := s.LimitSpeed(int64(*hz)); err != nil {
				return err
			}
		}
		if dev, err = bme280.NewSPI(s, &opts); err != nil {
			return err
		}
	} else {
		i, err := i2creg.Open(*i2cID)
		if err != nil {
			return err
		}
		defer i.Close()
		if p, ok := i.(i2c.Pins); ok {
			printPin("SCL", p.SCL())
			printPin("SDA", p.SDA())
		}
		if *hz != 0 {
			if err := i.SetSpeed(int64(*hz)); err != nil {
				return err
			}
		}
		if dev, err = bme280.NewI2C(i, &opts); err != nil {
			return err
		}
	}
	return run(dev, *interval)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "bme280: %s.\n", err)
		os.Exit(1)
	}
}
