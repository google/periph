// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"io"

	"periph.io/x/periph/host/fs"
)

var ioctlOpen = ioctlOpenDefault

func ioctlOpenDefault(path string, flag int) (ioctlCloser, error) {
	f, err := fs.Open(path, flag)
	if err != nil {
		return nil, err
	}
	return f, nil
}

var fileIOOpen = fileIOOpenDefault

func fileIOOpenDefault(path string, flag int) (fileIO, error) {
	f, err := fs.Open(path, flag)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type ioctlCloser interface {
	io.Closer
	fs.Ioctler
}

type fileIO interface {
	Fd() uintptr
	fs.Ioctler
	io.Closer
	io.Reader
	io.Seeker
	io.Writer
}

// seekRead seeks to the beginning of a file and reads it.
func seekRead(f fileIO, b []byte) (int, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return 0, err
	}
	return f.Read(b)
}

// seekWrite seeks to the beginning of a file and writes to it.
func seekWrite(f fileIO, b []byte) error {
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	_, err := f.Write(b)
	return err
}
