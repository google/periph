// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pin

import (
	"testing"
)

func TestInvalid(t *testing.T) {
	if s := INVALID.String(); s != "INVALID" {
		t.Fatal(s)
	}
}

func TestBasicPin(t *testing.T) {
	b := BasicPin{N: "Pin1"}
	if s := b.String(); s != "Pin1" {
		t.Fatal(s)
	}
	if err := b.Halt(); err != nil {
		t.Fatal(err)
	}
	if s := b.Name(); s != "Pin1" {
		t.Fatal(s)
	}
	if n := b.Number(); n != -1 {
		t.Fatal(-1)
	}
	if s := b.Function(); s != "" {
		t.Fatal(s)
	}
	if f := b.Func(); f != FuncNone {
		t.Fatal(f)
	}
	if f := b.SupportedFuncs(); len(f) != 0 {
		t.Fatal(f)
	}
	if err := b.SetFunc(Func("Out/Low")); err == nil {
		t.Fatal("expected failure")
	}
}

func TestV3_3(t *testing.T) {
	if f := V3_3.Func(); f != FuncNone {
		t.Fatal(f)
	}
	if f := V3_3.SupportedFuncs(); len(f) != 0 {
		t.Fatal(f)
	}
}
