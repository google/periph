// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package sysfs

import (
	"errors"
	"os"
)

type event struct{}

func (e event) wait(ep, timeoutms int) (int, error) {
	return 0, errors.New("unreachable code")
}

func (e event) makeEvent(f *os.File) (int, error) {
	return 0, errors.New("unreachable code")
}

func isErrBusy(err error) bool {
	// This function is not used on non-linux.
	return false
}
