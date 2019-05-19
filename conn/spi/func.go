// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spi

import "periph.io/x/periph/conn/pin"

// Well known pin functionality.
const (
	CLK  pin.Func = "SPI_CLK"  // Clock
	CS   pin.Func = "SPI_CS"   // Chip select
	MISO pin.Func = "SPI_MISO" // Master in
	MOSI pin.Func = "SPI_MOSI" // Master out
)
