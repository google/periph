// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// cap1xxx sense touches.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/devices/cap1xxx"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use")
	i2cAddr := flag.Uint("ia", 0x29, "I²C bus address to use, Pimoroni's Drum Hat is 0x2c")
	var hz physic.Frequency
	flag.Var(&hz, "hz", "I²C bus/SPI port speed")
	verbose := flag.Bool("v", false, "verbose mode")
	alertPinName := flag.String("alert", "GPIO25", "Name of the alert/interrupt pin")
	resetPinName := flag.String("reset", "GPIO21", "Name of the reset pin")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	opts := cap1xxx.DefaultOpts
	if *i2cAddr != 0 {
		if *i2cAddr < 0 || *i2cAddr > 65535 {
			return errors.New("invlaid -i2c value")
		}
		opts.I2CAddr = uint16(*i2cAddr)
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	var dev *cap1xxx.Dev
	i2cBus, err := i2creg.Open(*i2cID)
	if err != nil {
		return fmt.Errorf("couldn't open the i2c bus - %s", err)
	}
	defer i2cBus.Close()
	if p, ok := i2cBus.(i2c.Pins); ok {
		printPin("SCL", p.SCL())
		printPin("SDA", p.SDA())
	}

	if hz != 0 {
		if err := i2cBus.SetSpeed(hz); err != nil {
			return fmt.Errorf("couldn't set the i2c bus speed - %s", err)
		}
	}
	// The alert pin is the pin connected to the IRQ/interrupt pin and indicates
	// when a touch event occurs.
	alertPin := gpioreg.ByName(*alertPinName)
	if alertPin == nil {
		return errors.New("invalid alert GPIO pin number")
	}
	if err := alertPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		return err
	}
	log.Printf("cap1xxx: alert pin: %#v", alertPin)

	resetPin := gpioreg.ByName(*resetPinName)
	if resetPin == nil {
		return errors.New("invalid reset GPIO pin number")
	}
	opts.AlertPin = alertPin
	opts.ResetPin = resetPin
	if *verbose {
		opts.Debug = true
	}

	if dev, err = cap1xxx.NewI2C(i2cBus, &opts); err != nil {
		return fmt.Errorf("couldn't open cap1xxx - %s", err)
	}

	userAskedToLinkLEDs := opts.LinkedLEDs
	// unlinked LED demo
	if err := dev.LinkLEDs(false); err != nil {
		log.Printf("Failed to unlink leds: %v", err)
	}
	for i := 0; i < 8; i++ {
		if err := dev.SetLED(i, true); err != nil {
			return err
		}
		time.Sleep(75 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
	if err := dev.AllLEDs(false); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := dev.AllLEDs(true); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err := dev.AllLEDs(false); err != nil {
		return err
	}
	if userAskedToLinkLEDs {
		if err := dev.LinkLEDs(true); err != nil {
			log.Printf("Failed to relink leds: %v", err)
		}
	}

	if alertPin != nil {
		log.Println("Monitoring for touch events")
		var status [8]cap1xxx.TouchStatus
		for {
			if alertPin.WaitForEdge(-1) {
				if err := dev.InputStatus(status[:]); err != nil {
					log.Printf("Error reading inputs: %s", err)
				}
				printSensorsStatus(status[:])
				// We need to clear the interrupt so it can be triggered again.
				if err := dev.ClearInterrupt(); err != nil {
					log.Printf("%v", err)
				}
			}
		}
	}

	if err2 := dev.Halt(); err == nil {
		err = err2
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "cap1xxx: %s.\n", err)
		os.Exit(1)
	}
}

func printPin(fn string, p pin.Pin) {
	name, pos := pinreg.Position(p)
	if name != "" {
		log.Printf("  %-4s: %-10s found on header %s, #%d", fn, p, name, pos)
	} else {
		log.Printf("  %-4s: %-10s", fn, p)
	}
}

func printSensorsStatus(statuses []cap1xxx.TouchStatus) {
	for i, st := range statuses {
		fmt.Printf("#%d: %s", i, st)
		if i != len(statuses)-1 {
			fmt.Printf("\t")
		}
	}
	fmt.Printf("\n")
}
