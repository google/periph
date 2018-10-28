// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package epd_test

import (
	"image"
	"log"

	"periph.io/x/periph/experimental/devices/epd"

	"periph.io/x/periph/conn/spi/spireg"

	"periph.io/x/periph/devices/ssd1306/image1bit"
	"periph.io/x/periph/host"
)

func Example() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use spireg SPI bus registry to find the first available SPI bus.
	b, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	dev, err := epd.NewSPIHat(b, &epd.EPD2in13) // Display config and size
	if err != nil {
		log.Fatalf("failed to initialize epd: %v", err)
	}

	// Draw on it.
	img := image1bit.NewVerticalLSB(dev.Bounds())
	// [start of example 1]
	// Note: this code is commented out so periph does not depend on:
	//    "golang.org/x/image/font"
	//    "golang.org/x/image/font/basicfont"
	//    "golang.org/x/image/math/fixed"
	//
	// f := basicfont.Face7x13
	// drawer := font.Drawer{
	// 	Dst:  img,
	// 	Src:  &image.Uniform{image1bit.On},
	// 	Face: f,
	// 	Dot:  fixed.P(0, img.Bounds().Dy()-1-f.Descent),
	// }
	// drawer.DrawString("Hello from periph!")
	// [end of example 1]

	// [start of example 2]
	// Note: this code is commented out so periph does not depend on:
	//    "github.com/fogleman/gg"
	//    "github.com/golang/freetype/truetype"
	//    "golang.org/x/image/font/gofont/goregular"
	//
	// bounds := dev.Bounds()
	// w := bounds.Dx()
	// h := bounds.Dy()
	// dc := gg.NewContext(w, h)
	// im, err := gg.LoadPNG("gopher.png")
	// if err != nil {
	// 	panic(err)
	// }
	// dc.SetRGB(1, 1, 1)
	// dc.Clear()
	// dc.SetRGB(0, 0, 0)
	// dc.Rotate(gg.Radians(90))
	// dc.Translate(0.0, -float64(h/2))
	// font, err := truetype.Parse(goregular.TTF)
	// if err != nil {
	// 	panic(err)
	// }
	// face := truetype.NewFace(font, &truetype.Options{
	// 	Size: 16,
	// })
	// dc.SetFontFace(face)
	// text := "Hello from periph!"
	// tw, th := dc.MeasureString(text)
	// dc.DrawImage(im, 120, 30)
	// padding := 8.0
	// dc.DrawRoundedRectangle(padding*2, padding*2, tw+padding*2, th+padding, 10)
	// dc.Stroke()
	// dc.DrawString(text, padding*3, padding*2+th)
	// for i := 0; i < 10; i++ {
	// 	dc.DrawCircle(float64(30+(10*i)), 100, 5)
	// }
	// for i := 0; i < 10; i++ {
	// 	dc.DrawRectangle(float64(30+(10*i)), 80, 5, 5)
	// }
	// dc.Fill()
	// [end of example 2]
	if err := dev.Draw(dev.Bounds(), img, image.Point{}); err != nil {
		log.Fatal(err)
	}
	dev.DisplayFrame() // After drawing on the display, you have to show the frame
}
