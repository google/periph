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
	i2cAddr := f.Uint("ia", 0x40, "I²C bus address use: 0x40 to 0x4f")
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

	// open default i2c bus
	bus, err := i2creg.Open(*i2cID)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := bus.Close(); err == nil {
			err = err2
		}
	}()

	// create a new power sensor a sense resistor of 100 mΩ
	options := []ina219.Option{
		ina219.Address(uint8(*i2cAddr)),
		ina219.SenseResistor(100 * physic.MilliOhm),
		ina219.MaxCurrent(3200 * physic.MilliAmpere),
	}

	sensor, err := ina219.New(bus, options...)
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
