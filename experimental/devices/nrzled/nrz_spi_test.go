// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package nrzled

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spitest"
)

// ToRGB converts a slice of color.NRGBA to a byte stream of RGB pixels.
//
// Ignores alpha.
func ToRGB(p []color.NRGBA) []byte {
	b := make([]byte, 0, len(p)*3)
	for _, c := range p {
		b = append(b, c.R, c.G, c.B)
	}
	return b
}

func TestDevEmpty(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 150}
	o.NumPixels = 0
	d, _ := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if n, err := d.Write([]byte{}); n != 0 || err != nil {
		t.Fatalf("%d %v", n, err)
	}
	if expected := []byte{0x0, 0x0, 0x0}; !bytes.Equal(expected, buf.Bytes()) {
		t.Fatalf("\nGot:  %#02v\nWant: %#02v\n", buf.Bytes(), expected)
	}
	if got, expected := d.String(), "nrzled: {0, recordraw}"; got != expected {
		t.Fatalf("\nGot:  %s\nWant: %s\n", got, expected)
	}
}

func TestConnectFail(t *testing.T) {
	if d, err := NewSPI(&configFail{}, &Opts{NumPixels: 150}); d != nil || err == nil {
		t.Fatal("Connect() call have failed")
	}
}

func TestDevLen(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 150}
	o.NumPixels = 1
	d, _ := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if n, err := d.Write([]byte{0}); n != 0 || err == nil {
		t.Fatalf("%d %v", n, err)
	}
	if expected := []byte{}; !bytes.Equal(expected, buf.Bytes()) {
		t.Fatalf("\nGot:  %#02v\nWant: %#02v\n", buf.Bytes(), expected)
	}
}

