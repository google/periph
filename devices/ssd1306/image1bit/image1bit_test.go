// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package image1bit

import (
	"image"
	"image/color"
	"testing"
)

func TestBit(t *testing.T) {
	if r, g, b, a := On.RGBA(); r != 65535 || g != r || b != r || a != r {
		t.Fatal(r, g, b, a)
	}
	if r, g, b, a := Off.RGBA(); r != 0 || g != r || b != r || a != 65535 {
		t.Fatal(r, g, b, a)
	}
	if On.String() != "On" || Off.String() != "Off" {
		t.Fail()
	}
	if On != convertBit(On) {
		t.Fail()
	}
}

func TestVerticalLSB_NewVerticalLSB(t *testing.T) {
	data := []struct {
		r      image.Rectangle
		l      int
		stride int
	}{
		// Empty.
		{
			image.Rect(0, 0, 0, 1),
			0,
			0,
		},
		// Empty.
		{
			image.Rect(0, 0, 1, 0),
			0,
			1,
		},
		// 1 horizontal band of 1px high, 1px wide.
		{
			image.Rect(0, 0, 1, 1),
			1,
			1,
		},
		{
			image.Rect(0, 1, 1, 2),
			1,
			1,
		},
		// 1 horizontal band of 8px high, 1px wide.
		{
			image.Rect(0, 0, 1, 8),
			1,
			1,
		},
		// 1 horizontal band of 1px high, 9px wide.
		{
			image.Rect(0, 0, 9, 1),
			9,
			9,
		},
		// 2 horizontal bands of 1px high, 1px wide.
		{
			image.Rect(0, 0, 1, 9),
			2,
			1,
		},
		// 2 horizontal bands, 1px wide.
		{
			image.Rect(0, 1, 1, 9),
			2,
			1,
		},
		// 2 horizontal bands, 1px wide.
		{
			image.Rect(0, 7, 1, 9),
			2,
			1,
		},
		// 2 horizontal bands, 1px wide.
		{
			image.Rect(0, 7, 1, 16),
			2,
			1,
		},
		// 3 horizontal bands, 1px wide.
		{
			image.Rect(0, 7, 1, 17),
			3,
			1,
		},
		// 3 horizontal bands, 1px wide.
		{
			image.Rect(0, 7, 1, 17),
			3,
			1,
		},
		// 3 horizontal bands, 9px wide.
		{
			image.Rect(0, 7, 9, 17),
			3 * 9,
			9,
		},
		// Negative X.
		{
			image.Rect(-1, 0, 0, 1),
			1,
			1,
		},
		// Negative Y.
		{
			image.Rect(0, -1, 1, 0),
			1,
			1,
		},
		{
			image.Rect(0, -1, 1, 1),
			2,
			1,
		},
	}
	for i, line := range data {
		img := NewVerticalLSB(line.r)
		if r := img.Bounds(); r != line.r {
			t.Fatalf("#%d: expected %v; actual %v", i, line.r, r)
		}
		if l := len(img.Pix); l != line.l {
			t.Fatalf("#%d: len(img.Pix) expected %v; actual %v for %v", i, line.l, l, line.r)
		}
		if img.Stride != line.stride {
			t.Fatalf("#%d: img.Stride expected %v; actual %v for %v", i, line.stride, img.Stride, line.r)
		}
	}
}

func TestVerticalLSB_At(t *testing.T) {
	img := NewVerticalLSB(image.Rect(0, 0, 1, 1))
	img.SetBit(0, 0, On)
	c := img.At(0, 0)
	if b, ok := c.(Bit); !ok || b != On {
		t.Fatal(c, b)
	}
	c = img.At(0, 1)
	if b, ok := c.(Bit); !ok || b != Off {
		t.Fatal(c, b)
	}
}

