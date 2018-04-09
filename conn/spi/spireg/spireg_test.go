// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spireg

import (
	"errors"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/spi"
)

func TestOpen(t *testing.T) {
	defer reset()
	if _, err := Open(""); err == nil {
		t.Fatal("no port registered")
	}
	if err := Register("a", []string{"x"}, 1, getFakePort); err != nil {
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
	if err := Register("a", nil, -1, getFakePort); err != nil {
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
	if err := Register("a", nil, 1, getFakePort); err != nil {
		t.Fatal(err)
	}
	if err := Register("b", nil, 2, getFakePort); err != nil {
		t.Fatal(err)
	}
	if a := All(); len(a) != 2 {
		t.Fatal(a)
	}
}

func TestRef(t *testing.T) {
	out := insertRef(nil, &Ref{Name: "b"})
	out = insertRef(out, &Ref{Name: "d"})
	out = insertRef(out, &Ref{Name: "c"})
	out = insertRef(out, &Ref{Name: "a"})
	for i, l := range []string{"a", "b", "c", "d"} {
		if out[i].Name != l {
			t.Fatal(out)
		}
	}
}

func TestRegister(t *testing.T) {
	defer reset()
	if err := Register("a", []string{"b"}, 42, getFakePort); err != nil {
		t.Fatal(err)
	}
	if Register("a", nil, -1, getFakePort) == nil {
		t.Fatal("same port name")
	}
	if Register("b", nil, -1, getFakePort) == nil {
		t.Fatal("same port alias name")
	}
	if Register("c", nil, 42, getFakePort) == nil {
		t.Fatal("same port number")
	}
	if Register("c", []string{"a"}, -1, getFakePort) == nil {
		t.Fatal("same port alias")
	}
	if Register("c", []string{"b"}, -1, getFakePort) == nil {
		t.Fatal("same port alias")
	}
}

func TestRegister_fail(t *testing.T) {
	defer reset()
	if Register("a", nil, -1, nil) == nil {
		t.Fatal("missing Opener")
	}
	if Register("a", nil, -2, getFakePort) == nil {
		t.Fatal("bad port number")
	}
	if Register("", nil, 42, getFakePort) == nil {
		t.Fatal("missing name")
	}
	if Register("1", nil, 42, getFakePort) == nil {
		t.Fatal("numeric name")
	}
	if Register("a:b", nil, 42, getFakePort) == nil {
		t.Fatal("':' in name")
	}
	if Register("a", []string{"a"}, 0, getFakePort) == nil {
		t.Fatal("\"a\" is already registered")
	}
	if Register("a", []string{""}, 0, getFakePort) == nil {
		t.Fatal("empty alias")
	}
	if Register("a", []string{"1"}, 0, getFakePort) == nil {
		t.Fatal("numeric alias")
	}
	if Register("a", []string{"a:b"}, 0, getFakePort) == nil {
		t.Fatal("':' in alias")
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
	if err := Register("a", []string{"b"}, 0, getFakePort); err != nil {
		t.Fatal(err)
	}
	if err := Unregister("a"); err != nil {
		t.Fatal(err)
	}
}

//

func getFakePort() (spi.PortCloser, error) {
	return &fakePort{}, nil
}

type fakePort struct {
}

func (f *fakePort) String() string {
	return "fake"
}

func (f *fakePort) Close() error {
	return errors.New("not implemented")
}

func (f *fakePort) Tx(w, r []byte) error {
	return errors.New("not implemented")
}

func (f *fakePort) Duplex() conn.Duplex {
	return conn.DuplexUnknown
}

func (f *fakePort) LimitSpeed(maxHz int64) error {
	return errors.New("not implemented")
}

func (f *fakePort) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	return f, errors.New("not implemented")
}

func (f *fakePort) TxPackets(p []spi.Packet) error {
	return errors.New("not implemented")
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byName = map[string]*Ref{}
	byNumber = map[int]*Ref{}
	byAlias = map[string]*Ref{}
}
