// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains the CPU detection logic that determines whether we have an Allwinner CPU and
// if so, which exact model. Sadly there is no science behind this, it's more of a trial and error
// using as many boards and OS flavors as possible.

package allwinner

import (
	"strings"

	"github.com/google/pio/host/distro"
)

// Present detects whether we have an Allwinner cpu by looking for one of the more specific
// models. (There doesn't seem to be a generic way to detect Allwinner cpus.)
//
// https://en.wikipedia.org/wiki/Allwinner_Technology
func Present() bool {
	return IsR8() || IsA64()
}

// IsR8 detects whether we have an Allwinner R8 cpu by looking into the device tree
// to see whether /proc/device-tree/compatible contains "sun5i-r8".
func IsR8() bool {
	if isArm {
		for _, c := range distro.DTCompatible() {
			if strings.Contains(c, "sun5i-r8") {
				return true
			}
		}
	}
	return false
}

// IsA64 detects whether we have an Allwinner A64 cpu by looking into the device tree
// to see whether /proc/device-tree/compatible contains "sun50iw1p1"
func IsA64() bool {
	if isArm {
		for _, c := range distro.DTCompatible() {
			if strings.Contains(c, "sun50iw1p1") {
				return true
			}
		}
	}
	return false
}
