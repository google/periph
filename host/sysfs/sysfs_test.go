// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import "periph.io/x/periph/host/fs"

func init() {
	fs.Inhibit()
}

type ioctlClose int

func (i ioctlClose) Ioctl(op uint, data uintptr) error {
	return nil
}

func (i ioctlClose) Close() error {
	return nil
}
