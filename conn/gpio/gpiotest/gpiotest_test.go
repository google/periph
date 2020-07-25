// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiotest

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
)

func TestPin(t *testing.T) {
	p := &Pin{N: "GPIO1", Num: 10, Fn: "I2C1_SDA"}
	// conn.Resource
	if s := p.String(); s != "GPIO1(10)" {
		t.Fatal(s)
	}
	if err := p.Halt(); err != nil {
		t.Fatal(err)
	}
	// pin.Pin
	if n := p.Number(); n != 10 {
		t.Fatal(n)
	}
	if n := p.Name(); n != "GPIO1" {
		t.Fatal(n)
	}
	if f := p.Function(); f != "I2C1_SDA" {
		t.Fatal(f)
	}
	// pin.PinFunc
	if f := p.Func(); f != i2c.SDA.Specialize(1, -1) {
		t.Fatal(f)
	}
	if f := p.SupportedFuncs(); !reflect.DeepEqual(f, []pin.Func{gpio.IN, gpio.OUT}) {
		t.Fatal(f)
	}
	if err := p.SetFunc(i2c.SCL); err == nil {
		t.Fatal("expected failure")
	}
	// gpio.PinIn
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
	if pull := p.DefaultPull(); pull != gpio.PullUp {
		t.Fatal(pull)
	}
	// gpio.PinOut
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}
	if err := p.PWM(gpio.DutyHalf, physic.KiloHertz); err != nil {
		t.Fatalf("unexpected failure: %v", err)
	}
}

func TestPin_edge(t *testing.T) {
	p := &Pin{N: "GPIO1", Num: 1, Fn: "I2C1_SDA", EdgesChan: make(chan gpio.Level, 1)}
	p.EdgesChan <- gpio.High
	if !p.WaitForEdge(-1) {
		t.Fatal("expected edge")
	}
	if l := p.Read(); l != gpio.High {
		t.Fatalf("unexpected %s", l)
	}
	if p.WaitForEdge(time.Millisecond) {
		t.Fatal("unexpected edge")
	}
	p.EdgesChan <- gpio.Low
	if !p.WaitForEdge(time.Minute) {
		t.Fatal("expected edge")
	}
	if l := p.Read(); l != gpio.Low {
		t.Fatalf("unexpected %s", l)
	}
}

func TestPin_fail(t *testing.T) {
	p := &Pin{N: "GPIO1", Num: 1, Fn: "I2C1_SDA"}
	if err := p.In(gpio.Float, gpio.BothEdges); err == nil {
		t.Fatal()
	}
}

func TestLogPinIO(t *testing.T) {
	p := &Pin{}
	l := &LogPinIO{p}
	if l.Real() != p {
		t.Fatal("unexpected real pin")
	}
	// gpio.PinIn
	if err := l.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if v := l.Read(); v != gpio.Low {
		t.Fatalf("unexpected level %v", v)
	}
	if l.Pull() != gpio.PullNoChange {
		t.Fatal("unexpected pull")
	}
	if l.WaitForEdge(0) {
		t.Fatal("unexpected edge")
	}
	// gpio.PinOut
	if err := l.Out(gpio.High); err != nil {
		t.Fatal(err)
	}
	if v := l.Read(); v != gpio.High {
		t.Fatalf("unexpected level %v", v)
	}
	if err := l.PWM(gpio.DutyHalf, physic.KiloHertz); err != nil {
		t.Fatalf("unexpected failure: %v", err)
	}
}

func TestAll(t *testing.T) {
	if len(gpioreg.All()) != 2 {
		t.Fatal("expected two pins registered for test")
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
	if err := p.PWM(gpio.DutyHalf, physic.KiloHertz); err != nil {
		t.Fatalf("unexpected failure: %v", err)
	}
}

//

var (
	gpio2 = &Pin{N: "GPIO2", Num: 2, Fn: "I2C1_SDA"}
	gpio3 = &Pin{N: "GPIO3", Num: 3, Fn: "I2C1_SCL"}
)

func init() {
	if err := gpioreg.Register(gpio2); err != nil {
		panic(err)
	}
	if err := gpioreg.Register(gpio3); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias("2", "GPIO2"); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias("3", "GPIO3"); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias(string(gpio2.Func()), gpio2.Name()); err != nil {
		panic(err)
	}
	if err := gpioreg.RegisterAlias(string(gpio3.Func()), gpio3.Name()); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	os.Exit(m.Run())
}
