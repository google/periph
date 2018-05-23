// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package image14bit

import (
	"image"
	"image/color"
	"testing"
)

func TestNewGray14(t *testing.T) {
	data := []struct {
		r      image.Rectangle
		l      int
		stride int
	}{
		// Empty.
		{
			image.Rect(0, 0, 0, 0),
			0,
			0,
		},
		// Empty
		{
			image.Rect(0, 0, 1, 0),
			0,
			1,
		},
		// Empty
		{
			image.Rect(0, 0, 0, 1),
			0,
			0,
		},
		// 1x1
		{
			image.Rect(0, 0, 1, 1),
			1,
			1,
		},
		// Zero-based
		{
			image.Rect(0, 0, 9, 17),
			9 * 17,
			9,
		},
		// Non-zero-based
		{
			image.Rect(1, 7, 9, 17),
			8 * 10,
			8,
		},
		// Negative X
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
	}
	for i, line := range data {
		img := NewGray14(line.r)
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

func TestAt(t *testing.T) {
	img := NewGray14(image.Rect(0, 0, 1, 1))
	img.SetIntensity14(0, 0, Intensity14(16383))
	c := img.At(0, 0)
	if g, ok := c.(Intensity14); !ok || g != Intensity14(16383) {
		t.Fatal(c, g)
	}
	// Out of bounds.
	c = img.At(0, 1)
	if g, ok := c.(Intensity14); !ok || g != Intensity14(0) {
		t.Fatal(c, g)
	}
}

func TestIntensity14At(t *testing.T) {
	img := NewGray14(image.Rect(0, 0, 1, 1))
	img.SetIntensity14(0, 0, Intensity14(16383))
	if g := img.Intensity14At(0, 0); g != Intensity14(16383) {
		t.Fatal(g)
	}
	// Out of bounds.
	if g := img.Intensity14At(0, 1); g != Intensity14(0) {
		t.Fatal(g)
	}
}

func TestColorModel(t *testing.T) {
	img := NewGray14(image.Rect(0, 0, 1, 8))
	if v := img.ColorModel(); v != Intensity14Model {
		t.Fatalf("%s", v)
	}
	if v := img.ColorModel().Convert(color.NRGBA{0x00, 0x00, 0x00, 0xFF}).(Intensity14); v != Intensity14(0) {
		t.Fatalf("%s", v)
	}
	if v := img.ColorModel().Convert(color.NRGBA{0x7F, 0x7F, 0x7F, 0xFF}).(Intensity14); v != Intensity14(8159) {
		t.Fatalf("%s", v)
	}
	if v := img.ColorModel().Convert(color.NRGBA{0xFF, 0xFF, 0xFF, 0xFF}).(Intensity14); v != Intensity14(16383) {
		t.Fatalf("%s", v)
	}
}

func TestOpaque(t *testing.T) {
	if !NewGray14(image.Rect(0, 0, 1, 8)).Opaque() {
		t.Fatal("image is always opaque")
	}
}

func TestPixOffset(t *testing.T) {
	data := []struct {
		r      image.Rectangle
		x, y   int
		offset int
	}{
		{
			image.Rect(0, 0, 1, 1),
			0, 0,
			0,
		},
		{
			image.Rect(0, 0, 1, 8),
			0, 1,
			1,
		},
		{
			image.Rect(0, 0, 3, 16),
			1, 5,
			16,
		},
		{
			image.Rect(-1, -1, 3, 16),
			1, 5,
			26,
		},
	}
	for i, line := range data {
		img := NewGray14(line.r)
		offset := img.PixOffset(line.x, line.y)
		if offset != line.offset {
			t.Fatalf("#%d: expected offset:%v, actual offset:%v", i, line.offset, offset)
		}
	}
}

func TestSetIntensity14(t *testing.T) {
	img := NewGray14(image.Rect(0, 0, 1, 1))
	if img.Pix[0] != 0 {
		t.Fatal(img.Pix)
	}
	if img.SetIntensity14(0, 0, Intensity14(16383)); img.Pix[0] != 16383 {
		t.Fatal(img.Pix)
	}
	if img.SetIntensity14(0, 0, Intensity14(0)); img.Pix[0] != 0 {
		t.Fatal(img.Pix)
	}
}

func TestSet(t *testing.T) {
	img := NewGray14(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.NRGBA{0x80, 0x80, 0x80, 0xFF})
	if img.Pix[0] != 8224 {
		t.Fatal(img.Pix)
	}
}
