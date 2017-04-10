// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package conntest

import (
	"bytes"
	"testing"

	"periph.io/x/periph/conn"
)

func TestRecordRaw(t *testing.T) {
	b := bytes.Buffer{}
	r := RecordRaw{W: &b}
	if d := r.Duplex(); d != conn.Half {
		t.Fatal(d)
	}
	if s := r.String(); s != "recordraw" {
		t.Fatal(s)
	}
	if r.Tx(nil, []byte{0}) == nil {
		t.Fatal("cannot accept read buffer")
	}
	if err := r.Tx([]byte{'a'}, nil); err != nil {
		t.Fatal(err)
	}
	if s := b.String(); s != "a" {
		t.Fatal(s)
	}
}

func TestRecord_empty(t *testing.T) {
	r := Record{}
	if s := r.String(); s != "record" {
		t.Fatal(s)
	}
	if r.Tx(nil, []byte{'a'}) == nil {
		t.Fatal("Bus is nil")
	}
	if d := r.Duplex(); d != conn.DuplexUnknown {
		t.Fatal(d)
	}
}

func TestRecord_Tx_empty(t *testing.T) {
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
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPlayback_Close_panic(t *testing.T) {
	p := Playback{Ops: []IO{{W: []byte{10}}}}
	defer func() {
		v := recover()
		err, ok := v.(error)
		if !ok {
			t.Fatal("expected error")
		}
		if !IsErr(err) {
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
				W: []byte{10},
				R: []byte{12},
			},
		},
		DontPanic: true,
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

func TestPlayback_Tx_panic_count(t *testing.T) {
	p := Playback{}
	defer func() {
		v := recover()
		err, ok := v.(error)
		if !ok {
			t.Fatal("expected error")
		}
		if !IsErr(err) {
			t.Fatalf("unexpected error: %v", err)
		}
	}()
	p.Tx([]byte{0}, nil)
	t.Fatal("shouldn't run")
}

func TestPlayback_Tx_panic_write(t *testing.T) {
	p := Playback{Ops: []IO{{W: []byte{1}}}}
	defer func() {
		v := recover()
		err, ok := v.(error)
		if !ok {
			t.Fatal("expected error")
		}
		if !IsErr(err) {
			t.Fatalf("unexpected error: %v", err)
		}
	}()
	p.Tx([]byte{0}, nil)
	t.Fatal("shouldn't run")
}

func TestPlayback_Tx_panic_read(t *testing.T) {
	p := Playback{Ops: []IO{{R: []byte{1}}}}
	defer func() {
		v := recover()
		err, ok := v.(error)
		if !ok {
			t.Fatal("expected error")
		}
		if !IsErr(err) {
			t.Fatalf("unexpected error: %v", err)
		}
	}()
	p.Tx(nil, []byte{0, 1})
	t.Fatal("shouldn't run")
}

func TestRecord_Playback(t *testing.T) {
	r := Record{
		Conn: &Playback{
			Ops: []IO{
				{
					W: []byte{10},
					R: []byte{12},
				},
			},
			D:         conn.Full,
			DontPanic: true,
		},
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

func TestDiscard(t *testing.T) {
	d := Discard{D: conn.Half}
	if s := d.String(); s != "discard" {
		t.Fatal(s)
	}
	if v := d.Duplex(); v != conn.Half {
		t.Fatal(v)
	}
	if err := d.Tx(nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := d.Tx([]byte{0}, []byte{0}); err != nil {
		t.Fatal(err)
	}
}
