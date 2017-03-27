// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spitest

import (
	"bytes"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
)

func TestRecordRaw(t *testing.T) {
	b := bytes.Buffer{}
	r := NewRecordRaw(&b)
	if err := r.Speed(-100); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRecord_empty(t *testing.T) {
	r := Record{}
	if s := r.String(); s != "record" {
		t.Fatal(s)
	}
	if err := r.Speed(-100); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure(0, 0); err != nil {
		t.Fatal(err)
	}
	if r.Tx(nil, []byte{'a'}) == nil {
		t.Fatal("Bus is nil")
	}
	if s := r.CLK(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if s := r.MOSI(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if s := r.MISO(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if s := r.CS(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if d := r.Duplex(); d != conn.DuplexUnknown {
		t.Fatal(d)
	}
}

func TestRecord_empty_tx(t *testing.T) {
	r := Record{}
	if err := r.Tx(nil, nil); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 1 {
		t.Fatal(r.Ops)
	}
	if err := r.Tx([]byte{'a', 'b'}, nil); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 2 {
		t.Fatal(r.Ops)
	}
	if r.Tx([]byte{'a', 'b'}, []byte{'d'}) == nil {
		t.Fatal("Bus is nil")
	}
	if len(r.Ops) != 2 {
		t.Fatal(r.Ops)
	}
}

func TestPlayback(t *testing.T) {
	p := Playback{}
	if s := p.String(); s != "playback" {
		t.Fatal(s)
	}
	if err := p.Speed(-100); err != nil {
		t.Fatal(err)
	}
	if err := p.Configure(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPlayback_tx(t *testing.T) {
	p := Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				{
					Write: []byte{10},
					Read:  []byte{12},
				},
			},
		},
	}
	if p.Tx(nil, nil) == nil {
		t.Fatal("missing read and write")
	}
	if p.Close() == nil {
		t.Fatal("Ops is not empty")
	}
	v := [1]byte{}
	if p.Tx([]byte{10}, make([]byte, 2)) == nil {
		t.Fatal("invalid read size")
	}
	if err := p.Tx([]byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
	if err := p.Tx([]byte{10}, v[:]); err == nil {
		t.Fatal("Playback.Ops is empty")
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRecord_Playback(t *testing.T) {
	r := Record{
		Conn: &Playback{
			Playback: conntest.Playback{
				Ops: []conntest.IO{
					{
						Write: []byte{10},
						Read:  []byte{12},
					},
				},
				D: conn.Full,
			},
			CLKPin:  &gpiotest.Pin{N: "CLK"},
			MOSIPin: &gpiotest.Pin{N: "MOSI"},
			MISOPin: &gpiotest.Pin{N: "MISO"},
			CSPin:   &gpiotest.Pin{N: "CS"},
		},
	}
	if err := r.Speed(-100); err != nil {
		t.Fatal(err)
	}
	if err := r.Configure(0, 0); err != nil {
		t.Fatal(err)
	}
	if n := r.CLK().Name(); n != "CLK" {
		t.Fatal(n)
	}
	if n := r.MOSI().Name(); n != "MOSI" {
		t.Fatal(n)
	}
	if n := r.MISO().Name(); n != "MISO" {
		t.Fatal(n)
	}
	if n := r.CS().Name(); n != "CS" {
		t.Fatal(n)
	}
	if d := r.Duplex(); d != conn.Full {
		t.Fatal(d)
	}

	v := [1]byte{}
	if err := r.Tx([]byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
	if r.Tx([]byte{10}, v[:]) == nil {
		t.Fatal("Playback.Ops is empty")
	}
}
