// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

const (
	// 31:4 reserved
	timerM3 = 1 << 3 // M3
	timerM2 = 1 << 2 // M2
	timerM1 = 1 << 1 // M1
	timerM0 = 1 << 0 // M0
)

// Page 173
type timerCtl uint32
