// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package jtag

import "periph.io/x/periph/conn/pin"

// Well known pin functionality.
const (
	TCK  pin.Func = "JTAG_TCK"  // Test clock
	TDI  pin.Func = "JTAG_TDI"  // Test mode data input
	TDO  pin.Func = "JTAG_TDO"  // Test mode data output
	TMS  pin.Func = "JTAG_TMS"  // Test mode select
	TRST pin.Func = "JTAG_TRST" // Test reset
)
