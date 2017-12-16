// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio/gpiostream"
)

func TestRaster32Bits(t *testing.T) {
	b := gpiostream.BitStreamLSB{Res: time.Second, Bits: make(gpiostream.BitsLSB, 100)}
	// TODO(maruel): Test all code path, including filtering and all errors.
	var d32Set []uint32
	var d32Clear []uint32
	if err := raster32Bits(&b, 8*time.Millisecond, d32Set, d32Clear, 2); err == nil {
		t.FailNow()
	}
}

func TestRaster32Edges(t *testing.T) {
	e := gpiostream.EdgeStream{Res: time.Second, Edges: []time.Duration{time.Second, time.Millisecond}}
	// TODO(maruel): Test all code path, including filtering and all errors.
	var d32Set []uint32
	var d32Clear []uint32
	if err := raster32Edges(&e, 8*time.Millisecond, d32Set, d32Clear, 2); err == nil {
		t.FailNow()
	}
}
