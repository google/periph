// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package chip

import (
	"testing"

	"github.com/google/pio/host/allwinner"
)

func TestChipPresent(t *testing.T) {
	if !Present() {
		t.Fatalf("Did not detect presence of CHIP")
	}
	if !allwinner.Present() {
		t.Fatalf("Did not detect presence of Allwinner CPU")
	}
}
