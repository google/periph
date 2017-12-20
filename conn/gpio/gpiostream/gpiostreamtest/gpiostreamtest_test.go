// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiostreamtest

import (
	"reflect"
	"testing"
	"time"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiostream"
)

// PinIn

func TestPinIn(t *testing.T) {
	p := &PinIn{
		N:   "Yo",
		Ops: []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
	}
	b := gpiostream.BitStream{Res: time.Second, Bits: make([]byte, 1), LSBF: true}
	if err := p.StreamIn(gpio.PullNoChange, &b); err != nil {
		t.Fatal(err)
	}
	if s := p.String(); s != "Yo" {
		t.Fatal(s)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPinIn_fail_type(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.EdgeStream{Res: time.Minute, Edges: make([]uint16, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("unsupported EdgeStream")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinIn_fail_res(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Res: time.Minute, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinIn_fail_len(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Res: time.Second, Bits: make([]byte, 2), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different len")
	}
}

func TestPinIn_fail_LSBF(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Res: time.Second, Bits: make([]byte, 1), LSBF: false}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different LSBF")
	}
}

func TestPinIn_fail_pull(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Res: time.Second, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullDown, &b) == nil {
		t.Fatal("different pull")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinIn_fail_count(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		Count:     1,
		DontPanic: true,
	}
	b := gpiostream.BitStream{Res: time.Second, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("count too large")
	}
}

func TestPinIn_panic_res(t *testing.T) {
	p := &PinIn{
		Ops: []InOp{{BitStream: gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
	}
	defer func() {
		if err, ok := recover().(error); !ok {
			t.Fatal("expected conntest error, got nothing")
		} else if !conntest.IsErr(err) {
			t.Fatalf("expected conntest error, got %v", err)
		}
	}()
	b := gpiostream.BitStream{Res: time.Minute, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
}

// PinOutPlayback

func TestPinOutPlayback(t *testing.T) {
	p := &PinOutPlayback{N: "Yo", Ops: []gpiostream.Stream{&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}}}
	if err := p.StreamOut(&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}); err != nil {
		t.Fatal(err)
	}
	if s := p.String(); s != "Yo" {
		t.Fatal(s)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPinOutPlayback_fail(t *testing.T) {
	p := &PinOutPlayback{DontPanic: true}
	if p.StreamOut(&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}) == nil {
		t.Fatal("expected failure")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}}}
	if p.StreamOut(&gpiostream.BitStream{Res: time.Minute, Bits: []byte{0xCC}, LSBF: true}) == nil {
		t.Fatal("different Res")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}}}
	if p.Close() == nil {
		t.Fatal("expected failure")
	}
}

// PinOutRecord

func TestPinOutRecord(t *testing.T) {
	p := &PinOutRecord{N: "Yo"}
	data := []gpiostream.Stream{
		&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true},
		&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: false},
		&gpiostream.EdgeStream{Res: time.Second, Edges: []uint16{60, 120}},
		&gpiostream.Program{Parts: []gpiostream.Stream{&gpiostream.BitStream{Res: time.Second, Bits: []byte{0xCC}, LSBF: true}}, Loops: 2},
	}
	for _, line := range data {
		if err := p.StreamOut(line); err != nil {
			t.Fatal(err)
		}
	}
	for i := range p.Ops {
		if !reflect.DeepEqual(data[i], p.Ops[i]) {
			t.Fatalf("%d data not equal", i)
		}
	}
	if s := p.String(); s != "Yo" {
		t.Fatal(s)
	}
}

func TestPinOutRecord_fail(t *testing.T) {
	p := &PinOutRecord{DontPanic: true}
	if p.StreamOut(nil) == nil {
		t.Fatal("expected failure")
	}
	if p.StreamOut(&gpiostream.Program{Parts: []gpiostream.Stream{nil}}) == nil {
		t.Fatal("expected failure")
	}
}
