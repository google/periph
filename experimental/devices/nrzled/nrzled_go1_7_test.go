// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build go1.7

package nrzled

import (
	"strconv"
	"testing"
)

func TestNRZ(t *testing.T) {
	data := []struct {
		in       byte
		expected uint32
	}{
		{0x00, 0x924924},
		{0x01, 0x924926},
		{0x02, 0x924934},
		{0x04, 0x9249A4},
		{0x08, 0x924D24},
		{0x10, 0x926924},
		{0x20, 0x934924},
		{0x40, 0x9A4924},
		{0x80, 0xD24924},
		{0xFD, 0xDB6DA6},
		{0xFE, 0xDB6DB4},
		{0xFF, 0xDB6DB6},
	}
	for i, line := range data {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if actual := NRZ(line.in); line.expected != actual {
				t.Fatalf("NRZ(%X): 0x%X != 0x%X", line.in, line.expected, actual)
			}
		})
	}
}
