// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

import (
	"io/ioutil"
)

// DTModel returns platform model info from the Linux device tree (/proc/device-tree/model), and
// returns "unknown" on non-linux systems or if the file is missing.
func DTModel() string {
	lock.Lock()
	defer lock.Unlock()

	if dtModel == "" {
		dtModel = "unknown"
		if isLinux {
			// Read model from device tree.
			if bytes, err := ioutil.ReadFile("/proc/device-tree/model"); err == nil {
				if model := splitNull(bytes); len(model) > 0 {
					dtModel = model[0]
				}
			}
		}
	}
	return dtModel
}

// DTCompatible returns platform compatibility info from the Linux device tree
// (/proc/device-tree/compatible), and returns []{"unknown"} on non-linux systems or if the file is
// missing.
func DTCompatible() []string {
	lock.Lock()
	defer lock.Unlock()

	if dtCompatible == nil {
		dtCompatible = []string{}
		if isLinux {
			// Read compatible from device tree.
			if bytes, err := ioutil.ReadFile("/proc/device-tree/compatible"); err == nil {
				dtCompatible = splitNull(bytes)
			}
		}
	}
	return dtCompatible
}

var (
	dtModel      string   // cached /proc/device-tree/model
	dtCompatible []string // cached /proc/device-tree/compatible
)
