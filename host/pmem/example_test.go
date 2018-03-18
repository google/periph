// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem_test

import (
	"log"

	"periph.io/x/periph/host/pmem"
)

func ExampleMapAsPOD() {
	// Let's say the CPU has 4 x 32 bits memory mapped registers at the address
	// 0xDEADBEEF.
	var reg *[4]uint32
	if err := pmem.MapAsPOD(0xDEADBEAF, reg); err != nil {
		log.Fatal(err)
	}
	// reg now points to physical memory.
}
