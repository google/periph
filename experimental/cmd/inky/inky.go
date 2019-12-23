// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
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
	spiPort     = flag.String("spi", "SPI0.0", "Name or number of SPI port to open")
	path        = flag.String("image", "", "Path to a png file to display on the inky")
	dcPin       = flag.String("dc", "22", "Inky DC Pin")
	resetPin    = flag.String("reset", "27", "Inky Reset Pin")
	busyPin     = flag.String("busy", "17", "Inky Busy Pin")
	model       = flag.String("model", "PHAT", "Inky model (PHAT or WHAT)")
	modelColor  = flag.String("model-color", "red", "Inky model color (black, red or yellow)")
	borderColor = flag.String("border-color", "black", "Border color (black, white, red or yellow)")
)

func parseModel(s string) inky.Model {
	switch s {
	case "PHAT":
		return inky.PHAT
	case "WHAT":
		return inky.WHAT
	default:
		log.Fatalf("Unknown model %q: expected either PHAT or WHAT", s)
		return inky.PHAT
	}
}

func parseColor(s string) inky.Color {
	switch s {
	case "black":
		return inky.Black
	case "white":
		return inky.White
	case "red":
		return inky.Red
	case "yellow":
		return inky.Yellow
	default:
		log.Fatalf("Unknown color %q: expected black, white, red or yellow", s)
		return inky.Black
	}
}

func main() {
	flag.Parse()

	// Open and decode the image.
	f, err := os.Open(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Opening %s...", *spiPort)
	b, err := spireg.Open(*spiPort)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Opening pins...")
	dc := gpioreg.ByName(*dcPin)
	reset := gpioreg.ByName(*resetPin)
	busy := gpioreg.ByName(*busyPin)

	log.Printf("Creating inky...")
	dev, err := inky.New(b, dc, reset, busy, &inky.Opts{
		Model:       parseModel(*model),
		ModelColor:  parseColor(*modelColor),
		BorderColor: parseColor(*borderColor),
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Drawing image...")
	if err := dev.Draw(img.Bounds(), img, image.ZP); err != nil {
		log.Fatal(err)
	}
}
