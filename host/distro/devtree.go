// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

import (
	"encoding/binary"
	"io/ioutil"
)

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

// DTRevision returns the device revision (e.g. a02082 for the Raspberry Pi 3)
// from the Linux device tree, or 0 if the file is missing or malformed.
func DTRevision() uint32 {
	mu.Lock()
	defer mu.Unlock()

	if dtRevisionRead {
		return dtRevision
	}
	dtRevisionRead = true
	if b, _ := ioutil.ReadFile("/proc/device-tree/system/linux,revision"); len(b) >= 4 {
		dtRevision = binary.BigEndian.Uint32(b[:4])
	}
	return dtRevision
}

//

var (
	dtModel        string   // cached /proc/device-tree/model
	dtCompatible   []string // cached /proc/device-tree/compatible
	dtRevision     uint32   // cached /proc/device-tree/system/linux,revision
	dtRevisionRead bool
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
