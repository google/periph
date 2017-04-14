// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import "testing"

func TestAlloc(t *testing.T) {
	if m, err := Alloc(0); m != nil || err == nil {
		t.Fatal("0 bytes")
	}
	if m, err := Alloc(1); m != nil || err == nil {
		t.Fatal("not 4096 bytes")
	}
	// TODO(maruel): https://github.com/google/periph/issues/126
	/*
		if m, err := Alloc(4096); m != nil || err == nil {
			t.Fatal("not expected to succeed; e.g. it's known to be broken")
		}
	*/
}
