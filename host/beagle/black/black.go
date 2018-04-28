// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package black implements headers for the BeagleBone Black and BeagleBone
// Black Wireless micro-computers.
//
// Reference
//
// https://beagleboard.org/black
//
// Datasheet
//
// https://elinux.org/Beagleboard:BeagleBoneBlack
//
// https://github.com/CircuitCo/BeagleBone-Black/blob/rev_b/BBB_SRM.pdf
//
// https://elinux.org/Beagleboard:Cape_Expansion_Headers
package black

import (
	"strings"

	"periph.io/x/periph/host/distro"
)

// Present returns true if the host is a BeagleBone Black or BeagleBone Black
// Wireless.
func Present() bool {
	if isArm {
		return strings.HasPrefix(distro.DTModel(), "TI AM335x BeagleBone Black")
	}
	return false
}
