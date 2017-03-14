// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire

import "testing"

func TestCheckCRC(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5, 6, 7}
	c := CalcCRC(a)
	if c != 15 {
		t.Fatal(c)
	}
	b := append([]byte{}, a...)
	b = append(b, c)
	if !CheckCRC(b) {
		t.FailNow()
	}
	b[len(b)-1]++
	if CheckCRC(b) {
		t.FailNow()
	}
	if CheckCRC(nil) {
		t.FailNow()
	}
}
