// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package am335x exposes functionality for the Texas Instruments Sitara AM335x
// processor family.
//
// This processor family is found on the BeagleBone. PRU-ICSS functionality is
// implemented in package pru.
//
// The GPIO pins of the AM335x CPU are grouped into 3 groups of 32 pins: GPIO0,
// GPIO1, and GPIO2. The CPU documentation refers to GPIO in the form of
// GPIOx_y. To get the absolute number, as exposed by sysfs, use 32*x+y to get
// the absolute number.
//
// Datasheet
//
// Technical Reference Manual
// https://www.ti.com/lit/ug/spruh73p/spruh73p.pdf
//
// Other
//
// Marketing page
// https://www.ti.com/processors/sitara/arm-cortex-a8/am335x/overview.html
//
// Family overview
// https://www.ti.com/lit/ds/symlink/am3359.pdf
package am335x
