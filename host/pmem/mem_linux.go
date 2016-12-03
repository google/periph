// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import "syscall"

const isLinux = true

func mmap(fd uintptr, offset int64, length int) ([]byte, error) {
	return syscall.Mmap(int(fd), offset, length, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
}

func munmap(b []byte) error {
	return syscall.Munmap(b)
}
