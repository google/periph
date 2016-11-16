// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// gpiosmoketest verifies that basic GPIO pin functionality work.
package gpiosmoketest

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/host/sysfs"
)

const (
	// At 1.2Ghz, a small capacitance and/or a long wire may cause a few cycles
	// of propagation delay. pin.Read() may take a single cycle to execute.
	//
	// Sleep for a short delay to workaround this problem.
	//
	// TODO(maruel): Should loop a few cycles instead of doing a syscall and
	// report when this occurs but not fail.
	shortDelay = time.Microsecond

	// Purely to help diagnose issues.
	longDelay = 2 * time.Second
)

// loggingPin logs when its state changes.
type loggingPin struct {
	gpio.PinIO
}

func (p *loggingPin) In(pull gpio.Pull, edge gpio.Edge) error {
	fmt.Printf("    %s.In(%s, %s)\n", p, pull, edge)
	return p.PinIO.In(pull, edge)
}

func (p *loggingPin) Out(l gpio.Level) error {
	fmt.Printf("    %s.Out(%s)\n", p, l)
	return p.PinIO.Out(l)
}

func getPin(s string, useSysfs bool) (gpio.PinIO, error) {
	number, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	var p gpio.PinIO
	if useSysfs {
		ok := false
		if p, ok = sysfs.Pins[number]; !ok {
			return nil, fmt.Errorf("pin %s is not exported by sysfs", p)
		}
	} else {
		p = gpio.ByNumber(number)
	}
	if p == nil {
		return nil, errors.New("invalid pin number")
	}
	return p, nil
}

func slowSleep(do bool) {
	if do {
		fmt.Printf("  Sleep(%s)\n", longDelay)
		time.Sleep(longDelay)
	}
}

// preparePins sets p1 as input without pull and p2 as output low.
func preparePins(p1, p2 gpio.PinIO) error {
	if err := p1.In(gpio.Float, gpio.None); err != nil {
		return err
	}
	return p2.Out(gpio.Low)
}

// testBasic ensures basic operation works.
func testBasic(p1, p2 gpio.PinIO, slow bool) error {
	fmt.Printf("  Testing basic functionality\n")
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	time.Sleep(shortDelay)
	fmt.Printf("    -> %s: %s\n    -> %s: %s\n", p1, p1.Function(), p2, p2.Function())
	if l := p1.Read(); l != gpio.Low {
		return fmt.Errorf("%s: expected to read %s but got %s", p1, gpio.Low, l)
	}

	slowSleep(slow)
	if err := p2.Out(gpio.High); err != nil {
		return err
	}
	time.Sleep(shortDelay)
	fmt.Printf("    -> %s: %s\n    -> %s: %s\n", p1, p1.Function(), p2, p2.Function())
	if l := p1.Read(); l != gpio.High {
		return fmt.Errorf("%s: expected to read %s but got %s", p1, gpio.High, l)
	}
	return nil
}

func waitForEdge(p gpio.PinIO) <-chan bool {
	c := make(chan bool)
	// A timeout inherently makes this test flaky but there's a inherent
	// assumption that the CPU edge trigger wakes up this process within a
	// reasonable amount of time; in term of latency.
	go func() {
		b := p.WaitForEdge(100 * time.Millisecond)
		fmt.Printf("    -> WaitForEdge(%s) -> %t\n", p, b)
		c <- b
	}()
	return c
}

func togglePin(p gpio.PinIO, levels ...gpio.Level) error {
	for i, l := range levels {
		if err := p.Out(l); err != nil {
			return err
		}
		if i != len(levels)-1 {
			// In that case, the switch can be very fast (a mere few CPU cycles) so
			// sleep a bit more as we really want to test if the CPU detected these
			// or not.
			time.Sleep(shortDelay)
		}
	}
	return nil
}

