// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"periph.io/x/periph/conn/physic"
)

// errClockRegister is returned in a situation where the clock memory is not
// working as expected. It is mocked in tests.
var errClockRegister = errors.New("can't write to clock divisor CPU register")

// Clock sources frequency in hertz.
const (
	clk19dot2MHz = 19200 * physic.KiloHertz
	clk500MHz    = 500 * physic.MegaHertz
)

const (
	// 31:24 password
	clockPasswdCtl clockCtl = 0x5A << 24 // PASSWD
	// 23:11 reserved
	clockMashMask clockCtl = 3 << 9 // MASH
	clockMash0    clockCtl = 0 << 9 // src_freq / divI  (ignores divF)
	clockMash1    clockCtl = 1 << 9
	clockMash2    clockCtl = 2 << 9
	clockMash3    clockCtl = 3 << 9 // will cause higher spread
	clockFlip     clockCtl = 1 << 8 // FLIP
	clockBusy     clockCtl = 1 << 7 // BUSY
	// 6 reserved
	clockKill          clockCtl = 1 << 5   // KILL
	clockEnable        clockCtl = 1 << 4   // ENAB
	clockSrcMask       clockCtl = 0xF << 0 // SRC
	clockSrcGND        clockCtl = 0        // 0Hz
	clockSrc19dot2MHz  clockCtl = 1        // 19.2MHz
	clockSrcTestDebug0 clockCtl = 2        // 0Hz
	clockSrcTestDebug1 clockCtl = 3        // 0Hz
	clockSrcPLLA       clockCtl = 4        // 0Hz
	clockSrcPLLC       clockCtl = 5        // 1000MHz (changes with overclock settings)
	clockSrcPLLD       clockCtl = 6        // 500MHz
	clockSrcHDMI       clockCtl = 7        // 216MHz; may be disabled
	// 8-15 == GND.
)

// clockCtl controls the clock properties.
//
// It must not be changed while busy is set or a glitch may occur.
//
// Page 107
type clockCtl uint32

func (c clockCtl) String() string {
	var out []string
	if c&0xFF000000 == clockPasswdCtl {
		c &^= 0xFF000000
		out = append(out, "PWD")
	}
	switch c & clockMashMask {
	case clockMash1:
		out = append(out, "Mash1")
	case clockMash2:
		out = append(out, "Mash2")
	case clockMash3:
		out = append(out, "Mash3")
	default:
	}
	c &^= clockMashMask
	if c&clockFlip != 0 {
		out = append(out, "Flip")
		c &^= clockFlip
	}
	if c&clockBusy != 0 {
		out = append(out, "Busy")
		c &^= clockBusy
	}
	if c&clockKill != 0 {
		out = append(out, "Kill")
		c &^= clockKill
	}
	if c&clockEnable != 0 {
		out = append(out, "Enable")
		c &^= clockEnable
	}
	switch x := c & clockSrcMask; x {
	case clockSrcGND:
		out = append(out, "GND(0Hz)")
	case clockSrc19dot2MHz:
		out = append(out, "19.2MHz")
	case clockSrcTestDebug0:
		out = append(out, "Debug0(0Hz)")
	case clockSrcTestDebug1:
		out = append(out, "Debug1(0Hz)")
	case clockSrcPLLA:
		out = append(out, "PLLA(0Hz)")
	case clockSrcPLLC:
		out = append(out, "PLLD(1000MHz)")
	case clockSrcPLLD:
		out = append(out, "PLLD(500MHz)")
	case clockSrcHDMI:
		out = append(out, "HDMI(216MHz)")
	default:
		out = append(out, fmt.Sprintf("GND(%d)", x))
	}
	c &^= clockSrcMask
	if c != 0 {
		out = append(out, fmt.Sprintf("clockCtl(0x%0x)", uint32(c)))
	}
	return strings.Join(out, "|")
}

const (
	// 31:24 password
	clockPasswdDiv clockDiv = 0x5A << 24 // PASSWD
	// Integer part of the divisor
	clockDiviShift          = 12
	clockDiviMax            = (1 << 12) - 1
	clockDiviMask  clockDiv = clockDiviMax << clockDiviShift // DIVI
	// Fractional part of the divisor
	clockDivfMask clockDiv = (1 << 12) - 1 // DIVF
)

