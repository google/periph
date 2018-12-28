// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fs

import (
	"testing"
)

func TestEpollEvent_String(t *testing.T) {
	if s := (epollIN | epollOUT).String(); s != "IN|OUT" {
		t.Fatal(s)
	}
	if s := (epollERR | epollEvent(0x1000)).String(); s != "ERR|0x1000" {
		t.Fatal(s)
	}
	if s := epollEvent(0).String(); s != "0" {
		t.Fatal(s)
	}
}
