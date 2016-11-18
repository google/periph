// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire

import (
	"fmt"
	"testing"
)

func ExampleAll() {
	fmt.Print("1-wire buses available:\n")
	for name := range All() {
		fmt.Printf("- %s\n", name)
	}
}

func TestInvalid(t *testing.T) {
	if _, err := New(-1); err == nil {
		t.Fail()
	}
}

func TestAreInOnewireTest(t *testing.T) {
	// Real tests are in onewiretest due to cyclic dependency.
}
