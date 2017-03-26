// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2c

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"testing"

	"periph.io/x/periph/conn"
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
	if i := d.Duplex(); i != conn.Half {
		t.Fatal(i)
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

func TestOpenByNumber(t *testing.T) {
	defer reset()
	if _, err := OpenByNumber(-1); err == nil {
		t.Fatal("no bus registered")
	}

	if err := Register("a", nil, 42, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if v, err := OpenByNumber(-1); err != nil || v == nil {
		t.Fatal(v, err)
	}
	if v, err := OpenByNumber(42); err != nil || v == nil {
		t.Fatal(v, err)
	}
	if v, err := OpenByNumber(1); err == nil || v != nil {
		t.Fatal(v, err)
	}
}

func TestOpenByName(t *testing.T) {
	defer reset()
	if _, err := OpenByName(""); err == nil {
		t.Fatal("no bus registered")
	}
	if err := Register("a", []string{"x"}, 1, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if o, err := OpenByName(""); o == nil || err != nil {
		t.Fatal(o, err)
	}
	if o, err := OpenByName("1"); o == nil || err != nil {
		t.Fatal(o, err)
	}
	if o, err := OpenByName("x"); o == nil || err != nil {
		t.Fatal(o, err)
	}
	if o, err := OpenByName("y"); o != nil || err == nil {
		t.Fatal(o, err)
	}
}

func TestDefault_NoNumber(t *testing.T) {
	defer reset()
	if err := Register("a", nil, -1, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if o, err := OpenByName(""); o == nil || err != nil {
		t.Fatal(o, err)
	}
}

func TestAll(t *testing.T) {
	defer reset()
	if a := All(); len(a) != 0 {
		t.Fatal(a)
	}
	if err := Register("a", nil, 1, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if err := Register("b", nil, 2, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 2 {
		t.Fatal(a)
	}
}

func TestRefList(t *testing.T) {
	l := refList{&Ref{Name: "b"}, &Ref{Name: "a"}}
	sort.Sort(l)
	if l[0].Name != "a" || l[1].Name != "b" {
		t.Fatal(l)
	}
}

func TestRegister(t *testing.T) {
	defer reset()
	if err := Register("a", []string{"b"}, 42, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if Register("a", nil, -1, fakeBuser) == nil {
		t.Fatal("same bus name")
	}
	if Register("b", nil, -1, fakeBuser) == nil {
		t.Fatal("same bus alias name")
	}
	if Register("c", nil, 42, fakeBuser) == nil {
		t.Fatal("same bus number")
	}
	if Register("c", []string{"a"}, -1, fakeBuser) == nil {
		t.Fatal("same bus alias")
	}
	if Register("c", []string{"b"}, -1, fakeBuser) == nil {
		t.Fatal("same bus alias")
	}
}

func TestRegister_fail(t *testing.T) {
	defer reset()
	if Register("a", nil, -1, nil) == nil {
		t.Fatal("missing Opener")
	}
	if Register("a", nil, -2, fakeBuser) == nil {
		t.Fatal("bad bus number")
	}
	if Register("", nil, 42, fakeBuser) == nil {
		t.Fatal("missing name")
	}
	if Register("1", nil, 42, fakeBuser) == nil {
		t.Fatal("numeric name")
	}
	if Register("a", []string{"a"}, 0, fakeBuser) == nil {
		t.Fatal("\"a\" is already registered")
	}
	if Register("a", []string{""}, 0, fakeBuser) == nil {
		t.Fatal("empty alias")
	}
	if Register("a", []string{"1"}, 0, fakeBuser) == nil {
		t.Fatal("numeric alias")
	}
	if a := All(); len(a) != 0 {
		t.Fatal(a)
	}
}

func TestUnregister(t *testing.T) {
	defer reset()
	if Unregister("") == nil {
		t.Fatal("unregister empty")
	}
	if Unregister("a") == nil {
		t.Fatal("unregister non-existing")
	}
	if err := Register("a", []string{"b"}, 0, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if err := Unregister("a"); err != nil {
		t.Fatal(err)
	}
}

//

func fakeBuser() (BusCloser, error) {
	return &fakeBus{}, nil
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byName = map[string]*Ref{}
	byNumber = map[int]*Ref{}
	byAlias = map[string]*Ref{}
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
