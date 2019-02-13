// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fs

// These constants, variables and functions are ported from the Linux userland
// API header ioctl.h (commonly packaged at /usr/include/linux/ioctl.h which
// includes /usr/include/asm-generic/ioctl.h).

const (
	iocNrbits   uint = 8
	iocTypebits uint = 8

	iocNrshift uint = 0

	iocTypeshift = iocNrshift + iocNrbits
	iocSizeshift = iocTypeshift + iocTypebits
	iocDirshift  = iocSizeshift + iocSizebits
)

func ioc(dir, typ, nr, size uint) uint {
	return (dir << iocDirshift) |
		(typ << iocTypeshift) |
		(nr << iocNrshift) |
		(size << iocSizeshift)
}

// IO defines an ioctl with no parameters. It corresponds to _IO in the Linux
// userland API.
func IO(typ, nr uint) uint {
	return ioc(iocNone, typ, nr, 0)
}

// IOR defines an ioctl with read (userland perspective) parameters. It
// corresponds to _IOR in the Linux userland API.
func IOR(typ, nr, size uint) uint {
	return ioc(iocRead, typ, nr, size)
}

// IOW defines an ioctl with write (userland perspective) parameters. It
// corresponds to _IOW in the Linux userland API.
func IOW(typ, nr, size uint) uint {
	return ioc(iocWrite, typ, nr, size)
}

// IOWR defines an ioctl with both read and write parameters. It corresponds to
// _IOWR in the Linux userland API.
func IOWR(typ, nr, size uint) uint {
	return ioc(iocRead|iocWrite, typ, nr, size)
}
