// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// nrzled writes to a strip of LEDs using the NRZ protocol.
//
// This includes the ws2811/ws2812/ws2812b family LEDs and compatible ICs like
// sk6812 and ucs1903.
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

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices"
	"periph.io/x/periph/experimental/devices/nrzled"
	"periph.io/x/periph/host"
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
			// Use the images from cmd/apa102.
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

// resize is a simple but fast nearest neighbordhood implementation.
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
	pin := flag.String("p", "", "GPIO pin to use")

	numPixels := flag.Int("n", nrzled.DefaultOpts.NumPixels, "number of pixels on the strip")
	hz := flag.Int("s", int(nrzled.DefaultOpts.Freq/physic.Hertz), "speed in Hz")
	channels := flag.Int("channels", nrzled.DefaultOpts.Channels, "number of color channels, use 4 for RGBW")
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
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open the display device.
	p := gpioreg.ByName(*pin)
	if p == nil {
		return errors.New("specify a valid pin")
	}
	s, ok := p.(gpiostream.PinOut)
	if !ok {
		return fmt.Errorf("pin %s doesn't support arbitrary bit stream", p)
	}
	opts := nrzled.DefaultOpts
	opts.NumPixels = *numPixels
	opts.Freq = physic.Frequency(*hz) * physic.Hertz
	opts.Channels = *channels
	display, err := nrzled.New(s, &opts)
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
	rgba, err := strconv.ParseUint(*color, 16, 32)
	if err != nil {
		return err
	}
	var r, g, b, a byte
	if *channels == 4 && len(*color) == 8 {
		r = byte(rgba >> 24)
		g = byte(rgba >> 16)
		b = byte(rgba >> 8)
		a = byte(rgba)
	} else {
		r = byte(rgba >> 16)
		g = byte(rgba >> 8)
		b = byte(rgba)
	}
	buf := make([]byte, *numPixels*(*channels))
	for i := 0; i < len(buf); i += *channels {
		buf[i] = r
		buf[i+1] = g
		buf[i+2] = b
		if *channels == 4 {
			buf[i+3] = a
		}
	}
	_, err = display.Write(buf)
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "nrzled: %s.\n", err)
		os.Exit(1)
	}
}
