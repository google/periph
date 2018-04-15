// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinnersmoketest

import (
	"errors"
	"flag"
	"fmt"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host/allwinner"
)

// Benchmark is imported by periph-smoketest.
type Benchmark struct {
	p    *allwinner.Pin
	pull gpio.Pull
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
func (s *Benchmark) Run(f *flag.FlagSet, args []string) error {
	name := f.String("p", "", "Pin to use")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unsupported flags")
	}
	if !allwinner.Present() {
		f.Usage()
		return errors.New("this smoke test can only be run on an allwinner based host")
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
	ok := false
	if s.p, ok = p.(*allwinner.Pin); !ok {
		return fmt.Errorf("pin is not allwinner %q", *name)
	}
	s.pull = gpio.PullDown
	s.runFastGPIOBenchmark()
	return nil
}
