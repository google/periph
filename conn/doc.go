// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package conn defines core interfaces for protocols and connections.
//
// Subpackages implements all the supported interfaces that can be used to
// connects peripherals.
//
// Concepts
//
// periph uses 3 layered concepts for interfacing:
//
//     Bus → Port → Conn
//
// Not every subpackage expose all 3 concepts. In fact, most packages don't.
// For example, SPI doesn't exxpose Bus as the OSes generally only expose the
// Port, that is, a CS line must be selected right upfront to get an handle.
// For I²C, there's no Port to configure, so selecting a "slave" address is
// sufficient to jump directly from a Bus to a Conn.
//
// Bus
//
// A Bus is a multi-point communication channel where one "master" and multiple
// "slaves" communicate together. In the case of periph, the Bus handle is
// assumed to be the "master". The "master" generally initiates communications
// and selects the "slave" to talk to.
//
// As the "master", selecting a "slave" over a Bus converts the Bus into a Port.
//
// Examples include SPI, I²C and 1-wire. In each case, selecting a
// communication line (CS line for SPI, address for I²C or 1-wire) converts the
// Bus into a Port.
//
// Port
//
// A port is a point-to-point communication channel that is not initialized and
// thus, could be used with different parameters: number of bits per word,
// clock phase, stop bit (for UART), communication speed in Hertz, etc.
//
// Configuring a Port converts it into a Conn.
//
// Conn
//
// A Conn is a fully configured half or full duplex communication channel that
// is point-to-point, only between two devices. It is ready to use like any
// readable and/or writable pipe.
package conn
