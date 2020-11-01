// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package st7567_test implements an example for the GFX HAT from Pimoroni
//
// Datasheet
//
// https://www.newhavendisplay.com/appnotes/datasheets/LCDs/ST7567.pdf

package st7567_test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/st7567"
	"periph.io/x/periph/host"
)

func Example() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	conn, err := spireg.Open("SPI0.0")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	dc := gpioreg.ByName("6")
	reset := gpioreg.ByName("5")
	cs := gpioreg.ByName("8")

	o := &st7567.Opts{
		Bias:             st7567.Bias17,
		CommonDirection:  st7567.CommonDirReverse,
		SegmentDirection: st7567.SegmentDirNormal,
		Display:          st7567.DisplayNormal,
		RegulationRatio:  st7567.RegulationRatio{st7567.RegResistorRR0, st7567.RegResistorRR1},
		StartLine:        0,
		Contrast:         40,
	}

	dev, err := st7567.New(conn, dc, reset, cs, o)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		fmt.Println("halting device")
		dev.Halt()
	}()

	//Control-C trap
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("halting device")
		dev.Halt()
		os.Exit(1)
	}()

	for x := 0; x < st7567.Width; x++ {
		for y := 0; y < st7567.Height; y++ {
			dev.SetPixel(x, y, true)
		}

		if err = dev.Update(); err != nil {
			log.Fatal(err)
		}

		time.Sleep(40 * time.Millisecond)
	}

	for i := 0; i < 64; i++ {
		fmt.Printf("current contrast value: %d\n", i)

		if err = dev.SetContrast(byte(i)); err != nil {
			log.Fatal(err)
		}

		if err = dev.Update(); err != nil {
			log.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	if err = dev.Update(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("entering power save mode")
	time.Sleep(2 * time.Second)

	if err = dev.PowerSave(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	fmt.Println("leaving power save mode")
	if err = dev.WakeUp(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)
}
