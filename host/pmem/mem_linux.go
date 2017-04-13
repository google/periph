// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import "syscall"

const isLinux = true

func mmap(fd uintptr, offset int64, length int) ([]byte, error) {
	v, err := syscall.Mmap(int(fd), offset, length, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, wrapf("failed to memory map: %v", err)
	}
	return v, nil
}

func munmap(b []byte) error {
	if err := syscall.Munmap(b); err != nil {
		return wrapf("failed to unmap memory: %v", err)
	}
	return nil

}

func mlock(b []byte) error {
	if err := syscall.Mlock(b); err != nil {
		return wrapf("failed to lock memory: %v", err)
	}
	return nil
}

func munlock(b []byte) error {
	if err := syscall.Munlock(b); err != nil {
		return wrapf("failed to unlock memory: %v", err)
	}
	return nil
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
		return nil, wrapf("allocating %d bytes failed: %v", size, err)
	}
	return b, err
}
