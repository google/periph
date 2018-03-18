// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package uart_test

import (
	"fmt"
	"log"

	"periph.io/x/periph/experimental/conn/uart"
	"periph.io/x/periph/experimental/conn/uart/uartreg"
	"periph.io/x/periph/host"
)

func ExamplePins() {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	// Use uartreg UART port registry to find the first available UART port.
	p, err := uartreg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	// Prints out the gpio pin used.
	if p, ok := p.(uart.Pins); ok {
		fmt.Printf("  RX : %s", p.RX())
		fmt.Printf("  TX : %s", p.TX())
		fmt.Printf("  RTS: %s", p.RTS())
		fmt.Printf("  CTS: %s", p.CTS())
	}
}
