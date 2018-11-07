// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package analog

import "testing"

func TestINVALID(t *testing.T) {
	if INVALID.Number() != -1 {
		t.Fatal("Number")
	}
	if INVALID.Name() != "INVALID" {
		t.Fatal("Name")
	}
	if INVALID.String() != "INVALID" {
		t.Fatal("String")
	}
	if INVALID.Function() != "" {
		t.Fatal("Function")
	}
	if INVALID.Halt() == nil {
		t.Fatal("Halt")
	}
	INVALID.Range()
	if _, err := INVALID.Read(); err == nil {
		t.Fatal("Read")
	}
	if INVALID.Out(0) == nil {
		t.Fatal("Out")
	}
}
