// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package host

import (
	"testing"
)

func TestInit(t *testing.T) {
	if _, err := Init(); err != nil {
		t.Fatalf("failed to initialize periph: %v", err)
	}
}
