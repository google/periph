// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package apa102_test

import (
	"image"
	"image/color"
	"log"

	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/apa102"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use spireg SPI port registry to find the first available SPI bus.
	p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// Opens a 300 lights strip at 50% intensity with color temperature at
	// 3500°Kelvin.
	o := apa102.DefaultOpts
	o.NumPixels = 300
	o.Intensity = 127
	o.Temperature = 3500
	dev, err := apa102.New(p, &o)
	if err != nil {
		log.Fatalf("failed to open apa102: %v", err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, dev.Bounds().Dy(), 1))
	for x := 0; x < img.Rect.Max.X; x++ {
		img.SetNRGBA(x, 0, color.NRGBA{uint8(x), uint8(255 - x), 0, 255})
	}
	dev.Draw(dev.Bounds(), img, image.Point{})
}
