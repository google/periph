// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ssd1306smoketest is leveraged by periph-smoketest to verify that two
// SSD1306, one over I²C, one over SPI, can display the same output.
package ssd1306smoketest

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/conn/spi/spitest"
	"periph.io/x/periph/devices/ssd1306"
	"periph.io/x/periph/devices/ssd1306/image1bit"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
	delay   time.Duration
	devices []*ssd1306.Dev
	timings []time.Duration
}

func (s *SmokeTest) String() string {
	return s.Name()
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "ssd1306"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests SSD1306 over I²C and SPI by displaying multiple patterns that exercises all code paths"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(args []string) (err error) {
	s.delay = 2 * time.Second

	f := flag.NewFlagSet("buses", flag.ExitOnError)
	i2cName := f.String("i2c", "", "I²C bus to use")
	spiName := f.String("spi", "", "SPI bus to use")
	dcName := f.String("dc", "", "DC pin to use in 4-wire SPI mode")

	w := f.Int("w", 128, "Display width")
	h := f.Int("h", 64, "Display height")
	rotated := f.Bool("rotated", false, "Rotate the displays by 180°")

	record := f.Bool("record", false, "record operation (for playback unit testing)")
	f.Parse(args)

	i2cBus, err2 := i2creg.Open(*i2cName)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := i2cBus.Close(); err == nil {
			err = err2
		}
	}()

	spiBus, err2 := spireg.Open(*spiName)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := spiBus.Close(); err == nil {
			err = err2
		}
	}()

	var dc gpio.PinOut
	if len(*dcName) != 0 {
		dc = gpioreg.ByName(*dcName)
	}
	if !*record {
		return s.run(i2cBus, spiBus, dc, *w, *h, *rotated)
	}

	i2cRecorder := i2ctest.Record{Bus: i2cBus}
	spiRecorder := spitest.Record{Conn: spiBus}
	err = s.run(&i2cRecorder, &spiRecorder, dc, *w, *h, *rotated)
	if len(i2cRecorder.Ops) != 0 {
		fmt.Printf("I²C recorder Addr: 0x%02X\n", i2cRecorder.Ops[0].Addr)
	} else {
		fmt.Print("I²C recorder\n")
	}
	for _, op := range i2cRecorder.Ops {
		fmt.Print("  Write: ")
		for i, b := range op.Write {
			if i != 0 {
				fmt.Print(", ")
			}
			fmt.Printf("0x%02X", b)
		}
		fmt.Print("\n   Read: ")
		for i, b := range op.Read {
			if i != 0 {
				fmt.Print(", ")
			}
			fmt.Printf("0x%02X", b)
		}
		fmt.Print("\n")
	}
	fmt.Print("\nSPI recorder\n")
	for _, op := range spiRecorder.Ops {
		fmt.Print("  Write: ")
		if len(op.Read) != 0 {
			// Read data.
			fmt.Printf("0x%02X\n   Read: ", op.Write[0])
			// first byte is dummy.
			for i, b := range op.Read[1:] {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Printf("0x%02X", b)
			}
		} else {
			// Write-only command.
			for i, b := range op.Write {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Printf("0x%02X", b)
			}
			fmt.Print("\n   Read: ")
		}
		fmt.Print("\n")
	}
	return err
}

