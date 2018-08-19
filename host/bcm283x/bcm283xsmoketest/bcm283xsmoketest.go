// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bcm283xsmoketest verifies that bcm283x specific functionality work.
//
// This test assumes GPIO6 and GPIO13 are connected together. GPIO6 implements
// CLK2 and GPIO13 imlements PWM1.
package bcm283xsmoketest

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
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
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) error {
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}
	if !bcm283x.Present() {
		f.Usage()
		return errors.New("this smoke test can only be run on a bcm283x based host")
	}

	start := time.Now()
	pClk := &loggingPin{bcm283x.GPIO6, start}
	pPWM := &loggingPin{bcm283x.GPIO13, start}
	// First make sure they are connected together.
	if err := ensureConnectivity(pClk, pPWM); err != nil {
		return err
	}
	// Confirmed they are connected. Now ready to test.
	if err := s.testPWMbyDMA(pClk, pPWM); err != nil {
		return err
	}
	if err := s.testPWM(pPWM, pClk); err != nil {
		return err
	}
	if err := s.testFunc(pClk); err != nil {
		return err
	}
	return s.testStreamIn(pPWM, pClk)
	// TODO(simokawa): test StreamOut.
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

// testPWMbyDMA tests .PWM() for a PWM pin driven by DMA.
func (s *SmokeTest) testPWMbyDMA(p1, p2 *loggingPin) error {
	fmt.Printf("- Testing DMA PWM\n")
	const freq = 5 * physic.KiloHertz
	if err := p2.In(gpio.PullDown, gpio.BothEdges); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	if err := p1.PWM(0, freq); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected Low", p1)
	}

	if err := p1.PWM(gpio.DutyMax, freq); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.High {
		return fmt.Errorf("unexpected %s value; expected High", p1)
	}

	// DMA PWM supports arbitrary duty cycle.
	if err := p1.PWM(gpio.DutyHalf/2, freq); err != nil {
		return err
	}

	if err := p1.PWM(gpio.DutyHalf, freq); err != nil {
		return err
	}

	return p1.Halt()
}

// testPWM tests .PWM() for a PWM pin.
func (s *SmokeTest) testPWM(p1, p2 *loggingPin) error {
	const freq = 5 * physic.KiloHertz
	fmt.Printf("- Testing PWM\n")
	if err := p2.In(gpio.PullDown, gpio.BothEdges); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)

	if err := p1.PWM(0, freq); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.Low {
		return fmt.Errorf("unexpected %s value; expected Low", p1)
	}

	if err := p1.PWM(gpio.DutyMax, freq); err != nil {
		return err
	}
	time.Sleep(time.Microsecond)
	if p2.Read() != gpio.High {
		return fmt.Errorf("unexpected %s value; expected High", p1)
	}

	// A real PWM supports arbitrary duty cycle.
	if err := p1.PWM(gpio.DutyHalf/2, freq); err != nil {
		return err
	}

	if err := p1.PWM(gpio.DutyHalf, freq); err != nil {
		return err
	}

	if err := p1.Halt(); err != nil {
		return err
	}
	return p1.Out(gpio.Low)
}

// testFunc tests .Func(), .SetFunc().
func (s *SmokeTest) testFunc(p *loggingPin) error {
	if string(p.Func()) != p.Function() {
		return fmt.Errorf("Func %q != Function %q", p.Func(), p.Function())
	}
	// This is dependent on testPWM() succeeding.
	if p.Func() != gpio.IN_LOW {
		return fmt.Errorf("Expected %q, got %q", gpio.IN_LOW, p.Func())
	}
	if f := p.SupportedFuncs(); !reflect.DeepEqual(f, []pin.Func{gpio.IN, gpio.OUT, gpio.CLK.Specialize(-1, 2)}) {
		return fmt.Errorf("Unexpected functions %q", f)
	}
	if err := p.SetFunc(gpio.CLK); err != nil {
		return fmt.Errorf("Failed to set %q", gpio.CLK)
	}
	if err := p.SetFunc(gpio.CLK.Specialize(-1, 2)); err != nil {
		return fmt.Errorf("Failed to set %q", gpio.CLK.Specialize(-1, 2))
	}
	return p.Halt()
}

// testStreamIn tests gpiostream.StreamIn and gpio.PWM.
func (s *SmokeTest) testStreamIn(p1, p2 *loggingPin) (err error) {
	const freq = 5 * physic.KiloHertz
	fmt.Printf("- Testing StreamIn\n")
	defer func() {
		if err2 := p2.Halt(); err == nil {
			err = err2
		}
	}()
	if err = p2.PWM(gpio.DutyHalf, freq); err != nil {
		return err
	}
	// Gather 0.1 second of readings at 10kHz sampling rate.
	// TODO(maruel): Support >64kb buffer.
	b := &gpiostream.BitStream{
		Bits: make([]byte, 1000),
		Freq: freq * 2,
		LSBF: true,
	}
	if err = p1.StreamIn(gpio.PullDown, b); err != nil {
		fmt.Printf("%x\n", b.Bits)
		return err
	}

	// Sum the bits, it should be close to 50%.
	v := 0
	for _, x := range b.Bits {
		for j := 0; j < 8; j++ {
			v += int((x >> uint(j)) & 1)
		}
	}
	fraction := (100 * v) / (8 * len(b.Bits))
	fmt.Println("fraction", fraction)
	if fraction < 45 || fraction > 55 {
		return fmt.Errorf("reading clock lead to %d%% bits On, expected 50%%", fraction)
	}

	// TODO(maruel): There should be 10 streaks.
	return nil
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

func (p *loggingPin) PWM(duty gpio.Duty, f physic.Frequency) error {
	fmt.Printf("  %s %s.PWM(%s, %s)\n", since(p.start), p, duty, f)
	return p.Pin.PWM(duty, f)
}

func (p *loggingPin) StreamIn(pull gpio.Pull, s gpiostream.Stream) error {
	fmt.Printf("  %s %s.StreamIn(%s, %s)\n", since(p.start), p, pull, s)
	return p.Pin.StreamIn(pull, s)
}

func (p *loggingPin) StreamOut(s gpiostream.Stream) error {
	fmt.Printf("  %s %s.StreamOut(%s)\n", since(p.start), p, s)
	return p.Pin.StreamOut(s)
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
