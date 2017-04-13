// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package pmem

const isLinux = false

func mmap(fd uintptr, offset int64, length int) ([]byte, error) {
	return nil, wrapf("syscall.Mmap() not implemented on this OS")
}

func munmap(b []byte) error {
	return wrapf("syscall.Munmap() not implemented on this OS")
}

func mlock(b []byte) error {
	return wrapf("syscall.Mlock() not implemented on this OS")
}

func munlock(b []byte) error {
	return wrapf("syscall.Munlock() not implemented on this OS")
}

// uallocMem allocates user space memory.
func uallocMem(size int) ([]byte, error) {
	return make([]byte, size), nil
}
