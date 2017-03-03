// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

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

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/devices/ssd1306"
	"periph.io/x/periph/devices/ssd1306/image1bit"
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

func demo(s *ssd1306.Dev) error {
	if err := s.Scroll(ssd1306.Left, ssd1306.FrameRate2); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	if err := s.Scroll(ssd1306.Right, ssd1306.FrameRate2); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	if err := s.Scroll(ssd1306.UpLeft, ssd1306.FrameRate2); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	if err := s.Scroll(ssd1306.UpRight, ssd1306.FrameRate2); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	if err := s.StopScroll(); err != nil {
		return err
	}
	if err := s.SetContrast(0); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	if err := s.SetContrast(0xFF); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

// drawText draws text at the bottom right of img.
func drawText(img draw.Image, text string) {
	f := basicfont.Face7x13
	advance := font.MeasureString(f, text).Ceil()
	bounds := img.Bounds()
	if advance > bounds.Dx() {
		advance = 0
	} else {
		advance = bounds.Dx() - advance
	}
	drawer := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: f,
		Dot:  fixed.P(advance, bounds.Dy()-1-f.Descent),
	}
	drawer.DrawString(text)
}

// convert resizes and converts to black and white an image while keeping
// aspect ratio, put it in a centered image of the same size as the display.
func convert(s *ssd1306.Dev, src image.Image) (*image1bit.Image, error) {
	src = resize(src, s.W, s.H)
	img, err := image1bit.New(image.Rect(0, 0, s.W, s.H))
	if err != nil {
		return nil, err
	}
	r := src.Bounds()
	r = r.Add(image.Point{(s.W - r.Max.X) / 2, (s.H - r.Max.Y) / 2})
	draw.Draw(img, r, src, image.Point{}, draw.Src)
	return img, nil
}

// patterns runs a number of test patterns to verify that the basics are working.
func patterns(s *ssd1306.Dev) error {
	// Create synthetic images using a raw array. Each byte corresponds to 8
	// vertical pixels, and then the array scans horizontally and down.
	var img [128 * 64 / 8]byte

	// Fill with broad stripes.
	for y := 0; y < 8; y++ {
		// Horizontal stripes.
		for x := 0; x < 64; x++ {
			img[x+128*y] = byte((y & 1) * 0xff)
		}
		// Vertical stripes.
		for x := 64; x < 128; x++ {
			img[x+128*y] = byte(((x / 8) & 1) * 0xff)
		}
	}
	if _, err := s.Write(img[:]); err != nil {
		return err
	}

	// Display off and back on.
	log.Printf("off & on")
	time.Sleep(500 * time.Millisecond)
	if err := s.Enable(false); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	if err := s.Enable(true); err != nil {
		return err
	}

	// Display inverted and back.
	log.Printf("inverted and back")
	time.Sleep(500 * time.Millisecond)
	if err := s.Invert(true); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	if err := s.Invert(false); err != nil {
		return err
	}

	// Change the contrast around.
	log.Printf("contrast ramp")
	for c := 0; c < 256; c++ {
		if err := s.SetContrast(byte(c)); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.SetContrast(0xff)

	// Fill display with binary 0..255 pattern.
	for i := 0; i < len(img); i++ {
		img[i] = byte(i)
	}
	if _, err := s.Write(img[:]); err != nil {
		return err
	}

	// Display inverted and back.
	log.Printf("inverted and back")
	time.Sleep(500 * time.Millisecond)
	if err := s.Invert(true); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	if err := s.Invert(false); err != nil {
		return err
	}

	return nil
}

func mainImpl() error {
	i2cID := flag.Int("i2c", -1, "specify I²C bus to use")
	spiID := flag.Int("spi", -1, "specify SPI bus to use")
	csID := flag.Int("cs", 0, "specify SPI chip select (CS) to use")
	speed := flag.Int("speed", 0, "specify SPI speed in Hz to use")
	h := flag.Int("h", 64, "display height")
	imgName := flag.String("i", "ballerine.gif", "image to load; try bunny.gif")
	text := flag.String("t", "periph is awesome", "text to display")
	w := flag.Int("w", 128, "display width")
	demoMode := flag.Bool("d", false, "demo scrolling")
	rotated := flag.Bool("r", false, "Rotate the display by 180°")
	pattern := flag.Bool("p", false, "Display test patterns")
	verbose := flag.Bool("v", false, "verbose mode")
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

	// Open the device on the right bus.
	var s *ssd1306.Dev
	if *spiID >= 0 {
		bus, err := spi.New(*spiID, *csID)
		if err != nil {
			return err
		}
		defer bus.Close()
		if *speed != 0 {
			if err := bus.Speed(int64(*speed)); err != nil {
				return err
			}
		}
		if p, ok := bus.(spi.Pins); ok {
			// TODO(maruel): Print where the pins are located.
			log.Printf("Using pins CLK: %s  MOSI: %s  CS: %s", p.CLK(), p.MOSI(), p.CS())
		}
		s, err = ssd1306.NewSPI(bus, *w, *h, *rotated)
		if err != nil {
			return err
		}
	} else {
		bus, err := i2c.New(*i2cID)
		if err != nil {
			return err
		}
		defer bus.Close()
		if p, ok := bus.(i2c.Pins); ok {
			// TODO(maruel): Print where the pins are located.
			log.Printf("Using pins SCL: %s  SDA: %s", p.SCL(), p.SDA())
		}
		s, err = ssd1306.NewI2C(bus, *w, *h, *rotated)
		if err != nil {
			return err
		}
	}

	// Run test patterns, if requested.
	if *pattern {
		if err := patterns(s); err != nil {
			return err
		}
	}

	// Load image.
	src, g, err := loadImg(*imgName)
	if err != nil {
		return err
	}
	// If an animated GIF, draw it in a loop.
	// TODO: this probably shouldn't loop forever...
	if g != nil {
		// Resize all the images up front to save on CPU processing.
		imgs := make([]*image1bit.Image, len(g.Image))
		for i := range g.Image {
			imgs[i], err = convert(s, g.Image[i])
			drawText(imgs[i], *text)
			if err != nil {
				return err
			}
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
		src, err = image1bit.New(image.Rect(0, 0, s.W, s.H))
		if err != nil {
			return err
		}
	}

	img, err := convert(s, src)
	if err != nil {
		return err
	}
	drawText(img, *text)
	s.Draw(img.Bounds(), img, image.Point{})
	if *demoMode {
		if err := demo(s); err != nil {
			return err
		}
	}
	if err := s.Enable(false); err != nil {
		return err
	}
	return err
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "ssd1306: %s.\n", err)
		os.Exit(1)
	}
}
