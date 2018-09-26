// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mt7688

import (
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/host/distro"
)

// Present returns true if a mt7688 processor is detected.
func Present() bool {
	if isMIPS {
		sysType, ok := distro.CPUInfo()["system type"]
		return ok && strings.HasPrefix(sysType, "MediaTek MT7688")
	}
	return false
}

func init() {
	// Since isMIPS is a compile time constant, the compile can strip the
	// unnecessary code and unused private symbols.
	if isMIPS {
		periph.MustRegister(&drvGPIO)
	}
}
