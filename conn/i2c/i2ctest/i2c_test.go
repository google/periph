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
	d := i2c.DevReg8{i2c.Dev{Bus: &p, Addr: 23}, nil}
	v, err := d.ReadRegUint8(10)
	if err != nil {
		t.Fatal(err)
	}
	if v != 12 {
		t.Fail()
	}
}
