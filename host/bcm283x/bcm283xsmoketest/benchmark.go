// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283xsmoketest

import (
	"errors"
	"flag"
	"fmt"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host/bcm283x"
)

// Benchmark is imported by periph-smoketest.
type Benchmark struct {
	short bool
	p     *bcm283x.Pin
	pull  gpio.Pull
}

// Name implements the SmokeTest interface.
func (s *Benchmark) Name() string {
	return "bcm283x-benchmark"
}

// Description implements the SmokeTest interface.
func (s *Benchmark) Description() string {
	return "Benchmarks bcm283x functionality"
}

// Run implements the SmokeTest interface.
func (s *Benchmark) Run(f *flag.FlagSet, args []string) error {
	name := f.String("p", "", "Pin to use")
	f.BoolVar(&s.short, "short", false, "Skip many partially redundant benchmarks")
	if err := f.Parse(args); err != nil {
		return err
	}

	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unsupported flags")
	}
	if !bcm283x.Present() {
		f.Usage()
		return errors.New("this smoke test can only be run on a bcm283x based host")
	}
	if *name == "" {
		f.Usage()
		return errors.New("-p is required")
	}
	p := gpioreg.ByName(*name)
	if p == nil {
		return fmt.Errorf("invalid pin %q", *name)
	}
	if r, ok := p.(gpio.RealPin); ok {
		p = r.Real()
	}
	var ok bool
	if s.p, ok = p.(*bcm283x.Pin); !ok {
		return fmt.Errorf("pin is not bcm283x %q", *name)
	}
	s.pull = gpio.PullDown
	s.runFastGPIOBenchmark()
	return nil
}
