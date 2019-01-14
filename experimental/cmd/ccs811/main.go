// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"periph.io/x/periph/conn/physic"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/ccs811"
	"periph.io/x/periph/host"
)

func justOneCommand(*flag.Flag) {

}

func main() {
	i2cID := flag.String("i2c", "", "I²C bus to use (default, uses the first I²C found)")
	i2cAddr := flag.Uint("ia", 0x5A, "I²C bus address to use; either 0x5A (90, default) or 0x5B (91)")
	status := flag.Bool("status", false, "command displays status register of sensor")
	rawData := flag.Bool("rawdata", false, "command displays current and voltage of sensors measurement resistor")
	baseline := flag.Bool("baseline", false, "command displays value used for correction of measurement")
	sense := flag.Bool("sense", false, "command performs one time measurement")
	readContinuously := flag.Bool("readcontinuously", false, "command performs continous measuremnt with interval of one second")
	fwInfo := flag.Bool("fwinfo", false, "command displays different versions of hardware, boot and firmware")
	appStart := flag.Bool("appstart", false, "command starts sensor's application - move it from boot to application mode")
	printMeasureMode := flag.Bool("printmeasuremode", false, "command shows current measurement mode")
	setMeasureMode := flag.Uint("setmeasuremode", 1, "command sets mode, valid values are 0-4, default is periodic 1s reading")
	flag.Parse()

	numberOfCommandFlags := 0
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "i2c" || f.Name == "ia" {
			return
		}
		numberOfCommandFlags++
	})

	if numberOfCommandFlags != 1 {
		fmt.Println("Please use exactly one command.")
		flag.PrintDefaults()
		return
	}

	if flag.NArg() != 0 {
		fmt.Println("Unexpected argument")
		flag.PrintDefaults()
		return
	}

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use i2creg I²C bus registry to find the first available I²C bus.
	b, err := i2creg.Open(*i2cID)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer b.Close()

	d, err := ccs811.New(b, &ccs811.Opts{Addr: uint16(*i2cAddr), MeasurementMode: ccs811.MeasurementModeConstant1000})
	if err != nil {
		log.Fatalf("Device creation failed: %v", err)
	}

	if *status {
		status, err := d.ReadStatus()
		if err != nil {
			fmt.Println("Error getting status:", err)
		}
		fmt.Print("Status: ")
		printByteAsNibble(status)
		fmt.Println()

		fmt.Print("Firmware mode: ")
		if status&0x80 == 0 {
			fmt.Println("Boot")
		} else {
			fmt.Println("Application")
		}
		fmt.Print("Application valid: ")
		if status&0x10 == 0 {
			fmt.Println("No")
		} else {
			fmt.Println("Yes")
		}
		fmt.Print("Data ready: ")
		if status&0x08 == 0 {
			fmt.Println("No")
		} else {
			fmt.Println("Yes")
		}
		fmt.Print("Error occured: ")
		if status&0x01 == 0 {
			fmt.Println("No")
		} else {
			fmt.Println("Yes")
		}
	}
	if *rawData {
		i, u, err := d.ReadRawData()
		if err != nil {
			fmt.Println("Error getting raw data:", err)
		}
		fmt.Printf("Current raw data: %duA, %.1fmV\n", i, float32(u)*1.65/1023)
	}
	if *baseline {
		baseline, err := d.GetBaseline()
		if err != nil {
			fmt.Println("Error getting baseline:", err)
		}
		fmt.Printf("Baseline: %X %X\n", baseline[0], baseline[1])
	}
	if *sense {
		values := &ccs811.SensorValues{}
		if err := d.Sense(values); err != nil {
			fmt.Println("Error getting data:", err)
			return
		}
		fmt.Printf("Sensor values: \neCO2: %d\nVOC: %d\n", values.ECO2, values.VOC)
		fmt.Print("Status: ")
		printByteAsNibble(values.Status)
		fmt.Println()
		fmt.Print("Error: ")
		if values.Status&1 == 1 {
			printByteAsNibble(byte(values.ErrorID))
		} else {
			fmt.Print("N/A")
		}
		fmt.Println()
		fmt.Printf("Current: %s\n", values.RawDataCurrent*physic.MicroAmpere)
		fmt.Printf("Voltage: %s\n", values.RawDataVoltage*physic.Volt)
	}
	if *readContinuously {
		values := &ccs811.SensorValues{}
		for {
			err := d.SensePartial(ccs811.ReadCO2VOCStatus, values)
			if err != nil {
				log.Println("Error during measurement, waiting for next value", err)
			} else {
				fmt.Println("eCO2:", values.ECO2, "VOC:", values.VOC)
			}
			time.Sleep(1200 * time.Millisecond)
		}
	}
	if *fwInfo {
		fw, err := d.GetFirmwareData()
		if err != nil {
			fmt.Println("Error getting firmware versions:", err)
		}
		fmt.Printf("Versions: %+v\n", fw)
	}
	if *appStart {
		err := d.StartSensorApp()
		if err != nil {
			fmt.Println("Error starting sensor app:", err)
		}
	}
	if *printMeasureMode {
		mode, err := d.GetMeasurementModeRegister()
		if err != nil {
			fmt.Println("Can't get measurement mode", err)
			return
		}
		fmt.Println("Measurement mode:", mode.MeasurementMode)
		fmt.Println("Generate interrupt:", mode.GenerateInterrupt)
		fmt.Println("Use thresholds:", mode.UseThreshold)

	}
	if *setMeasureMode > 4 {

		fmt.Println("Setting measurement mode to", *setMeasureMode)
		i, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println("Can't convert measurement mode to number (0-4)")
		}

		mode := ccs811.MeasurementMode(i)
		d.SetMeasurementModeRegister(mode, false, false)
	}

}

func printByteAsNibble(b byte) {
	for c := 0; c < 8; c++ {
		fmt.Printf("%d", (b >> 0 & 1))
		if c == 3 {
			fmt.Print(" ")
		}
	}
}
