// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pin_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/pin"
)

func ExampleBasicPin() {
	// Declare a basic pin, that is not a GPIO, for registration on an header.
	b := &pin.BasicPin{N: "Exotic"}
	fmt.Println(b)

	// Output:
	// Exotic
}

func ExampleFunc_Specialize() {
	// Specializes both bus and line.
	fmt.Println(pin.Func("SPI_CS").Specialize(1, 2))
	// Specializes only bus.
	fmt.Println(pin.Func("SPI_MOSI").Specialize(1, -1))
	// Specializes only line.
	fmt.Println(pin.Func("CSI_D").Specialize(-1, 3))
	// Specializes neither.
	fmt.Println(pin.Func("INVALID").Specialize(-1, -1))
	// Output:
	// SPI1_CS2
	// SPI1_MOSI
	// CSI_D3
	// INVALID
}

func ExampleFunc_Generalize() {
	fmt.Println(pin.Func("SPI1_CS2").Generalize())
	fmt.Println(pin.Func("SPI1_MOSI").Generalize())
	fmt.Println(pin.Func("CSI_D3").Generalize())
	fmt.Println(pin.Func("INVALID").Generalize())
	// Output:
	// SPI_CS
	// SPI_MOSI
	// CSI_D
	// INVALID
}

func ExamplePinFunc() {
	p := gpioreg.ByName("GPIO14")
	if p == nil {
		log.Fatal("not running on a raspberry pi")
	}
	pf, ok := p.(pin.PinFunc)
	if !ok {
		log.Fatal("pin.PinFunc is not implemented")
	}
	if err := pf.SetFunc(pin.Func("UART1_TX")); err != nil {
		log.Fatal(err)
	}
}
