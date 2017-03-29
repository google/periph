// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pinreg

import (
	"testing"

	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/pin"
)

func TestAll(t *testing.T) {
	if len(allHeaders) != len(All()) {
		t.Fail()
	}
}

func TestIsConnected(t *testing.T) {
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

var (
	gpio2 = &gpiotest.Pin{N: "GPIO2", Num: 2, Fn: "I2C1_SDA"}
	gpio3 = &gpiotest.Pin{N: "GPIO3", Num: 3, Fn: "I2C1_SCL"}
)

func init() {
	if err := Register("P1", [][]pin.Pin{
		{pin.GROUND, pin.V3_3},
		{gpio2, gpio3},
	}); err != nil {
		panic(err)
	}
}
