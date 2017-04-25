// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bmp180

import "testing"

func TestCompensate(t *testing.T) {
	c := calibration{
		AC1: 408,
		AC2: -72,
		AC3: -14383,
		AC4: 32741,
		AC5: 32757,
		AC6: 23153,
		B1:  6190,
		B2:  4,
		MB:  -32768,
		MC:  -8711,
		MD:  2868,
	}

	if temp := c.compensateTemp(27898); temp != 150 {
		t.Errorf("temperature is wrong, want %v, got %v", 150, temp)
	}

	if pressure := c.compensatePressure(23843, 27898, 0); pressure != 69964 {
		t.Errorf("pressure is wrong, want %v, got %v", 69964, pressure)
	}
}
