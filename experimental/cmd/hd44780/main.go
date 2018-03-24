// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/experimental/devices/hd44780"
	"periph.io/x/periph/host"
)

func mainFunc() error {
	rsPin := flag.String("rs", "", "Register select pin")
	ePin := flag.String("e", "", "Strobe pin")
	data := flag.String("data", "", "Data pins, comma-separated")
	text := flag.String("text", "", "Text to display, could be multiline")
	flag.Parse()

	if _, err := host.Init(); err != nil {
		return err
	}

	const pinPattern = "no %s pin specified. Please provide the pin via '%s' flag, for example '%s'"

	if *rsPin == "" {
		return fmt.Errorf(pinPattern, "register select", "-rs", "-rs 25")
	}
	if *ePin == "" {
		return fmt.Errorf(pinPattern, "strobe pin", "-e", "-e 26")
	}
	if *data == "" {
		return fmt.Errorf(pinPattern, "data pins", "-data", "-data 6,13,17,22")
	}

	pinsStr := strings.Split(*data, ",")
	if len(pinsStr) != 4 {
		return errors.New("please provide 4 pins for DB4-DB7 pins")
	}

	rsPinReg := gpioreg.ByName(*rsPin)
	if rsPinReg == nil {
		return fmt.Errorf("Register select pin %s can not be found", *rsPin)
	}
	ePinReg := gpioreg.ByName(*ePin)
	if ePin == nil {
		return fmt.Errorf("Strobe pin %s can not be found", *ePin)
	}

	var dataPins [4]gpio.PinOut
	for i, pinName := range pinsStr {
		if dataPins[i] = gpioreg.ByName(pinName); dataPins[i] == nil {
			return fmt.Errorf("Data pin %s can not be found", pinName)
		}
	}

	dev, err := hd44780.New(dataPins[:], rsPinReg, ePinReg)
	if err != nil {
		return err
	}

	if *text == "" {
		return dev.Halt()
	}

	strs := strings.Split(*text, "\n")

	for i := 0; i < len(strs) && i < 2; i++ {
		if err := dev.SetCursor(uint8(i), 0); err != nil {
			return err
		}
		if err := dev.Print(strs[i]); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := mainFunc(); err != nil {
		fmt.Fprintf(os.Stderr, "hd44780: %s.\n", err)
		os.Exit(1)
	}
}
