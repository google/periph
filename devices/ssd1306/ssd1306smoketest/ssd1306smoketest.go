// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ssd1306smoketest is leveraged by periph-smoketest to verify that two
// SSD1306, one over I²C, one over SPI, can display the same output.
package ssd1306smoketest

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"time"

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
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) (err error) {
	s.delay = 2 * time.Second
	i2cID := f.String("i2c", "", "I²C bus to use")
	spiID := f.String("spi", "", "SPI port to use")
	dcName := f.String("dc", "", "DC pin to use in 4-wire SPI mode")

	w := f.Int("w", 128, "Display width")
	h := f.Int("h", 64, "Display height")
	rotated := f.Bool("rotated", false, "Rotate the displays by 180°")

	record := f.Bool("record", false, "record operation (for playback unit testing)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}

	i2cBus, err2 := i2creg.Open(*i2cID)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := i2cBus.Close(); err == nil {
			err = err2
		}
	}()

	spiPort, err2 := spireg.Open(*spiID)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := spiPort.Close(); err == nil {
			err = err2
		}
	}()

	var dc gpio.PinOut
	if len(*dcName) != 0 {
		dc = gpioreg.ByName(*dcName)
	}
	opts := ssd1306.Opts{W: *w, H: *h, Rotated: *rotated}
	if !*record {
		return s.run(i2cBus, spiPort, dc, &opts)
	}

	i2cRecorder := i2ctest.Record{Bus: i2cBus}
	spiRecorder := spitest.Record{Port: spiPort}
	err = s.run(&i2cRecorder, &spiRecorder, dc, &opts)
	if len(i2cRecorder.Ops) != 0 {
		fmt.Printf("I²C recorder Addr: 0x%02X\n", i2cRecorder.Ops[0].Addr)
	} else {
		fmt.Print("I²C recorder\n")
	}
	for _, op := range i2cRecorder.Ops {
		fmt.Print("  W: ")
		for i, b := range op.W {
			if i != 0 {
				fmt.Print(", ")
			}
			fmt.Printf("0x%02X", b)
		}
		fmt.Print("\n   R: ")
		for i, b := range op.R {
			if i != 0 {
				fmt.Print(", ")
			}
			fmt.Printf("0x%02X", b)
		}
		fmt.Print("\n")
	}
	fmt.Print("\nSPI recorder\n")
	for _, op := range spiRecorder.Ops {
		fmt.Print("  W: ")
		if len(op.R) != 0 {
			// Read data.
			fmt.Printf("0x%02X\n   R: ", op.W[0])
			// first byte is dummy.
			for i, b := range op.R[1:] {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Printf("0x%02X", b)
			}
		} else {
			// Write-only command.
			for i, b := range op.W {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Printf("0x%02X", b)
			}
			fmt.Print("\n   R: ")
		}
		fmt.Print("\n")
	}
	return err
}

