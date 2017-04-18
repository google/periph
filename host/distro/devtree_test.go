// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

import (
	"reflect"
	"testing"
)

func TestDTModel_fail(t *testing.T) {
	defer reset()
	DTModel()
}

func TestDTModel(t *testing.T) {
	defer reset()
	readFile = func(filename string) ([]byte, error) {
		if filename != "/proc/device-tree/model" {
			t.Fatal(filename)
		}
		return []byte("SUPER-FOO\000"), nil
	}
	DTModel()
	if c := makeDTModelLinux(); c != "SUPER-FOO" {
		t.Fatal(c)
	}
}

func TestDTCompatible_fail(t *testing.T) {
	defer reset()
	DTCompatible()
}

func TestDTCompatible(t *testing.T) {
	defer reset()
	readFile = func(filename string) ([]byte, error) {
		if filename != "/proc/device-tree/compatible" {
			t.Fatal(filename)
		}
		return []byte("SUPER\000FOO\000"), nil
	}
	DTCompatible()
	if c := makeDTCompatible(); !reflect.DeepEqual(c, []string{"SUPER", "FOO"}) {
		t.Fatal(c)
	}
}
