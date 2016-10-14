// Copyright 2016 Thorsten von Eicken. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package chip

import "github.com/maruel/dlibox/go/pio"

func init() {
	pio.MustRegister(&driver{})
}
