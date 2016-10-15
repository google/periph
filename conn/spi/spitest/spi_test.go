// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spitest

import (
	"testing"

	"github.com/google/pio/conn/conntest"
)

func TestBasic(t *testing.T) {
	p := Playback{}
	p.Ops = []conntest.IO{
		{
			Write: []byte{10},
			Read:  []byte{12},
		},
	}
	r := make([]byte, 1)
	if err := p.Tx([]byte{10}, r); err != nil {
		t.Fatal(err)
	}
	if r[0] != 12 {
		t.Fail()
	}
	if err := p.Speed(0); err != nil {
		t.Fatal(err)
	}
}
