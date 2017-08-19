// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package conn

import (
	"log"
	"testing"
)

func ExampleConn() {
	// Get a connection from one of the registries, for example:
	//   b, _ := spireg.Open("SPI0.0")
	//   c, _ := b.DevParams(1000000, spi.Mode3, 8)
	var c Conn

	// Send a command over the connection without reading.
	cmd := []byte("command")
	if err := c.Tx(cmd, nil); err != nil {
		log.Fatal(err)
	}
}

func TestDuplex(t *testing.T) {
	if Half.String() != "Half" || Duplex(10).String() != "Duplex(10)" {
		t.Fatal()
	}
}
