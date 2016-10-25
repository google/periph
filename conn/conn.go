// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package conn

import (
	"fmt"
	"io"
)

// Conn defines the interface for a connection on a point-to-point
// communication channel.
//
// The channel may either be write-only or read-write, either half-duplex or
// full duplex.
//
// This is the lowest common denominator for all point-to-point communication
// channels.
type Conn interface {
	// Every connection has a name, e.g. "SPI0.1", "I2C1.76", "COM6", etc.
	fmt.Stringer
	// io.Writer can be used for a write-only device.
	io.Writer
	// Tx does a single transaction.
	//
	// For full duplex protocols (SPI, UART), the two buffers must have the same
	// length as both reading and writing happen simultaneously.
	//
	// For half duplex protocols (I²C), there is no restriction as reading
	// happens after writing, and r can be nil.
	Tx(w, r []byte) error
}
