// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"testing"

	"periph.io/x/periph/host/fs"
)

func ExampleMapAsPOD() {
	// Let's say the CPU has 4 x 32 bits memory mapped registers at the address
	// 0xDEADBEEF.
	var reg *[4]uint32
	if err := MapAsPOD(0xDEADBEAF, reg); err != nil {
		log.Fatal(err)
	}
	// reg now points to physical memory.
}

//

func TestSlice(t *testing.T) {
	s := Slice([]byte{4, 3, 2, 1})
	if !bytes.Equal([]byte(s), s.Bytes()) {
		t.Fatal("Slice.Bytes() is the slice")
	}

	// TODO(maruel): Assumes binary.LittleEndian. Correct if this code is ever
	// run on BigEndian.
	expected := binary.LittleEndian.Uint32(s)
	{
		v := s.Uint32()
		if len(v) != 1 || v[0] != expected {
			t.Fatalf("%v", v)
		}
		var a *[1]uint32
		if err := s.AsPOD(&a); err != nil {
			t.Fatalf("%v", err)
		}
		if a[0] != v[0] {
			t.Fatalf("%x != %x", a[0], v[0])
		}
	}

	{
		var v *simpleStruct
		if err := s.AsPOD(&v); err != nil {
			t.Fatalf("%v", err)
		}
		if v == nil {
			t.Fatal("v is nil")
		}
		if v.u != expected {
			t.Fatalf("%v", v.u)
		}
	}

	{
		var v *uint32
		if err := s.AsPOD(&v); err != nil {
			t.Fatalf("%v", err)
		}
		if *v != expected {
			t.Fatalf("%v", v)
		}
	}

	{
		var v []uint32
		if err := s.AsPOD(&v); err != nil {
			t.Fatalf("%v", err)
		}
	}
}

func TestSlice_Errors4(t *testing.T) {
	s := Slice([]byte{4, 3, 2, 1})

	if s.AsPOD(nil) == nil {
		t.Fatal("nil is not a valid type")
	}

	{
		var v simpleStruct
		if s.AsPOD(v) == nil {
			t.Fatal("must be Ptr to Ptr")
		}
		if s.AsPOD(&v) == nil {
			t.Fatal("must be Ptr to Ptr")
		}
	}

	{
		var v *uint32
		if s.AsPOD(v) == nil {
			t.Fatal("must be Ptr to Ptr")
		}
	}

	{
		var v ***[1]uint32
		if s.AsPOD(v) == nil {
			t.Fatal("buffer is not large enough")
		}
	}

	{
		var v *int
		if s.AsPOD(&v) == nil {
			t.Fatal("must be Ptr to sized type")
		}
	}

	{
		v := []uint32{1}
		if s.AsPOD(&v) == nil {
			t.Fatal("buffer is not large enough")
		}
	}

	{
		var v []interface{}
		if s.AsPOD(&v) == nil {
			t.Fatal("slice of non-POD")
		}
	}

	{
		var v *struct {
			A interface{}
		}
		if s.AsPOD(&v) == nil {
			t.Fatal("struct of non-POD")
		}
	}
}

func TestSlice_Errors1(t *testing.T) {
	s := Slice([]byte{1})
	{
		var v simpleStruct
		if s.AsPOD(&v) == nil {
			t.Fatal("not large enough")
		}
	}

	{
		var v *[1]uint32
		if s.AsPOD(&v) == nil {
			t.Fatal("buffer is not large enough")
		}
	}

	{
		var v []uint32
		if s.AsPOD(&v) == nil {
			t.Fatal("buffer is not large enough")
		}
	}
}

// These are really just exercising code, not real tests.

func TestMapGPIO(t *testing.T) {
	defer reset()
	// It can fail, depending on the platform.
	_, _ = MapGPIO()
}

func TestMap(t *testing.T) {
	defer reset()
	if v, err := Map(0, 0); v != nil || err == nil {
		t.Fatal("0 size")
	}
}

func TestMapAsPOD(t *testing.T) {
	defer reset()
	if MapAsPOD(0, nil) == nil {
		t.Fatal("0 size")
	}
	var i *int
	if MapAsPOD(0, i) == nil {
		t.Fatal("not pointer to pointer")
	}
	x := 0
	i = &x
	if MapAsPOD(0, &i) == nil {
		t.Fatal("pointer is not nil")
	}

	var v *simpleStruct
	if MapAsPOD(0, &v) == nil {
		t.Fatal("file I/O is inhibited; otherwise it would have worked")
	}
}

func TestView(t *testing.T) {
	defer reset()
	v := View{}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
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
