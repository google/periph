// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2c

import (
	"bytes"
	"errors"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/physic"
)

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
		t.Fatal(err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal(b.w)
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
		t.Fatal(err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal(b.w)
	}
	if b.addr != 12 {
		t.Fatal(b.addr)
	}
}

//

type fakeBus struct {
	freq physic.Frequency
	err  error
	addr uint16
	w, r []byte
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

func (f *fakeBus) SetSpeed(freq physic.Frequency) error {
	f.freq = freq
	return f.err
}

func TestAddr_Set(t *testing.T) {

	tests := []struct {
		str  string
		want Addr
		err  error
	}{
		{"0x18", 0x18, nil},
		{"24", 24, nil},
		{"65535", 65535, nil},
		{"65536", 0, errI2CSetError},
		{"-1", 0, errI2CSetError},
	}

	for _, tt := range tests {
		var a Addr
		// fmt.Println(tt)
		if err := a.Set(tt.str); err != tt.err {
			t.Errorf("i2cAddr.Set(%s) error %v", tt.str, err)
		}
		if tt.err == nil && a != tt.want {
			t.Errorf("i2cAddr.Set(%s) expected %d but got %d", tt.str, tt.want, a)
		}
	}

}
