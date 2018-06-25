// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build go1.8

package bcm283x

import (
	"fmt"
	"testing"

	"periph.io/x/periph/conn/physic"
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

func TestCalcSource_exact(t *testing.T) {
	t.Parallel()
	data := []struct {
		desired            physic.Frequency
		maxWaitCycles      uint32
		src                clockCtl
		clkDiv, waitCycles uint32
	}{
		{
			// Lowest clean exact clock.
			150 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 4000, 32,
		},
		{
			200 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 4000, 24,
		},
		{
			250 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3840, 20,
		},
		{
			300 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 4000, 16,
		},
		{
			1500 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3200, 4,
		},
		{
			1000 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3840, 5,
		},
		{
			2000 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3200, 3,
		},
		{
			2500 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3840, 2,
		},
		{
			3 * physic.KiloHertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 3200, 2,
		},
		{
			10 * physic.KiloHertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 1920, 1,
		},
		{
			100 * physic.KiloHertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 192, 1,
		},
		{
			120 * physic.KiloHertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz, 160, 1,
		},
		{
			125 * physic.KiloHertz,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 4000, 1,
		},
		{
			1 * physic.MegaHertz,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 500, 1,
		},
		{
			10 * physic.MegaHertz,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 50, 1,
		},
		{
			25 * physic.MegaHertz,
			dmaWaitcyclesMax + 1,
			clockSrcPLLD, 20, 1,
		},
		{
			25 * physic.MegaHertz,
			125,
			clockSrcPLLD, 20, 1,
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()
			src, clkDiv, waitCycles, f, err := calcSource(line.desired, line.maxWaitCycles)
			if src != line.src || line.clkDiv != clkDiv || line.waitCycles != waitCycles || line.desired != f || err != nil {
				t.Fatalf("calcSource(%s, %d) = %s / %d / %d = %s  expected %s / %d / %d = %s",
					line.desired, line.maxWaitCycles,
					src, clkDiv, waitCycles, f,
					line.src, line.clkDiv, line.waitCycles, line.desired)
			}
		})
	}
}

func TestCalcSource_oversample(t *testing.T) {
	t.Parallel()
	data := []struct {
		desired            physic.Frequency
		maxWaitCycles      uint32
		src                clockCtl
		clkDiv, waitCycles uint32
		expected           physic.Frequency
	}{
		{
			// 150x
			1 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 32,
			150 * physic.Hertz,
		},
		{
			// 75x
			2 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 32,
			150 * physic.Hertz,
		},
		{
			// 15x
			10 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 32,
			150 * physic.Hertz,
		},
		{
			// 2x
			100 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			clockSrc19dot2MHz,
			4000, 24,
			200 * physic.Hertz,
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()
			src, clkDiv, waitCycles, f, err := calcSource(line.desired, line.maxWaitCycles)
			if src != line.src || line.clkDiv != clkDiv || line.waitCycles != waitCycles || line.expected != f || err != nil {
				t.Fatalf("calcSource(%s, %d) = %s / %d / %d = %s  expected %s / %d / %d = %s",
					line.desired, line.maxWaitCycles,
					src, clkDiv, waitCycles, f,
					line.src, line.clkDiv, line.waitCycles, line.expected)
			}
		})
	}
}

func TestCalcSource_inexact(t *testing.T) {
	t.Parallel()
	// clockDiviMax is 4095, an odd number.
	data := []struct {
		desired       physic.Frequency
		maxWaitCycles uint32
		//src                clockCtl
		//clkDiv, waitCycles int
		//expected           physic.Frequency
	}{
		{
			// 2930/1465 = 2x
			clk19dot2MHz / clockDiviMax / (dmaWaitcyclesMax + 1) * physic.Hertz,
			dmaWaitcyclesMax + 1,
			//clockSrc19dot2MHz, 2121, 31, 292,
		},
		{
			// 93795/46886 = 2.0004 (error: 0.025%)
			clk19dot2MHz / clockDiviMax * physic.Hertz,
			dmaWaitcyclesMax + 1,
			//clockSrcPLLD, 2051, 26, 9376,
		},
		{
			// 1465.2014/7 = 209.31x (error: 0.15%)
			7 * physic.Hertz,
			dmaWaitcyclesMax + 1,
			//clockSrc19dot2MHz, 4095, 32, 146, // 146.52014
		},
	}
	for i, line := range data {
		line := line
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()
			src, clkDiv, waitCycles, f, err := calcSource(line.desired, line.maxWaitCycles)
			if src != 0 || clkDiv != 0 || waitCycles != 0 || f != 0 || err == nil {
				t.Fatalf("inexact match is not yet implemented: %s, %d, %d, %d", src, clkDiv, waitCycles, f)
			}
		})
	}
}