func (s *SmokeTest) run(i2cBus i2c.Bus, spiBus spi.ConnCloser, dc gpio.PinOut, w, h int, rotated bool) (err error) {
	s.timings = make([]time.Duration, 2)
	start := time.Now()
	i2cDev, err2 := ssd1306.NewI2C(i2cBus, w, h, rotated)
	s.timings[0] = time.Since(start)
	if err2 != nil {
		return err2
	}
	start = time.Now()
	spiDev, err2 := ssd1306.NewSPI(spiBus, dc, w, h, rotated)
	s.timings[1] = time.Since(start)
	if err2 != nil {
		return err2
	}

	s.devices = []*ssd1306.Dev{i2cDev, spiDev}
	fmt.Printf("%s: Devices:   %s,   %s\n", s, s.devices[0], s.devices[1])
	s.printStr("NewXXX() durations")

	// Preparations.
	imgBunnyNRGBA, err := gif.Decode(bytes.NewReader(bunny))
	if err != nil {
		return err
	}
	// Right format but not the right size.
	imgBunny1bit := image1bit.NewVerticalLSB(imgBunnyNRGBA.Bounds())
	draw.Src.Draw(imgBunny1bit, imgBunnyNRGBA.Bounds(), imgBunnyNRGBA, image.Point{})
	// Right format, right size
	imgBunny1bitLarge := image1bit.NewVerticalLSB(i2cDev.Bounds())
	center := imgBunny1bit.Bounds()
	draw.Src.Draw(imgBunny1bitLarge, center.Add(image.Point{X: (w - center.Dx()) / 2}), imgBunny1bit, image.Point{})
	imgClear := make([]byte, w*h/8)

	for i, d := range s.devices {
		start := time.Now()
		if _, err := d.Write(imgClear); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.printStr("Clear")

	for i, d := range s.devices {
		start := time.Now()
		d.Draw(d.Bounds(), imgBunnyNRGBA, image.Point{})
		if err := d.Err(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Bunny NRGBA")

	for i, d := range s.devices {
		start := time.Now()
		d.Draw(d.Bounds(), imgBunny1bitLarge, image.Point{})
		if err := d.Err(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Bunny image1bit exact frame")

	for i, d := range s.devices {
		start := time.Now()
		d.Draw(d.Bounds(), imgBunny1bit, image.Point{})
		if err := d.Err(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Bunny image1bit partial draw")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Scroll(ssd1306.Left, ssd1306.FrameRate2, 0, -1); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Scroll left: rate = 2")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Scroll(ssd1306.Right, ssd1306.FrameRate25, 0, -1); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Scroll right: rate = 25")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Scroll(ssd1306.UpLeft, ssd1306.FrameRate5, 0, -1); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Scroll up left: rate = 5")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Scroll(ssd1306.UpRight, ssd1306.FrameRate128, 0, -1); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Scroll up right: rate = 128")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Scroll(ssd1306.Left, ssd1306.FrameRate2, 0, 16); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Split scroll top 16 pixels")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Scroll(ssd1306.Right, ssd1306.FrameRate2, 16, -1); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Split scroll 16-64 pixels")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.StopScroll(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Stop scroll")

	for i, d := range s.devices {
		start := time.Now()
		d.Draw(d.Bounds(), imgBunny1bitLarge, image.Point{})
		if err := d.Err(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Redraw")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.SetContrast(0); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Contrast min")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.SetContrast(0xFF); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Contrast max")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Invert(true); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Invert")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Invert(false); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Restore")

	imgStripes := broadStripes(w, h)
	for i, d := range s.devices {
		start := time.Now()
		if _, err := d.Write(imgStripes); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("broad stripes: testing raw Write()")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Halt(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Off")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Invert(false); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("On")

	for i, d := range s.devices {
		start := time.Now()
		if _, err := d.Write(imgClear); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.printStr("Clear")

	for i, d := range s.devices {
		start := time.Now()
		if _, err := d.Write(imgClear); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.printStr("Clear (redundant)")

	imgPattern := binaryPattern(w, h)
	for i, d := range s.devices {
		start := time.Now()
		if _, err := d.Write(imgPattern); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Fill display with binary 0..255 pattern")

	imgPattern[w+h/2] ^= 0x10
	for i, d := range s.devices {
		start := time.Now()
		if _, err := d.Write(imgPattern); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Update a single pixel on second band")

	bmp := image1bit.NewVerticalLSB(i2cDev.Bounds())
	copy(bmp.Pix, imgPattern)
	drawText(bmp, "periph.io", 1)
	drawText(bmp, "is awesome!", 0)
	for i, d := range s.devices {
		start := time.Now()
		d.Draw(d.Bounds(), bmp, image.Point{})
		if err := d.Err(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Draw text")

	for i, d := range s.devices {
		start := time.Now()
		if err := d.Halt(); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Halt")
	return nil
}

func (s *SmokeTest) printStr(str string) {
	fmt.Printf("%s: %-50s:", s, str)
	for i, t := range s.timings {
		if i != 0 {
			fmt.Print(",")
		}
		fmt.Printf(" %s", round(t))
	}
	fmt.Print("\n")
}

func (s *SmokeTest) step(str string) {
	s.printStr(str)
	time.Sleep(s.delay)
}

// broadStripes() returns an image using a raw array. Each byte corresponds to 8
// vertical pixels, and then the array scans horizontally and down.
func broadStripes(w, h int) []byte {
	img := make([]byte, w*h/8)
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
	return img
}

func binaryPattern(w, h int) []byte {
	img := make([]byte, w*h/8)
	for i := 0; i < len(img); i++ {
		img[i] = 0
	}
	for i := 0; i < 256; i++ {
		offset := i % w
		band := ((i / w) * w) * 2
		img[band+offset] = byte(i)
	}
	return img
}

// round returns the duration rounded in µs.
func round(d time.Duration) string {
	µs := (d + time.Microsecond/2) / time.Microsecond
	ms := µs / 1000
	µs %= 1000
	return fmt.Sprintf("%3d.%03dms", ms, µs)
}

// drawText draws text at the bottom right of img.
func drawText(img draw.Image, text string, lastToBottom int) {
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
		Dot:  fixed.P(advance, bounds.Dy()-1-f.Descent-lastToBottom*f.Height),
	}
	drawer.DrawString(text)
}
