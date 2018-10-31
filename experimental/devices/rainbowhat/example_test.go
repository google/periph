// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package rainbowhat_test

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/apa102"
	"periph.io/x/periph/experimental/devices/rainbowhat"
	"periph.io/x/periph/host"
)

func Example() {

	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	hat, err := rainbowhat.NewRainbowHat(&apa102.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer hat.Halt()

	handleButton := func(btn gpio.PinIn, led gpio.PinOut) {
		ledState := false
		if err := led.Out(gpio.Low); err != nil {
			log.Fatal(err)
		}
		for {
			btn.WaitForEdge(-1)
			if btn.Read() == gpio.Low {
				if ledState {
					if err := led.Out(gpio.High); err != nil {
						log.Fatal(err)
					}
				} else {
					if err := led.Out(gpio.Low); err != nil {
						log.Fatal(err)
					}
				}
				ledState = !ledState
			}
		}
	}

	go handleButton(hat.GetButtonA(), hat.GetLedR())
	go handleButton(hat.GetButtonB(), hat.GetLedG())
	go handleButton(hat.GetButtonC(), hat.GetLedB())

	ledstrip := hat.GetLedStrip()
	ledstrip.Intensity = 50

	img := image.NewNRGBA(image.Rect(0, 0, ledstrip.Bounds().Dx(), 1))
	img.SetNRGBA(0, 0, color.NRGBA{148, 0, 211, 255})
	img.SetNRGBA(1, 0, color.NRGBA{75, 0, 130, 255})
	img.SetNRGBA(2, 0, color.NRGBA{0, 0, 255, 255})
	img.SetNRGBA(3, 0, color.NRGBA{0, 255, 0, 255})
	img.SetNRGBA(4, 0, color.NRGBA{255, 255, 0, 255})
	img.SetNRGBA(5, 0, color.NRGBA{255, 127, 0, 255})
	img.SetNRGBA(6, 0, color.NRGBA{255, 0, 0, 255})

	if err := ledstrip.Draw(ledstrip.Bounds(), img, image.Point{}); err != nil {
		log.Fatalf("failed to draw: %v", err)
	}

	display := hat.GetDisplay()
	sensor := hat.GetBmp280()
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for {
			var envi physic.Env
			if err := sensor.Sense(&envi); err != nil {
				log.Fatal(err)
			}

			temp := fmt.Sprintf("%5s", envi.Temperature)
			fmt.Printf("Pressure %8s \n", envi.Pressure)
			fmt.Printf("Temperature %8s \n", envi.Temperature)

			if _, err := display.WriteString(temp); err != nil {
				log.Fatal(err)
			}
			<-ticker.C
		}
	}()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigs // Wait for signal
		log.Println(sig)
		done <- true
	}()

	log.Println("Press ctrl+c to stop...")
	<-done // Wait
}