// clockDiv is a 12.12 fixed point value.
//
// The fractional part generates a significant amount of noise so it is
// preferable to not use it.
//
// Page 108
type clockDiv uint32

func (c clockDiv) String() string {
	i := (c & clockDiviMask) >> clockDiviShift
	c &^= clockDiviMask
	if c == 0 {
		return fmt.Sprintf("%d.0", i)
	}
	return fmt.Sprintf("%d.(%d/%d)", i, c, clockDiviMax)
}

// clock is a pair of clockCtl / clockDiv.
//
// It can be set to one of the sources: clockSrc19dot2MHz(19.2MHz) and
// clockSrcPLLD(500Mhz), then divided to a value to get the resulting clock.
// Per spec the resulting frequency should be under 25Mhz.
type clock struct {
	ctl clockCtl
	div clockDiv
}

// findDivisorExact finds the clock divisor and wait cycles to reduce src to
// desired hz.
//
// The clock divisor is capped to clockDiviMax.
//
// Returns clock divisor, wait cycles. Returns 0, 0 if no exact match is found.
// Favorizes high clock divisor value over high clock wait cycles. This means
// that the function is slower than it could be, but results in more stable
// clock.
func findDivisorExact(src, desired physic.Frequency, maxWaitCycles uint32) (uint32, uint32) {
	if src < desired || src%desired != 0 || src/physic.Frequency(maxWaitCycles*clockDiviMax) > desired {
		// Can't attain without oversampling (too low) or desired frequency is
		// higher than the source (too high) or is not a multiple.
		return 0, 0
	}
	factor := uint32(src / desired)
	// TODO(maruel): Only iterate over valid divisors to save a bit more
	// calculations. Since it's is only doing 32 loops, this is not a big deal.
	for wait := uint32(1); wait <= maxWaitCycles; wait++ {
		if rest := factor % wait; rest != 0 {
			continue
		}
		clk := factor / wait
		if clk == 0 {
			break
		}
		if clk <= clockDiviMax {
			return clk, wait
		}
	}
	return 0, 0
}

// findDivisorOversampled tries to find the lowest allowed oversampling to make
// desiredHz a multiple of srcHz.
//
// Allowed oversampling depends on the desiredHz. Cap oversampling because
// oversampling at 10x in the 1Mhz range becomes unreasonable in term of
// memory usage.
func findDivisorOversampled(src, desired physic.Frequency, maxWaitCycles uint32) (uint32, uint32, physic.Frequency) {
	//log.Printf("findDivisorOversampled(%s, %s, %d)", src, desired, maxWaitCycles)
	// There are 2 reasons:
	// - desired is so low it is not possible to lower src to this frequency
	// - not a multiple, there's a need for a prime number
	// TODO(maruel): Rewrite without a loop, this is not needed. Leverage primes
	// to reduce the number of iterations.
	for multiple := physic.Frequency(2); ; multiple++ {
		n := multiple * desired
		if n > 100*physic.KiloHertz && multiple > 10 {
			break
		}
		if clk, wait := findDivisorExact(src, n, maxWaitCycles); clk != 0 {
			return clk, wait, n
		}
	}
	return 0, 0, 0
}

// calcSource choose the best source to get the exact desired clock.
//
// It calculates the clock source, the clock divisor and the wait cycles, if
// applicable. Wait cycles is 'div minus 1'.
func calcSource(f physic.Frequency, maxWaitCycles uint32) (clockCtl, uint32, uint32, physic.Frequency, error) {
	if f < physic.Hertz {
		return 0, 0, 0, 0, fmt.Errorf("bcm283x-clock: desired frequency %s must be >1hz", f)
	}
	if f > 125*physic.MegaHertz {
		return 0, 0, 0, 0, fmt.Errorf("bcm283x-clock: desired frequency %s is too high", f)
	}
	// http://elinux.org/BCM2835_datasheet_errata states that clockSrc19dot2MHz
	// is the cleanest clock source so try it first.
	div, wait := findDivisorExact(clk19dot2MHz, f, maxWaitCycles)
	if div != 0 {
		return clockSrc19dot2MHz, div, wait, f, nil
	}
	// Try 500Mhz.
	div, wait = findDivisorExact(clk500MHz, f, maxWaitCycles)
	if div != 0 {
		return clockSrcPLLD, div, wait, f, nil
	}

	// Try with up to 10x oversampling. This is generally useful for lower
	// frequencies, below 10kHz. Prefer the one with less oversampling. Only for
	// non-aliased matches.
	div19, wait19, f19 := findDivisorOversampled(clk19dot2MHz, f, maxWaitCycles)
	div500, wait500, f500 := findDivisorOversampled(clk500MHz, f, maxWaitCycles)
	if div19 != 0 && (div500 == 0 || f19 < f500) {
		return clockSrc19dot2MHz, div19, wait19, f19, nil
	}
	if div500 != 0 {
		return clockSrcPLLD, div500, wait500, f500, nil
	}
	return 0, 0, 0, 0, errors.New("failed to find a good clock")
}

