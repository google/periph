// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"reflect"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio/gpiostream"
)

func TestRaster32Bits(t *testing.T) {
	b := gpiostream.BitStreamLSBF{
		Res:  time.Second,
		Bits: []byte{0x1, 0x40}}
	d32Clear := make([]uint32, 8*2)
	d32Set := make([]uint32, 8*2)
	if err := raster32Bits(&b, 1, d32Clear, d32Set, 2); err != nil {
		t.FailNow()
	}
	if !reflect.DeepEqual(d32Set, []uint32{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0}) {
		t.Errorf("unexpected d32Set %v", d32Set)
	}
	if !reflect.DeepEqual(d32Clear, []uint32{0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 0, 2}) {
		t.Errorf("unexpected d32Clear %v", d32Clear)
	}
}
