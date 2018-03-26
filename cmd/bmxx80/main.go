// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// bmxx80 reads environmental data from a BMP180/BME280/BMP280.
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
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/bmxx80"
)

func printPin(fn string, p pin.Pin) {
	name, pos := pinreg.Position(p)
	if name != "" {
		log.Printf("  %-4s: %-10s found on header %s, #%d\n", fn, p, name, pos)
	} else {
		log.Printf("  %-4s: %-10s\n", fn, p)
	}
}

func printEnv(e *physic.Env) {
	if e.Humidity == 0 {
		fmt.Printf("%8s %10s\n", e.Temperature, e.Pressure)
	} else {
		fmt.Printf("%8s %10s %9s\n", e.Temperature, e.Pressure, e.Humidity)
	}
}

func run(dev physic.SenseEnv, interval time.Duration) error {
	if interval == 0 {
		e := physic.Env{}
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
}

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use (default, uses the first I²C found)")
	i2cAddr := flag.Uint("ia", 0x76, "I²C bus address to use; either 0x76 (BMx280, the default) or 0x77 (BMP180)")
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
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

	s := bmxx80.O4x
	if *sample1x {
		s = bmxx80.O1x
	} else if *sample2x {
		s = bmxx80.O2x
	} else if *sample4x {
		s = bmxx80.O4x
	} else if *sample8x {
		s = bmxx80.O8x
	} else if *sample16x {
		s = bmxx80.O16x
	}
	opts := bmxx80.Opts{Temperature: s, Pressure: s, Humidity: s}
	if *filter2x {
		if *interval == 0 {
			return errors.New("-f2 only makes sense with -i")
		}
		opts.Filter = bmxx80.F2
	} else if *filter4x {
		if *interval == 0 {
			return errors.New("-f4 only makes sense with -i")
		}
		opts.Filter = bmxx80.F4
	} else if *filter8x {
		if *interval == 0 {
			return errors.New("-f8 only makes sense with -i")
		}
		opts.Filter = bmxx80.F8
	} else if *filter16x {
		if *interval == 0 {
			return errors.New("-f16 only makes sense with -i")
		}
		opts.Filter = bmxx80.F16
	}

	if _, err := hostInit(); err != nil {
		return err
	}

	var dev *bmxx80.Dev
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
		if dev, err = bmxx80.NewSPI(s, &opts); err != nil {
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
		if dev, err = bmxx80.NewI2C(i, uint16(*i2cAddr), &opts); err != nil {
			return err
		}
	}
	log.Printf("Found %s", dev)
	err := run(dev, *interval)
	if err2 := dev.Halt(); err == nil {
		err = err2
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "bmxx80: %s.\n", err)
		os.Exit(1)
	}
}
