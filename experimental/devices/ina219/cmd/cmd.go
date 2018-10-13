package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/ina219"
	"periph.io/x/periph/host"
)

func main() {
	address := flag.Uint("address", 0x40, "I²C address")
	i2cbus := flag.String("bus", "", "I²C bus (/dev/i2c-1)")

	flag.Parse()

	fmt.Println("Starting INA219 Current Sensor\nctrl+c to exit")
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// open default I²C bus
	bus, err := i2creg.Open(*i2cbus)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	// create a new power sensor a sense resistor of 100 mΩ, 3.2A
	options := []ina219.Option{
		ina219.Address(uint8(*address)),
		ina219.SenseResistor(100 * physic.MilliOhm),
		ina219.MaxCurrent(3200 * physic.MilliAmpere),
	}

	sensor, err := ina219.New(bus, options...)
	if err != nil {
		log.Fatalln(err)
	}

	// read values from sensor every second
	everySecond := time.NewTicker(time.Second).C
	var halt = make(chan os.Signal)
	signal.Notify(halt, syscall.SIGTERM)
	signal.Notify(halt, syscall.SIGINT)

	for {
		select {
		case <-everySecond:
			p, err := sensor.Sense()
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(p)
		case <-halt:
			os.Exit(0)
		}
	}
}
