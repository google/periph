// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/st7567"
	"periph.io/x/periph/host"
)

var (
	spiPort  = flag.String("spi", "SPI0.0", "Name or number of SPI port to open")
	dcPin    = flag.String("dc", "6", "DC pin")
	resetPin = flag.String("reset", "5", "Reset pin")
	csPin    = flag.String("cs", "8", "Chip select pin")

	contrast  = flag.Int("contrast", 58, "The contrast of the LCD valid values are between 0-63")
	startLine = flag.Int("startLine", 0, "The start line valid values are between 0-63")
	path      = flag.String("image", "", "The path to the PNG which should be painted on the LCD")

	bias               = st7567.Bias17
	segmentDirection   = st7567.SegmentDirNormal
	comDirection       = st7567.CommonDirReverse
	display            = st7567.DisplayNormal
	regulationResistor = st7567.RegulationRatio{st7567.RegResistorRR0, st7567.RegResistorRR1}
)

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "st7567: %s.\n", err)
		os.Exit(1)
	}
}

func mainImpl() error {
	flag.Var(&bias, "bias", "Selects bias setting (17 for 1/7 or 19 for 1/9)")
	flag.Var(&segmentDirection, "segmentDir", "The direction of the segments (normal or reverse")
	flag.Var(&comDirection, "commonDir", "Changes the vertical display direction (normal or reverse)")
	flag.Var(&display, "display", "the Display should be in normal or inverse mode (normal or inverse")
	flag.Var(&regulationResistor, "reg", "Comma-separated list of the regulation ratio of the built-in regulator. (RR0, RR1 or RR2)")
	flag.Parse()

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
		return fmt.Errorf("invalid reset pin name: %s", *resetPin)
	}

	cs := gpioreg.ByName(*csPin)
	if cs == nil {
		return fmt.Errorf("invalid chip select pin name: %s", *csPin)
	}

	if *contrast > 63 || *contrast < 0 {
		return fmt.Errorf("invalid contrast value %d, valid value must between 0-63", *contrast)
	}

	if *startLine > 63 || *startLine < 0 {
		return fmt.Errorf("invalid startLine value %d, valid value must between 0-63", *contrast)
	}

	log.Printf("Creating st7567...")
	dev, err := st7567.New(b, dc, reset, cs, &st7567.Opts{
		Bias:             bias,
		CommonDirection:  comDirection,
		SegmentDirection: segmentDirection,
		Display:          display,
		RegulationRatio:  regulationResistor,
		StartLine:        byte(*startLine),
		Contrast:         byte(*contrast),
	})

	if err != nil {
		return err
	}

	for x := 0; x < st7567.Width; x++ {
		for y := 0; y < st7567.Height; y++ {
			r, g, b, _ := img.At(x, y).RGBA()

			if r == 0 && g == 0 && b == 0 {
				dev.SetPixel(x, y, true)
			}
		}
	}

	return dev.Update()
}
