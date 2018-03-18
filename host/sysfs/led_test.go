// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
)

func TestLEDByName(t *testing.T) {
	if _, err := LEDByName("FOO"); err == nil {
		t.Fail()
	}
}

func TestLED(t *testing.T) {
	l := LED{number: 42, name: "Glow", root: "/tmp/led/priv/"}
	if s := l.String(); s != "Glow(42)" {
		t.Fatal(s)
	}
	if s := l.Name(); s != "Glow" {
		t.Fatal(s)
	}
	if n := l.Number(); n != 42 {
		t.Fatal(n)
	}
}

func TestLEDMock(t *testing.T) {
	l := LED{number: 42, name: "Glow", root: "/tmp/led/priv/"}
	if s := l.Function(); s != "LED/Off" {
		t.Fatal(s)
	}
	if err := l.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if l := l.Read(); l != gpio.Low {
		t.Fatal("need mock")
	}
	if err := l.Out(gpio.High); err == nil {
		t.Fatal("need mock")
	}
}

func TestLED_not_supported(t *testing.T) {
	l := LED{number: 42, name: "Glow", root: "/tmp/led/priv/"}
	if err := l.In(gpio.PullDown, gpio.NoEdge); err == nil {
		t.Fatal("sysfs-led no real In() support")
	}
	if l.WaitForEdge(-1) {
		t.Fatal("not supported")
	}
	if pull := l.Pull(); pull != gpio.PullNoChange {
		t.Fatal(pull)
	}
}

func TestLEDDriver(t *testing.T) {
	if len((&driverLED{}).Prerequisites()) != 0 {
		t.Fatal("unexpected LED prerequisites")
	}
}
