// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package host

import (
	// Make sure CPU and board drivers are registered.
	_ "github.com/google/periph/host/allwinner"
	_ "github.com/google/periph/host/bcm283x"
	_ "github.com/google/periph/host/chip"
	_ "github.com/google/periph/host/odroid_c1"
	// While this board is ARM64, it may run ARM 32 bits binaries so load it on
	// 32 bits builds too.
	_ "github.com/google/periph/host/pine64"
	_ "github.com/google/periph/host/rpi"
)