var writeTests = []struct {
	name   string
	pixels []byte
	want   []byte
	opts   Opts
}{
	{
		name: "1 pixel to #FFFFFF",
		pixels: ToRGB([]color.NRGBA{
			{0xFF, 0xFF, 0xFF, 0x00},
		}),
		want: []byte{
			/*FF*/ 0xEE, 0xEE, 0xEE, 0xEE /*FF*/, 0xEE, 0xEE, 0xEE, 0xEE /*FF*/, 0xEE, 0xEE, 0xEE, 0xEE,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "1 pixel to #FEFEFE",
		pixels: ToRGB([]color.NRGBA{
			{0xFE, 0xFE, 0xFE, 0x00},
		}),
		want: []byte{
			/*FE*/ 0xEE, 0xEE, 0xEE, 0xE8 /*FE*/, 0xEE, 0xEE, 0xEE, 0xE8 /*FE*/, 0xEE, 0xEE, 0xEE, 0xE8,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "1 pixel to #F0F0F0",
		pixels: ToRGB([]color.NRGBA{
			{0xF0, 0xF0, 0xF0, 0x00},
		}),
		want: []byte{
			/*F0*/ 0xEE, 0xEE, 0x88, 0x88 /*F0*/, 0xEE, 0xEE, 0x88, 0x88 /*F0*/, 0xEE, 0xEE, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "1 pixel to #808080",
		pixels: ToRGB([]color.NRGBA{
			{0x80, 0x80, 0x80, 0x00},
		}),
		want: []byte{
			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "1 pixel to #80FF00",
		pixels: ToRGB([]color.NRGBA{
			{0x80, 0xFF, 0x00, 0x00},
		}),
		want: []byte{
			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*FF*/, 0xEE, 0xEE, 0xEE, 0xEE /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "1 pixel to #800000",
		pixels: ToRGB([]color.NRGBA{
			{0x80, 0x00, 0x00, 0x00},
		}),
		want: []byte{
			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "1 pixel to #008000",
		pixels: ToRGB([]color.NRGBA{
			{0x00, 0x80, 0x00, 0x00},
		}),
		want: []byte{
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "1 pixel to #000080",
		pixels: ToRGB([]color.NRGBA{
			{0x00, 0x00, 0x80, 0x00},
		}),
		want: []byte{
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
		},
	},
	{
		name: "All at once",
		pixels: ToRGB([]color.NRGBA{
			{0xFF, 0xFF, 0xFF, 0x00},
			{0xFE, 0xFE, 0xFE, 0x00},
			{0xF0, 0xF0, 0xF0, 0x00},
			{0x80, 0x80, 0x80, 0x00},

			{0x80, 0x00, 0x00, 0x00},
			{0x00, 0x80, 0x00, 0x00},
			{0x00, 0x00, 0x80, 0x00},

			{0x00, 0x00, 0x10, 0x00},
			{0x00, 0x00, 0x01, 0x00},
			{0x00, 0x00, 0x00, 0x00},
		}),
		want: []byte{
			/*FF*/ 0xEE, 0xEE, 0xEE, 0xEE /*FF*/, 0xEE, 0xEE, 0xEE, 0xEE /*FF*/, 0xEE, 0xEE, 0xEE, 0xEE,
			/*FE*/ 0xEE, 0xEE, 0xEE, 0xE8 /*FE*/, 0xEE, 0xEE, 0xEE, 0xE8 /*FE*/, 0xEE, 0xEE, 0xEE, 0xE8,
			/*F0*/ 0xEE, 0xEE, 0x88, 0x88 /*F0*/, 0xEE, 0xEE, 0x88, 0x88 /*F0*/, 0xEE, 0xEE, 0x88, 0x88,
			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88,

			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88,

			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*10*/, 0x88, 0x8E, 0x88, 0x88,
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*01*/, 0x88, 0x88, 0x88, 0x8E,
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 10,
		},
	},
}

func TestWrites(t *testing.T) {
	for _, tt := range writeTests {
		buf := bytes.Buffer{}
		tt.opts.NumPixels = len(tt.pixels) / 3
		d, _ := NewSPI(spitest.NewRecordRaw(&buf), &tt.opts)
		n, err := d.Write(tt.pixels)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(tt.pixels) {
			t.Fatalf("%s: Got %d bytes result, want %d", tt.name, n, len(tt.pixels)*3)
		}
		if got := buf.Bytes(); !bytes.Equal(got, tt.want) {
			t.Logf("%s:\nGot:  (%d)%#02v\nWant: (%d)%#02v\n", tt.name, len(got), got, len(tt.want), tt.want)
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Logf("(%d) Got:  %#02v\tWant: %#02v\n", i, got[i], tt.want[i])
				}
			}
			t.Fatal("expectation failure")
		}
	}
}

func TestDevColor(t *testing.T) {
	if c := (&Dev{}).ColorModel(); c != color.NRGBAModel {
		t.Fatal(c)
	}
}

func TestDevLong(t *testing.T) {
	buf := bytes.Buffer{}
	colors := make([]color.NRGBA, 256)
	o := Opts{NumPixels: 150}
	o.NumPixels = len(colors)
	d, _ := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if n, err := d.Write(ToRGB(colors)); n != len(colors)*3 || err != nil {
		t.Fatalf("%d %v", n, err)
	}
	expected := make([]byte, 12*o.NumPixels+3)
	for i := 0; i < 12*o.NumPixels; i += 12 {
		//Each channel should be 0x00
		for j := 0; j < 12; j++ {
			expected[i+j] = 0x88
		}
	}
	trailer := expected[12*o.NumPixels:]
	for i := range trailer {
		trailer[i] = 0x00
	}
	if !bytes.Equal(expected, buf.Bytes()) {
		t.Fatalf("\nGot:  %#02v\nWant: %#02v\n", buf.Bytes(), expected)
	}
}

func TestDevWrite_Long(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 150}
	o.NumPixels = 1
	d, _ := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if n, err := d.Write([]byte{0, 0, 0, 1, 1, 1}); n != 0 || err == nil {
		t.Fatal(n, err)
	}
}

var drawTests = []struct {
	name string
	img  image.Image
	want []byte
	opts Opts
}{
	{
		name: "Draw NRGBA",
		img: func() image.Image {
			im := image.NewNRGBA(image.Rect(0, 0, 4, 1))
			for i := 0; i < 4; i++ {
				im.Pix[4*i+0] = 0x00
				im.Pix[4*i+1] = 0x80
				im.Pix[4*i+2] = 0xFF
				im.Pix[4*i+3] = 0
			}
			return im
		}(),
		want: func() []byte {
			var b []byte
			for i := 0; i < 4; i++ {
				b = append(b, 0x88, 0x88, 0x88, 0x88) //0x00
				b = append(b, 0xE8, 0x88, 0x88, 0x88) //0x80
				b = append(b, 0xEE, 0xEE, 0xEE, 0xEE) //0xFF
			}
			for i := 0; i < 3; i++ {
				b = append(b, 0x00)
			}
			return b
		}(),
		opts: Opts{
			NumPixels: 4,
		},
	},
	{
		name: "Draw Empty",
		img: func() image.Image {
			im := image.NewNRGBA(image.Rect(0, 0, 0, 0))
			return im
		}(),
		want: func() []byte {
			var b []byte
			return b
		}(),
		opts: Opts{
			NumPixels: 4,
		},
	},
}

func TestDraws(t *testing.T) {
	for _, tt := range drawTests {
		buf := bytes.Buffer{}
		d, _ := NewSPI(spitest.NewRecordRaw(&buf), &tt.opts)
		if err := d.Draw(d.Bounds(), tt.img, image.Point{}); err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		got := buf.Bytes()
		if !bytes.Equal(got, tt.want) {
			t.Logf("%s:\nGot:  (%d)%#02v\nWant: (%d)%#02v\n", tt.name, len(got), got, len(tt.want), tt.want)
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Logf("(%d) Got:  %#02v\tWant: %#02v\n", i, got[i], tt.want[i])
				}
			}
			t.Fatal("expectation failure")
		}
	}
}

var offsetDrawWant = []byte{
	0x00, 0x00, 0x00, 0x00,
	0xE1, 0x89, 0x79, 0x6B,
	0xE1, 0x9A, 0x88, 0x75,
	0xE1, 0xAD, 0x98, 0x82,
	0xE1, 0xC2, 0xAB, 0x92,
	0xE1, 0xDA, 0xC0, 0xA4,
	0xE1, 0xF5, 0xD9, 0xB9,
	0xE2, 0x89, 0x7A, 0x69,
	0xE2, 0x9A, 0x8A, 0x76,
	0xE2, 0xAC, 0x9B, 0x86,
	0xE2, 0xC0, 0xAE, 0x98,
	0xE2, 0xD5, 0xC3, 0xAC,
	0xE2, 0xED, 0xDA, 0xC2,
	0xE4, 0x83, 0x7A, 0x6E,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0xFF,
}

var offsetDrawTests = []struct {
	name   string
	img    image.Image
	point  image.Point
	offset image.Rectangle
	want   []byte
	opts   Opts
}{
	{
		name: "Offset Draw NRGBA",
		img: func() image.Image {
			im := image.NewNRGBA(image.Rect(0, 0, 16, 4))
			for x := 0; x < 16; x++ {
				for y := 0; y < 4; y++ {
					i := (y*16 + x) * 3
					im.Set(x, y, color.RGBA{R: uint8(i + 1), G: uint8(i + 2), B: uint8(i + 3), A: 0xFF})
				}
			}
			return im
		}(),
		point:  image.Point{X: 3, Y: 2},
		offset: image.Rect(0, 0, 16, 1),
		want:   offsetDrawWant,
		opts: Opts{
			NumPixels: 15,
		},
	},
	{
		name: "Both Offset Draw NRGBA",
		img: func() image.Image {
			im := image.NewNRGBA(image.Rect(0, 0, 16, 4))
			for x := 0; x < 16; x++ {
				for y := 0; y < 4; y++ {
					i := (y*16 + x) * 3
					im.Set(x, y, color.RGBA{R: uint8(i + 1), G: uint8(i + 2), B: uint8(i + 3), A: 0xFF})
				}
			}
			return im
		}(),
		point:  image.Point{X: 3, Y: 2},
		offset: image.Rect(2, 0, 16, 1),
		want: []byte{
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0xE1, 0x89, 0x79, 0x6B,
			0xE1, 0x9A, 0x88, 0x75,
			0xE1, 0xAD, 0x98, 0x82,
			0xE1, 0xC2, 0xAB, 0x92,
			0xE1, 0xDA, 0xC0, 0xA4,
			0xE1, 0xF5, 0xD9, 0xB9,
			0xE2, 0x89, 0x7A, 0x69,
			0xE2, 0x9A, 0x8A, 0x76,
			0xE2, 0xAC, 0x9B, 0x86,
			0xE2, 0xC0, 0xAE, 0x98,
			0xE2, 0xD5, 0xC3, 0xAC,
			0xE2, 0xED, 0xDA, 0xC2,
			0xE4, 0x83, 0x7A, 0x6E,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0xFF, 0xFF,
		},
		opts: Opts{
			NumPixels: 17,
		},
	},
}

func TestHalt(t *testing.T) {
	s := spitest.Playback{
		Playback: conntest.Playback{
			DontPanic: false,
			Count:     1,
			Ops: []conntest.IO{
				{},
				{W: []byte{
					0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88,
					0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88,
					0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88,
					0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88, 0x88,
					//End of frame
					0x00, 0x00, 0x00,
				}},
			},
		},
	}
	o := Opts{NumPixels: 150}
	o.NumPixels = 4
	d, _ := NewSPI(&s, &o)
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

type genColor func(int) [3]byte

func benchmarkWrite(b *testing.B, o Opts, length int, f genColor) {
	var pixels []byte
	for i := 0; i < length; i++ {
		c := f(i)
		pixels = append(pixels, c[:]...)
	}
	o.NumPixels = length
	b.ReportAllocs()
	d, _ := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &o)
	_, _ = d.Write(pixels[:])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Write(pixels[:])
	}
}

func BenchmarkWriteWhite(b *testing.B) {
	o := Opts{NumPixels: 150}
	benchmarkWrite(b, o, 150, func(i int) [3]byte { return [3]byte{0xFF, 0xFF, 0xFF} })
}

func BenchmarkWriteDim(b *testing.B) {
	o := Opts{NumPixels: 150}
	benchmarkWrite(b, o, 150, func(i int) [3]byte { return [3]byte{0x01, 0x01, 0x01} })
}

func BenchmarkWriteBlack(b *testing.B) {
	o := Opts{NumPixels: 150}
	benchmarkWrite(b, o, 150, func(i int) [3]byte { return [3]byte{0x0, 0x0, 0x0} })
}

func genColorfulPixel(x int) [3]byte {
	i := x * 3
	return [3]byte{uint8(i) + uint8(i>>8),
		uint8(i+1) + uint8(i+1>>8),
		uint8(i+2) + uint8(i+2>>8),
	}
}

func BenchmarkWriteColorful(b *testing.B) {
	o := Opts{NumPixels: 150}
	benchmarkWrite(b, o, 150, genColorfulPixel)
}

func BenchmarkWriteColorfulPassThru(b *testing.B) {
	o := Opts{
		NumPixels: 150,
	}
	benchmarkWrite(b, o, 150, genColorfulPixel)
}

func BenchmarkWriteColorfulVariation(b *testing.B) {
	// Continuously vary the lookup tables.
	b.ReportAllocs()
	pixels := [256 * 3]byte{}
	for i := range pixels {
		pixels[i] = uint8(i) + uint8(i>>8)
	}
	o := Opts{NumPixels: 150}
	o.NumPixels = len(pixels) / 3
	d, _ := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &o)
	_, _ = d.Write(pixels[:])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = d.Write(pixels[:])
	}
}

