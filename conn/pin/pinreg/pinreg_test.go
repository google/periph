// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pinreg

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/pin"
)

func TestAll(t *testing.T) {
	defer reset()
	gpio2 := &gpiotest.Pin{N: "GPIO2", Num: 2, Fn: "I2C1_SDA"}
	gpio3 := &gpiotest.Pin{N: "GPIO3", Num: 3, Fn: "I2C1_SCL"}
	p := [][]pin.Pin{
		{pin.GROUND, pin.V3_3},
		{gpio2, gpio3},
	}
	if err := Register("P1", p); err != nil {
		t.Fatal(err)
	}
	if len(allHeaders) != len(All()) {
		t.Fatal("unexpected register")
	}
}

func TestRegister_twice(t *testing.T) {
	defer reset()
	gpio2 := &gpiotest.Pin{N: "GPIO2", Num: 2, Fn: "I2C1_SDA"}
	if err := Register("P1", [][]pin.Pin{{gpio2}}); err != nil {
		t.Fatal(err)
	}
	if err := Register("P1", [][]pin.Pin{{gpio2}}); err == nil {
		t.Fatal("can't register twice")
	}
}

func TestRegister_nil(t *testing.T) {
	defer reset()
	if err := Register("P1", [][]pin.Pin{{nil}}); err == nil {
		t.Fatal("can't register nil pin")
	}
}

func TestIsConnected(t *testing.T) {
	defer reset()
	gpio2 := &gpiotest.Pin{N: "GPIO2", Num: 2, Fn: "I2C1_SDA"}
	gpio3 := &gpiotest.Pin{N: "GPIO3", Num: 3, Fn: "I2C1_SCL"}
	alias := &pinAlias{Pin: &gpiotest.Pin{N: "ALIAS", Num: 4}, alias: gpio2}
	p := [][]pin.Pin{
		{pin.GROUND, pin.V3_3},
		{gpio2, gpio3},
		{alias, alias},
	}
	if err := Register("P1", p); err != nil {
		t.Fatal(err)
	}
	if !IsConnected(pin.V3_3) {
		t.Fatal("V3_3 should be connected")
	}
	if IsConnected(pin.V5) {
		t.Fatal("V5 should not be connected")
	}
	if !IsConnected(gpio2) {
		t.Fatal("GPIO2 should be connected")
	}
}

//

func reset() {
	mu.Lock()
	defer mu.Unlock()
	allHeaders = map[string][][]pin.Pin{}
	byPin = map[string]position{}
}

type pinAlias struct {
	*gpiotest.Pin
	alias *gpiotest.Pin
}

func (p *pinAlias) Real() gpio.PinIO {
	return p.alias
}
