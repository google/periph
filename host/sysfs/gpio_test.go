// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"testing"

	"periph.io/x/periph/conn/gpio"
)

func TestPin_String(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if s := p.String(); s != "foo" {
		t.Fatal(s)
	}
}

func TestPin_Name(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if s := p.Name(); s != "foo" {
		t.Fatal(s)
	}
}

func TestPin_Number(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if n := p.Number(); n != 42 {
		t.Fatal(n)
	}
}

func TestPin_Function(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	// Fails because open is not mocked.
	if s := p.Function(); s != "ERR" {
		t.Fatal(s)
	}
	p = Pin{
		number:     42,
		name:       "foo",
		root:       "/tmp/gpio/priv/",
		fDirection: &fakeGPIOFile{},
	}
	if s := p.Function(); s != "ERR" {
		t.Fatal(s)
	}
	p.fDirection = &fakeGPIOFile{data: []byte("foo")}
	if s := p.Function(); s != "ERR" {
		t.Fatal(s)
	}
	p.fDirection = &fakeGPIOFile{data: []byte("in")}
	if s := p.Function(); s != "In/Low" {
		t.Fatal(s)
	}
	p.fDirection = &fakeGPIOFile{data: []byte("out")}
	if s := p.Function(); s != "Out/Low" {
		t.Fatal(s)
	}
}

func TestPin_In(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if p.In(gpio.PullNoChange, gpio.NoEdge) == nil {
		t.Fatal("can't open")
	}
	p = Pin{
		number:     42,
		name:       "foo",
		root:       "/tmp/gpio/priv/",
		fDirection: &fakeGPIOFile{},
	}
	if p.In(gpio.PullNoChange, gpio.NoEdge) == nil {
		t.Fatal("can't read direction")
	}

	p.fDirection = &fakeGPIOFile{data: []byte("out")}
	if err := p.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if p.In(gpio.PullDown, gpio.NoEdge) == nil {
		t.Fatal("pull not supported on sysfs-gpio")
	}
	if p.In(gpio.PullNoChange, gpio.BothEdges) == nil {
		t.Fatal("can't open edge")
	}

	p.fEdge = &fakeGPIOFile{}
	if p.In(gpio.PullNoChange, gpio.NoEdge) == nil {
		t.Fatal("edge I/O failed")
	}

	p.fEdge = &fakeGPIOFile{data: []byte("none")}
	if err := p.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.PullNoChange, gpio.RisingEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.PullNoChange, gpio.FallingEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		t.Fatal(err)
	}
}

func TestPin_Read(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if l := p.Read(); l != gpio.Low {
		t.Fatal("broken pin is always low")
	}
	p.fValue = &fakeGPIOFile{}
	if l := p.Read(); l != gpio.Low {
		t.Fatal("broken pin is always low")
	}
	p.fValue = &fakeGPIOFile{data: []byte("0")}
	if l := p.Read(); l != gpio.Low {
		t.Fatal("pin is low")
	}
	p.fValue = &fakeGPIOFile{data: []byte("1")}
	if l := p.Read(); l != gpio.High {
		t.Fatal("pin is high")
	}
	p.fValue = &fakeGPIOFile{data: []byte("2")}
	if l := p.Read(); l != gpio.Low {
		t.Fatal("pin is unknown")
	}
}

func TestPin_WaitForEdges(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if p.WaitForEdge(-1) {
		t.Fatal("broken pin doesn't have edge triggered")
	}
}

func TestPin_Pull(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if pull := p.Pull(); pull != gpio.PullNoChange {
		t.Fatal(pull)
	}
}

func TestPin_Out(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/", direction: dIn, edge: gpio.NoEdge}
	if p.Out(gpio.High) == nil {
		t.Fatal("can't open fake root")
	}
	p.fDirection = &fakeGPIOFile{}
	if p.Out(gpio.High) == nil {
		t.Fatal("failed to write to direction")
	}
	p.fDirection = &fakeGPIOFile{data: []byte("dummy")}
	if err := p.Out(gpio.High); err != nil {
		t.Fatal(err)
	}
	p.direction = dIn
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}
	p.direction = dIn
	p.edge = gpio.RisingEdge
	p.fEdge = &fakeGPIOFile{}
	if p.Out(gpio.High) == nil {
		t.Fatal("failed to write to edge")
	}
	p.edge = gpio.RisingEdge
	p.fEdge = &fakeGPIOFile{data: []byte("dummy")}
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}

	p.direction = dOut
	p.edge = gpio.NoEdge
	p.fValue = &fakeGPIOFile{}
	if p.Out(gpio.Low) == nil {
		t.Fatal("write to value failed")
	}
	p.fValue = &fakeGPIOFile{data: []byte("dummy")}
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}
	if err := p.Out(gpio.High); err != nil {
		t.Fatal(err)
	}
}

func TestPin_PWM(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if p.PWM(0) == nil {
		t.Fatal("sysfs-gpio doesn't support PWM")
	}
}

func TestPin_readInt(t *testing.T) {
	if _, err := readInt("/tmp/gpio/priv/invalid_file"); err == nil {
		t.Fatal("file is not expected to exist")
	}
}

//

type fakeGPIOFile struct {
	data []byte
}

func (f *fakeGPIOFile) Close() error {
	return nil
}

func (f *fakeGPIOFile) Fd() uintptr {
	return 0
}

func (f *fakeGPIOFile) read(b []byte) error {
	if f.data == nil {
		return errors.New("injected")
	}
	copy(b, f.data)
	return nil
}

func (f *fakeGPIOFile) write(b []byte) error {
	if f.data == nil {
		return errors.New("injected")
	}
	copy(f.data, b)
	return nil
}