// testEdgesBoth tests with gpio.Both.
//
// The following events are tested for:
// - Getting missing edges
// - No accumulation of edges (only trigger once)
// - No spurious edge
func testEdgesBoth(p1, p2 gpio.PinIO, slow bool) error {
	fmt.Printf("  Testing edges with %s\n", gpio.Both)
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	if err := p1.In(gpio.Float, gpio.Both); err != nil {
		return err
	}
	if c := waitForEdge(p1); <-c {
		return errors.New("there should be no edge right after setting a pin")
	}

	slowSleep(slow)
	c := waitForEdge(p1)
	if err := p2.Out(gpio.High); err != nil {
		return err
	}
	if !<-c {
		return errors.New("edge Low->High didn't trigger")
	}

	slowSleep(slow)
	c = waitForEdge(p1)
	if err := p2.Out(gpio.Low); err != nil {
		return err
	}
	if !<-c {
		return errors.New("edge High->Low didn't trigger")
	}

	// No edge
	slowSleep(slow)
	if <-waitForEdge(p1) {
		return errors.New("spurious edge 2")
	}

	// One accumulated edge.
	slowSleep(slow)
	if err := p2.Out(gpio.High); err != nil {
		return err
	}
	if !<-waitForEdge(p1) {
		return errors.New("edge Low->High didn't trigger")
	}

	// Two accumulated edge are merged.
	slowSleep(slow)
	if err := togglePin(p2, gpio.Low, gpio.High); err != nil {
		return err
	}
	if !<-waitForEdge(p1) {
		return errors.New("edge High->Low didn't trigger")
	}
	if <-waitForEdge(p1) {
		// BUG(maruel): Seen this to occur flakily. Need to investigate. :(
		//return errors.New("didn't expect for two edges to accumulate")
	}

	// Calling In() flush any accumulated event.
	slowSleep(slow)
	if err := p2.Out(gpio.Low); err != nil {
		return err
	}
	// At that point, there's an accumulated event.
	time.Sleep(shortDelay)
	// This flushes the event.
	if err := p1.In(gpio.Float, gpio.Both); err != nil {
		return err
	}
	if <-waitForEdge(p1) {
		return errors.New("accumulated event should have been flushed by In()")
	}

	return nil
}

// testEdgesSide tests with gpio.Rising or gpio.Falling.
//
// The following events are tested for:
// - Getting missing edges
// - No accumulation of edges (only trigger once)
// - No spurious edge
func testEdgesSide(p1, p2 gpio.PinIO, e gpio.Edge, slow bool) error {
	set := gpio.High
	idle := gpio.Low
	if e == gpio.Falling {
		set, idle = idle, set
	}
	fmt.Printf("  Testing edges with %s\n", e)
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	if err := p2.Out(idle); err != nil {
		return err
	}
	if err := p1.In(gpio.Float, e); err != nil {
		return err
	}
	if c := waitForEdge(p1); <-c {
		return errors.New("there should be no edge right after setting a pin")
	}

	slowSleep(slow)
	c := waitForEdge(p1)
	if err := p2.Out(set); err != nil {
		return err
	}
	if !<-c {
		return fmt.Errorf("edge %s->%s didn't trigger", idle, set)
	}

	// No edge
	slowSleep(slow)
	c = waitForEdge(p1)
	if err := p2.Out(idle); err != nil {
		return err
	}
	if <-c {
		return fmt.Errorf("edge %s->%s shouldn't trigger", set, idle)
	}
	if <-waitForEdge(p1) {
		return errors.New("spurious edge 2")
	}

	// One accumulated edge.
	slowSleep(slow)
	if err := p2.Out(set); err != nil {
		return err
	}
	if !<-waitForEdge(p1) {
		return fmt.Errorf("edge %s->%s didn't trigger", idle, set)
	}

	// Two accumulated edge are merged.
	slowSleep(slow)
	if err := togglePin(p2, idle, set, idle, set); err != nil {
		return err
	}
	if !<-waitForEdge(p1) {
		return fmt.Errorf("edge %s->%s didn't trigger", idle, set)
	}
	if <-waitForEdge(p1) {
		// BUG(maruel): Seen this to occur flakily. Need to investigate. :(
		//return errors.New("didn't expect for two edges to accumulate")
	}

	// Calling In() flush any accumulated event.
	slowSleep(slow)
	if err := togglePin(p2, idle, set); err != nil {
		return err
	}
	// At that point, there's an accumulated event.
	time.Sleep(shortDelay)
	// This flushes the event.
	if err := p1.In(gpio.Float, e); err != nil {
		return err
	}
	if <-waitForEdge(p1) {
		return errors.New("accumulated event should have been flushed by In()")
	}

	return nil
}

