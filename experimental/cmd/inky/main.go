package main

import (
	"log"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/inky"
	"periph.io/x/periph/host"
)

func main() {
	host.Init()
	port, err := spireg.Open("")
	if err != nil {
		log.Fatalf("inky: %v", err)
	}

	dc := gpioreg.ByName("22")
	reset := gpioreg.ByName("27")
	busy := gpioreg.ByName("17")

	dev, err := inky.New(port, dc, reset, busy)
	if err != nil {
		log.Fatalf("inky: %v", err)
	}

	dev.Update(inky.Red)
}
