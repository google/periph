// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

func TestDebounce_Err(t *testing.T) {
	defer mocktime(t, nil)()
	f := gpiotest.Pin{}
	if _, err := Debounce(&f, time.Second, 0, gpio.BothEdges); err == nil {
		t.Fatal("expected error")
	}
}

func TestDebounce_Zero(t *testing.T) {
	defer mocktime(t, nil)()
	f := gpiotest.Pin{}
	p, err := Debounce(&f, 0, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal("expected error")
	}
	if p1, ok := p.(*gpiotest.Pin); !ok || p1 != &f {
		t.Fatal("expected the pin to be returned as-is")
	}
}

func TestDebounce_In(t *testing.T) {
	defer mocktime(t, nil)()
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		t.Fatal(err)
	}
	if p.Halt() != nil {
		t.Fatal(err)
	}
}

func TestDebounce_Read_Low(t *testing.T) {
	defer mocktime(t, nil)()
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, time.Second, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if p.Read() != gpio.Low {
		t.Fatal("expected level")
	}
	if p.Read() != gpio.Low {
		t.Fatal("expected level")
	}
}

func TestDebounce_Read_High(t *testing.T) {
	defer mocktime(t, nil)()
	f := gpiotest.Pin{L: gpio.High, EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, time.Second, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if p.Read() != gpio.High {
		t.Fatal("expected level")
	}
	if p.Read() != gpio.High {
		t.Fatal("expected level")
	}
}

func TestDebounce_WaitForEdge_Got(t *testing.T) {
	offsets := []time.Duration{}
	defer mocktime(t, offsets)()
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level, 1)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	f.EdgesChan <- gpio.Low
	if !p.WaitForEdge(0) {
		t.Fatal("expected edge")
	}
}

func TestDebounce_WaitForEdge_Timeout(t *testing.T) {
	offsets := []time.Duration{}
	defer mocktime(t, offsets)()
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	if p.WaitForEdge(0) {
		t.Fatal("expected no edge")
	}
}

func TestDebounce_RealPin(t *testing.T) {
	defer mocktime(t, []time.Duration{})()
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := p.(gpio.RealPin)
	if !ok {
		t.Fatal("expected gpio.RealPin")
	}
	a, ok := r.Real().(*gpiotest.Pin)
	if !ok {
		t.Fatal("expected gpiotest.Pin")
	}
	if a != &f {
		t.Fatal("expected actual pin")
	}
}

func TestDebounce_RealPin_Deep(t *testing.T) {
	defer mocktime(t, []time.Duration{})()
	f := gpiotest.Pin{EdgesChan: make(chan gpio.Level)}
	p, err := Debounce(&f, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	p, err = Debounce(p, time.Second, 0, gpio.BothEdges)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := p.(gpio.RealPin)
	if !ok {
		t.Fatal("expected gpio.RealPin")
	}
	a, ok := r.Real().(*gpiotest.Pin)
	if !ok {
		t.Fatal("expected gpiotest.Pin")
	}
	if a != &f {
		t.Fatal("expected actual pin")
	}
}

//

func init() {
	resetNow()
}

func resetNow() {
	now = func() time.Time {
		panic("unexpected call")
	}
}

func mocktime(t *testing.T, offsets []time.Duration) func() {
	offset := 0
	d := time.Date(2000, 01, 01, 0, 0, 0, 0, time.UTC)
	now = func() time.Time {
		if offset == len(offsets) {
			t.Fatal("need one more offset")
		}
		v := d.Add(offsets[offset])
		offset++
		return v
	}
	return func() {
		resetNow()
		if offset != len(offsets) {
			t.Fatalf("expected to consume all time mocks; used %d, expectd %d", offset, len(offsets))
		}
	}
}
