// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package host

import "time"

const isLinux = false

func nanospinLinux(d time.Duration) {
}
