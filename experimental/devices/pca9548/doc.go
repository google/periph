// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package pca9548 is a driver for an 8 port I²C multiplexer that is avaliable
// from multiple venders. The main features of this multiplexer is that its has
// 8 channels and is capable of voltage level translation.
//
//
// Use with other I²C multiplexers
//
// Most I²C multiplexers work the same way as
// the pca9548 so it could be possible to use this driver with a range of other
// I²C multiplexers inclding 2 and 4 port versions.
//
//
// Limtations
//
// Interupts are not enabled. There can not be a conflict with the address of
// the multiplexer.
//
//
// Adjusting the Bus CLK
//
// The bus clock is slaved to the master bus clock, different clock for each
// port is currently not supported.
//
//
// Datasheet
//
// https://www.nxp.com/docs/en/data-sheet/PCA9548A.pdf
package pca9548
