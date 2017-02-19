// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package driver_skeleton

import (
	"fmt"
	"log"
	"testing"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/host"
)

func Example() {
	// FIXME: Make sure to expose a simple use case.
	if _, err := host.Init(); err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}
	bus, err := i2c.New(-1)
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()
	dev, err := New(bus)
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
	fmt.Printf("%s\n", dev.Read())
}

func TestDriverSkeleton(t *testing.T) {
	// FIXME: Try to include basic code coverage. You can use "replay" tests by
	// leveraging i2ctest and spitest.
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Initial detection in New().
			{Addr: 42, Write: []byte("in"), Read: []byte("IN")},
			// Read().
			{Addr: 42, Write: []byte("what"), Read: []byte("Hello world!")},
		},
	}
	dev, err := New(&bus)
	if err != nil {
		t.Fatal(err)
	}

	if data := dev.Read(); data != "Hello world!" {
		t.Fatalf("unexpected %#v", data)
	}
}
