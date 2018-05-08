// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner

import (
	"fmt"
	"strings"
	"time"
)

const pwmClock = 24000000
const pwmMaxPeriod = 0x10000

// prescalers is the value for pwm0Prescale*
var prescalers = []struct {
	freq   uint32
	scaler pwmPrescale
}{
	// Base frequency (min freq is half that) / PWM clock at pwmMaxPeriod
	{pwmClock, pwmPrescale1},             //  24MHz / 366Hz
	{pwmClock / 120, pwmPrescale120},     // 200kHz / 3Hz
	{pwmClock / 180, pwmPrescale180},     // 133kHz / 2Hz
	{pwmClock / 240, pwmPrescale240},     // 100kHz / 1.5Hz
	{pwmClock / 360, pwmPrescale360},     //  66kHz / 1.01Hz
	{pwmClock / 480, pwmPrescale480},     //  50kHz / 0.7Hz
	{pwmClock / 12000, pwmPrescale12000}, //   2kHz
	{pwmClock / 24000, pwmPrescale24000}, //   1kHz
	{pwmClock / 36000, pwmPrescale36000}, // 666 Hz
	{pwmClock / 48000, pwmPrescale48000}, // 500 Hz
	{pwmClock / 72000, pwmPrescale72000}, // 333 Hz / 0.005Hz
}

const (
	// 31:29 reserved
	pwmBusy pwmCtl = 1 << 28 // PWM0_RDY
	// 27:10 reserved (used for pwm1)
	pwm0Mask       pwmCtl = (1 << 10) - 1
	pwm0Bypass     pwmCtl = 1 << 9 // PWM0_BYPASS (marked as unused on some drivers?)
	pwm0PulseStart pwmCtl = 1 << 8 // PWM_CH0_PUL_START
	pwm0ModePulse  pwmCtl = 1 << 7 // PWM_CHANNEL0_MODE
	pwm0SCLK       pwmCtl = 1 << 6 // SCLK_CH0_GATING
	pwm0Polarity   pwmCtl = 1 << 5 // PWM_CH0_ACT_STA
	pwm0Enable     pwmCtl = 1 << 4 // PWM_CH0_EN
	// 3:0
	pwm0PrescaleMask pwmCtl = pwmCtl(pwmPrescaleMask) // PWM_CH0_PRESCAL
	pwm0Prescale120  pwmCtl = pwmCtl(pwmPrescale120)
	pwm0Prescale180  pwmCtl = pwmCtl(pwmPrescale180)
	pwm0Prescale240  pwmCtl = pwmCtl(pwmPrescale240)
	pwm0Prescale360  pwmCtl = pwmCtl(pwmPrescale360)
	pwm0Prescale480  pwmCtl = pwmCtl(pwmPrescale480)
	// 5, 6, 7 reserved
	pwm0Prescale12000 pwmCtl = pwmCtl(pwmPrescale12000)
	pwm0Prescale24000 pwmCtl = pwmCtl(pwmPrescale24000)
	pwm0Prescale36000 pwmCtl = pwmCtl(pwmPrescale36000)
	pwm0Prescale48000 pwmCtl = pwmCtl(pwmPrescale48000)
	pwm0Prescale72000 pwmCtl = pwmCtl(pwmPrescale72000)
	// 13, 14 reserved
	pwm0Prescale1 pwmCtl = pwmCtl(pwmPrescale1)
)

// A64: Pages 194-195.
// R8: Pages 83-84.
type pwmCtl uint32

