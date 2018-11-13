// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire

import (
	"bytes"
	"errors"
	"testing"

	"periph.io/x/periph/conn"
)

func TestPullUp(t *testing.T) {
	if s := WeakPullup.String(); s != "Weak" {
		t.Fatal(s)
	}
	if s := StrongPullup.String(); s != "Strong" {
		t.Fatal(s)
	}
}

func TestNoDevicesError(t *testing.T) {
	e := noDevicesError("no")
	if !e.NoDevices() {
		t.Fatal("expected NoDevices")
	}
	if s := e.Error(); s != "no" {
		t.Fatal(s)
	}
}

func TestShortedBusError(t *testing.T) {
	e := shortedBusError("no")
	if !e.IsShorted() {
		t.Fatal("expected IsShorted")
	}
	if !e.BusError() {
		t.Fatal("expected BusError")
	}
	if s := e.Error(); s != "no" {
		t.Fatal(s)
	}
}

func TestBusError(t *testing.T) {
	e := busError("no")
	if !e.BusError() {
		t.Fatal("expected BusError")
	}
	if s := e.Error(); s != "no" {
		t.Fatal(s)
	}
}

func TestDevString(t *testing.T) {
	d := Dev{&fakeBus{}, 12}
	if s := d.String(); s != "fake(0x000000000000000c)" {
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
	expected := []byte{85, 12, 0, 0, 0, 0, 0, 0, 0, 3, 4, 5}
	if !bytes.Equal(b.w, expected) {
		t.Fatal(b.w)
	}
	expected = []byte{1, 2, 3}
	if !bytes.Equal(r, expected) {
		t.Fatalf("r: %v != %v", b.r, expected)
	}
	if i := d.Duplex(); i != conn.Half {
		t.Fatal(i)
	}
}

func TestDevTxPower(t *testing.T) {
	b := nopBus("hi")
	d := Dev{Bus: &b, Addr: 12}
	if s := d.String(); s != "hi(0x000000000000000c)" {
		t.Fatal(s)
	}
	// TODO(maruel): Verify the output.
	if err := d.Tx([]byte{1}, nil); err != nil {
		t.Fatal(err)
	}
	if err := d.TxPower([]byte{1}, nil); err != nil {
		t.Fatal(err)
	}
	if v := d.Duplex(); v != conn.Half {
		t.Fatal(v)
	}
}

//

type fakeBus struct {
	power Pullup
	err   error
	w, r  []byte
}

func (f *fakeBus) Close() error {
	return nil
}

func (f *fakeBus) String() string {
	return "fake"
}

func (f *fakeBus) Tx(w, r []byte, power Pullup) error {
	f.power = power
	f.w = append(f.w, w...)
	copy(r, f.r)
	f.r = f.r[len(r):]
	return f.err
}

func (f *fakeBus) Search(alarmOnly bool) ([]Address, error) {
	return nil, errors.New("not implemented")
}

// nopBus implements Bus.
type nopBus string

func (b *nopBus) String() string                           { return string(*b) }
func (b *nopBus) Tx(w, r []byte, power Pullup) error       { return nil }
func (b *nopBus) Search(alarmOnly bool) ([]Address, error) { return nil, nil }
func (b *nopBus) Close() error                             { return nil }
