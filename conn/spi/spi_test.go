// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spi

import (
	"testing"
)

func TestMode_String(t *testing.T) {
	if s := Mode(^int(0)).String(); s != "Mode3|HalfDuplex|NoCS|LSBFirst|0xffffffffffffffe0" {
		t.Fatal(s)
	}
	if s := Mode0.String(); s != "Mode0" {
		t.Fatal(s)
	}
	if s := Mode1.String(); s != "Mode1" {
		t.Fatal(s)
	}
	if s := Mode2.String(); s != "Mode2" {
		t.Fatal(s)
	}
}
