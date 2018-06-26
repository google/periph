// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package display_test

import (
	"image"
	"log"

	"periph.io/x/periph/conn/display"
	"periph.io/x/periph/host"
)

func ExampleDrawer() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Get a display output device, like an apa102 or ssd1306. For example:
	//   s, _ := spireg.Open("")
	//   d, _ := apa102.New(s, &apa102.DefaultOpts)
	var d display.Drawer

	// Get an image. You could load a PNG. Resize it to the device display size.
	img := image.NewNRGBA(d.Bounds())

	// Render the image. The normal use case is:
	// - Use d.Bounds() as the dstRect, to cover the whole screen.
	// - Use image.ZP/image.Point{} as 'srcPts' unless you want to offset inside
	//   the image.
	if err := d.Draw(d.Bounds(), img, image.ZP); err != nil {
		log.Fatal(err)
	}
}
