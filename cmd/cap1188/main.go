package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/experimental/devices/cap1188"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	i2cID := flag.String("i2c", "", "I²C bus to use")
	i2cADDR := flag.Uint("ia", 0x2c, "I²C bus address to use, defaults to Pimoroni's Drum Hat")
	hz := flag.Int("hz", 0, "I²C bus/SPI port speed")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	} else {
		cap1188.Debug = true
	}
	log.SetFlags(log.Lmicroseconds)

	opts := cap1188.DefaultOpts()
	if *i2cADDR != 0 {
		opts.Address = uint16(*i2cADDR)
	}

	if _, err := host.Init(); err != nil {
		return fmt.Errorf("couldn't init the host - %s", err)
	}

	var dev *cap1188.Dev
	i, err := i2creg.Open(*i2cID)
	if err != nil {
		return fmt.Errorf("couldn't open the i2c bus - %s", err)
	}
	defer i.Close()
	if p, ok := i.(i2c.Pins); ok {
		printPin("SCL", p.SCL())
		printPin("SDA", p.SDA())
	} else {
		fmt.Println("i.(i2c.Pins) failed")
	}

	if *hz != 0 {
		if err := i.SetSpeed(int64(*hz)); err != nil {
			return fmt.Errorf("couldn't set the i2c bus speed - %s", err)
		}
	}
	alertPin := gpioreg.ByName("GPIO25")
	if alertPin == nil {
		return errors.New("invalid alert GPIO pin number")
	}
	if err := alertPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		return err
	}
	if *verbose {
		fmt.Printf("cap1188: alert pin: %#v\n", alertPin)
	}

	resetPin := gpioreg.ByName("GPIO21")
	if resetPin == nil {
		return errors.New("invalid reset GPIO pin number")
	}
	opts.AlertPin = alertPin
	opts.ResetPin = resetPin

	if dev, err = cap1188.NewI2C(i, opts); err != nil {
		return fmt.Errorf("couldn't open cap1188 - %s", err)
	}
	time.Sleep(200 * time.Millisecond)

	userAskedToLinkLeds := opts.LinkedLEDs
	// unlinked LED demo
	if err := dev.UnlinkLeds(); err != nil {
		log.Println("Failed to unlink leds", err)
	}
	for i := 0; i < 8; i++ {
		dev.SetLed(i, true)
		time.Sleep(75 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
	dev.AllLedsOff()
	time.Sleep(100 * time.Millisecond)
	dev.AllLedsOn()
	time.Sleep(100 * time.Millisecond)
	dev.AllLedsOff()
	if userAskedToLinkLeds {
		if err := dev.LinkLeds(); err != nil {
			log.Println("Failed to relink leds", err)
		}
	}

	if alertPin != nil {
		fmt.Println("Monitoring for touch events")
		for {
			if alertPin.WaitForEdge(-1) {
				status, err := dev.InputStatus()
				if err != nil {
					fmt.Printf("Error reading inputs: %s\n", err)
				}
				printSensorsStatus(status)
				// we need to clear the interrupt so it can be triggered again
				if err := dev.ClearInterrupt(); err != nil {
					fmt.Println(err, "while clearing the interrupt")
				}
			}
		}
	}

	err2 := dev.Halt()
	if err != nil {
		return err
	}
	return err2
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "cap1188: %s.\n", err)
		os.Exit(1)
	}
}

func printPin(fn string, p pin.Pin) {
	name, pos := pinreg.Position(p)
	if name != "" {
		log.Printf("  %-4s: %-10s found on header %s, #%d\n", fn, p, name, pos)
	} else {
		log.Printf("  %-4s: %-10s\n", fn, p)
	}
}

func printSensorsStatus(statuses []cap1188.TouchStatus) {
	for i, st := range statuses {
		fmt.Printf("#%d: %s\t", i, st)
	}
	fmt.Println()
}
