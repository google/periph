// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package nrzled

import "testing"

func TestNRZMSB3(t *testing.T) {
	for i := 0; i < 256; i++ {
		x := nrzMSB3Algo(byte(i))
		a := byte(x >> 16)
		b := byte((x >> 8))
		c := byte(x)
		if v := nrzMSB3[i]; a != v[0] || b != v[1] || c != v[2] {
			t.Fatalf("#%d: 0x%X != 0x%X || 0x%X != 0x%X || 0x%X != 0x%X", i, a, v[0], b, v[1], c, v[2])
		}
	}
}

func TestNRZMSB4(t *testing.T) {
	for i := 0; i < 256; i++ {
		x := nrzMSB4Algo(byte(i))
		a := byte(x >> 24)
		b := byte((x >> 16))
		c := byte((x >> 8))
		d := byte(x)
		if v := nrzMSB4[i]; a != v[0] || b != v[1] || c != v[2] || d != v[3] {
			t.Fatalf("#%d: 0x%X != 0x%X || 0x%X != 0x%X || 0x%X != 0x%X || 0x%X != 0x%X", i, a, v[0], b, v[1], c, v[2], d, v[3])
		}
	}
}
