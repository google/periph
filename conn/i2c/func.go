// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2c

import "periph.io/x/periph/conn/pin"

// Well known pin functionality.
const (
	SCL pin.Func = "I2C_SCL" // Clock
	SDA pin.Func = "I2C_SDA" // Data
)
