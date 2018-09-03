// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner

import (
	"strings"
	"sync"

	"periph.io/x/periph/host/distro"
)

// Present detects whether the host CPU is an Allwinner CPU.
//
// https://en.wikipedia.org/wiki/Allwinner_Technology
func Present() bool {
	detection.do()
	return detection.isAllwinner
}

// IsR8 detects whether the host CPU is an Allwinner R8 CPU.
//
// It looks for the string "sun5i-r8" in /proc/device-tree/compatible.
func IsR8() bool {
	detection.do()
	return detection.isR8
}

// IsA20 detects whether the host CPU is an Allwinner A20 CPU.
//
// It first looks for the string "sun71-a20" in /proc/device-tree/compatible,
// and if that fails it checks for "Hardware : sun7i" in /proc/cpuinfo.
func IsA20() bool {
	detection.do()
	return detection.isA20
}

// IsA64 detects whether the host CPU is an Allwinner A64 CPU.
//
// It looks for the string "sun50iw1p1" in /proc/device-tree/compatible.
func IsA64() bool {
	detection.do()
	return detection.isA64
}

//

type detectionS struct {
	mu          sync.Mutex
	done        bool
	isAllwinner bool
	isR8        bool
	isA20       bool
	isA64       bool
}

var detection detectionS

// do contains the CPU detection logic that determines whether we have an
// Allwinner CPU and if so, which exact model.
//
// Sadly there is no science behind this, it's more of a trial and error using
// as many boards and OS flavors as possible.
func (d *detectionS) do() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.done {
		d.done = true
		if isArm {
			for _, c := range distro.DTCompatible() {
				if strings.Contains(c, "sun50iw1p1") {
					d.isA64 = true
				}
				if strings.Contains(c, "sun5i-r8") {
					d.isR8 = true
				}
				if strings.Contains(c, "sun7i-a20") {
					d.isA20 = true
				}
			}
			d.isAllwinner = d.isA64 || d.isR8 || d.isA20

			if !d.isAllwinner {
				// The kernel in the image that comes pre-installed on the pcDuino3 Nano
				// is an old 3.x kernel that doesn't expose the device-tree in procfs,
				// so do an extra check in cpuinfo as well if we haven't detected
				// anything yet.
				// Distros based on 4.x kernels do expose it.
				if hw, ok := distro.CPUInfo()["Hardware"]; ok {
					if hw == "sun7i" {
						d.isA20 = true
					}
				}
			}
		}
	}
}
