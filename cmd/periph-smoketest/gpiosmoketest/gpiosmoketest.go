// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpiosmoketest is leveraged by periph-smoketest to verify that basic
// GPIO pin functionality work.
package gpiosmoketest

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host/allwinner"
	"periph.io/x/periph/host/bcm283x"
	"periph.io/x/periph/host/sysfs"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
	// start is to display the delta in µs.
	start time.Time

	// noEdge to skip edge testing.
	noEdge bool

	// noPull is set when input pull resistor are not testable.
	noPull bool

	// slow is inserted to slow down the test, purely to help diagnose issues.
	slow time.Duration

	// At 1.2Ghz, a small capacitance and/or a long wire may cause a few cycles
	// of propagation delay. pin.Read() may take a single cycle to execute.
	//
	// Sleep for a short delay to workaround this problem.
	//
	// 1µs is sufficient on a Raspberry Pi 3 (lower values would likely be fine)
	// but 20µs is necessary on a Pine64. There's quality right there.
	shortDelay time.Duration

	// Time to wait for an edge.
	edgeWait time.Duration
}

// Name implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Name() string {
	return "gpio"
}

// Description implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Description() string {
	return "Tests basic functionality, edge detection and input pull resistors"
}

// Run implements periph-smoketest.SmokeTest.
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) error {
	pin1 := f.String("pin1", "", "first pin to use")
	pin2 := f.String("pin2", "", "second pin to use")
	slow := f.Bool("s", false, "slow; insert a second between each step")
	useSysfs := f.Bool("sysfs", false, "force the use of sysfs")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}
	if *pin1 == "" || *pin2 == "" {
		f.Usage()
		return errors.New("-pin1 and -pin2 are required and they must be connected together")
	}

	s.edgeWait = 10 * time.Millisecond
	if *slow {
		s.slow = 2 * time.Second
	}

	if bcm283x.Present() {
		// 1µs is sufficient on a Raspberry Pi 3 (lower values would likely be fine)
		s.shortDelay = time.Microsecond
	} else {
		// 20µs is necessary on a Pine64. There's quality right there.
		s.shortDelay = 20 * time.Microsecond
	}
	if allwinner.IsA64() {
		// For now, skip edge testing on the Allwinner A64 (pine64).
		// https://periph.io/x/periph/issues/54
		s.noEdge = true
	}
	// On certain Allwinner CPUs, it's a good idea to test specifically the PLx
	// pins, since they use a different register memory block (driver
	// "allwinner_pl") than groups PB to PH (driver "allwinner").
	p1, err := getPin(*pin1, *useSysfs)
	if err != nil {
		return err
	}
	p2, err := getPin(*pin2, *useSysfs)
	if err != nil {
		return err
	}

	// Disable pull testing when using sysfs because it is not supported.
	if s.noPull = isSysfsPin(p1) || isSysfsPin(p2); s.noPull {
		fmt.Printf("Skipping input pull resistor on sysfs\n")
	}

	fmt.Printf("Using pins and their current state:\n")
	printPin(p1)
	printPin(p2)
	s.start = time.Now()
	pl1 := &loggingPin{p1, s.start}
	pl2 := &loggingPin{p2, s.start}
	if err = s.testCycle(pl1, pl2); err == nil {
		err = s.testCycle(pl2, pl1)
	}
	fmt.Printf("<terminating>\n")
	if err2 := pl1.In(gpio.PullNoChange, gpio.NoEdge); err2 != nil {
		fmt.Printf("(Exit) Failed to reset %s as input: %s\n", pl1, err2)
	}
	if err2 := pl2.In(gpio.PullNoChange, gpio.NoEdge); err2 != nil {
		fmt.Printf("(Exit) Failed to reset %s as input: %s\n", pl1, err2)
	}
	return err
}

func isSysfsPin(p gpio.PinIO) bool {
	if r, ok := p.(gpio.RealPin); ok {
		p = r.Real()
	}
	_, ok := p.(*sysfs.Pin)
	return ok
}

func (s *SmokeTest) slowSleep() {
	if s.slow != 0 {
		fmt.Printf("  Sleep(%s)\n", s.slow)
		time.Sleep(s.slow)
	}
}

// Returns a channel that will return one bool, true if a edge was detected,
// false otherwise.
func (s *SmokeTest) waitForEdge(p gpio.PinIO) <-chan bool {
	c := make(chan bool)
	// A timeout inherently makes this test flaky but there's a inherent
	// assumption that the CPU edge trigger wakes up this process within a
	// reasonable amount of time; in term of latency.
	go func() {
		b := p.WaitForEdge(s.edgeWait)
		// Author note: the test intentionally doesn't call p.Read() to test that
		// reading is not necessary.
		fmt.Printf("    %s -> WaitForEdge(%s) -> %t\n", since(s.start), p, b)
		c <- b
	}()
	return c
}

