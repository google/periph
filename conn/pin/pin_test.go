// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pin

import (
	"fmt"
	"testing"
)

func ExampleBasicPin() {
	// Declare a basic pin, that is not a GPIO, for registration on an header.
	b := &BasicPin{N: "Exotic"}
	fmt.Printf("%s\n", b)

	// Output:
	// Exotic
}

//

func TestInvalid(t *testing.T) {
	if INVALID.String() != "INVALID" {
		t.Fail()
	}
}

func TestBasicPin(t *testing.T) {
	b := BasicPin{N: "Pin1"}
	if s := b.String(); s != "Pin1" {
		t.Fatal(s)
	}
	if s := b.Name(); s != "Pin1" {
		t.Fatal(s)
	}
	if n := b.Number(); n != -1 {
		t.Fatal(-1)
	}
	if s := b.Function(); s != "" {
		t.Fatal(s)
	}
}
