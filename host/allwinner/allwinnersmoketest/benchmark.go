// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinnersmoketest

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host/allwinner"
)

// Benchmark is imported by periph-smoketest.
type Benchmark struct {
	p *allwinner.Pin
}

// Name implements the SmokeTest interface.
func (s *Benchmark) Name() string {
	return "allwinner-benchmark"
}

// Description implements the SmokeTest interface.
func (s *Benchmark) Description() string {
	return "Benchmarks allwinner functionality"
}

// Run implements the SmokeTest interface.
func (s *Benchmark) Run(args []string) error {
	if !allwinner.Present() {
		return errors.New("this smoke test can only be used on a allwinner based host")
	}
	f := flag.NewFlagSet(s.Name(), flag.ExitOnError)
	name := f.String("p", "", "Pin to use")
	f.Parse(args)
	if f.NArg() != 0 {
		return errors.New("unsupported flags")
	}

	if *name == "" {
		return errors.New("-p is required")
	}
	ok := false
	s.p, ok = gpioreg.ByName(*name).(*allwinner.Pin)
	if !ok {
		return fmt.Errorf("invalid pin %q", *name)
	}
	printBench("Out()", testing.Benchmark(s.benchmarkOut))
	return nil
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
	if r.Bytes != 0 {
		fmt.Fprintf(os.Stderr, "unexpected %d bytes written\n", r.Bytes)
		return
	}
	if r.MemAllocs != 0 || r.MemBytes != 0 {
		fmt.Fprintf(os.Stderr, "unexpected %d bytes allocated as %d calls\n", r.MemBytes, r.MemAllocs)
		return
	}
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
