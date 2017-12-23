// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewiretest

import (
	"encoding/binary"
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/onewire"
)

func TestRecord_empty(t *testing.T) {
	r := Record{}
	if s := r.String(); s != "record" {
		t.Fatal(s)
	}
	if r.Tx(nil, []byte{'a'}, onewire.WeakPullup) == nil {
		t.Fatal("Bus is nil")
	}
	if s := r.Q(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if a, err := r.Search(false); len(a) != 0 || err != nil {
		t.Fatal(a, err)
	}
}

func TestRecord_Tx_empty(t *testing.T) {
	r := Record{}
	if err := r.Tx(nil, nil, onewire.WeakPullup); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 1 {
		t.Fatal(r.Ops)
	}
	if err := r.Tx([]byte{'a', 'b'}, nil, onewire.WeakPullup); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 2 {
		t.Fatal(r.Ops)
	}
	if r.Tx([]byte{'a', 'b'}, []byte{'d'}, onewire.WeakPullup) == nil {
		t.Fatal("Bus is nil")
	}
	if len(r.Ops) != 2 {
		t.Fatal(r.Ops)
	}
}

func TestPlayback_empty(t *testing.T) {
	p := Playback{
		QPin:      &gpiotest.Pin{N: "Q"},
		DontPanic: true,
	}
	if s := p.String(); s != "playback" {
		t.Fatal(s)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if n := p.Q().Name(); n != "Q" {
		t.Fatal(n)
	}

	// Empty
	if len(p.Ops) != 0 {
		t.Fatal(p.Ops)
	}
	if p.Tx([]byte{20, 21}, []byte{20, 21}, onewire.WeakPullup) == nil {
		t.Fatal("Playback.Ops is empty")
	}
}

func TestPlayback(t *testing.T) {
	p := Playback{
		Ops: []IO{
			{
				W:    []byte{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 10, 11},
				R:    []byte{12, 13},
				Pull: onewire.WeakPullup,
			},
		},
		QPin:      &gpiotest.Pin{N: "Q"},
		DontPanic: true,
	}
	if len(p.Ops) != 1 {
		t.Fatal(p.Ops)
	}
	if p.Close() == nil {
		t.Fatal("Playblack.Ops is not empty")
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
		if !conntest.IsErr(err) {
			t.Fatalf("unexpected error: %v", err)
		}
	}()
	_ = p.Close()
	t.Fatal("shouldn't run")
}

func TestPlayback_searchbit(t *testing.T) {
	p := Playback{searchBit: 64, DontPanic: true}
	if _, err := p.SearchTriplet(0); err == nil {
		t.Fatal("invalid search triplet")
	}
}

func TestPlayback_inactive(t *testing.T) {
	p := Playback{inactive: []bool{true}, DontPanic: true}
	if _, err := p.SearchTriplet(0); err == nil {
		t.Fatal("invalid search inactive devices")
	}
}

func TestRecord_Playback(t *testing.T) {
	r := Record{
		Bus: &Playback{
			Ops: []IO{
				{
					W:    []byte{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 10, 11},
					R:    []byte{12, 13},
					Pull: onewire.WeakPullup,
				},
				{
					W:    []byte{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 20, 21},
					R:    []byte{22, 23},
					Pull: onewire.StrongPullup,
				},
			},
			QPin:      &gpiotest.Pin{N: "Q"},
			DontPanic: true,
		},
	}
	if n := r.Q().Name(); n != "Q" {
		t.Fatal(n)
	}

	d := onewire.Dev{Bus: &r, Addr: 0x740000070e41ac28}
	buf := []byte{0, 0}

	if d.Tx([]byte{0, 0}, buf) == nil {
		t.Fatal("not writing expected bytes")
	}
	if d.Tx([]byte{10, 11}, buf[:1]) == nil {
		t.Fatal("read buffer not expected size")
	}
	if r.Tx([]byte{0x55, 0x28, 0xac, 0x41, 0xe, 0x7, 0x0, 0x0, 0x74, 10, 11}, buf, onewire.StrongPullup) == nil {
		t.Fatal("not expected pull")
	}

	// Test Tx.
	if err := d.Tx([]byte{10, 11}, buf); err != nil {
		t.Fatal(err)
	}
	if buf[0] != 12 || buf[1] != 13 {
		t.Errorf("expected 12 & 13, got %d %d", buf[0], buf[1])
	}

	// Test TxPower.
	if err := d.TxPower([]byte{20, 21}, buf); err != nil {
		t.Fatal(err)
	}
	if buf[0] != 22 || buf[1] != 23 {
		t.Errorf("expected 12 & 13, got %d %d", buf[0], buf[1])
	}

	// Empty
	if r.Tx([]byte{20, 21}, []byte{20, 21}, onewire.WeakPullup) == nil {
		t.Fatal("Playback.Ops is empty")
	}
}

// TestSearch is the same as ../search_test.go.
func TestSearch(t *testing.T) {
	p := Playback{
		Devices: []onewire.Address{
			0x0000000000000000,
			0x0000000000000001,
			0x0010000000000000,
			0x0000100000000000,
			0xffffffffffffffff,
			0xfc0000013199a928,
			0xf100000131856328,
		},
	}
	// Fix-up the CRC byte for each device.
	var buf [8]byte
	for i := range p.Devices {
		binary.LittleEndian.PutUint64(buf[:], uint64(p.Devices[i]))
		crc := onewire.CalcCRC(buf[:7])
		p.Devices[i] = (onewire.Address(crc) << 56) | (p.Devices[i] & 0x00ffffffffffffff)
	}

	// We're doing one search operation per device, plus a last one.
	p.Ops = make([]IO, len(p.Devices)+1)
	for i := 0; i < len(p.Ops); i++ {
		p.Ops[i] = IO{W: []byte{0xf0}, Pull: onewire.WeakPullup}
	}

	// Start search.
	if err := p.Tx([]byte{0xf0}, nil, onewire.WeakPullup); err != nil {
		t.Fatal(err)
	}

	// Perform search.
	addrs, err := p.Search(false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we got all devices.
	if len(addrs) != len(p.Devices) {
		t.Fatalf("expected %d devices, got %d", len(p.Devices), len(addrs))
	}
match:
	for _, ai := range p.Devices {
		for _, aj := range addrs {
			if ai == aj {
				continue match
			}
		}
		t.Errorf("expected to find %#x but didn't", ai)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}
