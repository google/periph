// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bmx280smoketest is leveraged by periph-smoketest to verify that two
// BME280/BMP280, one over I²C, one over SPI, read roughly the same temperature,
// humidity and pressure.
package bmx280smoketest

import (
	"errors"
	"flag"
	"fmt"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/conn/spi/spitest"
	"periph.io/x/periph/devices/bmxx80"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "bmx280"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests BMx280 over I²C and SPI"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) (err error) {
	i2cID := f.String("i2c", "", "I²C bus to use")
	i2cAddr := f.Uint("ia", 0x76, "I²C bus address to use; either 0x76 (BMx280, the default) or 0x77 (BMP180)")
	spiID := f.String("spi", "", "SPI port to use")
	record := f.Bool("r", false, "record operation (for playback unit testing)")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}

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
		return run(i2cBus, uint16(*i2cAddr), spiPort)
	}

	i2cRecorder := i2ctest.Record{Bus: i2cBus}
	spiRecorder := spitest.Record{Port: spiPort}
	err = run(&i2cRecorder, uint16(*i2cAddr), &spiRecorder)
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

func run(i2cBus i2c.Bus, i2cAddr uint16, spiPort spi.PortCloser) (err error) {
	opts := &bmxx80.Opts{
		Temperature: bmxx80.O16x,
		Pressure:    bmxx80.O16x,
		Humidity:    bmxx80.O16x,
		Filter:      bmxx80.NoFilter,
	}

	i2cDev, err2 := bmxx80.NewI2C(i2cBus, i2cAddr, opts)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := i2cDev.Halt(); err == nil {
			err = err2
		}
	}()

	spiDev, err2 := bmxx80.NewSPI(spiPort, opts)
	if err2 != nil {
		return err2
	}
	defer func() {
		if err2 := spiDev.Halt(); err == nil {
			err = err2
		}
	}()

	i2cEnv := physic.Env{}
	spiEnv := physic.Env{}
	if err2 := i2cDev.Sense(&i2cEnv); err2 != nil {
		return err2
	}
	printEnv(i2cDev, &i2cEnv)
	if err2 = spiDev.Sense(&spiEnv); err2 != nil {
		return err2
	}
	printEnv(spiDev, &spiEnv)
	delta := physic.Env{
		Temperature: i2cEnv.Temperature - spiEnv.Temperature,
		Pressure:    i2cEnv.Pressure - spiEnv.Pressure,
		Humidity:    i2cEnv.Humidity - spiEnv.Humidity,
	}
	printEnv("Delta", &delta)

	// 1°C
	if delta.Temperature > 1000 || delta.Temperature < -1000 {
		return fmt.Errorf("Temperature delta higher than expected (%s): I²C got %s; SPI got %s", delta.Temperature, i2cEnv.Temperature, spiEnv.Temperature)
	}
	// 0.1kPa
	if delta.Pressure > 100 || delta.Pressure < -100 {
		return fmt.Errorf("Pressure delta higher than expected (%s): I²C got %s; SPI got %s", delta.Pressure, i2cEnv.Pressure, spiEnv.Pressure)
	}
	// 4%rH
	if delta.Humidity > 400 || delta.Humidity < -400 {
		return fmt.Errorf("Humidity delta higher than expected (%s): I²C got %s; SPI got %s", delta.Humidity, i2cEnv.Humidity, spiEnv.Humidity)
	}
	return nil
}

func printEnv(dev interface{}, e *physic.Env) {
	fmt.Printf("%-18s: %8s %10s %9s\n", dev, e.Temperature, e.Pressure, e.Humidity)
}
