// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// apa102 writes to a strip of APA102 LED.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices"
	"periph.io/x/periph/devices/apa102"
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
			if p2 := filepath.Join(p, "src/periph.io/x/periph/cmd/apa102", name); access(p2) {
				return p2
			}
		}
	}
	return ""
}

// loadImg loads an image from disk.
func loadImg(name string) (image.Image, error) {
	p := findFile(name)
	if len(p) == 0 {
		return nil, fmt.Errorf("couldn't find file %s", name)
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	log.Printf("Image %s", name)
	return img, nil
}

// resize is a simple but fast nearest neighbor implementation.
//
// If you need something better, please use one of the various high quality
// (slower!) Go packages available on github.
func resize(src image.Image, width, height int) *image.NRGBA {
	srcMax := src.Bounds().Max
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		sY := (y*srcMax.Y + height/2) / height
		for x := 0; x < width; x++ {
			dst.Set(x, y, src.At((x*srcMax.X+width/2)/width, sY))
		}
	}
	return dst
}

func showImage(display devices.Display, img image.Image, sleep time.Duration, loop bool, height int) {
	r := display.Bounds()
	w := r.Dx()
	orig := img.Bounds().Size()
	if height == 0 {
		height = img.Bounds().Dy()
	}
	p := image.Point{}
	now := time.Now()
	img = resize(img, w, height)
	log.Printf("Resizing %dx%d -> %dx%d took %s", orig.X, orig.Y, w, height, time.Since(now))
	now = time.Now()
	for {
		for p.Y = 0; p.Y < height; p.Y++ {
			c := time.After(sleep)
			display.Draw(r, img, p)
			if p.Y == height-1 && !loop {
				log.Printf("done %s", time.Since(now))
				return
			}
			<-c
		}
	}
}

func mainImpl() error {
	verbose := flag.Bool("v", false, "verbose mode")
	spiID := flag.String("spi", "", "SPI port to use")

	numPixels := flag.Int("n", 150, "number of pixels on the strip")
	intensity := flag.Int("l", 127, "light intensity [1-255]")
	temperature := flag.Int("t", 5000, "light temperature in Kelvin [3500-7500]")
	hz := flag.Int("hz", 0, "SPI port speed")
	color := flag.String("color", "208020", "hex encoded color to show")
	imgName := flag.String("img", "", "image to load")
	lineMs := flag.Int("linems", 2, "number of ms to show each line of the image")
	imgLoop := flag.Bool("imgloop", false, "loop the image")
	imgHeight := flag.Int("imgh", 0, "resize the Y axis of the image")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}
	if *intensity > 255 {
		return errors.New("max intensity is 255")
	}
	if *temperature > 65535 {
		return errors.New("max temperature is 65535")
	}
	if _, err := hostInit(); err != nil {
		return err
	}

	// Open the display device.
	s, err := spireg.Open(*spiID)
	if err != nil {
		return err
	}
	defer s.Close()
	if *hz != 0 {
		if err := s.LimitSpeed(int64(*hz)); err != nil {
			return err
		}
	}
	if p, ok := s.(spi.Pins); ok {
		// TODO(maruel): Print where the pins are located.
		log.Printf("Using pins CLK: %s  MOSI: %s  MISO: %s", p.CLK(), p.MOSI(), p.MISO())
	}
	o := apa102.DefaultOpts
	o.NumPixels = *numPixels
	o.Intensity = uint8(*intensity)
	o.Temperature = uint16(*temperature)
	display, err := apa102.New(s, &o)
	if err != nil {
		return err
	}

	// Load an image and make it loop through the pixels.
	if len(*imgName) != 0 {
		img, err := loadImg(*imgName)
		if err != nil {
			return err
		}
		showImage(display, img, time.Duration(*lineMs)*time.Millisecond, *imgLoop, *imgHeight)
		return nil
	}

	// Shows how to create a color array.
	rgb, err := strconv.ParseUint(*color, 16, 32)
	if err != nil {
		return err
	}
	r := byte(rgb >> 16)
	g := byte(rgb >> 8)
	b := byte(rgb)
	buf := make([]byte, *numPixels*3)
	for i := 0; i < len(buf); i += 3 {
		buf[i] = r
		buf[i+1] = g
		buf[i+2] = b
	}
	_, err = display.Write(buf)
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "apa102: %s.\n", err)
		os.Exit(1)
	}
}
