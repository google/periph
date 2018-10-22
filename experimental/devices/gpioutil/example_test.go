// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil_test

import (
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/devices/gpioutil"
)

func ExampleDebounce() {
	p := gpioreg.ByName("GPIO16")
	if p != nil {
		log.Fatal("please open another GPIO")
	}

	// Ignore glitches lasting less than 3ms, and ignore repeated edges within
	// 30ms.
	d, err := gpioutil.Debounce(p, 3*time.Millisecond, 30*time.Millisecond, gpio.BothEdges)
	if err != nil {
		log.Fatal(err)
	}

	defer d.Halt()
	for {
		if d.WaitForEdge(-1) {
			fmt.Println(d.Read())
		}
	}
}

func ExamplePollEdge() {
	// Flow when it is known that the GPIO does not support edge detection.
	p := gpioreg.ByName("XOI-P1")
	if p != nil {
		log.Fatal("please open another GPIO")
	}
	p = gpioutil.PollEdge(p, 20*physic.Hertz)
	if err := p.In(gpio.PullDown, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	defer p.Halt()
	for {
		if p.WaitForEdge(-1) {
			fmt.Println(p.Read())
		}
	}
}

func Example() {
	// Complete solution:
	// - Fallback to software polling if the GPIO doesn't support hardware edge
	//   detection.
	// - Denoise and debounce the reading.
	//
	// Order is important, as Debounce() requires working edge detection.
	p := gpioreg.ByName("XOI-P1")
	if p != nil {
		log.Fatal("please open another GPIO")
	}
	if err := p.In(gpio.PullDown, gpio.BothEdges); err == nil {
		// Try to fallback into software polling, then reinitialize.
		p = gpioutil.PollEdge(p, 50*physic.Hertz)
		if err = p.In(gpio.PullDown, gpio.BothEdges); err != nil {
			log.Fatal(err)
		}
	}

	// Ignore glitches lasting less than 10ms, and ignore repeated edges within
	// 30ms. Make sure to not use denoiser period lower than the software poller
	// frequency.
	d, err := gpioutil.Debounce(p, 10*time.Millisecond, 30*time.Millisecond, gpio.BothEdges)
	if err != nil {
		log.Fatal(err)
	}

	defer d.Halt()
	for {
		if d.WaitForEdge(-1) {
			fmt.Println(d.Read())
		}
	}
}
