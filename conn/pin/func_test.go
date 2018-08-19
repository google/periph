// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pin

import (
	"testing"
)

func TestFunc(t *testing.T) {
	if v := FuncNone.Specialize(-1, -1); v != FuncNone {
		t.Fatal(v)
	}
	if v := FuncNone.Specialize(1, -1); v != FuncNone {
		t.Fatal(v)
	}
	if v := FuncNone.Specialize(-1, 1); v != FuncNone {
		t.Fatal(v)
	}
	if v := Func("A").Specialize(-1, 1); v != Func("A1") {
		t.Fatal(v)
	}
	if v := Func("A").Specialize(1, -1); v != FuncNone {
		t.Fatal(v)
	}
}
