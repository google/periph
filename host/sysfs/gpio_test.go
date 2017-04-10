// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
)

func TestPin(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if s := p.String(); s != "foo" {
		t.Fatal(s)
	}
	if s := p.Name(); s != "foo" {
		t.Fatal(s)
	}
	if n := p.Number(); n != 42 {
		t.Fatal(n)
	}
}

func TestPinMock(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	// Fails because open is not mocked.
	if s := p.Function(); s != "ERR" {
		t.Fatal(s)
	}
	if err := p.In(gpio.PullNoChange, gpio.NoEdge); err == nil {
		t.Fatal("need mock")
	}
	if l := p.Read(); l != gpio.Low {
		t.Fatal("broken pin is always low")
	}
	if p.WaitForEdge(-1) {
		t.Fatal("broken pin doesn't have edge triggered")
	}
	if pull := p.Pull(); pull != gpio.PullNoChange {
		t.Fatal(pull)
	}
	if err := p.Out(gpio.High); err == nil {
		t.Fatal("need mock")
	}
}

func TestPin_not_supported(t *testing.T) {
	p := Pin{number: 42, name: "foo", root: "/tmp/gpio/priv/"}
	if err := p.In(gpio.PullDown, gpio.NoEdge); err == nil {
		t.Fatal("pull not supported on sysfs-gpio")
	}
	if err := p.PWM(0); err == nil {
		t.Fatal("sysfs-gpio doens't support PWM")
	}
}
