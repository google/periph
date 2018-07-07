// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
)

func TestContent(t *testing.T) {
	actual, err := ioutil.ReadFile("content_prod.go")
	if err != nil {
		t.Fatal(err)
	}
	c := exec.Command("go", "run", "internal/gen.go")
	c.Stderr = os.Stderr
	expected, err := c.Output()
	if err != nil {
		t.Fatal(string(expected), err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatal("Please run go generate")
	}
}
