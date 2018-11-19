// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package pca9548 is a driver for an 8 port I²C multiplexer that is available
// from multiple vendors. The main features of this multiplexer is that its has
// 8 channels and is capable of voltage level translation.
//
//
// Adjusting the Bus CLK
//
// The bus clock is slaved to the master bus clock, different clock for each
// port is currently not supported. The Maximum clock for this device is 400kHz.
//
//
// Datasheet
//
// https://www.nxp.com/docs/en/data-sheet/PCA9548A.pdf
package pca9548
