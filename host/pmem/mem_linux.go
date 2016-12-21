// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"fmt"
	"syscall"
)

const isLinux = true

func mmap(fd uintptr, offset int64, length int) ([]byte, error) {
	return syscall.Mmap(int(fd), offset, length, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
}

func munmap(b []byte) error {
	return syscall.Munmap(b)
}

func mlock(b []byte) error {
	return syscall.Mlock(b)
}

func munlock(b []byte) error {
	return syscall.Munlock(b)
}

// uallocMem allocates user space memory.
func uallocMem(size int) ([]byte, error) {
	b, err := syscall.Mmap(
		0,
		0,
		size,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_ANONYMOUS|syscall.MAP_LOCKED|syscall.MAP_NORESERVE|syscall.MAP_SHARED)
	// syscall.MAP_HUGETLB / MAP_HUGE_2MB
	// See /sys/kernel/mm/hugepages but both C.H.I.P. running Jessie and Raspbian
	// Jessie do not expose huge pages. :(
	if err != nil {
		return nil, fmt.Errorf("phys: allocating %d bytes failed: %v", size, err)
	}
	return b, err
}
