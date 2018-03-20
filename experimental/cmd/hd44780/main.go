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
	lcd "periph.io/x/periph/experimental/devices/hd44780"
	"periph.io/x/periph/host"
)

func mainFunc() error {
	rsPin := flag.String("rs", "", "Register select pin")
	ePin := flag.String("e", "", "Strobe pin")
	data := flag.String("data", "", "Data pins, comma-separated")
	text := flag.String("text", "", "Text to display, could be multiline")
	flag.Parse()

	host.Init()

	if *rsPin == "" {
		return errors.New("No register select pin configured")
	}

	if *ePin == "" {
		return errors.New("No strobe pin configured")
	}

	if *data == "" {
		return errors.New("No data pins configured")
	}

	pinsStr := strings.Split(*data, ",")

	if len(pinsStr) != 4 {
		return errors.New("Please provide 4 pins for DB4-DB7 pins")
	}

	rsPinReg := gpioreg.ByName(*rsPin)
	if rsPinReg == nil {
		return fmt.Errorf("Register select pin %s can not be found", *rsPin)
	}
	ePinReg := gpioreg.ByName(*ePin)
	if ePin == nil {
		return fmt.Errorf("Strobe pin %s can not be found", *ePin)
	}

	dataPins := make([]gpio.PinOut, 4)
	for i, pinName := range pinsStr {
		tmp := gpioreg.ByName(pinName)
		if tmp == nil {
			return fmt.Errorf("Data pin %s can not be found", tmp)
		}
		dataPins[i] = tmp
	}

	rpi, err := lcd.NewLCD4Bit(dataPins, rsPinReg, ePinReg)

	if err != nil {
		return err
	}

	if *text == "" {
		return nil
	}

	strs := strings.Split(*text, "\n")

	for i := 0; i < len(strs) && i < 2; i++ {
		if err := rpi.SetCursor(uint8(i), 0); err != nil {
			return err
		}
		if err := rpi.Print(strs[i]); err != nil {
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
