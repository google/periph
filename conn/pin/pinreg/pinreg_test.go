// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pinreg

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/pin"
)

func TestAll(t *testing.T) {
	defer reset(t)
	gpio2 := &gpiotest.Pin{N: "IMPROBABLE_PIN2", Num: 2, Fn: "I2C1_SDA"}
	gpio3 := &gpiotest.Pin{N: "IMPROBABLE_PIN3", Num: 3, Fn: "I2C1_SCL"}
	p := [][]pin.Pin{
		{pin.GROUND, pin.V3_3},
		{gpio2, gpio3},
	}
	if err := Register("IMPROBABLE_HEADER", p); err != nil {
		t.Fatal(err)
	}
	if len(allHeaders) != len(All()) {
		t.Fatal("unexpected register")
	}
	if err := Unregister("IMPROBABLE_HEADER"); err != nil {
		t.Fatal(err)
	}
}

func TestRegister_twice(t *testing.T) {
	defer reset(t)
	gpio2 := &gpiotest.Pin{N: "IMPROBABLE_PIN2", Num: 2, Fn: "I2C1_SDA"}
	if err := Register("IMPROBABLE_HEADER", [][]pin.Pin{{gpio2}}); err != nil {
		t.Fatal(err)
	}
	if err := Register("IMPROBABLE_HEADER", [][]pin.Pin{{gpio2}}); err == nil {
		t.Fatal("can't register twice")
	}
	if err := Unregister("IMPROBABLE_HEADER"); err != nil {
		t.Fatal(err)
	}
}

func TestRegister_nil(t *testing.T) {
	defer reset(t)
	if err := Register("IMPROBABLE_HEADER", [][]pin.Pin{{nil}}); err == nil {
		t.Fatal("can't register nil pin")
	}
}

func TestRegister_bad_pin(t *testing.T) {
	defer reset(t)
	gpio2 := &gpiotest.Pin{N: "IMPROBABLE_HEADER_1", Num: 2, Fn: "I2C1_SDA"}
	if err := gpioreg.Register(gpio2); err != nil {
		t.Fatal(err)
	}
	if Register("IMPROBABLE_HEADER", [][]pin.Pin{{gpio2}}) == nil {
		t.Fatal("should have failed due to alias conflict")
	}
}

func TestIsConnected(t *testing.T) {
	defer reset(t)
	gpio2 := &gpiotest.Pin{N: "IMPROBABLE_PIN2", Num: 2, Fn: "I2C1_SDA"}
	gpio3 := &gpiotest.Pin{N: "IMPROBABLE_PIN3", Num: 3, Fn: "I2C1_SCL"}
	alias := &pinAlias{Pin: &gpiotest.Pin{N: "ALIAS", Num: 4}, alias: gpio2}
	p := [][]pin.Pin{
		{pin.GROUND, pin.V3_3},
		{gpio2, gpio3},
		{alias, alias},
	}
	if err := Register("IMPROBABLE_HEADER", p); err != nil {
		t.Fatal(err)
	}
	if !IsConnected(pin.V3_3) {
		t.Fatal("V3_3 should be connected")
	}
	if IsConnected(pin.V5) {
		t.Fatal("V5 should not be connected")
	}
	if !IsConnected(gpio2) {
		t.Fatal("IMPROBABLE_PIN2 should be connected")
	}
	if err := Unregister("IMPROBABLE_HEADER"); err != nil {
		t.Fatal(err)
	}
}

func TestUnregister(t *testing.T) {
	defer reset(t)
	gpio2 := &gpiotest.Pin{N: "IMPROBABLE_PIN2", Num: 2, Fn: "I2C1_SDA"}
	if err := Register("IMPROBABLE_HEADER", [][]pin.Pin{{gpio2}}); err != nil {
		t.Fatal(err)
	}
	if err := Unregister("IMPROBABLE_HEADER"); err != nil {
		t.Fatal(err)
	}
}

func TestUnregister_unknown(t *testing.T) {
	defer reset(t)
	if Unregister("IMPROBABLE_HEADER") == nil {
		t.Fatal("can't unregister unregistered header")
	}
}

func TestUnregister_missing_alias(t *testing.T) {
	defer reset(t)
	gpio2 := &gpiotest.Pin{N: "IMPROBABLE_PIN2", Num: 2, Fn: "I2C1_SDA"}
	if err := gpioreg.Register(gpio2); err != nil {
		t.Fatal(err)
	}
	if err := Register("IMPROBABLE_HEADER", [][]pin.Pin{{gpio2}}); err != nil {
		t.Fatal(err)
	}
	if err := gpioreg.Unregister("IMPROBABLE_HEADER_1"); err != nil {
		t.Fatal(err)
	}
	if Unregister("IMPROBABLE_HEADER") == nil {
		t.Fatal("Should fail unregistering aliases")
	}
	if err := gpioreg.Unregister("IMPROBABLE_PIN2"); err != nil {
		t.Fatal(err)
	}
}

//

func reset(t *testing.T) {
	mu.Lock()
	defer mu.Unlock()
	allHeaders = map[string][][]pin.Pin{}
	byPin = map[string]position{}
	// Take no chance, they could still be there, but make sure to fail the test
	// in this case.
	names := []string{"IMPROBABLE_HEADER_1", "IMPROBABLE_HEADER_2", "IMPROBABLE_HEADER_3", "IMPROBABLE_HEADER_4", "IMPROBABLE_PIN2", "IMPROBABLE_PIN3"}
	for _, n := range names {
		if gpioreg.Unregister(n) == nil {
			t.Fatalf("%q wasn't cleaned up correctly", n)
		}
	}
}

func init() {
	reset(nil)
}

type pinAlias struct {
	*gpiotest.Pin
	alias *gpiotest.Pin
}

func (p *pinAlias) Real() gpio.PinIO {
	return p.alias
}
