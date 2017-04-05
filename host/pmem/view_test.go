// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"bytes"
	"log"
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
	var r *[1]uint32
	if err := s.Struct(reflect.ValueOf(&r)); err != nil {
		t.Fatalf("%v", err)
	}
	if r[0] != u32[0] {
		t.Fatalf("%x != %x", r[0], u32[0])
	}
	if !bytes.Equal([]byte(s), s.Bytes()) {
		t.Fatal("Slice.Bytes() is the slice")
	}
}

func TestSliceErrors(t *testing.T) {
	s := Slice([]byte{4, 3, 2, 1})
	if s.Struct(reflect.ValueOf(nil)) == nil {
		t.FailNow()
	}
	var v *simpleStruct
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

func ExampleMapStruct() {
	// Let's say the CPU has 4 x 32 bits memory mapped registers at the address
	// 0xDEADBEEF.
	var reg *[4]uint32
	if err := MapStruct(0xDEADBEAF, reflect.ValueOf(reg)); err != nil {
		log.Fatal(err)
	}
	// reg now points to physical memory.
}
