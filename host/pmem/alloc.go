// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// ReadPageMap reads a physical address mapping for a virtual page address from
// /proc/self/pagemap.
//
// It returns the physical address that corresponds to the start of the virtual
// page within which the virtual address virtAddr is located.
//
// The meaning of the return value is documented at
// https://www.kernel.org/doc/Documentation/vm/pagemap.txt
func ReadPageMap(virtAddr uintptr) (uint64, error) {
	if !isLinux {
		return nil, errors.New("pmem: pagemap is not supported on this platform")
	}
	var b [8]byte
	mu.Lock()
	defer mu.Unlock()
	// Convert address to page number, then index in uint64 array.
	if _, err := pageMap.Seek(int64(virtAddr/4096*8), os.SEEK_SET); err != nil {
		return 0, err
	}
	n, err := pageMap.Read(b[:])
	if err != nil {
		return 0, err
	}
	if n != len(b) {
		return 0, fmt.Errorf("pmem: failed to read the amount of data %d", len(b))
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

//

var (
	pageMap    *os.File
	pageMapErr error
)

// openPageMapLinux() opens /proc/self/pagemap.
//
// It is a uint64 array where the index represents the virtual 4Kb page number
// and the value represents the physical page properties backing this virtual
// page.
func openPageMapLinux() (*os.File, error) {
	mu.Lock()
	defer mu.Unlock()
	if pageMap == nil && pageMapErr == nil {
		pageMap, pageMapErr = os.OpenFile("/proc/self/pagemap", os.O_RDONLY|os.O_SYNC, 0)
	}
	return pageMap, pageMapErr
}
