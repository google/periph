// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bcm283xsmoketest verifies that bcm283x specific functionality work.
//
// This test assumes GPIO6 and GPIO13 are connected together. GPIO6 implements
// GPCLK2 and GPIO13 imlements PWM1_OUT.
package bcm283xsmoketest

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/bcm283x"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
	// start is to display the delta in µs.
	start time.Time
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "bcm283x"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests advanced bcm283x functionality"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(args []string) error {
	if !bcm283x.Present() {
		return errors.New("this smoke test can only be used on a bcm283x based host")
	}
	f := flag.NewFlagSet("bcm283x", flag.ExitOnError)
	f.Parse(args)
	if f.NArg() != 0 {
		return errors.New("unsupported flags")
	}

	start := time.Now()
	pClk := &loggingPin{bcm283x.GPIO6, start}
	pPWM := &loggingPin{bcm283x.GPIO13, start}
	// First make sure they are connected together.
	if err := ensureConnectivity(pClk, pPWM); err != nil {
		return err
	}
	return nil
}

// waitForEdge returns a channel that will return one bool, true if a edge was
// detected, false otherwise.
func (s *SmokeTest) waitForEdge(p gpio.PinIO) <-chan bool {
	c := make(chan bool)
	// A timeout inherently makes this test flaky but there's a inherent
	// assumption that the CPU edge trigger wakes up this process within a
	// reasonable amount of time; in term of latency.
	go func() {
		b := p.WaitForEdge(time.Second)
		// Author note: the test intentionally doesn't call p.Read() to test that
		// reading is not necessary.
		fmt.Printf("    %s -> WaitForEdge(%s) -> %t\n", since(s.start), p, b)
		c <- b
	}()
	return c
}

//

func printPin(p gpio.PinIO) {
	fmt.Printf("- %s: %s", p, p.Function())
	if r, ok := p.(gpio.RealPin); ok {
		fmt.Printf("  alias for %s", r.Real())
	}
	fmt.Print("\n")
}

// since returns time in µs since the test start.
func since(start time.Time) string {
	µs := (time.Since(start) + time.Microsecond/2) / time.Microsecond
	ms := µs / 1000
	µs %= 1000
	return fmt.Sprintf("%3d.%03dms", ms, µs)
}

// loggingPin logs when its state changes.
type loggingPin struct {
	*bcm283x.Pin
	start time.Time
}

func (p *loggingPin) In(pull gpio.Pull, edge gpio.Edge) error {
	fmt.Printf("  %s %s.In(%s, %s)\n", since(p.start), p, pull, edge)
	return p.Pin.In(pull, edge)
}

func (p *loggingPin) Out(l gpio.Level) error {
	fmt.Printf("  %s %s.Out(%s)\n", since(p.start), p, l)
	return p.Pin.Out(l)
}

// ensureConnectivity makes sure they are connected together.
func ensureConnectivity(p1, p2 *loggingPin) error {
	if err := p1.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return err
	}
	if err := p2.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p1.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected low", p1)
	}
	if p2.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected low", p2)
	}
	if err := p2.In(gpio.PullUp, gpio.NoEdge); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p1.Read() != gpio.High {
		return fmt.Errorf("unexpected %s value; expected high", p1)
	}
	return nil
}
