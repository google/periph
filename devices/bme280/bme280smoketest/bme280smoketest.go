// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bme280smoketest is leveraged by periph-smoketest to verify that two
// BME280, one over I²C, one over SPI, read roughly the same temperature,
// humidity and pressure.
package bme280smoketest

import (
	"flag"
	"fmt"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/conn/spi/spitest"
	"periph.io/x/periph/devices"
	"periph.io/x/periph/devices/bme280"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "bme280"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests BME280 over I²C and SPI"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(args []string) (err error) {
	f := flag.NewFlagSet("buses", flag.ExitOnError)
	i2cID := f.String("i2c", "", "I²C bus to use")
	spiID := f.String("spi", "", "SPI port to use")
	record := f.Bool("r", false, "record operation (for playback unit testing)")
	f.Parse(args)

	i2cBus, err2 := i2creg.Open(*i2cID)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := i2cBus.Close(); err == nil {
			err = err2
		}
	}()

	spiPort, err2 := spireg.Open(*spiID)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := spiPort.Close(); err == nil {
			err = err2
		}
	}()
	if !*record {
		return run(i2cBus, spiPort)
	}

	i2cRecorder := i2ctest.Record{Bus: i2cBus}
	spiRecorder := spitest.Record{Port: spiPort}
	err = run(&i2cRecorder, &spiRecorder)
	if len(i2cRecorder.Ops) != 0 {
		fmt.Printf("I²C recorder Addr: 0x%02X\n", i2cRecorder.Ops[0].Addr)
	} else {
		fmt.Print("I²C recorder\n")
	}
	for _, op := range i2cRecorder.Ops {
		fmt.Print("  W: ")
		for i, b := range op.W {
			if i != 0 {
				fmt.Print(", ")
			}
			fmt.Printf("0x%02X", b)
		}
		fmt.Print("\n   R: ")
		for i, b := range op.R {
			if i != 0 {
				fmt.Print(", ")
			}
			fmt.Printf("0x%02X", b)
		}
		fmt.Print("\n")
	}
	fmt.Print("\nSPI recorder\n")
	for _, op := range spiRecorder.Ops {
		fmt.Print("  W: ")
		if len(op.R) != 0 {
			// Read data.
			fmt.Printf("0x%02X\n   R: ", op.W[0])
			// first byte is dummy.
			for i, b := range op.R[1:] {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Printf("0x%02X", b)
			}
		} else {
			// Write-only command.
			for i, b := range op.W {
				if i != 0 {
					fmt.Print(", ")
				}
				fmt.Printf("0x%02X", b)
			}
			fmt.Print("\n   R: ")
		}
		fmt.Print("\n")
	}
	return err
}

func run(i2cBus i2c.Bus, spiPort spi.PortCloser) (err error) {
	opts := &bme280.Opts{
		Temperature: bme280.O16x,
		Pressure:    bme280.O16x,
		Humidity:    bme280.O16x,
		Filter:      bme280.NoFilter,
	}

	i2cDev, err2 := bme280.NewI2C(i2cBus, opts)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := i2cDev.Halt(); err == nil {
			err = err2
		}
	}()

	spiDev, err2 := bme280.NewSPI(spiPort, opts)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := spiDev.Halt(); err == nil {
			err = err2
		}
	}()

	// TODO(maruel): Generally the first measurement is way off.
	i2cEnv := devices.Environment{}
	spiEnv := devices.Environment{}
	if err2 := i2cDev.Sense(&i2cEnv); err2 != nil {
		return err2
	}
	if err2 = spiDev.Sense(&spiEnv); err2 != nil {
		return err2
	}

	// TODO(maruel): Determine acceptable threshold.
	if d := i2cEnv.Temperature - spiEnv.Temperature; d > 1000 || d < -1000 {
		return fmt.Errorf("Temperature delta higher than expected (%s): I²C got %s; SPI got %s", d, i2cEnv.Temperature, spiEnv.Temperature)
	}
	if d := i2cEnv.Pressure - spiEnv.Pressure; d > 100 || d < -100 {
		return fmt.Errorf("Pressure delta higher than expected (%s): I²C got %s; SPI got %s", d, i2cEnv.Pressure, spiEnv.Pressure)
	}
	if d := i2cEnv.Humidity - spiEnv.Humidity; d > 100 || d < -100 {
		return fmt.Errorf("Humidity delta higher than expected (%s): I²C got %s; SPI got %s", d, i2cEnv.Humidity, spiEnv.Humidity)
	}
	return nil
}
