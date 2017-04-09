// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package videocore

import "testing"

func TestGenPacket(t *testing.T) {
	actual := genPacket(10, 12, 1, 2, 3)
	expected := []uint32{0x24, 0x0, 0xa, 0xc, 0xc, 0x1, 0x2, 0x3, 0x0}
	if !uint32Equals(actual, expected) {
		t.Fatal(actual)
	}
}

func uint32Equals(a []uint32, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
