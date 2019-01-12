package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
		fmt.Printf("Baseline: %X %X", baseline[0], baseline[1])

	case "sense":
		values, err := d.Sense(ccs811.ReadAll)
		if err != nil {
			fmt.Println("Error getting data:", err)
		}
		fmt.Println("Sensor values: \neCO2: %d\nVOC: %d\n")
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
		fmt.Printf("Current: %duA\n", values.RawDataCurrent)
		fmt.Printf("Voltage %.1fmV\n", float32(values.RawDataVoltage)*1.65/1023)

	case "readcontinuously":
		for {
			values, err := d.Sense(ccs811.ReadCO2VOCStatus)
			if err != nil {
				log.Println("Error during measurement, waiting for next value", err)
			} else {
				fmt.Println("eCO:", values.ECO2, "VOC:", values.VOC)
			}
			time.Sleep(1200 * time.Millisecond)
		}

	case "fwinfo":
		fw, err := d.GetFirmwareData()
		if err != nil {
			fmt.Println("Error getting firmware versions:", err)
		}
		fmt.Printf("Versions: %+v\n")

	case "appstart":
		err := d.StartSensorApp()

	case "setmeasuremod":
		if len(os.Args) < 3 {
			fmt.Println("Invalid measurement mode")
			return
		}
		fmt.Println("Setting measurement mode to ", os.args[2])
		i, e := strconv.Atoi(os.args[2])
		if e != nil {
			fmt.Println("Can't convert measurement mode to number (0-4)")
		}
		d.SetMeasurementMode(i, false, false)
	default:
		fmt.Println("Allowed commands:")
	}

}

func printByteAsNibble(b byte) {
	for c := 0; c < 8; c++ {
		fmt.Printf("%d", (b >> 0 & 1))
		if c == 4 {
			fmt.Print(" ")
		}
	}
}
