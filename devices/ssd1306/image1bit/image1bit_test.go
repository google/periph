// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package image1bit

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestBit(t *testing.T) {
	if r, g, b, a := On.RGBA(); r != 65535 || g != r || b != r || a != r {
		t.Fail()
	}
	if r, g, b, a := Off.RGBA(); r != 0 || g != r || b != r || a != r {
		t.Fail()
	}
	if On.String() != "On" || Off.String() != "Off" {
		t.Fail()
	}
	if On != convertBit(On) {
		t.Fail()
	}
}
func TestImageNew(t *testing.T) {
	if img, err := New(image.Rect(0, 0, 8, 7)); img != nil || err == nil {
		t.Fail()
	}
	if img, err := New(image.Rect(0, 0, 1, 8)); img == nil || err != nil {
		t.Fail()
	}
}

func TestImagePixels(t *testing.T) {
	img, _ := New(image.Rect(0, 0, 1, 8))
	if !bytes.Equal(img.Buf, []byte{0x00}) {
		t.Fatal("starts black")
	}
	img.SetAll()
	if !bytes.Equal(img.Buf, []byte{0xFF}) {
		t.Fatal("SetAll sets white")
	}
	img.Clear()
	if !bytes.Equal(img.Buf, []byte{0x00}) {
		t.Fatal("Clear sets black")
	}
	img.Set(0, 2, color.NRGBA{0x80, 0x80, 0x80, 0xFF})
	img.Set(1, 2, color.NRGBA{0x80, 0x80, 0x80, 0xFF})
	img.Inverse()
	if !bytes.Equal(img.Buf, []byte{0xFB}) {
		t.Fatalf("inverse %# v", img.Buf)
	}
	if img.At(0, 2).(Bit) != Off {
		t.Fail()
	}
	if r := img.Bounds(); r.Min.X != 0 || r.Min.Y != 0 || r.Max.X != 1 || r.Max.Y != 8 {
		t.Fail()
	}
}

func TestColorModel(t *testing.T) {
	img, _ := New(image.Rect(0, 0, 1, 8))
	if v := img.ColorModel().Convert(color.NRGBA{0x80, 0x80, 0x80, 0xFF}).(Bit); v != On {
		t.Fatalf("%s", v)
	}
	if v := img.ColorModel().Convert(color.NRGBA{0x7F, 0x7F, 0x7F, 0xFF}).(Bit); v != Off {
		t.Fatalf("%s", v)
	}
}
