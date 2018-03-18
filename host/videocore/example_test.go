// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package videocore_test

import (
	"log"

	"periph.io/x/periph/host/videocore"
)

func ExampleAlloc() {
	// Allocates physical memory on a Broadcom CPU by leveraging the GPU.
	// This memory can be leveraged to do DMA operations.
	m, err := videocore.Alloc(64536)
	if err != nil {
		log.Fatal(err)
	}
	// Use m
	if err := m.Close(); err != nil {
		log.Fatal(err)
	}
}
