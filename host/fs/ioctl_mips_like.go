// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build mips mipsle

package fs

const (
	iocNone  uint = 1
	iocRead  uint = 2
	iocWrite uint = 4

	iocSizebits uint = 13
	iocDirbits  uint = 3
)
