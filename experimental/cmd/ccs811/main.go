package main

import (
	"fmt"
	"log"
	"time"

	"github.com/DeziderMesko/periph/experimental/devices/ccs811"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

func main() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer b.Close()

	d, err := ccs811.New(b, &ccs811.Opts{Addr: 0x5A, MeasurementMode: ccs811.MeasurementModeConstant1000})
	if err != nil {
		log.Fatalf("Device creation failed: %v", err)
	}

	for {
		values, err := d.Sense(ccs811.ReadCO2VOCStatus)
		if err != nil {
			log.Println("Error during measurement, waiting for next value", err)
		} else {
			fmt.Println("eCO:", values.ECO2, "VOC:", values.VOC)
		}

		time.Sleep(1200 * time.Millisecond)
	}

}
