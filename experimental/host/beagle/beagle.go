// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package beagle

import (
	"strings"

	"periph.io/x/periph/host/distro"
)

// Present returns true if the host is a BeagleBone.
func Present() bool {
	if isArm {
		return strings.HasPrefix(distro.DTModel(), "TI AM335x BeagleBone")
	}
	return false
}
