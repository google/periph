// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package conn

import "fmt"

// Duplex declares whether communication can happen simultaneously both ways.
//
// Some protocol can be either depending on configuration settings, like UART.
type Duplex int

const (
	// DuplexUnknown is used when the duplex of a connection is yet to be known.
	//
	// Some protocol can be configured either as half-duplex or full-duplex and
	// the connection is not yet is a determinate state.
	DuplexUnknown Duplex = 0
	// Half means that communication can only occurs one way at a time.
	//
	// Examples include 1-wire and I²C.
	Half Duplex = 1
	// Full means that communication occurs simultaneously both ways in a
	// synchronized manner.
	//
	// Examples include SPI (except 3-wire variant).
	Full Duplex = 2
)

const duplexName = "DuplexUnknownHalfFull"

var duplexIndex = [...]uint8{0, 13, 17, 21}

func (i Duplex) String() string {
	if i < 0 || i >= Duplex(len(duplexIndex)-1) {
		return fmt.Sprintf("Duplex(%d)", i)
	}
	return duplexName[duplexIndex[i]:duplexIndex[i+1]]
}

// Conn defines the interface for a connection on a point-to-point
// communication channel.
//
// The connection can either be unidirectional (read-only, write-only) or
// bidirectional (read-write). It can either be half-duplex or full duplex.
//
// This is the lowest common denominator for all point-to-point communication
// channels.
//
// Implementation are expected but not required to also implement the following
// interfaces:
//
// - fmt.Stringer which returns something meaningful to the user like "SPI0.1",
// "I2C1.76", "COM6", etc.
//
// - io.Reader and io.Writer as a way to use io.Copy() for half duplex
// operation.
//
// - io.Closer for the owner of the communication channel.
type Conn interface {
	// Tx does a single transaction.
	//
	// For full duplex protocols (generally SPI, UART), the two buffers must have
	// the same length as both reading and writing happen simultaneously.
	//
	// For half duplex protocols (I²C), there is no restriction as reading
	// happens after writing, and r can be nil.
	//
	// Query Limits.MaxTxSize() to know if there is a limit on the buffer size
	// per Tx() call.
	Tx(w, r []byte) error
	// Duplex returns the current duplex setting for this point-to-point
	// connection.
	//
	// It is expected to be either Half or Full unless the connection itself is
	// in an unknown state.
	Duplex() Duplex
}

// Limits returns information about the connection's limits.
type Limits interface {
	// MaxTxSize returns the maximum allowed data size to be sent as a single
	// I/O.
	//
	// Returns 0 if undefined.
	MaxTxSize() int
}
