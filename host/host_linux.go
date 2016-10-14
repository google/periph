// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package host

import (
	// Make sure sysfs drivers are registered.
	"syscall"
	"time"

	_ "github.com/google/pio/host/sysfs"
)

const isLinux = true

func nanospinLinux(d time.Duration) {
	// runtime.nanotime() is not exported so it cannot be used to busy loop for
	// very short sleep (10µs or less).
	time := syscall.NsecToTimespec(d.Nanoseconds())
	leftover := syscall.Timespec{}
	for {
		if err := syscall.Nanosleep(&time, &leftover); err != nil {
			time = leftover
			continue
		}
		break
	}
}
