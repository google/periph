// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"bytes"
	"errors"
	"log"
	"reflect"
	"testing"

	"periph.io/x/periph/host/fs"
)

func ExampleMapStruct() {
	// Let's say the CPU has 4 x 32 bits memory mapped registers at the address
	// 0xDEADBEEF.
	var reg *[4]uint32
	if err := MapStruct(0xDEADBEAF, reflect.ValueOf(reg)); err != nil {
		log.Fatal(err)
	}
	// reg now points to physical memory.
}

//

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

// These are really just exercising code, not real tests.

func TestMapGPIO(t *testing.T) {
	defer reset()
	MapGPIO()
}

func TestMap(t *testing.T) {
	defer reset()
	if v, err := Map(0, 0); v != nil || err == nil {
		t.Fatal("0 size")
	}
}

func TestMapStruct(t *testing.T) {
	defer reset()
	if MapStruct(0, reflect.Value{}) == nil {
		t.Fatal("0 size")
	}
	var i *int
	if MapStruct(0, reflect.ValueOf(i)) == nil {
		t.Fatal("not pointer to pointer")
	}
	x := 0
	i = &x
	if MapStruct(0, reflect.ValueOf(&i)) == nil {
		t.Fatal("pointer is not nil")
	}

	type tmp struct {
		A int
	}
	var v *tmp
	if MapStruct(0, reflect.ValueOf(&v)) == nil {
		t.Fatal("not as root")
	}
}

func TestView(t *testing.T) {
	defer reset()
	v := View{}
	v.Close()
	v.PhysAddr()
}

//

type simpleStruct struct {
	u uint32
}

type simpleFile struct {
	data []byte
}

func (s *simpleFile) Close() error {
	return nil
}

func (s *simpleFile) Fd() uintptr {
	return 0
}

func (s *simpleFile) Read(b []byte) (int, error) {
	if s.data == nil {
		return 0, errors.New("injected")
	}
	copy(b, s.data)
	return len(s.data), nil
}

func (s *simpleFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

type failFile struct {
}

func (f *failFile) Close() error {
	return errors.New("injected error")
}

func (f *failFile) Fd() uintptr {
	return 0
}

func (f *failFile) Read(b []byte) (int, error) {
	return 0, errors.New("injected error")
}

func (f *failFile) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("injected error")
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	gpioMemErr = nil
	gpioMemView = nil
	devMem = nil
	devMemErr = nil
	openFile = openFileOrig
	pageMap = nil
	pageMapErr = nil
}

func init() {
	fs.Inhibit()
}
