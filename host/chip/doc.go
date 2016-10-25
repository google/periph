// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package chip contains header definitions for NextThing Co's C.H.I.P. board.
// CHIP uses the Allwinner R8 processor and thus the allwinner host package is
// automatically imported.
//
// This package exports the U13 header, which is opposite the power LED, and U14,
// which is right next to the power LED. Most of the pins are usable as GPIO and
// are directly to the processor. These can use memory-mapped GPIO, which is very
// fast. The XIO-P0 through XIO-P7 pins are attached to a pcf8574 i2c expander
// which has the result that all accesses to these pins have to go through the kernel
// and the i2c bus protocol, i.e., they're slow.
//
// GPIO edge detection (using interrupts) is only supported on a few of the processor's
// pins: AP-EINT1, AP-EINT3, CSIPCK, and CSICK. Edge detection is also supported on the
// XIO pins, but this feature is rather limited due to the device and the driver
// (for example, the driver interrupts on all edges).
//
// References
//
// http://www.chip-community.org/index.php/Hardware_Information and
// http://docs.getchip.com/chip.html#chip-hardware
//
// Detection
//
// In order to detect the presence of CHIP the following info coming from the
// device tree is expected:
//   root@chip2:/proc/device-tree# od -c compatible
//   0000000   n   e   x   t   t   h   i   n   g   ,   c   h   i   p  \0   a
//   0000020   l   l   w   i   n   n   e   r   ,   s   u   n   5   i   -   r
//   0000040   8  \0
//   root@chip2:/proc/device-tree# od -c model
//   0000000   N   e   x   t   T   h   i   n   g       C   .   H   .   I   .
//   0000020   P   .  \0
package chip
