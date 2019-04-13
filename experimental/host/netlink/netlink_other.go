// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package netlink

import "errors"

const isLinux = false

type connSocket struct{}

func newConnSocket() (*connSocket, error) {
	return nil, errors.New("netlink sockets are not supported")
}

func (*connSocket) send(_ []byte) error {
	return errors.New("not implemented")
}

func (*connSocket) recv(_ []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (*connSocket) close() error {
	return errors.New("not implemented")
}
