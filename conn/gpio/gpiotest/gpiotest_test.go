// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiotest

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
)

func TestPin(t *testing.T) {
	p := &Pin{N: "GPIO1", Num: 10, Fn: "I2C1_SDA"}
	if s := p.String(); s != "GPIO1(10)" {
		t.Fatal(s)
	}
	if n := p.Number(); n != 10 {
		t.Fatal(n)
	}
	if err := p.In(gpio.PullDown, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if l := p.Read(); l != gpio.Low {
		t.Fatal(l)
	}
	if err := p.In(gpio.PullUp, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if l := p.Read(); l != gpio.High {
		t.Fatal(l)
	}
	if pull := p.Pull(); pull != gpio.PullUp {
		t.Fatal(pull)
	}
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}
	if err := p.Halt(); err != nil {
		t.Fatal(err)
	}
}

func TestPin_edge(t *testing.T) {
	p := &Pin{N: "GPIO1", Num: 1, Fn: "I2C1_SDA", EdgesChan: make(chan gpio.Level, 1)}
	p.EdgesChan <- gpio.High
	if !p.WaitForEdge(-1) {
		t.Fail()
	}
	if p.Read() != gpio.High {
		t.Fail()
	}
	if p.WaitForEdge(time.Millisecond) {
		t.Fail()
	}
	p.EdgesChan <- gpio.Low
	if !p.WaitForEdge(time.Minute) {
		t.Fail()
	}
	if p.Read() != gpio.Low {
		t.Fail()
	}
}

func TestPin_fail(t *testing.T) {
	p := &Pin{N: "GPIO1", Num: 1, Fn: "I2C1_SDA"}
	if err := p.In(gpio.Float, gpio.BothEdges); err == nil {
		t.Fatal()
	}
}

func TestLogPinIO(t *testing.T) {
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
		defer log.SetOutput(os.Stderr)
	}
	p := &Pin{}
	l := &LogPinIO{p}
	if l.Real() != p {
		t.Fatal("unexpected real pin")
	}
	if err := l.Out(gpio.High); err != nil {
		t.Fatal(err)
	}
	if err := l.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if l.Read() != gpio.High {
		t.Fatal("unexpected level")
	}
	if l.Pull() != gpio.PullNoChange {
		t.Fatal("unexpected pull")
	}
	if l.WaitForEdge(0) {
		t.Fatal("unexpected edge")
	}
}

func TestAll(t *testing.T) {
	if 2 != len(gpioreg.All()) {
		t.Fail()
	}
}

func TestByName(t *testing.T) {
	if gpioreg.ByName("GPIO0") != nil {
		t.Fatal("GPIO0 doesn't exist")
	}
	if gpioreg.ByName("GPIO2") != gpio2 {
		t.Fatal("GPIO2 should have been found")
	}
	if gpioreg.ByName("1") != nil {
		t.Fatal("1 exist")
	}
	p := gpioreg.ByName("2")
	if p == nil {
		t.Fatal("2 missing")
	}
	if r, ok := p.(gpio.RealPin); !ok {
		t.Fatalf("unexpected alias: %v", r)
	}
	p = gpioreg.ByName("3")
	if p == nil {
		t.Fatal("3 missing")
	}
	r, ok := p.(gpio.RealPin)
	if !ok || r.Real().Name() != "GPIO3" {
		t.Fatalf("expected alias, got: %T", p)
	}
	p = r.Real()
	if err := p.(gpio.PinPWM).PWM(gpio.DutyHalf, time.Millisecond); err != nil {
		t.Fatalf("unexpected failure: %v", err)
	}
}

//

var (
	gpio2 = &Pin{N: "GPIO2", Num: 2, Fn: "I2C1_SDA"}
	gpio3 = &PinPWM{Pin: Pin{N: "GPIO3", Num: 3, Fn: "I2C1_SCL"}}
)

func init() {
	if err := gpioreg.Register(gpio2, true); err != nil {
		panic(err)
	}
	if err := gpioreg.Register(gpio3, false); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias("2", "GPIO2"); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias("3", "GPIO3"); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias(gpio2.Function(), gpio2.Name()); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias(gpio3.Function(), gpio3.Name()); err != nil {
		panic(err)
	}
}
