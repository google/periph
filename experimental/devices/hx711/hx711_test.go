// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package hx711

import (
	"errors"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

func TestNew(t *testing.T) {
	clk := gpiotest.Pin{N: "clk"}
	data := gpiotest.Pin{N: "data", EdgesChan: make(chan gpio.Level)}
	d, err := New(&clk, &data)
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "hx711{clk, data}" {
		t.Fatal(s)
	}
	if s := d.Name(); s != "hx711{clk, data}" {
		t.Fatal(s)
	}
	if n := d.Number(); n != -1 {
		t.Fatal(n)
	}
	if f := d.Function(); f != "ADC" {
		t.Fatal(f)
	}
	min, max := d.Range()
	// TODO(davidsansome): Is that the right values?
	if min.Raw != -8388608 {
		t.Fatal(min.Raw)
	}
	if max.Raw != 8388608 {
		t.Fatal(max.Raw)
	}
	if !d.IsReady() {
		t.Fatal("data is low")
	}
	if err := d.SetInputMode(CHANNEL_A_GAIN_128); err != nil {
		t.Fatal(err)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

func TestNew_Fail(t *testing.T) {
	ok := gpiotest.Pin{N: "ok", EdgesChan: make(chan gpio.Level)}
	fail := failPin{gpiotest.Pin{N: "fail"}}
	if _, err := New(&fail, &ok); err == nil {
		t.Fatal("expected failure")
	}
	if _, err := New(&ok, &fail); err == nil {
		t.Fatal("expected failure")
	}
}

func TestRead(t *testing.T) {
	clk := gpiotest.Pin{N: "clk"}
	data := gpiotest.Pin{N: "data", EdgesChan: make(chan gpio.Level)}
	d, err := New(&clk, &data)
	if err != nil {
		t.Fatal(err)
	}
	// TODO(davidsansome): Real testing.
	r, err := d.Read()
	if err != nil {
		t.Fatal(err)
	}
	if r.Raw != 0 {
		t.Fatal("we should implement something")
	}
}

func TestReadTimeout(t *testing.T) {
	clk := gpiotest.Pin{N: "clk"}
	data := gpiotest.Pin{N: "data", EdgesChan: make(chan gpio.Level)}
	d, err := New(&clk, &data)
	if err != nil {
		t.Fatal(err)
	}
	// TODO(davidsansome): Real testing.
	r, err := d.ReadTimeout(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if r != 0 {
		t.Fatal("we should implement something")
	}
}

func TestReadContinuous(t *testing.T) {
	clk := gpiotest.Pin{N: "clk"}
	data := gpiotest.Pin{N: "data", EdgesChan: make(chan gpio.Level)}
	d, err := New(&clk, &data)
	if err != nil {
		t.Fatal(err)
	}
	// TODO(davidsansome): Real testing.
	c := d.ReadContinuous()
	if c == nil {
		t.Fatal("expected chan")
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
}

//

type failPin struct {
	gpiotest.Pin
}

func (f *failPin) In(pull gpio.Pull, edge gpio.Edge) error {
	return errors.New("fail")
}

func (f *failPin) Out(l gpio.Level) error {
	return errors.New("fail")
}
