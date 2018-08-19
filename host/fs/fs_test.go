// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fs

import "testing"

func TestTranslateOpMIPS(t *testing.T) {
	// input, expected
	data := [][2]uint{
		// Dir = 0 Size = 0x2000, Type = 0x10, NR = 0x20
		{0<<(14+8+8) | 0x1000<<(8+8) | 0x10<<8 | 0x20, 0x30001020},
		// Dir = 1, Size = 0x2000, Type = 0x10, NR = 0x20
		{1<<(14+8+8) | 0x1000<<(8+8) | 0x10<<8 | 0x20, 0x90001020},
		// Dir = 2, Size = 0x2000, Type = 0x10, NR = 0x20
		{2<<(14+8+8) | 0x1000<<(8+8) | 0x10<<8 | 0x20, 0x50001020},
		// Dir = 0, Size = 0x0000, Type = 0x00, NR = 0x00
		{0 << (14 + 8 + 8), 0x20000000},
		// Dir = 1, Size = 0x0000, Type = 0x00, NR = 0x00
		{1 << (14 + 8 + 8), 0x80000000},
		// Dir = 2, Size = 0x0000, Type = 0x00, NR = 0x00
		{2 << (14 + 8 + 8), 0x40000000},
	}
	for i, line := range data {
		if actual, err := translateOpMIPS(line[0]); err != nil {
			t.Fatalf("#%d: error: %v", i, err)
		} else if line[1] != actual {
			t.Fatalf("#%d: expected %#x, got %#x", i, line[1], actual)
		}
	}
}

func TestTranslateOpMIPS_Error(t *testing.T) {
	// 14 bit size.
	if _, err := translateOpMIPS(1 << (13 + 8 + 8)); err == nil {
		t.Fatal("size")
	}
	if _, err := translateOpMIPS(3 << (14 + 8 + 8)); err == nil {
		t.Fatal("dir")
	}
}
