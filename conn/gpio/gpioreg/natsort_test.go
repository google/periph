// Copyright 2010 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Extracted from https://github.com/maruel/natural for code coverage.

package gpioreg

import (
	"testing"
)

func TestLessLess(t *testing.T) {
	data := [][2]string{
		{"", "a"},
		{"a", "b"},
		{"a", "aa"},
		{"a0", "a1"},
		{"a0", "a00"},
		{"a00", "a01"},
		{"a01", "a2"},
		{"a01x", "a2x"},
		// Only the last number matters.
		{"a0b00", "a00b1"},
		{"a0b00", "a00b01"},
		{"a00b0", "a0b00"},
		{"a00b00", "a0b01"},
		{"a00b00", "a0b1"},
	}
	for _, l := range data {
		if !lessNatural(l[0], l[1]) {
			t.Fatalf("Less(%q, %q) returned false", l[0], l[1])
		}
	}
}

func TestLessNot(t *testing.T) {
	data := [][2]string{
		{"a", ""},
		{"a", "a"},
		{"aa", "a"},
		{"b", "a"},
		{"a01", "a00"},
		{"a01", "a01"},
		{"a1", "a1"},
		{"a2", "a01"},
		{"a2x", "a01x"},
		{"a00b00", "a0b0"},
		{"a00b01", "a0b00"},
		{"a00b00", "a0b00"},
	}
	for _, l := range data {
		if lessNatural(l[0], l[1]) {
			t.Fatalf("Less(%q, %q) returned true", l[0], l[1])
		}
	}
}
