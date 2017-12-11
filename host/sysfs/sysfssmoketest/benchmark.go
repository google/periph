// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package sysfssmoketest verifies that sysfs specific functionality work.
package sysfssmoketest

import (
	"errors"
	"flag"
	"fmt"
	"sort"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/sysfs"
)

// Benchmark is imported by periph-smoketest.
type Benchmark struct {
	p *sysfs.Pin
}

// Name implements the SmokeTest interface.
func (s *Benchmark) Name() string {
	return "sysfs-benchmark"
}

// Description implements the SmokeTest interface.
func (s *Benchmark) Description() string {
	return "Benchmarks sysfs gpio functionality"
}

// Run implements the SmokeTest interface.
func (s *Benchmark) Run(args []string) error {
	f := flag.NewFlagSet(s.Name(), flag.ExitOnError)
	num := f.Int("p", -1, "Pin number to use")
	f.Parse(args)
	if f.NArg() != 0 {
		return errors.New("unsupported flags")
	}

	if *num == -1 {
		return errors.New("-p is required")
	}
	if s.p = sysfs.Pins[*num]; s.p == nil {
		list := make([]int, 0, len(sysfs.Pins))
		for i := range sysfs.Pins {
			list = append(list, i)
		}
		sort.Ints(list)
		valid := ""
		for i, v := range list {
			if i == 0 {
				valid += fmt.Sprintf("%d", v)
			} else {
				valid += fmt.Sprintf(", %d", v)
			}
		}
		return fmt.Errorf("invalid pin %d; valid: %s", *num, valid)
	}
	printBench("In()", testing.Benchmark(s.benchmarkIn))
	printBench("Out()", testing.Benchmark(s.benchmarkOut))
	return nil
}

func (s *Benchmark) benchmarkIn(b *testing.B) {
	p := s.p
	if err := p.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	l := gpio.Low
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l = p.Read()
	}
	b.StopTimer()
	b.Log(l)
}

func (s *Benchmark) benchmarkOut(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Out(gpio.High)
		p.Out(gpio.Low)
	}
}

//

func printBench(name string, r testing.BenchmarkResult) {
	fmt.Printf("%s \t%s\t%s\n", name, r, toHz(r.N, r.T))
}

func toHz(n int, t time.Duration) string {
	// Periph has a ban on float64 on the library but it's not too bad on unit
	// and smoke tests.
	hz := float64(n) * float64(time.Second) / float64(t)
	switch {
	case hz >= 1000000:
		return fmt.Sprintf("%.1fMhz", hz*0.000001)
	case hz >= 1000:
		return fmt.Sprintf("%.1fKhz", hz*0.001)
	default:
		return fmt.Sprintf("%.1fhz", hz)
	}
}
