// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package conn

import "testing"

func TestDuplex(t *testing.T) {
	if Half.String() != "Half" || Duplex(10).String() != "Duplex(10)" {
		t.Fatal()
	}
}
