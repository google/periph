// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire

import (
	"fmt"
	"testing"
)

func ExampleAll() {
	fmt.Print("1-wire buses available:\n")
	for name := range All() {
		fmt.Printf("- %s\n", name)
	}
}

func TestPullUp(t *testing.T) {
	if WeakPullup.String() != "Weak" || StrongPullup.String() != "Strong" {
		t.FailNow()
	}
}

func TestNoDevicesError(t *testing.T) {
	e := noDevicesError("no")
	if !e.NoDevices() || e.Error() != "no" {
		t.FailNow()
	}
}

func TestShortedBusError(t *testing.T) {
	e := shortedBusError("no")
	if !e.IsShorted() || !e.BusError() || e.Error() != "no" {
		t.FailNow()
	}
}

func TestBusError(t *testing.T) {
	e := busError("no")
	if !e.BusError() || e.Error() != "no" {
		t.FailNow()
	}
}

func TestDev(t *testing.T) {
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
}

func TestNoneRegistered(t *testing.T) {
	if _, err := New(-1); err == nil {
		t.Fail()
	}
}

// TestRegDereg tests registration and deregistration of buses.
func TestRegDereg(t *testing.T) {
	defer reset()
	opener1 := func() (BusCloser, error) {
		b := nopBus("bus1")
		return &b, nil
	}
	opener2 := func() (BusCloser, error) {
		b := nopBus("bus2")
		return &b, nil
	}

	if Register("", 1, opener1) == nil {
		t.FailNow()
	}
	if Register("a", -1, opener1) == nil {
		t.FailNow()
	}
	if Register("a", 1, nil) == nil {
		t.FailNow()
	}
	if Unregister("", 1) == nil {
		t.FailNow()
	}

	// Register a first bus.
	if err := Register("bus1", 4, opener1); err != nil {
		t.Errorf("Failed to register the first bus: %s", err)
	}

	// Try to register a clashing bus.
	if err := Register("bus1", 5, opener2); err == nil {
		t.Errorf("Expected re-registration with the same name to fail")
	}
	if err := Register("bus2", 4, opener2); err == nil {
		t.Errorf("Expected re-registration with the same number to fail")
	}

	// Register a second bus.
	if err := Register("bus2", 15, opener2); err != nil {
		t.Errorf("Failed to register the second bus: %s", err)
	}

	// Ensure queries work.
	a := All()
	if len(a) != 2 {
		t.Errorf("Expected All() to return 2 buses, got %d", len(a))
	}
	if b, _ := a["bus1"](); b.String() != "bus1" {
		t.Errorf("Expected All() to return bus1, got %v", b.String())
	}
	if b, _ := a["bus2"](); b.String() != "bus2" {
		t.Errorf("Expected All() to return bus2, got %v", b.String())
	}

	// Quick test of New.
	n, err := New(4)
	if err != nil {
		t.Errorf("Expected New(4) to succeed, got %s", err)
	}
	if n.String() != "bus1" {
		t.Errorf("Expected New(4) to return bus1, got %v", n.String())
	}
	if n, _ := New(-1); n.String() != "bus1" {
		t.Errorf("Expected New(-1) to return bus1, got %v", n.String())
	}

	// Failures
	if o, err := New(42); o != nil || err == nil {
		t.FailNow()
	}
	if Unregister("bus1", 42) == nil {
		t.FailNow()
	}
	if Unregister("", 4) == nil {
		t.FailNow()
	}

	// Deregister the first bus.
	if err := Unregister("bus1", 4); err != nil {
		t.Errorf("Expected unregister of bus1 to succeed, got %s", err)
	}
	a = All()
	if len(a) != 1 {
		t.Errorf("Expected All() to return 1 buses, got %d", len(a))
	}

	// Verify that first got reassigned.
	if n, _ := New(-1); n.String() != "bus2" {
		t.Errorf("Expected New(-1) to return bus2, got %v", n.String())
	}

	// Deregister the second bus.
	if err := Unregister("bus2", 15); err != nil {
		t.Errorf("Expected unregister of bus2 to succeed, got %s", err)
	}
	a = All()
	if len(a) != 0 {
		t.Errorf("Expected All() to return 0 buses, got %d", len(a))
	}
}

//

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byName = map[string]Opener{}
	byNumber = map[int]Opener{}
}

// nopBus implements Bus.
type nopBus string

func (b *nopBus) String() string                           { return string(*b) }
func (b *nopBus) Tx(w, r []byte, power Pullup) error       { return nil }
func (b *nopBus) Search(alarmOnly bool) ([]Address, error) { return nil, nil }
func (b *nopBus) Close() error                             { return nil }
