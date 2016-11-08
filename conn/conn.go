// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package conn

// Conn defines the interface for a connection on a point-to-point
// communication channel.
//
// The channel may either be write-only or read-write, either half-duplex or
// full duplex.
//
// This is the lowest common denominator for all point-to-point communication
// channels.
//
// Implementation are expected to also implement the following interfaces:
// - fmt.Stringer which returns something meaningful to the user like "SPI0.1",
//   "I2C1.76", "COM6", etc.
// - io.Writer as an way to use io.Copy() on a write-only device.
type Conn interface {
	// Tx does a single transaction.
	//
	// For full duplex protocols (SPI, UART), the two buffers must have the same
	// length as both reading and writing happen simultaneously.
	//
	// For half duplex protocols (I²C), there is no restriction as reading
	// happens after writing, and r can be nil.
	Tx(w, r []byte) error
}
