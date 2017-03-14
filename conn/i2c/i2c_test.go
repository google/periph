// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2c

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func ExampleAll() {
	fmt.Print("I²C buses available:\n")
	for name := range All() {
		fmt.Printf("- %s\n", name)
	}
}

//

func TestDevString(t *testing.T) {
	d := Dev{&fakeBus{}, 12}
	if s := d.String(); s != "fake(12)" {
		t.Fatalf("got %s", s)
	}
}

func TestDevTx(t *testing.T) {
	exErr := errors.New("yes")
	b := &fakeBus{err: exErr, r: []byte{1, 2, 3}}
	d := Dev{b, 12}
	r := make([]byte, 3)
	w := []byte{3, 4, 5}
	if err := d.Tx(w, r); exErr != err {
		t.Fatalf("got %s", err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal("w")
	}
	expected := []byte{1, 2, 3}
	if !bytes.Equal(r, expected) {
		t.Fatalf("r: %v != %v", b.r, expected)
	}
	if b.addr != 12 {
		t.Fatalf("got %d", b.addr)
	}
}

func TestDevWrite(t *testing.T) {
	b := &fakeBus{}
	d := Dev{b, 12}
	w := []byte{3, 4, 5}
	if n, err := d.Write(w); err != nil || n != 3 {
		t.Fatalf("got %s", err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal("w")
	}
	if b.addr != 12 {
		t.Fatalf("got %d", b.addr)
	}
}

func TestDevWriteErr(t *testing.T) {
	exErr := errors.New("yes")
	b := &fakeBus{err: exErr}
	d := Dev{b, 12}
	w := []byte{3, 4, 5}
	if n, err := d.Write(w); err != exErr || n != 0 {
		t.Fatalf("got %s", err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal("w")
	}
	if b.addr != 12 {
		t.Fatalf("got %d", b.addr)
	}
}

//

func TestAll(t *testing.T) {
	defer reset()
	byName = map[string]Opener{"foo": nil}
	actual := All()
	if len(actual) != 1 {
		t.Fatalf("%v", actual)
	}
	if _, ok := actual["foo"]; !ok {
		t.FailNow()
	}
}

func TestNew(t *testing.T) {
	defer reset()
	if _, err := New(-1); err == nil {
		t.FailNow()
	}

	byNumber = map[int]Opener{42: fakeBusOpener}
	if v, err := New(-1); err != nil || v == nil {
		t.FailNow()
	}
	if v, err := New(42); err != nil || v == nil {
		t.FailNow()
	}
	if v, err := New(1); err == nil || v != nil {
		t.FailNow()
	}
}

func TestRegister(t *testing.T) {
	defer reset()
	if Unregister("", 42) == nil {
		t.FailNow()
	}
	if Register("a", 42, nil) == nil {
		t.FailNow()
	}
	if Register("", 42, fakeBusOpener) == nil {
		t.FailNow()
	}
	if err := Register("a", 42, fakeBusOpener); err != nil {
		t.Fatal(err)
	}
	if Register("a", 42, fakeBusOpener) == nil {
		t.FailNow()
	}
	if Register("b", 42, fakeBusOpener) == nil {
		t.FailNow()
	}
	if Unregister("", 42) == nil {
		t.FailNow()
	}
	if Unregister("a", 0) == nil {
		t.FailNow()
	}
	if err := Unregister("a", 42); err != nil {
		t.Fatal(err)
	}
}

//

func fakeBusOpener() (BusCloser, error) {
	return &fakeBus{}, nil
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byName = map[string]Opener{}
	byNumber = map[int]Opener{}
}

type fakeBus struct {
	speed int64
	err   error
	addr  uint16
	w, r  []byte
}

func (f *fakeBus) Close() error {
	return nil
}

func (f *fakeBus) String() string {
	return "fake"
}

func (f *fakeBus) Tx(addr uint16, w, r []byte) error {
	f.addr = addr
	f.w = append(f.w, w...)
	copy(r, f.r)
	f.r = f.r[len(r):]
	return f.err
}

func (f *fakeBus) Speed(hz int64) error {
	f.speed = hz
	return f.err
}
