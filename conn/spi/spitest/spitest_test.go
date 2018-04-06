// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spitest

import (
	"bytes"
	"io/ioutil"
	"log"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/spi"
)

func TestRecordRaw(t *testing.T) {
	b := bytes.Buffer{}
	r := NewRecordRaw(&b)
	if err := r.LimitSpeed(-100); err != nil {
		t.Fatal(err)
	}
	c, err := r.Connect(0, spi.Mode0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.TxPackets(nil); err == nil {
		t.Fatal("not yet implemented")
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
	if err := r.LimitSpeed(-100); err != nil {
		t.Fatal(err)
	}
	c, err := r.Connect(0, spi.Mode0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if c.Tx(nil, []byte{'a'}) == nil {
		t.Fatal("Port is nil")
	}
	if err := c.TxPackets(nil); err == nil {
		t.Fatal("not yet implemented")
	}
	if d := c.Duplex(); d != conn.DuplexUnknown {
		t.Fatal(d)
	}
	p := c.(spi.Pins)
	if s := p.CLK(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if s := p.MOSI(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if s := p.MISO(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if s := p.CS(); s != gpio.INVALID {
		t.Fatal(s)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRecord_Tx_empty(t *testing.T) {
	r := Record{}
	c, err := r.Connect(0, spi.Mode0, 8)
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Tx(nil, nil); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 1 {
		t.Fatal(r.Ops)
	}
	if err := c.Tx([]byte{'a', 'b'}, nil); err != nil {
		t.Fatal(err)
	}
	if len(r.Ops) != 2 {
		t.Fatal(r.Ops)
	}
	if c.Tx([]byte{'a', 'b'}, []byte{'d'}) == nil {
		t.Fatal("Port is nil")
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
	if err := p.LimitSpeed(-100); err != nil {
		t.Fatal(err)
	}
	c, err := p.Connect(0, spi.Mode0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.TxPackets(nil); err == nil {
		t.Fatal("not yet implemented")
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPlayback_Tx_err(t *testing.T) {
	p := Playback{
		Playback: conntest.Playback{
			Ops:       []conntest.IO{{W: []byte{10}, R: []byte{12}}},
			DontPanic: true,
		},
	}
	c, err := p.Connect(0, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	if c.Tx(nil, nil) == nil {
		t.Fatal("missing read and write")
	}
	if p.Close() == nil {
		t.Fatal("Ops is not empty")
	}
	if c.Tx([]byte{10}, make([]byte, 2)) == nil {
		t.Fatal("invalid read size")
	}
}

func TestPlayback_Tx_empty(t *testing.T) {
	p := Playback{Playback: conntest.Playback{DontPanic: true}}
	c, err := p.Connect(0, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Tx([]byte{0}, []byte{0}); err == nil {
		t.Fatal("Playback.Ops is empty")
	}
}

func TestPlayback_Tx(t *testing.T) {
	p := Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{{W: []byte{10}, R: []byte{12}}},
		},
	}
	c, err := p.Connect(0, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	v := [1]byte{}
	if err := c.Tx([]byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRecord_Playback(t *testing.T) {
	r := Record{
		Port: &Playback{
			Playback: conntest.Playback{
				Ops:       []conntest.IO{{W: []byte{10}, R: []byte{12}}},
				D:         conn.Full,
				DontPanic: true,
			},
			CLKPin:  &gpiotest.Pin{N: "CLK"},
			MOSIPin: &gpiotest.Pin{N: "MOSI"},
			MISOPin: &gpiotest.Pin{N: "MISO"},
			CSPin:   &gpiotest.Pin{N: "CS"},
		},
	}
	if err := r.LimitSpeed(-100); err != nil {
		t.Fatal(err)
	}
	c, err := r.Connect(0, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	if d := c.Duplex(); d != conn.Full {
		t.Fatal(d)
	}
	p := c.(spi.Pins)
	if n := p.CLK().Name(); n != "CLK" {
		t.Fatal(n)
	}
	if n := p.MOSI().Name(); n != "MOSI" {
		t.Fatal(n)
	}
	if n := p.MISO().Name(); n != "MISO" {
		t.Fatal(n)
	}
	if n := p.CS().Name(); n != "CS" {
		t.Fatal(n)
	}

	v := [1]byte{}
	if err := c.Tx([]byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
	if c.Tx([]byte{10}, v[:]) == nil {
		t.Fatal("Playback.Ops is empty")
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLog_Playback(t *testing.T) {
	r := Log{
		PortCloser: &Playback{
			Playback: conntest.Playback{
				Ops:       []conntest.IO{{W: []byte{10}, R: []byte{12}}},
				D:         conn.Full,
				DontPanic: true,
			},
			CLKPin:  &gpiotest.Pin{N: "CLK"},
			MOSIPin: &gpiotest.Pin{N: "MOSI"},
			MISOPin: &gpiotest.Pin{N: "MISO"},
			CSPin:   &gpiotest.Pin{N: "CS"},
		},
	}
	if err := r.LimitSpeed(-100); err != nil {
		t.Fatal(err)
	}
	c, err := r.Connect(0, spi.Mode0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.TxPackets(nil); err == nil {
		t.Fatal("not yet implemented")
	}
	if d := c.Duplex(); d != conn.Full {
		t.Fatal(d)
	}

	v := [1]byte{}
	if err := c.Tx([]byte{10}, v[:]); err != nil {
		t.Fatal(err)
	}
	if v[0] != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
	if c.Tx([]byte{10}, v[:]) == nil {
		t.Fatal("Playback.Ops is empty")
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

//

func init() {
	log.SetOutput(ioutil.Discard)
}
