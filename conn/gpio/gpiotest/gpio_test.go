// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiotest

import (
	"testing"

	"github.com/google/periph/conn/gpio"
)

func TestAll(t *testing.T) {
	if 2 != len(gpio.All()) {
		t.Fail()
	}
}

func TestByNumber(t *testing.T) {
	if gpio.ByNumber(1) != nil {
		t.Fatal("1 exist")
	}
	if gpio.ByNumber(2) != gpio2 {
		t.Fatal("2 missing")
	}
}

func TestByName(t *testing.T) {
	if gpio.ByName("GPIO0") != nil {
		t.Fatal("GPIO0 doesn't exist")
	}
	if gpio.ByName("GPIO2") != gpio2 {
		t.Fatal("GPIO2 should have been found")
	}
}

func TestByFunction(t *testing.T) {
	if gpio.ByFunction("SPI1_MOSI") != nil {
		t.Fatal("spi doesn't exist")
	}
	if gpio.ByFunction("I2C1_SDA") != gpio2 {
		t.Fatal("I2C1_SDA should have been found")
	}
}

//

var (
	gpio2  = &Pin{N: "GPIO2", Num: 2, Fn: "I2C1_SDA"}
	gpio2a = &Pin{N: "GPIO2a", Num: 2}
	gpio3  = &Pin{N: "GPIO3", Num: 3, Fn: "I2C1_SCL"}
)

func init() {
	if err := gpio.Register(gpio2, true); err != nil {
		panic(err)
	}
	if err := gpio.Register(gpio2a, false); err != nil {
		panic(err)
	}
	if err := gpio.Register(gpio3, false); err != nil {
		panic(err)
	}
	gpio.MapFunction(gpio2.Function(), gpio2.Number())
	gpio.MapFunction(gpio3.Function(), gpio3.Number())
}
