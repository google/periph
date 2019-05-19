// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2s

import "periph.io/x/periph/conn/pin"

// Well known pin functionality.
const (
	SCK  pin.Func = "I2S_SCK"  // Clock; occasionally named BCLK
	WS   pin.Func = "I2S_WS"   // Word (Function) select; occasionally named FS or LRCLK
	IN   pin.Func = "I2S_DIN"  // Data in (e.g. microphone)
	OUT  pin.Func = "I2S_DOUT" // Data out (e.g. speakers)
	MCLK pin.Func = "I2S_MCLK" // Master clock; rarely used
)
