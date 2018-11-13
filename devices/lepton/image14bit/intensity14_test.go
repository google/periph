// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package image14bit

import (
	"image/color"
	"testing"
)

func TestIntensity14(t *testing.T) {
	if r, g, b, a := Intensity14(16383).RGBA(); r != 65535 || g != r || b != r || a != r {
		t.Fatal(r, g, b, a)
	}
	if r, g, b, a := Intensity14(0).RGBA(); r != 0 || g != r || b != r || a != 65535 {
		t.Fatal(r, g, b, a)
	}
	if s := Intensity14(16383).String(); s != "Intensity14(16383)" {
		t.Fatal(s)
	}
	if s := Intensity14(0).String(); s != "Intensity14(0)" {
		t.Fatal(s)
	}
	if Intensity14(8192) != convertIntensity14(Intensity14(8192)) {
		t.Fatal("failed to convert Intensity14 correctly")
	}
	if Intensity14(8224) != convertIntensity14(color.NRGBA{0x80, 0x80, 0x80, 0xFF}) {
		t.Fatal("failed to convert color.NRGBA correctly")
	}
}
