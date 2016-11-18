// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds248x

import (
	"testing"

	"github.com/google/periph/conn/i2c"
)

const testDeviceID = uint64(0x3300000131892a28)

func TestSearch(t *testing.T) {
	bus, err := i2c.New(1)
	if err != nil {
		t.Fatal(err)
	}
	dev, err := NewI2C(bus, nil)
	if err != nil {
		t.Fatal(err)
	}

	devices, err := dev.Search(false)
	if err != nil {
		t.Fatal(err)
	}
	if len(devices) != 1 {
		t.Fatalf("search found %d devices, expected 1", len(devices))
	}
	if devices[0] != testDeviceID {
		t.Errorf("found device %#x, expected %#x", devices[0], testDeviceID)
	}
}
