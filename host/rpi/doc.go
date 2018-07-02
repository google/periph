// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package rpi contains Raspberry Pi hardware logic. It is intrinsically
// related to package bcm283x.
//
// Assumes Raspbian but does not directly depend on the distro being Raspbian.
// Windows IoT is currently not supported.
//
// Physical
//
// The physical pin out is based on http://www.raspberrypi.org information but
// http://pinout.xyz/ has a nice interactive web page.
package rpi
