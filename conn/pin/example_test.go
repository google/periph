// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pin_test

import (
	"fmt"

	"periph.io/x/periph/conn/pin"
)

func ExampleBasicPin() {
	// Declare a basic pin, that is not a GPIO, for registration on an header.
	b := &pin.BasicPin{N: "Exotic"}
	fmt.Println(b)

	// Output:
	// Exotic
}
