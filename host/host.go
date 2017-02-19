// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package host

import "periph.io/x/periph"

// Init calls periph.Init() and returns it as-is.
//
// The only difference is that by calling host.Init(), you are guaranteed to
// have all the drivers implemented in this library to be implicitly loaded.
func Init() (*periph.State, error) {
	return periph.Init()
}
