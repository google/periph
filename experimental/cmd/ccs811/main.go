// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
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

	if len(os.Args) < 2 {
		printHelp()
		return
	}
	switch os.Args[1] {
	case "status":
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

	case "rawdata":
		i, u, err := d.ReadRawData()
		if err != nil {
			fmt.Println("Error getting raw data:", err)
		}
		fmt.Printf("Current raw data: %duA, %.1fmV\n", i, float32(u)*1.65/1023)

	case "baseline":
		baseline, err := d.GetBaseline()
		if err != nil {
			fmt.Println("Error getting baseline:", err)
		}
		fmt.Printf("Baseline: %X %X\n", baseline[0], baseline[1])

	case "sense":
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

	case "readcontinuously":
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

	case "fwinfo":
		fw, err := d.GetFirmwareData()
		if err != nil {
			fmt.Println("Error getting firmware versions:", err)
		}
		fmt.Printf("Versions: %+v\n", fw)

	case "appstart":
		err := d.StartSensorApp()
		if err != nil {
			fmt.Println("Error starting sensor app:", err)
		}

	case "measuremode":
		if len(os.Args) < 3 {
			mode, err := d.GetMeasurementModeRegister()
			if err != nil {
				fmt.Println("Can't get measurement mode", err)
				return
			}
			fmt.Println("Measurement mode:", mode.MeasurementMode)
			fmt.Println("Generate interrupt:", mode.GenerateInterrupt)
			fmt.Println("Use thresholds:", mode.UseThreshold)

		} else {
			fmt.Println("Setting measurement mode to", os.Args[2])
			i, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println("Can't convert measurement mode to number (0-4)")
			}

			mode := ccs811.MeasurementMode(i)
			d.SetMeasurementModeRegister(mode, false, false)
		}
	default:
		printHelp()
	}

}

func printHelp() {
	fmt.Println(`Allowed commands:
	status: displays status register of sensor
	rawdata: displays current and voltage of sensors measurement resistor
	baseline: displays value used for correction of measurement
	sense: perform one time measurement
	readcontinuously: perform continous measuremnt with interval of one second
	fwinfo: display different versions of hardware, boot and firmware
	appstart: start sensors application - move it from boot to application mode
	measuremode [0-4]: without parameter just shows current measurement mode, with parameter sets the mode`)
}

func printByteAsNibble(b byte) {
	for c := 0; c < 8; c++ {
		fmt.Printf("%d", (b >> 0 & 1))
		if c == 3 {
			fmt.Print(" ")
		}
	}
}
