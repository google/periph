// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/inky"
	"periph.io/x/periph/host"
)

var (
	spiPort  = flag.String("spi", "SPI0.0", "Name or number of SPI port to open")
	path     = flag.String("image", "", "Path to a png file to display on the inky")
	dcPin    = flag.String("dc", "22", "Inky DC Pin")
	resetPin = flag.String("reset", "27", "Inky Reset Pin")
	busyPin  = flag.String("busy", "17", "Inky Busy Pin")

	model       = inky.PHAT
	modelColor  = inky.Red
	borderColor = inky.Black
)

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "inky: %s.\n", err)
		os.Exit(1)
	}
}

func mainImpl() error {
	flag.Var(&model, "model", "Inky model (PHAT or WHAT)")
	flag.Var(&modelColor, "model-color", "Inky model color (black, red or yellow)")
	flag.Var(&borderColor, "border-color", "Border color (black, white, red or yellow)")
	flag.Parse()

	// Open and decode the image.
	f, err := os.Open(*path)
	if err != nil {
		return err
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return err
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	log.Printf("Opening %s...", *spiPort)
	b, err := spireg.Open(*spiPort)
	if err != nil {
		return err
	}

	log.Printf("Opening pins...")
	dc := gpioreg.ByName(*dcPin)
	if dc == nil {
		return fmt.Errorf("invalid DC pin name: %s", *dcPin)
	}

	reset := gpioreg.ByName(*resetPin)
	if reset == nil {
		return fmt.Errorf("invalid Reset pin name: %s", *resetPin)
	}

	busy := gpioreg.ByName(*busyPin)
	if busy == nil {
		return fmt.Errorf("invalid Busy pin name: %s", *busyPin)
	}

	log.Printf("Creating inky...")
	dev, err := inky.New(b, dc, reset, busy, &inky.Opts{
		Model:       model,
		ModelColor:  modelColor,
		BorderColor: borderColor,
	})
	if err != nil {
		return err
	}

	log.Printf("Drawing image...")
	return dev.Draw(img.Bounds(), img, image.ZP)
}
