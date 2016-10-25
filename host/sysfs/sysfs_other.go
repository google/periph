// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package sysfs

import "errors"

const isLinux = false

func ioctl(f uintptr, op uint, arg uintptr) error {
	return errors.New("ioctl not supported on non-linux")
}