// testBasic ensures basic operation works.
func (s *SmokeTest) testBasic(p1, p2 gpio.PinIO) error {
	fmt.Printf("  Testing basic functionality\n")
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	fmt.Printf("    %s -> %s: %s\n", since(s.start), p1, p1.Function())
	fmt.Printf("    %s -> %s: %s\n", since(s.start), p2, p2.Function())
	if l := p1.Read(); l != gpio.Low {
		return fmt.Errorf("%s: expected to read %s but got %s", p1, gpio.Low, l)
	}

	s.slowSleep()
	if err := p2.Out(gpio.High); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	fmt.Printf("    %s -> %s: %s\n", since(s.start), p1, p1.Function())
	fmt.Printf("    %s -> %s: %s\n", since(s.start), p2, p2.Function())
	if l := p1.Read(); l != gpio.High {
		return fmt.Errorf("%s: expected to read %s but got %s", p1, gpio.High, l)
	}
	return nil
}

func (s *SmokeTest) togglePin(p gpio.PinIO, levels ...gpio.Level) error {
	for i, l := range levels {
		if err := p.Out(l); err != nil {
			return err
		}
		if i != len(levels)-1 {
			// In that case, the switch can be very fast (a mere few CPU cycles) so
			// sleep a bit more as we really want to test if the CPU detected these
			// or not.
			time.Sleep(s.shortDelay)
		}
	}
	return nil
}

// testEdgesBoth tests with gpio.BothEdges.
//
// The following events are tested for:
// - Getting missing edges
// - No accumulation of edges (only trigger once)
// - No spurious edge
func (s *SmokeTest) testEdgesBoth(p1, p2 gpio.PinIO) error {
	fmt.Printf("  Testing edges with %s\n", gpio.BothEdges)
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	if err := p1.In(gpio.Float, gpio.BothEdges); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	if c := s.waitForEdge(p1); <-c {
		// The linux kernel makes it hard to enforce this. :(
		fmt.Printf("    warning: there should be no edge right after setting a pin\n")
	}

	s.slowSleep()
	c := s.waitForEdge(p1)
	if err := p2.Out(gpio.High); err != nil {
		return err
	}
	if !<-c {
		return errors.New("edge Low->High didn't trigger")
	}

	s.slowSleep()
	c = s.waitForEdge(p1)
	if err := p2.Out(gpio.Low); err != nil {
		return err
	}
	if !<-c {
		return errors.New("edge High->Low didn't trigger")
	}

	// No edge
	s.slowSleep()
	if <-s.waitForEdge(p1) {
		return errors.New("spurious edge 2")
	}

	// One accumulated edge.
	s.slowSleep()
	if err := p2.Out(gpio.High); err != nil {
		return err
	}
	if !<-s.waitForEdge(p1) {
		return errors.New("edge Low->High didn't trigger")
	}

	// Two accumulated edge are merged.
	s.slowSleep()
	if err := s.togglePin(p2, gpio.Low, gpio.High); err != nil {
		return err
	}
	if !<-s.waitForEdge(p1) {
		return errors.New("edge High->Low didn't trigger")
	}
	if <-s.waitForEdge(p1) {
		// BUG(maruel): Seen this to occur flakily. Need to investigate. :(
		//return errors.New("didn't expect for two edges to accumulate")
	}

	// Calling In() flush any accumulated event.
	s.slowSleep()
	if err := p2.Out(gpio.Low); err != nil {
		return err
	}
	// At that point, there's an accumulated event.
	time.Sleep(s.shortDelay)
	// This flushes the event.
	if err := p1.In(gpio.Float, gpio.BothEdges); err != nil {
		return err
	}
	if <-s.waitForEdge(p1) {
		// The linux kernel makes it hard to enforce this. :(
		fmt.Printf("    warning: accumulated event should have been flushed by In()\n")
	}

	return nil
}

