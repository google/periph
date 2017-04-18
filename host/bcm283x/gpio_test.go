// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
)

func TestPresent(t *testing.T) {
	Present()
	if gpioMemory != nil {
		t.Fatal("gpioMemory should not be initialized")
	}
}

func TestPin(t *testing.T) {
	p := Pin{name: "Foo", number: 42, defaultPull: gpio.PullDown}

	if s := p.String(); s != "Foo" {
		t.Fatal(s)
	}
	if s := p.Name(); s != "Foo" {
		t.Fatal(s)
	}
	if n := p.Number(); n != 42 {
		t.Fatal(n)
	}
	if d := p.DefaultPull(); d != gpio.PullDown {
		t.Fatal(d)
	}
	if s := p.Function(); s != "UART1_RTS" {
		t.Fatal(s)
	}
	if p.In(gpio.PullNoChange, gpio.NoEdge) == nil {
		t.Fatal("not initialized")
	}
	if d := p.Read(); d != gpio.Low {
		t.Fatal(d)
	}
	if d := p.Pull(); d != gpio.PullNoChange {
		t.Fatal(d)
	}
	if p.WaitForEdge(-1) {
		t.Fatal("edge not initialized")
	}
	if p.Out(gpio.Low) == nil {
		t.Fatal("not initialized")
	}

	defer func() {
		gpioMemory = nil
	}()
	gpioMemory = &gpioMap{}

	if err := p.In(gpio.PullDown, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.PullUp, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.Float, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if s := p.Function(); s != "In/Low" {
		t.Fatal(s)
	}
	if d := p.Read(); d != gpio.Low {
		t.Fatal(d)
	}
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}
	if s := p.Function(); s != "Out/Low" {
		t.Fatal(s)
	}
	if err := p.Out(gpio.High); err != nil {
		t.Fatal(err)
	}

	p.number = 25
	p.setFunction(alt0)
	if s := p.Function(); s != "<Alt0>" {
		t.Fatal(s)
	}
	p.setFunction(alt1)
	if s := p.Function(); s != "<Alt1>" {
		t.Fatal(s)
	}
	p.setFunction(alt2)
	if s := p.Function(); s != "<Alt2>" {
		t.Fatal(s)
	}
	p.setFunction(alt3)
	if s := p.Function(); s != "<Alt3>" {
		t.Fatal(s)
	}
	p.setFunction(alt4)
	if s := p.Function(); s != "<Alt4>" {
		t.Fatal(s)
	}
	p.setFunction(alt5)
	if s := p.Function(); s != "<Alt5>" {
		t.Fatal(s)
	}

	p.number = 45
	p.setFunction(alt0)
	if s := p.Function(); s != "PWM1_OUT" {
		t.Fatal(s)
	}
	p.setFunction(alt1)
	if s := p.Function(); s != "I2C0_SCL" {
		t.Fatal(s)
	}
	p.setFunction(alt2)
	if s := p.Function(); s != "I2C1_SCL" {
		t.Fatal(s)
	}
	p.setFunction(alt4)
	if s := p.Function(); s != "SPI2_CS2" {
		t.Fatal(s)
	}

	p.number = 33
	p.setFunction(alt3)
	if s := p.Function(); s != "UART0_RXD" {
		t.Fatal(s)
	}
	p.setFunction(alt5)
	if s := p.Function(); s != "UART1_RXD" {
		t.Fatal(s)
	}
}

func TestDriver(t *testing.T) {
	d := driverGPIO{}
	if s := d.String(); s != "bcm283x-gpio" {
		t.Fatal(s)
	}
	if s := d.Prerequisites(); s != nil {
		t.Fatal(s)
	}
	d.Init()
}

func TestSetSpeed(t *testing.T) {
	if setSpeed(1000) == nil {
		t.Fatal("cannot change live")
	}
}
