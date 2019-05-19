// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpio

import "periph.io/x/periph/conn/pin"

// Well known pin functionality.
const (
	// Inputs
	IN      pin.Func = "IN"      // Input
	IN_HIGH pin.Func = "In/High" // Read high
	IN_LOW  pin.Func = "In/Low"  // Read low

	// Outputs
	OUT      pin.Func = "OUT"      // Output, drive
	OUT_OC   pin.Func = "OUT_OPEN" // Output, open collector/drain
	OUT_HIGH pin.Func = "Out/High" // Drive high
	OUT_LOW  pin.Func = "Out/Low"  // Drive low; open collector low

	FLOAT pin.Func = "FLOAT" // Input float or Output open collector high

	CLK pin.Func = "CLK" // Clock is a subset of a PWM, with a 50% duty cycle
	PWM pin.Func = "PWM" // Pulse Width Modulation, which is a clock with variable duty cycle
)
