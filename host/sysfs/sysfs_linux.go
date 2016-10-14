// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import "syscall"

const isLinux = true

func ioctl(f uintptr, op uint, arg uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f, uintptr(op), arg); errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}