// testEdgesSide tests with gpio.RisingEdge or gpio.FallingEdge.
//
// The following events are tested for:
// - Getting missing edges
// - No accumulation of edges (only trigger once)
// - No spurious edge
func (s *SmokeTest) testEdgesSide(p1, p2 gpio.PinIO, e gpio.Edge) error {
	set := gpio.High
	idle := gpio.Low
	if e == gpio.FallingEdge {
		set, idle = idle, set
	}
	fmt.Printf("  Testing edges with %s\n", e)
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	if err := p2.Out(idle); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	if err := p1.In(gpio.Float, e); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	if c := s.waitForEdge(p1); <-c {
		// The linux kernel makes it hard to enforce this. :(
		fmt.Printf("    warning: there should be no edge right after setting a pin\n")
	}

	s.slowSleep()
	c := s.waitForEdge(p1)
	if err := p2.Out(set); err != nil {
		return err
	}
	if !<-c {
		return fmt.Errorf("edge %s->%s didn't trigger", idle, set)
	}

	// No edge
	s.slowSleep()
	c = s.waitForEdge(p1)
	if err := p2.Out(idle); err != nil {
		return err
	}
	if <-c {
		return fmt.Errorf("edge %s->%s shouldn't trigger", set, idle)
	}
	if <-s.waitForEdge(p1) {
		return errors.New("spurious edge 2")
	}

	// One accumulated edge.
	s.slowSleep()
	if err := p2.Out(set); err != nil {
		return err
	}
	if !<-s.waitForEdge(p1) {
		return fmt.Errorf("edge %s->%s didn't trigger", idle, set)
	}

	// Two accumulated edge are merged.
	s.slowSleep()
	if err := s.togglePin(p2, idle, set, idle, set); err != nil {
		return err
	}
	if !<-s.waitForEdge(p1) {
		return fmt.Errorf("edge %s->%s didn't trigger", idle, set)
	}
	if <-s.waitForEdge(p1) {
		// The linux kernel makes it hard to enforce this. :(
		fmt.Printf("    warning: didn't expect for two edges to accumulate\n")
	}

	// Calling In() flush any accumulated event.
	s.slowSleep()
	if err := s.togglePin(p2, idle, set); err != nil {
		return err
	}
	// At that point, there's an accumulated event.
	time.Sleep(s.shortDelay)
	// This flushes the event.
	if err := p1.In(gpio.Float, e); err != nil {
		return err
	}
	if <-s.waitForEdge(p1) {
		// The linux kernel makes it hard to enforce this. :(
		fmt.Printf("    warning: accumulated event should have been flushed by In()\n")
	}

	return nil
}

// testEdges ensures edge based triggering works.
func (s *SmokeTest) testEdges(p1, p2 gpio.PinIO) error {
	// Test for:
	// - FallingEdge, RisingEdge, BothEdges
	// - NoEdge
	if err := s.testEdgesBoth(p1, p2); err != nil {
		return err
	}
	if err := s.testEdgesSide(p1, p2, gpio.RisingEdge); err != nil {
		return err
	}
	return s.testEdgesSide(p1, p2, gpio.FallingEdge)
}

// testPull ensures input pull resistor works.
func (s *SmokeTest) testPull(p1, p2 gpio.PinIO) error {
	fmt.Printf("  Testing input pull resistor\n")
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	if err := p2.In(gpio.PullDown, gpio.NoEdge); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	fmt.Printf("    -> %s: %s\n    -> %s: %s\n", p1, p1.Function(), p2, p2.Function())
	if p1.Read() != gpio.Low {
		return errors.New("read pull down failure")
	}

	s.slowSleep()
	if err := p2.In(gpio.PullUp, gpio.NoEdge); err != nil {
		return err
	}
	time.Sleep(s.shortDelay)
	fmt.Printf("    -> %s: %s\n    -> %s: %s\n", p1, p1.Function(), p2, p2.Function())
	if p1.Read() != gpio.High {
		return errors.New("read pull up failure")
	}
	return nil
}

func (s *SmokeTest) testCycle(p1, p2 gpio.PinIO) error {
	fmt.Printf("Testing %s -> %s\n", p2, p1)
	if err := s.testBasic(p1, p2); err != nil {
		return err
	}
	if !s.noEdge {
		if err := s.testEdges(p1, p2); err != nil {
			return err
		}
	}
	if !s.noPull {
		if err := s.testPull(p1, p2); err != nil {
			return err
		}
	}
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

func getPin(s string, useSysfs bool) (gpio.PinIO, error) {
	if useSysfs {
		number, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		p, ok := sysfs.Pins[number]
		if !ok {
			return nil, fmt.Errorf("pin %s is not exported by sysfs", p)
		}
		return p, nil
	}
	p := gpioreg.ByName(s)
	if p == nil {
		return nil, errors.New("invalid pin number")
	}
	return p, nil
}

// preparePins sets p1 as input without pull and p2 as output low.
func preparePins(p1, p2 gpio.PinIO) error {
	if err := p1.In(gpio.Float, gpio.NoEdge); err != nil {
		return err
	}
	return p2.Out(gpio.Low)
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
	gpio.PinIO
	start time.Time
}

func (p *loggingPin) In(pull gpio.Pull, edge gpio.Edge) error {
	fmt.Printf("    %s %s.In(%s, %s)\n", since(p.start), p, pull, edge)
	return p.PinIO.In(pull, edge)
}

func (p *loggingPin) Out(l gpio.Level) error {
	fmt.Printf("    %s %s.Out(%s)\n", since(p.start), p, l)
	return p.PinIO.Out(l)
}
