// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// tm1637 writes to a digits LED display.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/devices/tm1637"
	"github.com/google/pio/host"
)

func mainImpl() error {
	clk := flag.Int("c", 4, "CLK pin number")
	data := flag.Int("d", 5, "DIO pin number")
	off := flag.Bool("o", false, "set display as off")
	b1 := flag.Bool("b1", false, "set PWM to 1/16")
	b2 := flag.Bool("b2", false, "set PWM to 2/16")
	b4 := flag.Bool("b4", false, "set PWM to 4/16")
	b10 := flag.Bool("b10", false, "set PWM to 10/16 (default)")
	b12 := flag.Bool("b12", false, "set PWM to 12/16")
	b13 := flag.Bool("b13", false, "set PWM to 13/16")
	b14 := flag.Bool("b14", false, "set PWM to 14/16")
	verbose := flag.Bool("v", false, "verbose mode")
	asSeg := flag.Bool("s", false, "use hex encoded segments instead of numbers")
	asTime := flag.Bool("t", false, "expect two numbers representing time")
	showDot := flag.Bool("dot", false, "when -t is used, show dots")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	b := tm1637.Brightness10
	switch {
	case *off:
		b = tm1637.Off
	case *b1:
		b = tm1637.Brightness1
	case *b2:
		b = tm1637.Brightness2
	case *b4:
		b = tm1637.Brightness4
	case *b10:
		b = tm1637.Brightness10
	case *b12:
		b = tm1637.Brightness12
	case *b13:
		b = tm1637.Brightness13
	case *b14:
		b = tm1637.Brightness14
	}
	if flag.NArg() > 6 {
		return errors.New("too many digits")
	}
	if b != tm1637.Off && flag.NArg() == 0 {
		// Turn it off
		b = tm1637.Off
	}
	var segments []byte
	if *asTime {
		if flag.NArg() != 2 {
			return errors.New("provide hh and mm")
		}
		x, err := strconv.ParseUint(flag.Arg(0), 10, 8)
		if err != nil {
			return err
		}
		hour := int(x)
		x, err = strconv.ParseUint(flag.Arg(1), 10, 8)
		if err != nil {
			return err
		}
		minute := int(x)
		segments = tm1637.Clock(hour, minute, *showDot)
	} else if *asSeg {
		segments = make([]byte, flag.NArg())
		for i, d := range flag.Args() {
			x, err := strconv.ParseUint(d, 16, 8)
			if err != nil {
				return err
			}
			segments[i] = byte(x)
		}
	} else {
		digits := make([]int, flag.NArg())
		for i, d := range flag.Args() {
			x, err := strconv.ParseUint(d, 16, 8)
			if err != nil {
				return err
			}
			digits[i] = int(x)
		}
		segments = tm1637.Digits(digits...)
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	pClk := gpio.ByNumber(*clk)
	if pClk == nil {
		return errors.New("specify a valid pin for clock")
	}
	pData := gpio.ByNumber(*data)
	if pData == nil {
		return errors.New("specify a valid pin for data")
	}
	// TODO(maruel): Print where the pins are located.
	d, err := tm1637.New(pClk, pData)
	if err != nil {
		return err
	}
	if err = d.SetBrightness(b); err != nil {
		return err
	}
	_, err = d.Write(segments)
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "tm1637: %s.\n", err)
		os.Exit(1)
	}
}
