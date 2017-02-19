// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package host

import (
	// Make sure CPU and board drivers are registered.
	_ "periph.io/x/periph/host/allwinner"
	_ "periph.io/x/periph/host/bcm283x"
	_ "periph.io/x/periph/host/chip"
	_ "periph.io/x/periph/host/odroid_c1"
	// While this board is ARM64, it may run ARM 32 bits binaries so load it on
	// 32 bits builds too.
	_ "periph.io/x/periph/host/pine64"
	_ "periph.io/x/periph/host/rpi"
)
