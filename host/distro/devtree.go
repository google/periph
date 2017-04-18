// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

// DTModel returns platform model info from the Linux device tree (/proc/device-tree/model), and
// returns "unknown" on non-linux systems or if the file is missing.
func DTModel() string {
	mu.Lock()
	defer mu.Unlock()

	if dtModel == "" {
		dtModel = "<unknown>"
		if isLinux {
			dtModel = makeDTModelLinux()
		}
	}
	return dtModel
}

// DTCompatible returns platform compatibility info from the Linux device tree
// (/proc/device-tree/compatible), and returns []{"unknown"} on non-linux systems or if the file is
// missing.
func DTCompatible() []string {
	mu.Lock()
	defer mu.Unlock()

	if dtCompatible == nil {
		dtCompatible = []string{}
		if isLinux {
			dtCompatible = makeDTCompatible()
		}
	}
	return dtCompatible
}

//

var (
	dtModel      string   // cached /proc/device-tree/model
	dtCompatible []string // cached /proc/device-tree/compatible
)

func makeDTModelLinux() string {
	// Read model from device tree.
	if bytes, err := readFile("/proc/device-tree/model"); err == nil {
		if model := splitNull(bytes); len(model) > 0 {
			return model[0]
		}
	}
	return "<unknown>"
}

func makeDTCompatible() []string {
	// Read compatible from device tree.
	if bytes, err := readFile("/proc/device-tree/compatible"); err == nil {
		return splitNull(bytes)
	}
	return []string{}
}