func (s *SmokeTest) run(i2cBus i2c.Bus, spiPort spi.PortCloser, dc gpio.PinOut, opts *ssd1306.Opts) (err error) {
	s.timings = make([]time.Duration, 2)
	start := time.Now()
	i2cDev, err2 := ssd1306.NewI2C(i2cBus, opts)
	s.timings[0] = time.Since(start)
	if err2 != nil {
		return err2
	}
	start = time.Now()
	spiDev, err2 := ssd1306.NewSPI(spiPort, dc, opts)
	s.timings[1] = time.Since(start)
	if err2 != nil {
		return err2
	}

	s.devices = []*ssd1306.Dev{i2cDev, spiDev}
	fmt.Printf("%s: Devices:   %v,   %v\n", s, s.devices[0], s.devices[1])
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
	draw.Src.Draw(imgBunny1bitLarge, center.Add(image.Point{X: (opts.W - center.Dx()) / 2}), imgBunny1bit, image.Point{})
	imgClear := make([]byte, opts.W*opts.H/8)

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

	imgStripes := broadStripes(opts.W, opts.H)
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

	imgPattern := binaryPattern(opts.W, opts.H)
	for i, d := range s.devices {
		start := time.Now()
		if _, err := d.Write(imgPattern); err != nil {
			return err
		}
		s.timings[i] = time.Since(start)
	}
	s.step("Fill display with binary 0..255 pattern")

	imgPattern[opts.W+opts.H/2] ^= 0x10
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
	r := bmp.Bounds()
	r.Min = r.Max.Sub(periphImg.Rect.Max)
	draw.DrawMask(bmp, r, &image.Uniform{C: image1bit.On}, image.Point{}, &periphImg, image.Point{}, draw.Over)
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

// image1bit.Bit is not transparent, so it cannot be used with draw.DrawMask().
type bit bool

func (b bit) RGBA() (uint32, uint32, uint32, uint32) {
	if b {
		return 65535, 65535, 65535, 65535
	}
	return 0, 0, 0, 0
}

func convertBit(c color.Color) color.Color {
	r, g, b, _ := c.RGBA()
	return bit((r | g | b) >= 0x8000)
}

type alpha struct {
	image1bit.VerticalLSB
}

func (a *alpha) ColorModel() color.Model {
	return color.ModelFunc(convertBit)
}

func (a *alpha) At(x, y int) color.Color {
	return convertBit(a.VerticalLSB.At(x, y))
}

// periphImg is the text "periph.io\nis awesome !" at the bottom right of a
// 80x24 image encoded as .Pix.
//
// It is encoded here to not have to depend on golang.org/x/image/...
var periphImg = alpha{
	image1bit.VerticalLSB{
		Pix: []byte{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xfc, 0x28, 0x44,
			0x44, 0x44, 0x38, 0, 0x78, 0x94, 0x94, 0x94, 0x94, 0x58, 0, 0x4, 0xf8,
			0x4, 0x4, 0x4, 0x8, 0, 0, 0x80, 0x84, 0xfd, 0x80, 0x80, 0, 0xfc, 0x28,
			0x44, 0x44, 0x44, 0x38, 0, 0xff, 0x8, 0x4, 0x4, 0x4, 0xf8, 0, 0, 0, 0x80,
			0xc0, 0x80, 0, 0, 0, 0x80, 0x84, 0xfd, 0x80, 0x80, 0, 0x78, 0x84, 0x84,
			0x84, 0x84, 0x78, 0, 0, 0, 0, 0, 0, 0x80, 0xa0, 0, 0, 0, 0, 0x80, 0x80,
			0x80, 0x80, 0, 0, 0x3, 0, 0, 0, 0, 0, 0, 0, 0x80, 0x80, 0x80, 0x80, 0, 0,
			0, 0x80, 0, 0, 0, 0x80, 0, 0, 0x80, 0x80, 0x80, 0x80, 0, 0, 0x3, 0x80,
			0x80, 0x80, 0x80, 0, 0, 0, 0x80, 0x80, 0x80, 0x80, 0, 0, 0, 0x80, 0x80,
			0x1, 0x80, 0, 0, 0, 0x80, 0x80, 0x80, 0x80, 0, 0, 0, 0, 0, 0xf0, 0, 0, 0,
			0, 0, 0, 0, 0x10, 0x10, 0x1f, 0x10, 0x10, 0, 0x9, 0x12, 0x12, 0x14, 0x14,
			0x9, 0, 0, 0, 0, 0, 0, 0, 0, 0xc, 0x12, 0x12, 0x12, 0xa, 0x1f, 0, 0, 0xf,
			0x10, 0xe, 0x10, 0xf, 0, 0xf, 0x12, 0x12, 0x12, 0x12, 0xb, 0, 0x9, 0x12,
			0x12, 0x14, 0x14, 0x9, 0, 0xf, 0x10, 0x10, 0x10, 0x10, 0xf, 0, 0, 0x1f,
			0, 0xf, 0, 0x1f, 0, 0xf, 0x12, 0x12, 0x12, 0x12, 0xb, 0, 0, 0, 0, 0x17,
			0, 0, 0,
		},
		Stride: 80,
		Rect:   image.Rectangle{Max: image.Point{80, 24}},
	},
}
