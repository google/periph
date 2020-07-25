// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package mcp9808smoketest implements a smoke test for the mcp9808.
package mcp9808smoketest

import (
	"errors"
	"flag"
	"fmt"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/mcp9808"
	"periph.io/x/periph/host"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "mcp9808"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests MCP9808 over I²C"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) (err error) {
	i2cID := f.String("i2c", "", "I²C bus to use")
	i2cAddr := f.Int("ia", 0x18, "I²C bus address use: 0x18 to 0x1f")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}

	fmt.Println("Starting MCP9808 Temperature Sensor\nctrl+c to exit")
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open default i2c bus.
	bus, err := i2creg.Open(*i2cID)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := bus.Close(); err == nil {
			err = err2
		}
	}()

	// Create a new temperature sensor a with maximum resolution.
	config := mcp9808.Opts{
		Addr: *i2cAddr,
		Res:  mcp9808.Maximum,
	}

	sensor, err := mcp9808.New(bus, &config)
	if err != nil {
		return err
	}
	t, err := sensor.SenseTemp()
	if err != nil {
		return err
	}
	fmt.Println(t)

	return nil
}
