// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !linux

package distro

// DTModel returns platform model info from the Linux device tree (/proc/device-tree/model), and
// returns "" on non-linux systems
func DTModel() string { return "" }

// DTCompatible returns platform compatibility info from the Linux device tree
// (/proc/device-tree/compatible), and returns nil on non-linux systems
func DTCompatible() []string { return nil }
