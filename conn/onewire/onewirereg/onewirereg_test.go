// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewirereg

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"periph.io/x/periph/conn/onewire"
)

func ExampleAll() {
	// Enumerate all 1-wire buses available and the corresponding pins.
	fmt.Print("1-wire buses available:\n")
	for _, ref := range All() {
		fmt.Printf("- %s\n", ref.Name)
		if ref.Number != -1 {
			fmt.Printf("  %d\n", ref.Number)
		}
		if len(ref.Aliases) != 0 {
			fmt.Printf("  %s\n", strings.Join(ref.Aliases, " "))
		}

		b, err := ref.Open()
		if err != nil {
			fmt.Printf("  Failed to open: %v", err)
		}
		if p, ok := b.(onewire.Pins); ok {
			fmt.Printf("  Q: %s", p.Q())
		}
		if err := b.Close(); err != nil {
			fmt.Printf("  Failed to close: %v", err)
		}
	}
}

//

func TestOpen(t *testing.T) {
	defer reset()
	if _, err := Open(""); err == nil {
		t.Fatal("no bus registered")
	}
	if err := Register("a", []string{"x"}, 1, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if o, err := Open(""); o == nil || err != nil {
		t.Fatal(o, err)
	}
	if o, err := Open("1"); o == nil || err != nil {
		t.Fatal(o, err)
	}
	if o, err := Open("x"); o == nil || err != nil {
		t.Fatal(o, err)
	}
	if o, err := Open("y"); o != nil || err == nil {
		t.Fatal(o, err)
	}
}

func TestDefault_NoNumber(t *testing.T) {
	defer reset()
	if err := Register("a", nil, -1, fakeBuser); err != nil {
		t.Fatal(err)
	}
	if o, err := Open(""); o == nil || err != nil {
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

func fakeBuser() (onewire.BusCloser, error) {
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
	power onewire.Pullup
	err   error
	w, r  []byte
}

func (f *fakeBus) Close() error {
	return nil
}

func (f *fakeBus) String() string {
	return "fake"
}

func (f *fakeBus) Tx(w, r []byte, power onewire.Pullup) error {
	f.power = power
	f.w = append(f.w, w...)
	copy(r, f.r)
	f.r = f.r[len(r):]
	return f.err
}

func (f *fakeBus) Search(alarmOnly bool) ([]onewire.Address, error) {
	return nil, errors.New("not implemented")
}
