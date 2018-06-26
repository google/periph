// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package conn defines core interfaces for protocols and connections.
//
// This package and its subpackages describe the base interfaces to connect the
// software with the real world. It doesn't contain any implementation but
// includes registries to enable the application to discover the available
// hardware.
//
// Concepts
//
// periph uses 3 layered concepts for interfacing:
//
//     Bus → Port → Conn
//
// Not every subpackage expose all 3 concepts. In fact, most packages don't.
// For example, SPI doesn't expose Bus as the OSes generally only expose the
// Port, that is, a Chip Select (CS) line must be selected right upfront to get
// an handle.  For I²C, there's no Port to configure, so selecting a "slave"
// address is sufficient to jump directly from a Bus to a Conn.
//
// periph doesn't have yet a concept of star-like communication network, like
// an IP network.
//
// Bus
//
// A Bus is a multi-point communication channel where one "master" and multiple
// "slaves" communicate together. In the case of periph, the Bus handle is
// assumed to be the "master". The "master" generally initiates communications
// and selects the "slave" to talk to.
//
// As the "master" selects a "slave" over a bus, a virtual Port is
// automatically created.
//
// Examples include SPI, I²C and 1-wire. In each case, selecting a
// communication line (Chip Select (CS) line for SPI, address for I²C or
// 1-wire) converts the Bus into a Port.
//
// Port
//
// A port is a point-to-point communication channel that is yet to be
// initialized. It cannot be used for communication until it is connected and
// transformed into a Conn. Configuring a Port converts it into a Conn. Not all
// Port need configuration.
//
// Conn
//
// A Conn is a fully configured half or full duplex communication channel that
// is point-to-point, only between two devices. It is ready to use like any
// readable and/or writable pipe.
//
// Subpackages
//
// Most connection-type specific subpackages include subpackages:
//
// → XXXreg: registry as that is populated by the host drivers and that can be
// leveraged by applications.
//
// → XXXtest: fake implementation that can be leveraged when writing device
// driver unit test.
package conn
