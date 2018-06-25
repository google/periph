// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"testing"

	"periph.io/x/periph/conn/physic"
)

func TestClockDiv_String(t *testing.T) {
	if s := clockDiv(1 << 12).String(); s != "1.0" {
		t.Fatal(s)
	}
	if s := clockDiv(1<<12 | 1).String(); s != "1.(1/4095)" {
		t.Fatal(s)
	}
}

func TestFindDivisorExact(t *testing.T) {
	t.Parallel()
	m, n := findDivisorExact(clk19dot2MHz, 216*7, dmaWaitcyclesMax+1)
	if m != 0 || n != 0 {
		t.Fatalf("%d != %d || %d != %d", m, 0, n, 0)
	}
}

func TestCalcSource_err(t *testing.T) {
	if _, _, _, _, err := calcSource(0, dmaWaitcyclesMax+1); err == nil {
		t.Fatal("0 hz")
	}
	if _, _, _, _, err := calcSource(25000001, dmaWaitcyclesMax+1); err == nil {
		t.Fatal("0 hz")
	}
}

func TestClock(t *testing.T) {
	// Necessary to zap out setRaw failing on non-working fake CPU memory map.
	oldClockRawError := clockRawError
	clockRawError = nil
	defer func() {
		clockRawError = oldClockRawError
	}()
	c := clock{}
	if _, _, err := c.set(0, dmaWaitcyclesMax+1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.set(100*physic.Hertz, dmaWaitcyclesMax+1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.set(25000001*physic.Hertz, dmaWaitcyclesMax+1); err == nil {
		t.Fatal("freq too high")
	}
	if s := c.String(); s != "PWD|Enable|19.2MHz / 4000.(1509949440/4095)" {
		t.Fatal(s)
	}

	if c.setRaw(clockSrc19dot2MHz, 0) == nil {
		t.Fatal("invalid div")
	}
	if c.setRaw(0, 1) == nil {
		t.Fatal("invalid src")
	}
}

func TestClockMap(t *testing.T) {
	c := clockMap{}
	expected := "{\n  gp0: GND(0Hz) / 0.0,\n  gp1: GND(0Hz) / 0.0,\n  gp2: GND(0Hz) / 0.0,\n  pcm: GND(0Hz) / 0.0w,\n  pwm: GND(0Hz) / 0.0,\n}"
	if s := c.GoString(); s != expected {
		t.Fatal(s)
	}
}

//

func BenchmarkCalcSource_Exact(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		src, clkDiv, waitCycles, actual, err := calcSource(120*physic.KiloHertz, dmaWaitcyclesMax+1)
		if src != clockSrc19dot2MHz || clkDiv != 160 || waitCycles != 1 || actual != 120*physic.KiloHertz || err != nil {
			b.Fatal(src, clkDiv, waitCycles, actual, err)
		}
	}
}

func BenchmarkCalcSource_Oversample(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		src, clkDiv, waitCycles, actual, err := calcSource(10*physic.Hertz, dmaWaitcyclesMax+1)
		if src != clockSrc19dot2MHz || clkDiv != 4000 || waitCycles != 32 || actual != 150*physic.Hertz || err != nil {
			b.Fatal(clkDiv, waitCycles, actual)
		}
	}
}

func BenchmarkCalcSource_Inexact(b *testing.B) {
	// TODO(maruel): It is really too slow.
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		src, clkDiv, waitCycles, hz, err := calcSource(7, dmaWaitcyclesMax+1)
		if src != 0 || clkDiv != 0 || waitCycles != 0 || hz != 0 || err == nil {
			b.Fatalf("inexact match is not yet implemented")
		}
	}
}
