// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2ctest

import (
	"testing"

	"periph.io/x/periph/conn/i2c"
)

func TestDev(t *testing.T) {
	p := Playback{
		Ops: []IO{
			{
				Addr:  23,
				Write: []byte{10},
				Read:  []byte{12},
			},
		},
	}
	d := i2c.Dev{Bus: &p, Addr: 23}
	v := [1]byte{}
	if err := d.Tx([]byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fail()
	}
}
