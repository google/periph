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
