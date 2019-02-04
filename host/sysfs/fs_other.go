// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package sysfs

type eventsListener struct {
}

// events is the global events listener.
//
// It is not used outside linux.
var events eventsListener
