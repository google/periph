// Copyright 2016 Thorsten von Eicken. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

import (
	"io/ioutil"
	"strings"
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
		dtRead()
	}
	return dtModel
}

// DTCompatible returns platform compatibility info from the Linux device tree
// (/proc/device-tree/compatible)
func DTCompatible() []string {
	lock.Lock()
	defer lock.Unlock()

	if dtCompatible == nil {
		dtRead()
	}
	return dtCompatible
}

// dtRead reads the info we need from the device tree and caches it in local variables
func dtRead() {
	// read model from device tree
	dtModel = "unknown"
	if bytes, err := ioutil.ReadFile("/proc/device-tree/model"); err == nil {
		if model := splitNull(bytes); len(model) > 0 {
			dtModel = model[0]
		}
	}

	// read compatible from device tree
	dtCompatible = []string{}
	if bytes, err := ioutil.ReadFile("/proc/device-tree/compatible"); err == nil {
		dtCompatible = splitNull(bytes)
	}
}

// splitNull returns the null-terminated strings in the data
func splitNull(data []byte) []string {
	ss := strings.Split(string(data), "\x00")
	// the last string is typically null-terminated, so remove empty string from end of array
	if len(ss) > 0 && len(ss[len(ss)-1]) == 0 {
		ss = ss[:len(ss)-1]
	}
	return ss
}
