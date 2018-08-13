// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build darwin dragonfly openbsd

package serial

import "syscall"

var acceptedBauds = [][2]uint32{
	{50, syscall.B50},
	{75, syscall.B75},
	{110, syscall.B110},
	{134, syscall.B134},
	{150, syscall.B150},
	{200, syscall.B200},
	{300, syscall.B300},
	{600, syscall.B600},
	{1200, syscall.B1200},
	{1800, syscall.B1800},
	{2400, syscall.B2400},
	{4800, syscall.B4800},
	{9600, syscall.B9600},
	{19200, syscall.B19200},
	{38400, syscall.B38400},
	{57600, syscall.B57600},
	{115200, syscall.B115200},
	{230400, syscall.B230400},
}
