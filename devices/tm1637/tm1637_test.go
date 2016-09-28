// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package tm1637

import (
	"log"
	"testing"

	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/conn/gpio/gpiotest"
	"github.com/google/pio/host"
)

func TestNew(t *testing.T) {
	var clk, data gpiotest.Pin
	dev, err := New(&clk, &data)
	if err != nil {
		t.Fatalf("failed to initialize tm1637: %v", err)
	}
	if _, err := dev.Write(Clock(12, 00, true)); err != nil {
		log.Fatalf("failed to write to tm1637: %v", err)
	}
	// TODO(maruel): Check the state of the pins. That's hard since it has to
	// emulate the quasi-I²C protocol.
}

func Example() {
	if _, err := host.Init(); err != nil {
		log.Fatalf("failed to initialize pio: %v", err)
	}
	dev, err := New(gpio.ByNumber(6), gpio.ByNumber(12))
	if err != nil {
		log.Fatalf("failed to initialize tm1637: %v", err)
	}
	if err := dev.SetBrightness(Brightness10); err != nil {
		log.Fatalf("failed to set brightness on tm1637: %v", err)
	}
	if _, err := dev.Write(Clock(12, 00, true)); err != nil {
		log.Fatalf("failed to write to tm1637: %v", err)
	}
}