// set changes the clock frequency to the desired value or the closest one
// otherwise.
//
// f=0 means disabled.
//
// maxWaitCycles is the maximum oversampling via an additional wait cycles that
// can further divide the clock. Use 1 if no additional wait cycle is
// available. It is expected to be dmaWaitcyclesMax+1.
//
// Returns the actual clock used and divisor.
func (c *clock) set(f physic.Frequency, maxWaitCycles uint32) (physic.Frequency, uint32, error) {
	if f == 0 {
		c.ctl = clockPasswdCtl | clockKill
		for c.ctl&clockBusy != 0 {
		}
		return 0, 0, nil
	}
	ctl, div, div2, actual, err := calcSource(f, maxWaitCycles)
	if err != nil {
		return 0, 0, err
	}
	return actual, div2, c.setRaw(ctl, div)
}

// setRaw sets the clock speed with the clock source and the divisor.
func (c *clock) setRaw(ctl clockCtl, div uint32) error {
	if div < 1 || div > clockDiviMax {
		return errors.New("invalid clock divisor")
	}
	if ctl != clockSrc19dot2MHz && ctl != clockSrcPLLD {
		return errors.New("invalid clock control")
	}
	// Stop the clock.
	// TODO(maruel): Do not stop the clock if the current clock rate is the one
	// desired.
	for c.ctl&clockBusy != 0 {
		c.ctl = clockPasswdCtl | clockKill
	}
	d := clockDiv(div << clockDiviShift)
	c.div = clockPasswdDiv | d
	Nanospin(10 * time.Nanosecond)
	// Page 107
	c.ctl = clockPasswdCtl | ctl
	Nanospin(10 * time.Nanosecond)
	c.ctl = clockPasswdCtl | ctl | clockEnable
	if c.div != d {
		// This error is mocked out in tests, so the code path of set() callers can
		// follow on.
		return errClockRegister
	}
	return nil
}

func (c *clock) String() string {
	return fmt.Sprintf("%s / %s", c.ctl, c.div)
}

// clockMap is the memory mapped clock registers.
//
// The clock #1 must not be touched since it is being used by the ethernet
// controller.
//
// Page 107 for gp0~gp2.
// https://scribd.com/doc/127599939/BCM2835-Audio-clocks for PCM/PWM.
type clockMap struct {
	reserved0 [0x70 / 4]uint32          //
	gp0       clock                     // CM_GP0CTL+CM_GP0DIV; 0x70-0x74 (125MHz max)
	gp1ctl    uint32                    // CM_GP1CTL+CM_GP1DIV; 0x78-0x7A must not use (used by ethernet)
	gp1div    uint32                    // CM_GP1CTL+CM_GP1DIV; 0x78-0x7A must not use (used by ethernet)
	gp2       clock                     // CM_GP2CTL+CM_GP2DIV; 0x80-0x84 (125MHz max)
	reserved1 [(0x98 - 0x88) / 4]uint32 // 0x88-0x94
	pcm       clock                     // CM_PCMCTL+CM_PCMDIV 0x98-0x9C
	pwm       clock                     // CM_PWMCTL+CM_PWMDIV 0xA0-0xA4
}

func (c *clockMap) GoString() string {
	return fmt.Sprintf(
		"{\n  gp0: %s,\n  gp1: %s,\n  gp2: %s,\n  pcm: %sw,\n  pwm: %s,\n}",
		&c.gp0, &clock{clockCtl(c.gp1ctl), clockDiv(c.gp1div)}, &c.gp2, &c.pcm, &c.pwm)
}
