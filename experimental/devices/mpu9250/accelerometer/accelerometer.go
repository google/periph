// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package accelerometer

// Valid accelerator values.
const (
	ACCEL_FS_SEL_2G  = 0
	ACCEL_FS_SEL_4G  = 8
	ACCEL_FS_SEL_8G  = 0x10
	ACCEL_FS_SEL_16G = 0x18
)

// Sensitivity returns the sensitivity as float32.
func Sensitivity(selector int) float32 {
	switch selector {
	default:
		fallthrough
	case ACCEL_FS_SEL_2G:
		return 2.0 / 32768.0
	case ACCEL_FS_SEL_4G:
		return 4.0 / 32768.0
	case ACCEL_FS_SEL_8G:
		return 8.0 / 32768.0
	case ACCEL_FS_SEL_16G:
		return 16.0 / 32768.0
	}
}
