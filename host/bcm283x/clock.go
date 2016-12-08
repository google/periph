// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

const (
	// 31:24 password
	passwdCtl clockCtl = 0x5A << 24 // PASSWD
	// 23:11 reserved
	mashMask clockCtl = 3 << 9 // MASH
	mash0    clockCtl = 0 << 9 // src_freq / divI  (ignores divF)
	mash1    clockCtl = 1 << 9
	mash2    clockCtl = 2 << 9
	mash3    clockCtl = 3 << 9 // will cause higher spread
	flip     clockCtl = 1 << 8 // FLIP
	busy     clockCtl = 1 << 7 // BUSY
	// 6 reserved
	kill          clockCtl = 1 << 5   // KILL
	enabClk       clockCtl = 1 << 4   // ENAB
	srcMask       clockCtl = 0xF << 0 //SRC
	srcGND        clockCtl = 0        // 0Hz
	srcOscillator clockCtl = 1        // 19.2MHz
	srcTestDebug0 clockCtl = 2        // 0Hz
	srcTestDebug1 clockCtl = 3        // 0Hz
	srcPLLA       clockCtl = 4        // 0Hz
	srcPLLC       clockCtl = 5        // 1000MHz (changes with overclock settings)
	srcPLLD       clockCtl = 6        // 500MHz
	srcHDMI       clockCtl = 7        // 216MHz
	// 8-15 == GND.
)

// clockCtl controls the clock properties.
//
// It must not be changed while busy is set or a glitch may occur.
//
// Page 107
type clockCtl uint32

const (
	// 31:24 password
	passwdDiv clockDiv = 0x5A << 24 // PASSWD
	// Integer part of the divisor
	diviShift          = 12
	diviMax   clockDiv = (1 << 12) - 1
	diviMask  clockDiv = diviMax << diviShift // DIVI
	// Fractional part of the divisor
	divfMask clockDiv = (1 << 12) - 1 // DIVF
)

// clockDiv is a 12.12 fixed point value.
//
// Page 108
type clockDiv uint32
