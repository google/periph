// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ssd1306

import (
	"errors"
	"image"
	"image/color"
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spitest"
	"periph.io/x/periph/devices/ssd1306/image1bit"
)

func TestNewI2C_fail(t *testing.T) {
	bus := i2ctest.Playback{DontPanic: true}
	if d, err := NewI2C(&bus, &Opts{H: 64}); d != nil || err == nil {
		t.Fatal(d, err)
	}
	if d, err := NewI2C(&bus, &Opts{W: 64}); d != nil || err == nil {
		t.Fatal(d, err)
	}
	if d, err := NewI2C(&bus, &Opts{W: 64, H: 64, Rotated: true}); d != nil || err == nil {
		t.Fatal(d, err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_ColorModel(t *testing.T) {
	bus := getI2CPlayback()
	dev, err := NewI2C(bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if c := dev.ColorModel(); c != image1bit.BitModel {
		t.Fatal(c)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_String(t *testing.T) {
	bus := getI2CPlayback()
	dev, err := NewI2C(bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	expected := "ssd1360.Dev{playback(60), (128,64)}"
	if s := dev.String(); s != expected {
		t.Fatalf("%q != %q", expected, s)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Draw_VerticalLSD_fast(t *testing.T) {
	// Exercise the fast path.
	buf := make([]byte, 1025)
	buf[0] = i2cData
	buf[23] = 1
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
			// Actual draw buffer.
			{Addr: 0x3c, W: buf},
		},
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	img := image1bit.NewVerticalLSB(dev.Bounds())
	img.Pix[22] = 1
	dev.Draw(dev.Bounds(), img, image.Point{})
	if err := dev.Err(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Halt_Write(t *testing.T) {
	// Exercise the fast path.
	buf := make([]byte, 1025)
	buf[0] = i2cData
	buf[23] = 1
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
			// Halt()
			{Addr: 0x3c, W: []byte{0x0, 0xae}},
			// transparent resume
			{Addr: 0x3c, W: []byte{0x0, 0xaf}},
			// Actual draw buffer.
			{Addr: 0x3c, W: buf},
		},
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	pix := make([]byte, 1024)
	pix[22] = 1
	if n, err := dev.Write(pix); n != len(pix) || err != nil {
		t.Fatal(n, err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Halt_resume_fail(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
			// Halt()
			{Addr: 0x3c, W: []byte{0x0, 0xae}},
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	if n, err := dev.Write(make([]byte, 1024)); n != 0 || !conntest.IsErr(err) {
		t.Fatalf("expected conntest error: %v", err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Write_invalid_size(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
		},
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := dev.Write([]byte{1}); n != 0 || err == nil {
		t.Fatal("expected failure")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Write_fail(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := dev.Write(make([]byte, 1024)); n != 0 || !conntest.IsErr(err) {
		t.Fatalf("expected conntest error: %v", err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Draw_fail(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	dev.Draw(dev.Bounds(), makeGrayCheckboard(dev.Bounds()), image.Point{})
	if err := dev.Err(); !conntest.IsErr(err) {
		t.Fatalf("expected conntest error: %v", err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_DrawGray(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
			// Actual draw buffer.
			{Addr: 0x3c, W: append([]byte{i2cData}, grayCheckboard()...)},
		},
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	dev.Draw(dev.Bounds(), makeGrayCheckboard(dev.Bounds()), image.Point{0, 0})
	if err := dev.Err(); err != nil {
		t.Fatal(err)
	}
	// No-op (skip path).
	dev.Draw(dev.Bounds(), makeGrayCheckboard(dev.Bounds()), image.Point{0, 0})
	if err := dev.Err(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Scroll(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x3c, W: initCmdI2C()},
			// Scroll Left.
			{Addr: 0x3c, W: []byte{0x0, 0x27, 0x0, 0x0, 0x6, 0x7, 0x0, 0xff, 0x2f}},
			// Scroll UpRight.
			{Addr: 0x3c, W: []byte{0x0, 0x29, 0x0, 0x0, 0x6, 0x0, 0x1, 0x2f}},
			// StopScroll.
			{Addr: 0x3c, W: []byte{0x0, 0x2e}},
		},
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if dev.Scroll(Left, FrameRate25, 1, 8) == nil {
		t.Fatal("invalid start")
	}
	if dev.Scroll(Left, FrameRate25, 8, 0) == nil {
		t.Fatal("reversed start and end")
	}
	if dev.Scroll(Left, FrameRate25, 0, 9) == nil {
		t.Fatal("invalid end")
	}
	if err := dev.Scroll(Left, FrameRate25, 0, -1); err != nil {
		t.Fatal(err)
	}
	if err := dev.Scroll(UpRight, FrameRate25, 0, 8); err != nil {
		t.Fatal(err)
	}
	if err := dev.StopScroll(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_SetContrast(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x3c, W: initCmdI2C()},
			{Addr: 0x3c, W: []byte{0x0, 0x81, 0x0}},
			{Addr: 0x3c, W: []byte{0x0, 0x81, 0x7f}},
			{Addr: 0x3c, W: []byte{0x0, 0x81, 0xff}},
		},
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.SetContrast(0); err != nil {
		t.Fatal(err)
	}
	if err := dev.SetContrast(127); err != nil {
		t.Fatal(err)
	}
	if err := dev.SetContrast(255); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Invert_Halt_resume(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x3c, W: initCmdI2C()},
			// Invert(true)
			{Addr: 0x3c, W: []byte{0x0, 0xa7}},
			// Halt()
			{Addr: 0x3c, W: []byte{0x0, 0xae}},
			// transparent resume + Invert(false)
			{Addr: 0x3c, W: []byte{0x0, 0xaf, 0xa6}},
		},
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Invert(true); err != nil {
		t.Fatal(err)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := dev.Invert(false); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_Halt(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x3c, W: initCmdI2C()},
			// Halt()
			{Addr: 0x3c, W: []byte{0x0, 0xae}},
			// transparent resume + StopScroll()
			{Addr: 0x3c, W: []byte{0x0, 0xaf, 0x2e}},
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := dev.StopScroll(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

//

func TestNewSPI_fail(t *testing.T) {
	if d, err := NewSPI(&spitest.Playback{}, nil, &Opts{H: 64}); d != nil || err == nil {
		t.Fatal(d, err)
	}
	if d, err := NewSPI(&configFail{}, nil, &Opts{W: 64, H: 64}); d != nil || err == nil {
		t.Fatal(d, err)
	}
	if d, err := NewSPI(&spitest.Playback{}, gpio.INVALID, &DefaultOpts); d != nil || err == nil {
		t.Fatal(d, err)
	}
	if d, err := NewSPI(&spitest.Playback{}, &failPin{fail: true}, &DefaultOpts); d != nil || err == nil {
		t.Fatal(d, err)
	}
}

func TestSPI_3wire(t *testing.T) {
	// Not supported yet.
	if dev, err := NewSPI(&spitest.Playback{}, nil, &DefaultOpts); dev != nil || err == nil {
		t.Fatal("SPI 3-wire is not supported")
	}
}

func TestSPI_4wire_String(t *testing.T) {
	port := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{{W: getInitCmd(128, 64, false)}},
		},
	}
	dev, err := NewSPI(&port, &gpiotest.Pin{N: "pin1", Num: 42}, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	expected := "ssd1360.Dev{playback, pin1(42), (128,64)}"
	if s := dev.String(); s != expected {
		t.Fatalf("%q != %q", expected, s)
	}
	if err := port.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSPI_4wire_Write_differential(t *testing.T) {
	buf1 := make([]byte, 1024)
	buf1[130] = 1
	buf2 := make([]byte, 128)
	buf2[130-128] = 1
	buf2[131-128] = 2
	port := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				{W: getInitCmd(128, 64, false)},
				{W: buf1},
				// Reset to write only to the first page.
				{W: []byte{0x21, 0x0, 0x7f, 0x22, 0x1, 0x1}},
				{W: buf2},
			},
		},
	}
	dev, err := NewSPI(&port, &gpiotest.Pin{N: "pin1", Num: 42}, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	pix := make([]byte, 1024)
	pix[130] = 1
	if n, err := dev.Write(pix); n != len(pix) || err != nil {
		t.Fatal(n, err)
	}
	pix[131] = 2
	if n, err := dev.Write(pix); n != len(pix) || err != nil {
		t.Fatal(n, err)
	}
	if err := port.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSPI_4wire_Write_differential_fail(t *testing.T) {
	buf1 := make([]byte, 1024)
	buf1[130] = 1
	port := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				{W: getInitCmd(128, 64, false)},
				{W: buf1},
			},
			DontPanic: true,
		},
	}
	dev, err := NewSPI(&port, &gpiotest.Pin{N: "pin1", Num: 42}, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	pix := make([]byte, 1024)
	pix[130] = 1
	if n, err := dev.Write(pix); n != len(pix) || err != nil {
		t.Fatal(n, err)
	}
	pix[131] = 2
	if n, err := dev.Write(pix); n != 0 || !conntest.IsErr(err) {
		t.Fatalf("expected conntest error: %v", err)
	}
	if err := port.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSPI_4wire_gpio_fail(t *testing.T) {
	port := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{{W: getInitCmd(128, 64, false)}},
		},
	}
	pin := &failPin{fail: false}
	dev, err := NewSPI(&port, pin, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	// GPIO suddenly fail.
	pin.fail = true
	if n, err := dev.Write(make([]byte, 1024)); n != 0 || err == nil || err.Error() != "injected error" {
		t.Fatalf("expected gpio error: %v", err)
	}
	if err := dev.Halt(); err == nil || err.Error() != "injected error" {
		t.Fatalf("expected gpio error: %v", err)
	}
	if err := port.Close(); err != nil {
		t.Fatal(err)
	}
}

//

func initCmdI2C() []byte {
	return append([]byte{0}, getInitCmd(128, 64, false)...)
}

var preludeI2C = []byte{
	0x0, 0x21, 0x0, 0x7f, 0x22, 0x0, 0x7,
}

func getI2CPlayback() *i2ctest.Playback {
	return &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Startup initialization.
			{Addr: 0x3c, W: initCmdI2C()},
		},
	}
}

func grayCheckboard() []byte {
	buf := make([]byte, 1024)
	for i := range buf {
		if i&1 == 0 {
			buf[i] = 0xaa
		} else {
			buf[i] = 0x55
		}
	}
	return buf
}

func makeGrayCheckboard(r image.Rectangle) image.Image {
	img := image.NewGray(r)
	c := color.Gray{255}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if (y+x)&1 != 0 {
				img.SetGray(x, y, c)
			}
		}
	}
	return img
}

type configFail struct {
	spitest.Record
}

func (c *configFail) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	return nil, errors.New("injected error")
}

type failPin struct {
	gpiotest.Pin
	fail bool
}

func (f *failPin) Out(l gpio.Level) error {
	if f.fail {
		return errors.New("injected error")
	}
	return nil
}
