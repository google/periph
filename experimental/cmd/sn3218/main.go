// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"log"
	"time"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/experimental/devices/sn3218"
	"periph.io/x/periph/host"
)

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	b, err := i2creg.Open("")

	defer b.Close()

	if err != nil {
		log.Fatal(err)
	}

	l, err := sn3218.New(b)

	if err != nil {
		log.Fatal(err)
	}

	l.Enable()

	log.Println("Switch each LED on one by one")
	l.SetGlobalBrightness(1)
	for i := 0; i < 18; i++ {
		err := l.SwitchLed(i, true)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Set brightness of each LED to 10 one by one")
	for i := 0; i < 18; i++ {
		l.SetBrightness(i, 10)
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Change between even and odd LEDs a couple of times")
	l.SetGlobalBrightness(1)
	time.Sleep(100 * time.Millisecond)
	l.SetGlobalBrightness(1)
	for x := 0; x < 10; x++ {
		for i := 0; i < 18; i++ {
			l.SwitchLed(i, i%2 == 0)
		}
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < 18; i++ {
			l.SwitchLed(i, i%2 == 1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("And now all on and off together")
	state := true
	for i := 0; i < 10; i++ {
		state = !state
		l.SwitchAllLeds(state)
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("Cleanup: Reset register and switch off")
	l.Halt()
}