func benchmarkDraw(b *testing.B, o Opts, img draw.Image, f genColor) {
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			pix := f(x)
			c := color.NRGBA{R: pix[0], G: pix[1], B: pix[2], A: 255}
			img.Set(x, y, c)
		}
	}
	o.NumPixels = img.Bounds().Max.X
	b.ReportAllocs()
	d, _ := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &o)
	r := d.Bounds()
	p := image.Point{}
	if err := d.Draw(r, img, p); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.Draw(r, img, p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDrawNRGBAColorful(b *testing.B) {
	o := Opts{NumPixels: 150}
	benchmarkDraw(b, o, image.NewNRGBA(image.Rect(0, 0, 150, 1)), genColorfulPixel)
}

func BenchmarkDrawNRGBAWhite(b *testing.B) {
	o := Opts{NumPixels: 150}
	benchmarkDraw(b, o, image.NewNRGBA(image.Rect(0, 0, 150, 1)), func(i int) [3]byte { return [3]byte{0xFF, 0xFF, 0xFF} })
}

func BenchmarkDrawRGBAColorful(b *testing.B) {
	o := Opts{NumPixels: 150}
	benchmarkDraw(b, o, image.NewRGBA(image.Rect(0, 0, 256, 1)), genColorfulPixel)
}

func BenchmarkDrawSlowpath(b *testing.B) {
	// Should be an image type that doesn't have a fast path
	img := image.NewGray(image.Rect(0, 0, 150, 1))
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			pix := genColorfulPixel(x)
			img.Set(x, y, color.Gray{pix[0]})
		}
	}
	o := Opts{NumPixels: 150}
	o.NumPixels = img.Bounds().Max.X
	b.ReportAllocs()
	d, _ := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &o)
	r := d.Bounds()
	p := image.Point{}
	if err := d.Draw(r, img, p); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.Draw(r, img, p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHalt(b *testing.B) {
	b.ReportAllocs()
	d, _ := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &Opts{NumPixels: 150})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.Halt(); err != nil {
			b.Fatal(err)
		}
	}
}

//

type configFail struct {
	spitest.Record
}

func (c *configFail) Connect(f physic.Frequency, mode spi.Mode, bits int) (spi.Conn, error) {
	return nil, errors.New("injected error")
}

func equalUint16(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
