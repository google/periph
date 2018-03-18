// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package tm1637

import (
	"bytes"
	"errors"
	"log"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

func TestNew(t *testing.T) {
	var clk, data gpiotest.Pin
	dev, err := New(&clk, &data)
	if err != nil {
		t.Fatalf("failed to initialize tm1637: %v", err)
	}
	if s := dev.String(); s != "TM1637{clk:(0), data:(0)}" {
		t.Fatal(s)
	}
	if _, err := dev.Write(Clock(12, 00, true)); err != nil {
		log.Fatalf("failed to write to tm1637: %v", err)
	}
	if err := dev.SetBrightness(Brightness10); err != nil {
		log.Fatalf("failed to write to tm1637: %v", err)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	// TODO(maruel): Check the state of the pins. That's hard since it has to
	// emulate the quasi-I²C protocol.
}

func TestDigits(t *testing.T) {
	expected := []byte{0x3F, 0x06}
	if b := Digits(0, 1); !bytes.Equal(b, expected) {
		t.Fatalf("%v != %v", b, expected)
	}
}

func TestNew_clk_fail(t *testing.T) {
	clk := failPin{fail: true}
	data := gpiotest.Pin{}
	if dev, err := New(&clk, &data); dev != nil || err == nil {
		t.Fatal("data pin is not usable")
	}
}

func TestNew_data_fail(t *testing.T) {
	clk := gpiotest.Pin{}
	data := failPin{fail: true}
	if dev, err := New(&clk, &data); dev != nil || err == nil {
		t.Fatal("data pin is not usable")
	}
}

func TestWrite_fail(t *testing.T) {
	dev, err := New(&gpiotest.Pin{}, &gpiotest.Pin{})
	if err != nil {
		t.Fatalf("failed to initialize tm1637: %v", err)
	}
	if n, err := dev.Write(make([]byte, 7)); n != 0 || err == nil {
		t.Fatal("buffer too long")
	}
}

//

type failPin struct {
	gpiotest.Pin
	fail bool
}

func (f *failPin) Out(l gpio.Level) error {
	if f.fail {
		return errors.New("injected error")
	}
	return nil
}

func init() {
	spin = func(time.Duration) {}
}
