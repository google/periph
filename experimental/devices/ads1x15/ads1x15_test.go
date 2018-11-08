// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ads1x15

import (
	"testing"

	"periph.io/x/periph/conn/i2c/i2ctest"
)

func TestChannel_String(t *testing.T) {
	// Mainly to increase test coverage...
	data := []struct {
		c        Channel
		expected string
	}{
		{Channel0, "0"},
		{Channel1, "1"},
		{Channel2, "2"},
		{Channel3, "3"},
		{Channel0Minus1, "0-1"},
		{Channel0Minus3, "0-3"},
		{Channel1Minus3, "1-3"},
		{Channel2Minus3, "2-3"},
		{Channel(-1), "Invalid"},
	}
	for _, line := range data {
		if actual := line.c.String(); actual != line.expected {
			t.Fatalf("%s != %s", line.expected, actual)
		}
	}
}

func TestChannel_number(t *testing.T) {
	// Mainly to increase test coverage...
	data := []struct {
		c        Channel
		expected int
	}{
		{Channel0, 0},
		{Channel1, 1},
		{Channel2, 2},
		{Channel3, 3},
		{Channel0Minus1, 4},
		{Channel0Minus3, 5},
		{Channel1Minus3, 6},
		{Channel2Minus3, 7},
		{Channel(-1), -1},
	}
	for _, line := range data {
		if actual := line.c.number(); actual != line.expected {
			t.Fatalf("%d != %d", line.expected, actual)
		}
	}
}

func TestDev_String(t *testing.T) {
	b := i2ctest.Playback{}
	d, err := NewADS1115(&b, &DefaultOpts)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "ADS1115" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}
