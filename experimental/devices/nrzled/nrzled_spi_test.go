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

// toRGB converts a slice of color.NRGBA to a byte stream of RGB pixels.
//
// Ignores alpha.
func toRGB(p []color.NRGBA) []byte {
	b := make([]byte, 0, len(p)*3)
	for _, c := range p {
		b = append(b, c.R, c.G, c.B)
	}
	return b
}

func TestSPI_Empty(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 0, Channels: 3, Freq: 2500 * physic.KiloHertz}
	s := spitest.Playback{
		Playback: conntest.Playback{
			Count: 1,
			Ops:   []conntest.IO{{W: []byte{0x00, 0x00, 0x00}}},
		},
	}
	d, err := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if err != nil {
		t.Fatal(err)
	}
	if got, expected := d.String(), "nrzled{recordraw}"; got != expected {
		t.Fatalf("\nGot:  %s\nWant: %s\n", got, expected)
	}

	if n, err := d.Write([]byte{}); n != 0 || err != nil {
		t.Fatalf("%d %v", n, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSPI_fail(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 1, Channels: 3, Freq: 1 * physic.KiloHertz}
	if _, err := NewSPI(spitest.NewRecordRaw(&buf), &o); err == nil {
		t.Fatal("invalid Freq")
	}

	o = Opts{NumPixels: 1, Channels: 0, Freq: 2500 * physic.KiloHertz}
	if _, err := NewSPI(spitest.NewRecordRaw(&buf), &o); err == nil {
		t.Fatal("invalid Channels")
	}

	o = Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	if d, err := NewSPI(&configFail{}, &o); d != nil || err == nil {
		t.Fatal("Connect() call have failed")
	}

	o = Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	if d, err := NewSPI(&limitLow{}, &o); d != nil || err == nil {
		t.Fatal("MaxTxSize() is too small")
	}
}

func TestSPI_Len(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 1, Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if err != nil {
		t.Fatal(err)
	}
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
		pixels: toRGB([]color.NRGBA{
			{0xFF, 0xFF, 0xFF, 0x00},
		}),
		want: []byte{
			/*FF*/ 0xEE, 0xEE, 0xEE, 0xEE /*FF*/, 0xEE, 0xEE, 0xEE, 0xEE /*FF*/, 0xEE, 0xEE, 0xEE, 0xEE,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "1 pixel to #FEFEFE",
		pixels: toRGB([]color.NRGBA{
			{0xFE, 0xFE, 0xFE, 0x00},
		}),
		want: []byte{
			/*FE*/ 0xEE, 0xEE, 0xEE, 0xE8 /*FE*/, 0xEE, 0xEE, 0xEE, 0xE8 /*FE*/, 0xEE, 0xEE, 0xEE, 0xE8,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "1 pixel to #F0F0F0",
		pixels: toRGB([]color.NRGBA{
			{0xF0, 0xF0, 0xF0, 0x00},
		}),
		want: []byte{
			/*F0*/ 0xEE, 0xEE, 0x88, 0x88 /*F0*/, 0xEE, 0xEE, 0x88, 0x88 /*F0*/, 0xEE, 0xEE, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "1 pixel to #808080",
		pixels: toRGB([]color.NRGBA{
			{0x80, 0x80, 0x80, 0x00},
		}),
		want: []byte{
			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "1 pixel to #80FF00",
		pixels: toRGB([]color.NRGBA{
			{0x80, 0xFF, 0x00, 0x00},
		}),
		want: []byte{
			/*FF*/ 0xEE, 0xEE, 0xEE, 0xEE /*80*/, 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "1 pixel to #800000",
		pixels: toRGB([]color.NRGBA{
			{0x80, 0x00, 0x00, 0x00},
		}),
		want: []byte{
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "1 pixel to #008000",
		pixels: toRGB([]color.NRGBA{
			{0x00, 0x80, 0x00, 0x00},
		}),
		want: []byte{
			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "1 pixel to #000080",
		pixels: toRGB([]color.NRGBA{
			{0x00, 0x00, 0x80, 0x00},
		}),
		want: []byte{
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 1,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
	{
		name: "All at once",
		pixels: toRGB([]color.NRGBA{
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

			/*00*/ 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*80*/ 0xE8, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*80*/, 0xE8, 0x88, 0x88, 0x88,

			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*10*/, 0x88, 0x8E, 0x88, 0x88,
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*01*/, 0x88, 0x88, 0x88, 0x8E,
			/*00*/ 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88 /*00*/, 0x88, 0x88, 0x88, 0x88,
			/*EOF*/ 0x00, 0x00, 0x00,
		},
		opts: Opts{
			NumPixels: 10,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
}

func TestSPI_Writes(t *testing.T) {
	for _, tt := range writeTests {
		buf := bytes.Buffer{}
		tt.opts.NumPixels = len(tt.pixels) / 3
		d, err := NewSPI(spitest.NewRecordRaw(&buf), &tt.opts)
		if err != nil {
			t.Fatal(err)
		}
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

func TestSPI_Color(t *testing.T) {
	if c := (&Dev{}).ColorModel(); c != color.NRGBAModel {
		t.Fatal(c)
	}
}

func TestSPI_Long(t *testing.T) {
	buf := bytes.Buffer{}
	colors := make([]color.NRGBA, 256)
	o := Opts{NumPixels: len(colors), Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := d.Write(toRGB(colors)); n != len(colors)*3 || err != nil {
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

func TestSPI_Write_Long(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 1, Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if err != nil {
		t.Fatal(err)
	}
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
				b = append(b, 0xE8, 0x88, 0x88, 0x88) //0x80
				b = append(b, 0x88, 0x88, 0x88, 0x88) //0x00
				b = append(b, 0xEE, 0xEE, 0xEE, 0xEE) //0xFF
			}
			for i := 0; i < 3; i++ {
				b = append(b, 0x00)
			}
			return b
		}(),
		opts: Opts{
			NumPixels: 4,
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
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
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
}

func TestSPI_Draws(t *testing.T) {
	for _, tt := range drawTests {
		buf := bytes.Buffer{}
		d, err := NewSPI(spitest.NewRecordRaw(&buf), &tt.opts)
		if err != nil {
			t.Fatal(err)
		}
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

func TestSPI_Draw_DstEmpty(t *testing.T) {
	buf := bytes.Buffer{}
	o := Opts{NumPixels: 4, Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(spitest.NewRecordRaw(&buf), &o)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	if err := d.Draw(image.Rect(0, 0, 0, 0), img, image.Point{}); err != nil {
		t.Fatal(err)
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
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
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
			Channels:  3,
			Freq:      2500 * physic.KiloHertz,
		},
	},
}

func TestSPI_Halt(t *testing.T) {
	s := spitest.Playback{
		Playback: conntest.Playback{
			Count: 1,
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
	o := Opts{NumPixels: 4, Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(&s, &o)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSPI_Halt_fail(t *testing.T) {
	s := spitest.Playback{Playback: conntest.Playback{DontPanic: true}}
	o := Opts{NumPixels: 4, Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(&s, &o)
	if err != nil {
		t.Fatal(err)
	}
	if d.Halt() == nil {
		t.Fatal("expected failure")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

type genColor func(int) [3]byte

func benchmarkSPIWrite(b *testing.B, o Opts, length int, f genColor) {
	var pixels []byte
	for i := 0; i < length; i++ {
		c := f(i)
		pixels = append(pixels, c[:]...)
	}
	o.NumPixels = length
	b.ReportAllocs()
	d, err := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &o)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := d.Write(pixels[:]); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = d.Write(pixels[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSPI_WriteWhite(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIWrite(b, o, 150, func(i int) [3]byte { return [3]byte{0xFF, 0xFF, 0xFF} })
}

func BenchmarkSPI_WriteDim(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIWrite(b, o, 150, func(i int) [3]byte { return [3]byte{0x01, 0x01, 0x01} })
}

func BenchmarkSPI_WriteBlack(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIWrite(b, o, 150, func(i int) [3]byte { return [3]byte{0x0, 0x0, 0x0} })
}

func genColorfulPixel(x int) [3]byte {
	i := x * 3
	return [3]byte{uint8(i) + uint8(i>>8),
		uint8(i+1) + uint8(i+1>>8),
		uint8(i+2) + uint8(i+2>>8),
	}
}

func BenchmarkSPI_WriteColorful(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIWrite(b, o, 150, genColorfulPixel)
}

func BenchmarkSPI_WriteColorfulPassThru(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIWrite(b, o, 150, genColorfulPixel)
}

func BenchmarkSPI_WriteColorfulVariation(b *testing.B) {
	// Continuously vary the lookup tables.
	b.ReportAllocs()
	pixels := [256 * 3]byte{}
	for i := range pixels {
		pixels[i] = uint8(i) + uint8(i>>8)
	}
	o := Opts{NumPixels: len(pixels) / 3, Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &o)
	if err != nil {
		b.Fatal(err)
	}
	if _, err = d.Write(pixels[:]); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err = d.Write(pixels[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkSPIDraw(b *testing.B, o Opts, img draw.Image, f genColor) {
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

func BenchmarkSPI_DrawNRGBAColorful(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIDraw(b, o, image.NewNRGBA(image.Rect(0, 0, 150, 1)), genColorfulPixel)
}

func BenchmarkSPI_DrawNRGBAWhite(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIDraw(b, o, image.NewNRGBA(image.Rect(0, 0, 150, 1)), func(i int) [3]byte { return [3]byte{0xFF, 0xFF, 0xFF} })
}

func BenchmarkDrawRGBAColorful(b *testing.B) {
	o := Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	benchmarkSPIDraw(b, o, image.NewRGBA(image.Rect(0, 0, 256, 1)), genColorfulPixel)
}

func BenchmarkSPI_DrawSlowpath(b *testing.B) {
	// Should be an image type that doesn't have a fast path
	img := image.NewGray(image.Rect(0, 0, 150, 1))
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			pix := genColorfulPixel(x)
			img.Set(x, y, color.Gray{pix[0]})
		}
	}
	o := Opts{NumPixels: img.Bounds().Max.X, Channels: 3, Freq: 2500 * physic.KiloHertz}
	b.ReportAllocs()
	d, err := NewSPI(spitest.NewRecordRaw(ioutil.Discard), &o)
	if err != nil {
		b.Fatal(err)
	}
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

func BenchmarkSPI_Halt(b *testing.B) {
	b.ReportAllocs()
	o := &Opts{NumPixels: 150, Channels: 3, Freq: 2500 * physic.KiloHertz}
	d, err := NewSPI(spitest.NewRecordRaw(ioutil.Discard), o)
	if err != nil {
		b.Fatal(err)
	}
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

type limitLow struct {
	spitest.Record
}

func (c *limitLow) MaxTxSize() int {
	return 1
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
