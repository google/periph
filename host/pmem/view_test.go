// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"reflect"
	"testing"
)

type simpleStruct struct {
	u uint32
}

func TestSlice(t *testing.T) {
	// Assumes binary.LittleEndian. Correct if this code is ever run on BigEndian.
	s := Slice([]byte{4, 3, 2, 1})
	u32 := s.Uint32()
	if len(u32) != 1 || u32[0] != 0x01020304 {
		t.Fatalf("%v", u32)
	}
	var v *simpleStruct
	if err := s.Struct(reflect.ValueOf(&v)); err != nil {
		t.Fatalf("%v", err)
	}
	if v == nil {
		t.FailNow()
	}
	if v.u != 0x01020304 {
		t.Fatalf("%v", v.u)
	}
	if s.Struct(reflect.ValueOf(nil)) == nil {
		t.FailNow()
	}
	if s.Struct(reflect.ValueOf(v)) == nil {
		t.FailNow()
	}
	var p *uint32
	if s.Struct(reflect.ValueOf(p)) == nil {
		t.FailNow()
	}
	if s.Struct(reflect.ValueOf(&p)) == nil {
		t.FailNow()
	}
	s = Slice([]byte{1})
	if s.Struct(reflect.ValueOf(&v)) == nil {
		t.FailNow()
	}
}
