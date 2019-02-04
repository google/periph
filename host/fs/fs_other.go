// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package fs

import "errors"

const isLinux = false

func ioctl(f uintptr, op uint, arg uintptr) error {
	return errors.New("fs: ioctl not supported on non-linux")
}

type event struct{}

func (e *event) makeEvent(f uintptr) error {
	return errors.New("fs: unreachable code")
}

func (e *event) wait(timeoutms int) (int, error) {
	return 0, errors.New("fs: unreachable code")
}