// testEdges ensures edge based triggering works.
func testEdges(p1, p2 gpio.PinIO, slow bool) error {
	// Test for:
	// - Falling, Rising, Both
	// - None
	if err := testEdgesBoth(p1, p2, slow); err != nil {
		return err
	}
	if err := testEdgesSide(p1, p2, gpio.Rising, slow); err != nil {
		return err
	}
	if err := testEdgesSide(p1, p2, gpio.Falling, slow); err != nil {
		return err
	}
	return nil
}

// testPull ensures input pull resistor works.
func testPull(p1, p2 gpio.PinIO, slow bool) error {
	fmt.Printf("  Testing input pull resistor\n")
	if err := preparePins(p1, p2); err != nil {
		return err
	}
	if err := p2.In(gpio.Down, gpio.None); err != nil {
		return err
	}
	time.Sleep(shortDelay)
	fmt.Printf("    -> %s: %s\n    -> %s: %s\n", p1, p1.Function(), p2, p2.Function())
	if p1.Read() != gpio.Low {
		return errors.New("read pull down failure")
	}

	slowSleep(slow)
	if err := p2.In(gpio.Up, gpio.None); err != nil {
		return err
	}
	time.Sleep(shortDelay)
	fmt.Printf("    -> %s: %s\n    -> %s: %s\n", p1, p1.Function(), p2, p2.Function())
	if p1.Read() != gpio.High {
		return errors.New("read pull up failure")
	}
	return nil
}

func testCycle(p1, p2 gpio.PinIO, noPull, slow bool) error {
	fmt.Printf("Testing %s -> %s\n", p2, p1)
	if err := testBasic(p1, p2, slow); err != nil {
		return err
	}
	if err := testEdges(p1, p2, slow); err != nil {
		return err
	}
	if !noPull {
		if err := testPull(p1, p2, slow); err != nil {
			return err
		}
	}
	return nil
}

type SmokeTest struct {
}

func (s *SmokeTest) Name() string {
	return "gpio"
}

func (s *SmokeTest) Description() string {
	return "Tests basic functionality, edge detection and input pull resistors"
}

func printPin(p gpio.PinIO) {
	fmt.Printf("- %s: %s", p, p.Function())
	if r, ok := p.(gpio.RealPin); ok {
		fmt.Printf("  alias for %s", r.Real())
	}
	fmt.Print("\n")
}

func (s *SmokeTest) Run(args []string) error {
	f := flag.NewFlagSet("gpio", flag.ExitOnError)
	slow := f.Bool("s", false, "slow; insert a second between each step")
	useSysfs := f.Bool("sysfs", false, "force the use of sysfs")
	f.Parse(args)
	if f.NArg() != 2 {
		// TODO(maruel): Find pins automatically:
		// - For each header in headers.All():
		//   - For each pin in header that are GPIO:
		//     - p.In(Down, Rising)
		//     - Start a goroutine to detect edge.
		//   - For each pin in header that are GPIO:
		//     - p.Out(High)
		//     - See if a pin triggered
		//     - p.In(Down, Rising)
		// This assumes that everything actually work in the first place.
		return errors.New("specify the two pins to use; they must be connected together")
	}

	// On certain Allwinner CPUs, it's a good idea to test specifically the PLx
	// pins, since they use a different register memory block (driver
	// "allwinner_pl") than groups PB to PH (driver "allwinner").
	p1, err := getPin(f.Args()[0], *useSysfs)
	if err != nil {
		return err
	}
	p2, err := getPin(f.Args()[1], *useSysfs)
	if err != nil {
		return err
	}

	// Disable pull testing when using sysfs because it is not supported.
	_, noPull := p1.(*sysfs.Pin)
	if !noPull {
		_, noPull = p2.(*sysfs.Pin)
	}
	if noPull {
		fmt.Printf("Skipping input pull resistor on sysfs\n")
	}

	fmt.Printf("Using pins and their current state:\n")
	printPin(p1)
	printPin(p2)
	pl1 := &loggingPin{p1}
	pl2 := &loggingPin{p2}
	err = testCycle(pl1, pl2, noPull, *slow)
	if err == nil {
		err = testCycle(pl2, pl1, noPull, *slow)
	}
	if err2 := pl1.In(gpio.PullNoChange, gpio.None); err2 != nil {
		fmt.Printf("(Exit) Failed to reset %s as input: %s\n", pl1, err2)
	}
	if err2 := pl2.In(gpio.PullNoChange, gpio.None); err2 != nil {
		fmt.Printf("(Exit) Failed to reset %s as input: %s\n", pl1, err2)
	}
	return err
}
