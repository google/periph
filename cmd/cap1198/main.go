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
	"periph.io/x/periph/experimental/devices/cap1198"
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
	}
	log.SetFlags(log.Lmicroseconds)

	opts := cap1198.DefaultOpts()
	if *i2cADDR != 0 {
		opts.Address = uint16(*i2cADDR)
	}

	if _, err := host.Init(); err != nil {
		return fmt.Errorf("couldn't init the host - %s", err)
	}

	var dev *cap1198.Dev
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

	resetPin := gpioreg.ByName("GPIO21")
	if resetPin == nil {
		return errors.New("invalid reset GPIO pin number")
	}
	opts.AlertPin = alertPin
	opts.ResetPin = resetPin

	if dev, err = cap1198.NewI2C(i, opts); err != nil {
		return fmt.Errorf("couldn't open cap1198 - %s", err)
	}

	time.Sleep(200 * time.Millisecond)
	fmt.Println(dev.InputStatus())

	if alertPin != nil {
		for {
			alertPin.WaitForEdge(-1)
			fmt.Println(dev.InputStatus())
		}
	}
	// TODO: wait for an interrupt

	err2 := dev.Halt()
	if err != nil {
		return err
	}
	return err2
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "cap1198: %s.\n", err)
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
