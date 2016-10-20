// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package chip

import "github.com/google/pio"

func init() {
	pio.MustRegister(&driver{})
}