// Copyright 2016 Google Inc. All rights reserved.
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

	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/conn/spi"
	"github.com/google/pio/devices"
	"github.com/google/pio/devices/apa102"
	"github.com/google/pio/experimental/devices/bitbang"
	"github.com/google/pio/host"
	"github.com/nfnt/resize"
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
			if p2 := filepath.Join(p, "src/github.com/google/pio/cmd/apa102", name); access(p2) {
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

func showImage(display devices.Display, img image.Image, sleep time.Duration, loop bool, height int) {
	r := display.Bounds()
	w := r.Dx()
	orig := img.Bounds().Size()
	if height == 0 {
		height = img.Bounds().Dy()
	}
	p := image.Point{}
	now := time.Now()
	img = resize.Resize(uint(w), uint(height), img, resize.Bilinear)
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
	busNumber := flag.Int("b", -1, "SPI bus to use")
	clk := flag.Int("c", -1, "clk pin for bitbanging")
	mosi := flag.Int("m", -1, "mosi pin for bitbanging")

	numLights := flag.Int("n", 150, "number of lights on the strip")
	intensity := flag.Int("l", 127, "light intensity [1-255]")
	temperature := flag.Int("t", 5000, "light temperature in °Kelvin [3500-7500]")
	speed := flag.Int("s", 0, "speed in Hz")
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
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open the display device.
	var bus spi.Conn
	if *clk != -1 && *mosi != -1 {
		pclk := gpio.ByNumber(*clk)
		pmosi := gpio.ByNumber(*mosi)
		b, err := bitbang.NewSPI(pclk, pmosi, nil, nil, int64(*speed))
		if err != nil {
			return err
		}
		bus = b
	} else {
		b, err := spi.New(*busNumber, 0)
		if err != nil {
			return err
		}
		defer b.Close()
		if *speed != 0 {
			if err := b.Speed(int64(*speed)); err != nil {
				return err
			}
		}
		bus = b
	}
	if p, ok := bus.(spi.Pins); ok {
		// TODO(maruel): Print where the pins are located.
		log.Printf("Using pins CLK: %s  MOSI: %s  MISO: %s", p.CLK(), p.MOSI(), p.MISO())
	}
	var err error
	display, err := apa102.New(bus, *numLights, uint8(*intensity), uint16(*temperature))
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
	buf := make([]byte, *numLights*3)
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
