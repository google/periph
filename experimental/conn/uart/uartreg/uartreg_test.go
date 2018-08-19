// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package uartreg

import (
	"errors"
	"sort"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/conn/uart"
)

func TestOpen(t *testing.T) {
	defer reset()
	if _, err := Open(""); err == nil {
		t.Fatal("no bus registered")
	}
	if err := Register("a", []string{"x"}, 1, fakePorter); err != nil {
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
	if err := Register("a", nil, -1, fakePorter); err != nil {
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
	if err := Register("a", nil, 1, fakePorter); err != nil {
		t.Fatal(err)
	}
	if err := Register("b", nil, 2, fakePorter); err != nil {
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
	if err := Register("a", []string{"b"}, 42, fakePorter); err != nil {
		t.Fatal(err)
	}
	if Register("a", nil, -1, fakePorter) == nil {
		t.Fatal("same bus name")
	}
	if Register("b", nil, -1, fakePorter) == nil {
		t.Fatal("same bus alias name")
	}
	if Register("c", nil, 42, fakePorter) == nil {
		t.Fatal("same bus number")
	}
	if Register("c", []string{"a"}, -1, fakePorter) == nil {
		t.Fatal("same bus alias")
	}
	if Register("c", []string{"b"}, -1, fakePorter) == nil {
		t.Fatal("same bus alias")
	}
}

func TestRegister_fail(t *testing.T) {
	defer reset()
	if Register("a", nil, -1, nil) == nil {
		t.Fatal("missing Opener")
	}
	if Register("a", nil, -2, fakePorter) == nil {
		t.Fatal("bad bus number")
	}
	if Register("", nil, 42, fakePorter) == nil {
		t.Fatal("missing name")
	}
	if Register("1", nil, 42, fakePorter) == nil {
		t.Fatal("numeric name")
	}
	if Register("a:b", nil, 42, fakePorter) == nil {
		t.Fatal("':' in name")
	}
	if Register("a", []string{"a"}, 0, fakePorter) == nil {
		t.Fatal("\"a\" is already registered")
	}
	if Register("a", []string{""}, 0, fakePorter) == nil {
		t.Fatal("empty alias")
	}
	if Register("a", []string{"1"}, 0, fakePorter) == nil {
		t.Fatal("numeric alias")
	}
	if Register("a", []string{"a:b"}, 0, fakePorter) == nil {
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
	if err := Register("a", []string{"b"}, 0, fakePorter); err != nil {
		t.Fatal(err)
	}
	if err := Unregister("a"); err != nil {
		t.Fatal(err)
	}
}

//

func fakePorter() (uart.PortCloser, error) {
	return &fakePort{}, nil
}

// fakePort implements uart.PortCloser.
type fakePort struct {
	conn fakeConn
}

func (f *fakePort) String() string {
	return "fake"
}

func (f *fakePort) Close() error {
	return errors.New("not implemented")
}

func (f *fakePort) LimitSpeed(freq physic.Frequency) error {
	return errors.New("not implemented")
}

func (f *fakePort) Connect(freq physic.Frequency, stopBit uart.Stop, parity uart.Parity, flow uart.Flow, bits int) (conn.Conn, error) {
	return &f.conn, nil
}

func (f *fakePort) RX() gpio.PinIn   { return f.conn.RX() }
func (f *fakePort) TX() gpio.PinOut  { return f.conn.TX() }
func (f *fakePort) RTS() gpio.PinOut { return f.conn.RTS() }
func (f *fakePort) CTS() gpio.PinIn  { return f.conn.CTS() }

// fakeConn implements conn.Conn.
type fakeConn struct {
}

func (f *fakeConn) String() string {
	return "fake"
}

func (f *fakeConn) Tx(w, r []byte) error {
	return errors.New("not implemented")
}

func (f *fakeConn) Duplex() conn.Duplex {
	return conn.Full
}

func (f *fakeConn) RX() gpio.PinIn   { return gpio.INVALID }
func (f *fakeConn) TX() gpio.PinOut  { return gpio.INVALID }
func (f *fakeConn) RTS() gpio.PinOut { return gpio.INVALID }
func (f *fakeConn) CTS() gpio.PinIn  { return gpio.INVALID }

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byName = map[string]*Ref{}
	byNumber = map[int]*Ref{}
	byAlias = map[string]*Ref{}
}

//

var _ uart.PortCloser = &fakePort{}
var _ uart.Pins = &fakePort{}
var _ conn.Conn = &fakeConn{}
var _ uart.Pins = &fakeConn{}
