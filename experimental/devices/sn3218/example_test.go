package sn3218_test

import (
	"log"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/sn3218"
	"periph.io/x/periph/host"
	"time"
)

func Example() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	b, err := i2creg.Open("")

	defer b.Close()

	if err != nil {
		log.Fatal(err)
	}

	d, err := sn3218.New(b)

	// to ensure that all registers are reset and all LEDs are switched off
	defer d.Halt()

	if err != nil {
		log.Fatal(err)
	}

	// By default, the device is disabled and brightness is 0 for all LEDs
	// So let's set the brightness to a low value and enable the device to
	// get started
	d.SetGlobalBrightness(1)
	d.Enable()

	// Switch LED 7 on
	if err := d.SwitchLed(7, true); err != nil {
		log.Fatal("Error while switching LED", err)
	}

	time.Sleep(1000 * time.Millisecond)
}
