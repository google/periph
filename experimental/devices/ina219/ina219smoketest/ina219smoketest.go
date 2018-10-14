// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ina219smoketest

import (
	"errors"
	"flag"
	"fmt"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ina219"
	"periph.io/x/periph/host"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "ina219"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests INA219 over I²C"
}

func (s *SmokeTest) Run(f *flag.FlagSet, args []string) (err error) {
	i2cID := f.String("i2c", "", "I²C bus to use")
	i2cAddr := f.Int("ia", 0x40, "I²C bus address use: 0x40 to 0x4f")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}

	fmt.Println("Starting INA219 Current Sensor\nctrl+c to exit")
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

	// Create a new power sensor a sense resistor of 100 mΩ.
	config := ina219.Opts{
		Address:       *i2cAddr,
		SenseResistor: 100 * physic.MilliOhm,
		MaxCurrent:    3200 * physic.MilliAmpere,
	}

	sensor, err := ina219.New(bus, config)
	if err != nil {
		return err
	}
	pm, err := sensor.Sense()
	if err != nil {
		return err
	}
	fmt.Println(pm)

	return nil
}
