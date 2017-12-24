// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"fmt"
	"testing"
)

func TestClockCtl_String(t *testing.T) {
	t.Parallel()
	data := []struct {
		c clockCtl
		s string
	}{
		{^clockCtl(0), "Mash3|Flip|Busy|Kill|Enable|GND(15)|clockCtl(0xfffff840)"},
		{clockPasswdCtl, "PWD|GND(0Hz)"},
		{clockMash1 | clockSrc19dot2MHz, "Mash1|19.2MHz"},
		{clockMash2 | clockSrcTestDebug0, "Mash2|Debug0(0Hz)"},
		{clockSrcTestDebug1, "Debug1(0Hz)"},
		{clockSrcPLLA, "PLLA(0Hz)"},
		{clockSrcPLLC, "PLLD(1000MHz)"},
		{clockSrcPLLD, "PLLD(500MHz)"},
		{clockSrcHDMI, "HDMI(216MHz)"},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()
			if s := line.c.String(); s != line.s {
				t.Fatal(s)
			}
		})
	}
}

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

func TestCalcSource_exact(t *testing.T) {
	t.Parallel()
	data := []struct {
		desiredHz          uint64
		maxWaitCycles      uint32
		src                clockCtl
		clkDiv, waitCycles uint32
	}{
		{
			// Lowest clean exact clock.
			150,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 4000, 32,
		},
		{
			200,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 4000, 24,
		},
		{
			250,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3840, 20,
		},
		{
			300,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 4000, 16,
		},
		{
			1500,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3200, 4,
		},
		{
			1000,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3840, 5,
		},
		{
			2000,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3200, 3,
		},
		{
			2500,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3840, 2,
		},
		{
			3000,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3200, 2,
		},
		{
			10000,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 1920, 1,
		},
		{
			100000,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 192, 1,
		},
		{
			120000,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 160, 1,
		},
		{
			125000,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 4000, 1,
		},
		{
			1000000,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 500, 1,
		},
		{
			10000000,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 50, 1,
		},
		{
			25000000,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 20, 1,
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()
			src, clkDiv, waitCycles, hz, err := calcSource(line.desiredHz, line.maxWaitCycles)
			if src != line.src || line.clkDiv != clkDiv || line.waitCycles != waitCycles || line.desiredHz != hz || err != nil {
				t.Fatalf("calcSource(%dHz, %d) = %s / %d / %d = %dHz  expected %s / %d / %d = %dHz",
					line.desiredHz, line.maxWaitCycles,
					src, clkDiv, waitCycles, hz,
					line.src, line.clkDiv, line.waitCycles, line.desiredHz)
			}
		})
	}
}

func TestCalcSource_oversample(t *testing.T) {
	t.Parallel()
	data := []struct {
		desiredHz          uint64
		maxWaitCycles      uint32
		src                clockCtl
		clkDiv, waitCycles uint32
		hz                 uint64
	}{
		{
			// 150x
			1,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 32, 150,
		},
		{
			// 75x
			2,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 32, 150,
		},
		{
			// 15x
			10,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 32, 150,
		},
		{
			// 2x
			100,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 24, 200,
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()
			src, clkDiv, waitCycles, hz, err := calcSource(line.desiredHz, line.maxWaitCycles)
			if src != line.src || line.clkDiv != clkDiv || line.waitCycles != waitCycles || line.hz != hz || err != nil {
				t.Fatalf("calcSource(%dHz, %d) = %s / %d / %d = %dHz  expected %s / %d / %d = %dHz",
					line.desiredHz, line.maxWaitCycles,
					src, clkDiv, waitCycles, hz,
					line.src, line.clkDiv, line.waitCycles, line.hz)
			}
		})
	}
}

func TestCalcSource_inexact(t *testing.T) {
	t.Parallel()
	// clockDiviMax is 4095, an odd number.
	data := []struct {
		desiredHz     uint64
		maxWaitCycles uint32
		//src           clockCtl
		//clkDiv, waitCycles int
		//hz                 uint64
	}{
		{
			// 2930/1465 = 2x
			clk19dot2MHz / clockDiviMax / (dmaWaitcyclesMax + 1),
			dmaWaitcyclesMax + 1,
			//clockSrc19dot2MHz, 2121, 31, 292,
		},
		{
			// 93795/46886 = 2.0004 (error: 0.025%)
			clk19dot2MHz / clockDiviMax,
			dmaWaitcyclesMax + 1,
			//clockSrcPLLD, 2051, 26, 9376,
		},
		{
			// 1465.2014/7 = 209.31x (error: 0.15%)
			7,
			dmaWaitcyclesMax + 1,
			//clockSrc19dot2MHz, 4095, 32, 146, // 146.52014
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()
			src, clkDiv, waitCycles, hz, err := calcSource(line.desiredHz, line.maxWaitCycles)
			if src != 0 || clkDiv != 0 || waitCycles != 0 || hz != 0 || err == nil {
				t.Fatalf("inexact match is not yet implemented: %s, %d, %d, %d", src, clkDiv, waitCycles, hz)
			}
		})
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
	c := clock{}
	if _, _, err := c.set(0, dmaWaitcyclesMax+1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := c.set(100, dmaWaitcyclesMax+1); err == nil {
		t.Fatal("it's not a register so it doesn't react")
	}
	if _, _, err := c.set(25000001, dmaWaitcyclesMax+1); err == nil {
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
	for i := 0; i < b.N; i++ {
		src, clkDiv, waitCycles, hz, err := calcSource(120000, dmaWaitcyclesMax+1)
		if src != clockSrc19dot2MHz || clkDiv != 160 || waitCycles != 1 || hz != 120000 || err != nil {
			b.Fatal(clkDiv, waitCycles, hz)
		}
	}
}

func BenchmarkCalcSource_Oversample(b *testing.B) {
	for i := 0; i < b.N; i++ {
		src, clkDiv, waitCycles, hz, err := calcSource(10, dmaWaitcyclesMax+1)
		if src != clockSrc19dot2MHz || clkDiv != 4000 || waitCycles != 32 || hz != 150 || err != nil {
			b.Fatal(clkDiv, waitCycles, hz)
		}
	}
}

func BenchmarkCalcSource_Inexact(b *testing.B) {
	// TODO(maruel): It is really too slow.
	for i := 0; i < b.N; i++ {
		src, clkDiv, waitCycles, hz, err := calcSource(7, dmaWaitcyclesMax+1)
		if src != 0 || clkDiv != 0 || waitCycles != 0 || hz != 0 || err == nil {
			b.Fatalf("inexact match is not yet implemented")
		}
	}
}
