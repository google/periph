// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package chip contains header definitions for NextThing Co's C.H.I.P. board.
//
// CHIP uses the Allwinner R8 processor and thus the allwinner host package is
// automatically imported.
//
// This package exports the U13 header, which is opposite the power LED, and
// U14, which is right next to the power LED. Most of the pins are usable as
// GPIO and are directly to the processor. These can use memory-mapped GPIO,
// which is very fast. The XIO-P0 through XIO-P7 pins are attached to a pcf8574
// I²C expander which has the result that all accesses to these pins have to go
// through the kernel and the I²C bus protocol, i.e., they're slow.
//
// GPIO edge detection (using interrupts) is only supported on a few of the
// processor's pins: AP-EINT1, AP-EINT3, CSIPCK, and CSICK. Edge detection is
// also supported on the XIO pins, but this feature is rather limited due to
// the device and the driver (for example, the driver interrupts on all edges).
//
// References
//
// http://www.chip-community.org/index.php/Hardware_Information
//
// http://docs.getchip.com/chip.html#chip-hardware
//
// A graphical view of the board headers is available at:
// http://docs.getchip.com/chip.html#pin-headers
package chip
