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
	if !isLinux || isWSL() {
		return 0, errors.New("pmem: pagemap is not supported on this platform")
	}
	return readPageMapLinux(virtAddr)
}

//

var (
	pageMap    fileIO
	pageMapErr error
)

func readPageMapLinux(virtAddr uintptr) (uint64, error) {
	var b [8]byte
	mu.Lock()
	defer mu.Unlock()
	if pageMap == nil && pageMapErr == nil {
		// Open /proc/self/pagemap.
		//
		// It is a uint64 array where the index represents the virtual 4Kb page
		// number and the value represents the physical page properties backing
		// this virtual page.
		pageMap, pageMapErr = openFile("/proc/self/pagemap", os.O_RDONLY|os.O_SYNC)
	}
	if pageMapErr != nil {
		return 0, pageMapErr
	}
	// Convert address to page number, then index in uint64 array.
	offset := int64(virtAddr / pageSize * 8)
	if _, err := pageMap.Seek(offset, os.SEEK_SET); err != nil {
		return 0, fmt.Errorf("pmem: failed to seek at 0x%x for 0x%x: %v", offset, virtAddr, err)
	}
	n, err := pageMap.Read(b[:])
	if err != nil {
		return 0, fmt.Errorf("pmem: failed to read at 0x%x for 0x%x: %v", offset, virtAddr, err)
	}
	if n != len(b) {
		return 0, fmt.Errorf("pmem: failed to read the amount of data %d", len(b))
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}
