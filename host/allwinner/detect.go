// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains the CPU detection logic that determines whether we have an Allwinner CPU and
// if so, which exact model. Sadly there is no science behind this, it's more of a trial and error
// using as many boards and OS flavors as possible.

package allwinner

import (
	"strings"

	"github.com/google/periph/host/distro"
)

// Present detects whether the host CPU is an Allwinner CPU.
//
// https://en.wikipedia.org/wiki/Allwinner_Technology
func Present() bool {
	// BUG(maruel): There doesn't seem to be a generic way to detect Allwinner
	// CPUs.
	return IsR8() || IsA64()
}

// IsR8 detects whether the host CPU is an Allwinner R8 CPU.
//
// It looks for the string "sun5i-r8" in /proc/device-tree/compatible.
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

// IsA64 detects whether the host CPU is an Allwinner A64 CPU.
//
// It looks for the string "sun50iw1p1" in /proc/device-tree/compatible.
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
