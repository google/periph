// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2ctest

import (
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

func TestRecord_empty(t *testing.T) {
	r := Record{}
	if s := r.String(); s != "record" {
		t.Fatal(s)
	}
	if err := r.Speed(-100); err != nil {
		t.Fatal(err)
	}
	if r.Tx(1, nil, []byte{'a'}) == nil {
		t.Fatal("Bus is nil")
	}
	if s := r.SCL(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if s := r.SDA(); s != gpio.INVALID {
		t.Fatal(s)
	}
}

func TestRecord_Tx_empty(t *testing.T) {
	r := Record{}
	if err := r.Tx(1, nil, nil); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 1 {
		t.Fatal(r.Ops)
	}
	if err := r.Tx(1, []byte{'a', 'b'}, nil); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 2 {
		t.Fatal(r.Ops)
	}
	if r.Tx(1, []byte{'a', 'b'}, []byte{'d'}) == nil {
		t.Fatal("Bus is nil")
	}
	if len(r.Ops) != 2 {
		t.Fatal(r.Ops)
	}
}

func TestPlayback(t *testing.T) {
	p := Playback{
		SDAPin: &gpiotest.Pin{N: "DA"},
		SCLPin: &gpiotest.Pin{N: "CL"},
	}
	if s := p.String(); s != "playback" {
		t.Fatal(s)
	}
	if err := p.Speed(-100); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if n := p.SDA().Name(); n != "DA" {
		t.Fatal(n)
	}
	if n := p.SCL().Name(); n != "CL" {
		t.Fatal(n)
	}
}

func TestPlayback_Close_panic(t *testing.T) {
	p := Playback{Ops: []IO{{Write: []byte{10}}}}
	defer func() {
		v := recover()
		err, ok := v.(error)
		if !ok {
			t.Fatal("expected error")
		}
		if !conntest.IsErr(err) {
			t.Fatalf("unexpected error: %v", err)
		}
	}()
	p.Close()
	t.Fatal("shouldn't run")
}

func TestPlayback_Tx(t *testing.T) {
	p := Playback{
		Ops: []IO{
			{
				Addr:  23,
				Write: []byte{10},
				Read:  []byte{12},
			},
		},
		DontPanic: true,
	}
	if p.Tx(23, nil, nil) == nil {
		t.Fatal("missing read and write")
	}
	if p.Close() == nil {
		t.Fatal("Ops is not empty")
	}
	v := [1]byte{}
	if p.Tx(42, []byte{10}, v[:]) == nil {
		t.Fatal("invalid address")
	}
	if p.Tx(23, []byte{10}, make([]byte, 2)) == nil {
		t.Fatal("invalid read size")
	}
	if err := p.Tx(23, []byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
	if err := p.Tx(23, []byte{10}, v[:]); err == nil {
		t.Fatal("Playback.Ops is empty")
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRecord_Playback(t *testing.T) {
	r := Record{
		Bus: &Playback{
			Ops: []IO{
				{
					Addr:  23,
					Write: []byte{10},
					Read:  []byte{12},
				},
			},
			DontPanic: true,
			SDAPin:    &gpiotest.Pin{N: "DA"},
			SCLPin:    &gpiotest.Pin{N: "CL"},
		},
	}
	if err := r.Speed(-100); err != nil {
		t.Fatal(err)
	}
	if n := r.SDA().Name(); n != "DA" {
		t.Fatal(n)
	}
	if n := r.SCL().Name(); n != "CL" {
		t.Fatal(n)
	}

	v := [1]byte{}
	if err := r.Tx(23, []byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
	if r.Tx(23, []byte{10}, v[:]) == nil {
		t.Fatal("Playback.Ops is empty")
	}
}
