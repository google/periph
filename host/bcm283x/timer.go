// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"time"

	"periph.io/x/periph/host/cpu"
)

// ReadTime returns the time on a monotonic 1Mhz clock (1µs resolution).
//
// It only works if bcm283x-dma successfully loaded. Otherwise it returns 0.
func ReadTime() time.Duration {
	if drvDMA.timerMemory == nil {
		return 0
	}
	return (time.Duration(drvDMA.timerMemory.high)<<32 | time.Duration(drvDMA.timerMemory.low)) * time.Microsecond
}

// Nanospin spins the CPU without calling into the kernel code if possible.
func Nanospin(t time.Duration) {
	start := ReadTime()
	if start == 0 {
		// Use the slow generic version.
		cpu.Nanospin(t)
		return
	}
	// TODO(maruel): Optimize code path for sub-1µs duration.
	for ReadTime()-start < t {
	}
}

//

const (
	// 31:4 reserved
	timerM3 = 1 << 3 // M3
	timerM2 = 1 << 2 // M2
	timerM1 = 1 << 1 // M1
	timerM0 = 1 << 0 // M0
)

// Page 173
type timerCtl uint32

// timerMap represents the registers to access the 1Mhz timer.
//
// Page 172
type timerMap struct {
	ctl  timerCtl // CS
	low  uint32   // CLO
	high uint32   // CHI
	c0   uint32   // 0
	c1   uint32   // C1
	c2   uint32   // C2
	c3   uint32   // C3
}