func TestVerticalLSB_BitAt(t *testing.T) {
	img := NewVerticalLSB(image.Rect(0, 0, 1, 1))
	img.SetBit(0, 0, On)
	if b := img.BitAt(0, 0); b != On {
		t.Fatal(b)
	}
	if b := img.BitAt(0, 1); b != Off {
		t.Fatal(b)
	}
}

func TestVerticalLSB_ColorModel(t *testing.T) {
	img := NewVerticalLSB(image.Rect(0, 0, 1, 8))
	if v := img.ColorModel(); v != BitModel {
		t.Fatalf("%s", v)
	}
	if v := img.ColorModel().Convert(color.NRGBA{0x80, 0x80, 0x80, 0xFF}).(Bit); v != On {
		t.Fatalf("%s", v)
	}
	if v := img.ColorModel().Convert(color.NRGBA{0x7F, 0x7F, 0x7F, 0xFF}).(Bit); v != Off {
		t.Fatalf("%s", v)
	}
}

func TestVerticalLSB_Opaque(t *testing.T) {
	if !NewVerticalLSB(image.Rect(0, 0, 1, 8)).Opaque() {
		t.Fatal("image is always opaque")
	}
}

func TestVerticalLSB_PixOffset(t *testing.T) {
	data := []struct {
		r      image.Rectangle
		x, y   int
		offset int
		mask   byte
	}{
		{
			image.Rect(0, 0, 1, 1),
			0, 0,
			0, 0x01,
		},
		{
			image.Rect(0, 0, 1, 8),
			0, 1,
			0, 0x02,
		},
		{
			image.Rect(0, 0, 3, 16),
			1, 5,
			1, 0x20,
		},
		{
			image.Rect(-1, -1, 3, 16),
			1, 5,
			6, 0x20,
		},
	}
	for i, line := range data {
		img := NewVerticalLSB(line.r)
		offset, mask := img.PixOffset(line.x, line.y)
		if offset != line.offset || mask != line.mask {
			t.Fatalf("#%d: expected offset:%v, mask:0x%02X; actual offset:%v, mask:0x%02X", i, line.offset, line.mask, offset, mask)
		}
	}
}

func TestVerticalLSB_SetBit1x1(t *testing.T) {
	img := NewVerticalLSB(image.Rect(0, 0, 1, 1))
	if img.Pix[0] != 0 {
		t.Fatal(img.Pix)
	}
	if img.SetBit(0, 1, On); img.Pix[0] != 0 {
		t.Fatal(img.Pix)
	}
	if img.SetBit(0, 0, On); img.Pix[0] != 1 {
		t.Fatal(img.Pix)
	}
	if img.SetBit(0, 0, Off); img.Pix[0] != 0 {
		t.Fatal(img.Pix)
	}
}

func TestVerticalLSB_SetBit1x8(t *testing.T) {
	img := NewVerticalLSB(image.Rect(0, 0, 1, 8))
	if img.Pix[0] != 0 {
		t.Fatal(img.Pix)
	}
	if img.SetBit(0, 7, On); img.Pix[0] != 0x80 {
		t.Fatal(img.Pix)
	}
	if img.SetBit(0, 0, On); img.Pix[0] != 0x81 {
		t.Fatal(img.Pix)
	}
	if img.SetBit(0, 7, Off); img.Pix[0] != 1 {
		t.Fatal(img.Pix)
	}
}

func TestVerticalLSB_Set(t *testing.T) {
	img := NewVerticalLSB(image.Rect(0, 0, 1, 8))
	img.Set(0, 0, color.NRGBA{0x80, 0x80, 0x80, 0xFF})
	img.Set(0, 1, color.NRGBA{0x7F, 0x80, 0x80, 0xFF})
	img.Set(0, 2, color.NRGBA{0x7F, 0x7F, 0x80, 0xFF})
	img.Set(0, 3, color.NRGBA{0x7F, 0x7F, 0x7F, 0xFF})
	img.Set(0, 4, color.NRGBA{0x80, 0x80, 0x80, 0x7F})
	if img.Pix[0] != 7 {
		t.Fatal(img.Pix)
	}
}
