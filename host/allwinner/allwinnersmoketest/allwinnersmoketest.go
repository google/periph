// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package allwinnersmoketest verifies that allwinner specific functionality
// work.
//
// This test assumes GPIO pins are connected together. The exact ones depends
// on the actual board. It is PB2 and PB3 for the C.H.I.P.
package allwinnersmoketest

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/allwinner"
	"periph.io/x/periph/host/chip"
	"periph.io/x/periph/host/pine64"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
	// start is to display the delta in µs.
	start time.Time
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "allwinner"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests advanced Allwinner functionality"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) error {
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}
	if !allwinner.Present() {
		f.Usage()
		return errors.New("this smoke test can only be run on an allwinner based host")
	}

	start := time.Now()
	var pwm *loggingPin
	var other *loggingPin
	if chip.Present() {
		pwm = &loggingPin{allwinner.PB2, start}
		other = &loggingPin{allwinner.PB3, start}
	} else if pine64.Present() {
		//pwm = &loggingPin{allwinner.PD22}
		return errors.New("implement and test for pine64")
	} else {
		return errors.New("implement and test for this host")
	}
	if err := ensureConnectivity(pwm, other); err != nil {
		return err
	}
	return nil
}

// Returns a channel that will return one bool, true if a edge was detected,
// false otherwise.
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
	*allwinner.Pin
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
	if err := p1.In(gpio.Float, gpio.NoEdge); err != nil {
		return err
	}
	if err := p2.In(gpio.Float, gpio.NoEdge); err != nil {
		return err
	}
	return nil
}
