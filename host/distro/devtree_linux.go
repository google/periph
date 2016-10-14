// Copyright 2016 Thorsten von Eicken. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

import (
	"io/ioutil"
)

var (
	dtModel      string
	dtCompatible []string
)

// DTModel returns platform model info from the Linux device tree (/proc/device-tree/model)
func DTModel() string {
	lock.Lock()
	defer lock.Unlock()

	if dtModel == "" {
		dtModel = "unknown"
		// read model from device tree
		if bytes, err := ioutil.ReadFile("/proc/device-tree/model"); err == nil {
			if model := splitNull(bytes); len(model) > 0 {
				dtModel = model[0]
			}
		}
	}
	return dtModel
}

// DTCompatible returns platform compatibility info from the Linux device tree
// (/proc/device-tree/compatible)
func DTCompatible() []string {
	lock.Lock()
	defer lock.Unlock()

	if dtCompatible == nil {
		dtCompatible = []string{}
		// read compatible from device tree
		if bytes, err := ioutil.ReadFile("/proc/device-tree/compatible"); err == nil {
			dtCompatible = splitNull(bytes)
		}
	}
	return dtCompatible
}