func (p pwmCtl) String() string {
	var out []string
	if p&pwmBusy != 0 {
		out = append(out, "PWM0_RDY")
		p &^= pwmBusy
	}
	if p&pwm0Bypass != 0 {
		out = append(out, "PWM0_BYPASS")
		p &^= pwm0Bypass
	}
	if p&pwm0PulseStart != 0 {
		out = append(out, "PWM0_CH0_PUL_START")
		p &^= pwm0PulseStart
	}
	if p&pwm0ModePulse != 0 {
		out = append(out, "PWM0_CHANNEL0_MODE")
		p &^= pwm0ModePulse
	}
	if p&pwm0SCLK != 0 {
		out = append(out, "SCLK_CH0_GATING")
		p &^= pwm0SCLK
	}
	if p&pwm0Polarity != 0 {
		out = append(out, "PWM_CH0_ACT_STA")
		p &^= pwm0Polarity
	}
	if p&pwm0Enable != 0 {
		out = append(out, "PWM_CH0_EN")
		p &^= pwm0Enable
	}
	out = append(out, pwmPrescale(p&pwm0PrescaleMask).String())
	p &^= pwm0PrescaleMask
	if p != 0 {
		out = append(out, fmt.Sprintf("Unknown(0x%08X)", uint32(p)))
	}
	return strings.Join(out, "|")
}

const (
	pwmPrescaleMask pwmPrescale = 0xF
	pwmPrescale120  pwmPrescale = 0
	pwmPrescale180  pwmPrescale = 1
	pwmPrescale240  pwmPrescale = 2
	pwmPrescale360  pwmPrescale = 3
	pwmPrescale480  pwmPrescale = 4
	// 5, 6, 7 reserved
	pwmPrescale12000 pwmPrescale = 8
	pwmPrescale24000 pwmPrescale = 9
	pwmPrescale36000 pwmPrescale = 10
	pwmPrescale48000 pwmPrescale = 11
	pwmPrescale72000 pwmPrescale = 12
	// 13, 14 reserved
	pwmPrescale1 pwmPrescale = 15
)

type pwmPrescale uint32

func (p pwmPrescale) String() string {
	switch p {
	case pwmPrescale120:
		return "/120"
	case pwmPrescale180:
		return "/180"
	case pwmPrescale240:
		return "/240"
	case pwmPrescale360:
		return "/360"
	case pwmPrescale480:
		return "/480"
	case pwmPrescale12000:
		return "/12k"
	case pwmPrescale24000:
		return "/24k"
	case pwmPrescale36000:
		return "/36k"
	case pwmPrescale48000:
		return "/48k"
	case pwmPrescale72000:
		return "/72k"
	case pwmPrescale1:
		return "/1"
	default:
		return fmt.Sprintf("InvalidScalar(%d)", p&pwmPrescaleMask)
	}
}

// A64: Page 195.
// R8: Page 84
type pwmPeriod uint32

func (p pwmPeriod) String() string {
	return fmt.Sprintf("%d/%d", p&0xFFFF, uint32((p>>16)&0xFFFF)+1)
}

func toPeriod(total uint32, active uint16) pwmPeriod {
	if total > pwmMaxPeriod {
		total = pwmMaxPeriod
	}
	return pwmPeriod(total-1)<<16 | pwmPeriod(active)
}

// getBestPrescale finds the best prescaler.
//
// Cycles must be between 2 and 0x10000/2.
func getBestPrescale(period time.Duration) pwmPrescale {
	// TODO(maruel): Rewrite this function, it is incorrect.
	for _, v := range prescalers {
		p := time.Second / time.Duration(v.freq)
		smallest := (period / pwmMaxPeriod)
		largest := (period / 2)
		if p > smallest && p < largest {
			return v.scaler
		}
	}
	// Period is longer than 196s.
	return pwmPrescale72000
}

// pwmMap represents the PWM memory mapped CPU registers.
//
// The base frequency is 24Mhz.
//
// TODO(maruel): Some CPU have 2 PWMs.
type pwmMap struct {
	ctl    pwmCtl    // PWM_CTRL_REG
	period pwmPeriod // PWM_CH0_PERIOD
}

func (p *pwmMap) String() string {
	return fmt.Sprintf("pwmMap{%s, %v}", p.ctl, p.period)
}
