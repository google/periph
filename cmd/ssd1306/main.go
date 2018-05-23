// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

//go:generate go run gen.go

// ssd1306 writes to a display driven by a ssd1306 controler.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/ssd1306"
	"periph.io/x/periph/devices/ssd1306/image1bit"
)

func access(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func findFile(name string) string {
	if access(name) {
		return name
	}
	for _, p := range strings.Split(os.Getenv("GOPATH"), ":") {
		if len(p) != 0 {
			if p2 := filepath.Join(p, "src/periph.io/x/periph/cmd/ssd1306", name); access(p2) {
				return p2
			}
		}
	}
	return ""
}

// loadImg loads an image from disk.
func loadImg(name string) (image.Image, *gif.GIF, error) {
	p := findFile(name)
	if len(p) == 0 {
		return nil, nil, fmt.Errorf("couldn't find file %s", name)
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// Try to decode as an animated GIF first, then fall back to generic decoding.
	if g, err := gif.DecodeAll(f); err == nil {
		if len(g.Image) > 1 {
			log.Printf("Image %s as animated GIF", name)
			return nil, g, nil
		}
		log.Printf("Image %s", name)
		return g.Image[0], nil, nil
	}
	if _, err = f.Seek(0, 0); err != nil {
		return nil, nil, err
	}
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, nil, err
	}
	log.Printf("Image %s", name)
	return img, nil, nil
}

// resize is a simple but fast nearest neighbor implementation.
//
// If you need something better, please use one of the various high quality
// (slower!) Go packages available on github.
func resize(src image.Image, size image.Point) *image.NRGBA {
	srcMax := src.Bounds().Max
	dst := image.NewNRGBA(image.Rectangle{Max: size})
	for y := 0; y < size.Y; y++ {
		sY := (y*srcMax.Y + size.Y/2) / size.Y
		for x := 0; x < size.X; x++ {
			dst.Set(x, y, src.At((x*srcMax.X+size.X/2)/size.X, sY))
		}
	}
	return dst
}

// drawTextBottomRight draws text at the bottom right of img.
func drawTextBottomRight(img draw.Image, text string) {
	advance := utf8.RuneCountInString(text) * 7
	bounds := img.Bounds()
	if advance > bounds.Dx() {
		advance = 0
	} else {
		advance = bounds.Dx() - advance
	}
	drawText(img, image.Point{advance, bounds.Dy() - 1 - 13}, text)
}

// convert resizes and converts to black and white an image while keeping
// aspect ratio, put it in a centered image of the same size as the display.
func convert(s *ssd1306.Dev, src image.Image) *image1bit.VerticalLSB {
	screenBounds := s.Bounds()
	size := screenBounds.Size()
	src = resize(src, size)
	img := image1bit.NewVerticalLSB(screenBounds)
	r := src.Bounds()
	r = r.Add(image.Point{(size.X - r.Max.X) / 2, (size.Y - r.Max.Y) / 2})
	draw.Draw(img, r, src, image.Point{}, draw.Src)
	return img
}

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use")
	spiID := flag.String("spi", "", "SPI port to use")
	dcName := flag.String("dc", "", "DC pin to use in 4-wire SPI mode")
	hz := flag.Int("hz", 0, "I²C bus/SPI port speed")

	h := flag.Int("h", 64, "display height")
	w := flag.Int("w", 128, "display width")
	rotated := flag.Bool("r", false, "Rotate the display by 180°")

	imgName := flag.String("i", "ballerine.gif", "image to load; try bunny.gif")
	text := flag.String("t", "periph is awesome", "text to display")

	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

	if _, err := hostInit(); err != nil {
		return err
	}

	// Open the device on the right bus.
	var s *ssd1306.Dev
	opts := ssd1306.Opts{W: *w, H: *h, Rotated: *rotated}
	if *spiID != "" {
		c, err := spireg.Open(*spiID)
		if err != nil {
			return err
		}
		defer c.Close()
		if *hz != 0 {
			if err := c.LimitSpeed(int64(*hz)); err != nil {
				return err
			}
		}
		if p, ok := c.(spi.Pins); ok {
			// TODO(maruel): Print where the pins are located.
			log.Printf("Using pins CLK: %s  MOSI: %s  CS: %s", p.CLK(), p.MOSI(), p.CS())
		}
		var dc gpio.PinOut
		if len(*dcName) != 0 {
			dc = gpioreg.ByName(*dcName)
		}
		s, err = ssd1306.NewSPI(c, dc, &opts)
		if err != nil {
			return err
		}
	} else {
		c, err := i2creg.Open(*i2cID)
		if err != nil {
			return err
		}
		defer c.Close()
		if *hz != 0 {
			if err := c.SetSpeed(int64(*hz)); err != nil {
				return err
			}
		}
		if p, ok := c.(i2c.Pins); ok {
			// TODO(maruel): Print where the pins are located.
			log.Printf("Using pins SCL: %s  SDA: %s", p.SCL(), p.SDA())
		}
		s, err = ssd1306.NewI2C(c, &opts)
		if err != nil {
			return err
		}
	}

	// Load image.
	src, g, err := loadImg(*imgName)
	if err != nil {
		return err
	}
	// If an animated GIF, draw it in a loop.
	if g != nil {
		// Resize all the images up front to save on CPU processing.
		imgs := make([]*image1bit.VerticalLSB, len(g.Image))
		for i := range g.Image {
			imgs[i] = convert(s, g.Image[i])
			drawTextBottomRight(imgs[i], *text)
		}
		for i := 0; g.LoopCount <= 0 || i < g.LoopCount*len(g.Image); i++ {
			index := i % len(g.Image)
			c := time.After(time.Duration(10*g.Delay[index]) * time.Millisecond)
			img := imgs[index]
			s.Draw(img.Bounds(), img, image.Point{})
			<-c
		}
		return nil
	}

	if src == nil {
		// Create a blank image.
		src = image1bit.NewVerticalLSB(s.Bounds())
	}

	img := convert(s, src)
	drawTextBottomRight(img, *text)
	s.Draw(img.Bounds(), img, image.Point{})
	return s.Halt()
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "ssd1306: %s.\n", err)
		os.Exit(1)
	}
}
