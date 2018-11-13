// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package apa102

import "testing"

func TestToRGBFast_limits(t *testing.T) {
	if r, g, b := toRGBFast(999); r != 255 || g != 83 || b != 0 {
		t.Fatal(r, g, b)
	}

	if r, g, b := toRGBFast(30000); r != 159 || g != 191 || b != 255 {
		t.Fatal(r, g, b)
	}
}

func BenchmarkToRGBFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if r, g, blue := toRGBFast(30000); r != 159 || g != 191 || blue != 255 {
			b.Fatal(r, g, blue)
		}
	}
}
