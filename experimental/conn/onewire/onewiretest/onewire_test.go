// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewiretest

import (
	"testing"

	"github.com/google/periph/experimental/conn/onewire"
)

func TestDevTx(t *testing.T) {
	p := Playback{
		Ops: []IO{
			{
				Addr:  0xa800000131994528,
				Write: []byte{10},
				Read:  []byte{12},
			},
		},
	}
	d := onewire.Dev{Bus: &p, Addr: 0xa800000131994528}
	buf := []byte{0}

	// Test Tx.
	err := d.Tx([]byte{10}, buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf[0] != 12 {
		t.Fail()
	}

	// Test TxPup.
	err := d.TxPup([]byte{10}, buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf[0] != 12 {
		t.Fail()
	}
}
