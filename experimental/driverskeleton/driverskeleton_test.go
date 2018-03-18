// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package driverskeleton

import (
	"strings"
	"testing"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/i2c/i2ctest"
)

func TestDriverSkeleton(t *testing.T) {
	// FIXME: Try to include basic code coverage. You can use "replay" tests by
	// leveraging i2ctest and spitest.
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Initial detection in New().
			{Addr: 42, W: []byte("in"), R: []byte("IN")},
			// Read().
			{Addr: 42, W: []byte("what"), R: []byte("Hello world!")},
		},
		DontPanic: true,
	}
	dev, err := New(&bus)
	if err != nil {
		t.Fatal(err)
	}

	if data := dev.Read(); data != "Hello world!" {
		t.Fatal(data)
	}

	// Playback is empty.
	if data := dev.Read(); !strings.HasPrefix(data, "i2ctest: unexpected Tx()") {
		t.Fatal(data)
	}
}

func TestDriverSkeleton_empty(t *testing.T) {
	if dev, err := New(&i2ctest.Playback{DontPanic: true}); dev != nil || err == nil {
		t.Fatal("Tx should have failed")
	}
}

func TestDriverSkeleton_init_failed(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 42, W: []byte("in"), R: []byte("xx")},
		},
	}
	if dev, err := New(&bus); dev != nil || err == nil {
		t.Fatal("New should have failed")
	}
}

func TestInit(t *testing.T) {
	if state, err := periph.Init(); err != nil {
		t.Fatal(state, err)
	}
}
