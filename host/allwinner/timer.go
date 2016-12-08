// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner

const (
	// 31:3 reserved
	timerPLL6            timerCtrl = 2 << 1 // CONT64_CLK_SRC_SEL; OSC24M if not set;
	timerReadLatchEnable timerCtrl = 1 << 1 // CONT64_RLATCH_EN; 1 to latch the counter to the registers
	timerClear                     = 1 << 0 // CONT64_CLR_EN; clears the counter
)

// R8: Page 96
type timerCtrl uint32
