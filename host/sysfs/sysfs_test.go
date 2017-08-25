// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"io"

	"periph.io/x/periph/host/fs"
)

func init() {
	fs.Inhibit()
	reset()
}

func reset() {
	fileIOOpen = fileIOOpenDefault
	ioctlOpen = ioctlOpenDefault
	// Soon.
	//fileIOOpen = fileIOOpenPanic
	//ioctlOpen = ioctlOpenPanic
}

func ioctlOpenPanic(path string, flag int) (ioctlCloser, error) {
	panic("don't forget to override fileIOOpen")
}

func fileIOOpenPanic(path string, flag int) (fileIO, error) {
	panic("don't forget to override fileIOOpen")
}

type ioctlClose struct{}

func (i *ioctlClose) Ioctl(op uint, data uintptr) error {
	return nil
}

func (i *ioctlClose) Close() error {
	return nil
}

type file struct {
	ioctlClose
}

func (f *file) Fd() uintptr {
	return 0xFFFFFFFF
}

func (f *file) Read(p []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		return 0, nil
	}
	return 0, errors.New("not implemented")
}

func (f *file) Write(p []byte) (int, error) {
	return 0, errors.New("not implemented")
}
