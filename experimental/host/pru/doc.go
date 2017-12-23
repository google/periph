// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package pru exposes the Programmable Real-Time Unit Subsystem and Industrial
// Communication Subsystem (PRUICSS) functionality found on many Texas
// Instruments processors.
//
// This one of the rare way of doing true realtime on a linux microcomputer.
//
// Each PRU is a 32 bits "5ns per instruction" processor running at 200MHz,
// with 8Kb of program memory and 8Kb of data memory.
//
// The PRU processor can be found on this non-exhaustive list of processors:
//
// - 2x PRUs on am3356, am3357, am3358, am3359 https://www.ti.com/product/am3359
//
// - 4x PRUs on am4376, am4377, am4378, am4379 https://www.ti.com/product/am4379
//
// - 4x PRUs on 66ak2g02; http://www.ti.com/product/66ak2g02
//
// Datasheet
//
// Technical Reference Manual starting at page 199:
// https://www.ti.com/lit/ug/spruh73p/spruh73p.pdf
//
// Help
//
// Hands-on videos
// http://beagleboard.org/pru
//
// https://elinux.org/Ti_AM33XX_PRUSSv2
package pru
