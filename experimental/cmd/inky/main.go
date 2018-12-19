package main

import (
	"flag"
	"image"
	_ "image/png"
	"log"
	"os"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/experimental/devices/inky"
	"periph.io/x/periph/host"
)

var img = flag.String("image", "", "Path to image")

func main() {
	flag.Parse()

	reader, err := os.Open(*img)
	if err != nil {
		log.Fatalf("Failed to open image %s: %v", *img, err)
	}
	defer reader.Close()

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatalf("Could not decode image: %v", err)
	}

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
	dev.SetBorder(inky.Black)

	dev.Draw(m.Bounds(), m, image.Point{0, 0}) 
}
