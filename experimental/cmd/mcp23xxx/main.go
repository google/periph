package main

import (
	"flag"
	"fmt"
	"os"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/mcp23xxx"
	"periph.io/x/periph/host"
)

func mainImpl() error {
	if _, err := host.Init(); err != nil {
		return err
	}
	address := flag.Int("address", 0x20, "I²C address")
	i2cbus := flag.String("bus", "", "I²C bus (/dev/i2c-1)")

	flag.Parse()

	fmt.Println("Starting MCP23xxx IO extender")
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open default I²C bus.
	bus, err := i2creg.Open(*i2cbus)
	if err != nil {
		return fmt.Errorf("failed to open I²C: %v", err)
	}
	defer bus.Close()

	device, err := mcp23xxx.NewI2C(bus, mcp23xxx.MCP23x17, uint16(*address))
	if err != nil {
		return fmt.Errorf("failed to open new device: %v", err)
	}

	for _, port := range device.Pins {
		for _, pin := range port {
			fmt.Printf("%s\t%s\t%s\n", pin.Name(), pin.Function(), pin.Read().String())
		}
	}

	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp23xxx: %s.\n", err)
		return
	}
}
