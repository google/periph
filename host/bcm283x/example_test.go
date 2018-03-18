// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/host"
	"periph.io/x/periph/host/bcm283x"
)

func ExamplePinsRead0To31() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Print out the state of 32 GPIOs with a single read that reads all these
	// pins all at once.
	bits := bcm283x.PinsRead0To31()
	fmt.Printf("bits: %#x\n", bits)
	suffixes := []string{"   ", "\n"}
	for i := uint(0); i < 32; i++ {
		fmt.Printf("GPIO%-2d: %d%s", i, (bits>>i)&1, suffixes[(i%4)/3])
	}
	// Output:
	// bits: 0x80011010
	// GPIO0 : 0   GPIO1 : 0   GPIO2 : 0   GPIO3 : 0
	// GPIO4 : 1   GPIO5 : 0   GPIO6 : 0   GPIO7 : 0
	// GPIO8 : 0   GPIO9 : 0   GPIO10: 0   GPIO11: 0
	// GPIO12: 1   GPIO13: 0   GPIO14: 0   GPIO15: 0
	// GPIO16: 1   GPIO17: 0   GPIO18: 0   GPIO19: 0
	// GPIO20: 0   GPIO21: 0   GPIO22: 0   GPIO23: 0
	// GPIO24: 0   GPIO25: 0   GPIO26: 0   GPIO27: 0
	// GPIO28: 0   GPIO29: 0   GPIO30: 0   GPIO31: 1
}

func ExamplePinsRead32To46() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Print out the state of 15 GPIOs with a single read that reads all these
	// pins all at once.
	bits := bcm283x.PinsRead32To46()
	fmt.Printf("bits: %#x\n", bits)
	suffixes := []string{"   ", "\n"}
	for i := uint(0); i < (47 - 32); i++ {
		fmt.Printf("GPIO%d: %d%s", i+32, (bits>>i)&1, suffixes[(i%4)/3])
	}
	// Output:
	// bits: 0x4101
	// GPIO32: 1   GPIO33: 0   GPIO34: 0   GPIO35: 0
	// GPIO36: 0   GPIO37: 0   GPIO38: 0   GPIO39: 0
	// GPIO40: 1   GPIO41: 0   GPIO42: 0   GPIO43: 0
	// GPIO44: 0   GPIO45: 0   GPIO46: 1
}

func ExamplePinsClear0To31() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Simultaneously clears GPIO4 and GPIO16 to gpio.Low.
	bcm283x.PinsClear0To31(1<<16 | 1<<4)
}

func ExamplePinsSet0To31() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Simultaneously sets GPIO4 and GPIO16 to gpio.High.
	bcm283x.PinsClear0To31(1<<16 | 1<<4)
}
