// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package host

import (
	// Make sure CPU and board drivers are registered.
	_ "github.com/google/pio/host/allwinner"
	_ "github.com/google/pio/host/allwinner_pl"
	_ "github.com/google/pio/host/bcm283x"
	_ "github.com/google/pio/host/pine64"
	_ "github.com/google/pio/host/rpi"
)
