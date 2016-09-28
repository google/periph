// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pins

import "testing"

func TestInvalid(t *testing.T) {
	if INVALID.String() != "INVALID" {
		t.Fail()
	}
}
