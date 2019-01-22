// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283xsmoketest

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/physic"
)

func TestToHz(t *testing.T) {
	data := []struct {
		N        int
		T        time.Duration
		expected physic.Frequency
	}{
		{
			0,
			time.Second,
			0,
		},
		{
			1,
			0,
			0,
		},
		{
			1,
			time.Millisecond,
			physic.KiloHertz,
		},
		{
			1,
			time.Second,
			physic.Hertz,
		},
		{
			3,
			7 * time.Millisecond,
			// 3/7 with perfect rounding.
			428571429 * physic.MicroHertz,
		},
		{
			3000,
			7 * time.Microsecond,
			// 3/7 with perfect rounding.
			428571428571429 * physic.MicroHertz,
		},
		{
			1000000,
			1000 * time.Second,
			physic.KiloHertz,
		},
		{
			1000000,
			time.Second,
			physic.MegaHertz,
		},
		{
			1000000,
			time.Millisecond,
			physic.GigaHertz,
		},
		{
			1000000000,
			1000 * time.Second,
			physic.MegaHertz,
		},
		{
			1000000000,
			time.Second,
			physic.GigaHertz,
		},
		{
			1234556000,
			// 2.3s.
			2345567891 * time.Nanosecond,
			// 10 digits of resolution for 526.336MHz.
			526335849711 * physic.MilliHertz,
		},
		{
			1000000000,
			time.Millisecond,
			physic.TeraHertz,
		},
		{
			300000000,
			7 * time.Millisecond,
			// 3/7 with pretty good rounding, keeping in mind that's 42.857GHz.
			42857142857143 * physic.MilliHertz,
		},
	}
	for i, line := range data {
		r := testing.BenchmarkResult{N: line.N, T: line.T}
		if actual := toHz(&r); actual != line.expected {
			t.Fatalf("#%d: toHz(%d, %s) = %s(%d); expected %s(%d)", i, line.N, line.T, actual, actual, line.expected, line.expected)
		}
	}
}
