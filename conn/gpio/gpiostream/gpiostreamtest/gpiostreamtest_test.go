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

// PinInLSB

func TestPinInLSB(t *testing.T) {
	p := &PinInLSB{
		N:   "Yo",
		Ops: []InOpLSB{{BitStreamLSB: gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}}, Pull: gpio.PullNoChange}},
	}
	b := gpiostream.BitStreamLSB{Res: time.Second, Bits: make(gpiostream.BitsLSB, 1)}
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

func TestPinInLSB_fail_res(t *testing.T) {
	p := &PinInLSB{
		Ops:       []InOpLSB{{BitStreamLSB: gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSB{Res: time.Minute, Bits: make(gpiostream.BitsLSB, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinInLSB_fail_len(t *testing.T) {
	p := &PinInLSB{
		Ops:       []InOpLSB{{BitStreamLSB: gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSB{Res: time.Second, Bits: make(gpiostream.BitsLSB, 2)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different len")
	}
}

func TestPinInLSB_fail_pull(t *testing.T) {
	p := &PinInLSB{
		Ops:       []InOpLSB{{BitStreamLSB: gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSB{Res: time.Second, Bits: make(gpiostream.BitsLSB, 1)}
	if p.StreamIn(gpio.PullDown, &b) == nil {
		t.Fatal("different pull")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinInLSB_fail_count(t *testing.T) {
	p := &PinInLSB{
		Ops:       []InOpLSB{{BitStreamLSB: gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}}, Pull: gpio.PullNoChange}},
		Count:     1,
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSB{Res: time.Second, Bits: make(gpiostream.BitsLSB, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("count too large")
	}
}

func TestPinInLSB_panic_res(t *testing.T) {
	p := &PinInLSB{
		Ops: []InOpLSB{{BitStreamLSB: gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}}, Pull: gpio.PullNoChange}},
	}
	defer func() {
		if err, ok := recover().(error); !ok {
			t.Fatal("expected conntest error, got nothing")
		} else if !conntest.IsErr(err) {
			t.Fatalf("expected conntest error, got %v", err)
		}
	}()
	b := gpiostream.BitStreamLSB{Res: time.Minute, Bits: make(gpiostream.BitsLSB, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
}

// PinInMSB

func TestPinInMSB(t *testing.T) {
	p := &PinInMSB{
		N:   "Yo",
		Ops: []InOpMSB{{BitStreamMSB: gpiostream.BitStreamMSB{Res: time.Second, Bits: gpiostream.BitsMSB{0xCC}}, Pull: gpio.PullNoChange}},
	}
	b := gpiostream.BitStreamMSB{Res: time.Second, Bits: make(gpiostream.BitsMSB, 1)}
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

func TestPinInMSB_fail_res(t *testing.T) {
	p := &PinInMSB{
		Ops:       []InOpMSB{{BitStreamMSB: gpiostream.BitStreamMSB{Res: time.Second, Bits: gpiostream.BitsMSB{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSB{Res: time.Minute, Bits: make(gpiostream.BitsMSB, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinInMSB_fail_len(t *testing.T) {
	p := &PinInMSB{
		Ops:       []InOpMSB{{BitStreamMSB: gpiostream.BitStreamMSB{Res: time.Second, Bits: gpiostream.BitsMSB{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSB{Res: time.Second, Bits: make(gpiostream.BitsMSB, 2)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different len")
	}
}

func TestPinInMSB_fail_pull(t *testing.T) {
	p := &PinInMSB{
		Ops:       []InOpMSB{{BitStreamMSB: gpiostream.BitStreamMSB{Res: time.Second, Bits: gpiostream.BitsMSB{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSB{Res: time.Second, Bits: make(gpiostream.BitsMSB, 1)}
	if p.StreamIn(gpio.PullDown, &b) == nil {
		t.Fatal("different pull")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinIn_fail_count(t *testing.T) {
	p := &PinInMSB{
		Ops:       []InOpMSB{{BitStreamMSB: gpiostream.BitStreamMSB{Res: time.Second, Bits: gpiostream.BitsMSB{0xCC}}, Pull: gpio.PullNoChange}},
		Count:     1,
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSB{Res: time.Second, Bits: make(gpiostream.BitsMSB, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("count too large")
	}
}

func TestPinInMSB_panic_res(t *testing.T) {
	p := &PinInMSB{
		Ops: []InOpMSB{{BitStreamMSB: gpiostream.BitStreamMSB{Res: time.Second, Bits: gpiostream.BitsMSB{0xCC}}, Pull: gpio.PullNoChange}},
	}
	defer func() {
		if err, ok := recover().(error); !ok {
			t.Fatal("expected conntest error, got nothing")
		} else if !conntest.IsErr(err) {
			t.Fatalf("expected conntest error, got %v", err)
		}
	}()
	b := gpiostream.BitStreamMSB{Res: time.Minute, Bits: make(gpiostream.BitsMSB, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
}

// PinOutPlayback

func TestPinOutPlayback(t *testing.T) {
	p := &PinOutPlayback{N: "Yo", Ops: []gpiostream.Stream{&gpiostream.BitStream{Res: time.Second, Bits: gpiostream.Bits{0xCC}}}}
	if err := p.StreamOut(&gpiostream.BitStream{Res: time.Second, Bits: gpiostream.Bits{0xCC}}); err != nil {
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
	if p.StreamOut(&gpiostream.BitStream{Res: time.Second, Bits: gpiostream.Bits{0xCC}}) == nil {
		t.Fatal("expected failure")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStream{Res: time.Second, Bits: gpiostream.Bits{0xCC}}}}
	if p.StreamOut(&gpiostream.BitStream{Res: time.Minute, Bits: gpiostream.Bits{0xCC}}) == nil {
		t.Fatal("different Res")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStream{Res: time.Second, Bits: gpiostream.Bits{0xCC}}}}
	if p.Close() == nil {
		t.Fatal("expected failure")
	}
}

// PinOutRecord

func TestPinOutRecord(t *testing.T) {
	p := &PinOutRecord{N: "Yo"}
	data := []gpiostream.Stream{
		&gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}},
		&gpiostream.BitStreamMSB{Res: time.Second, Bits: gpiostream.BitsMSB{0xCC}},
		&gpiostream.EdgeStream{Res: time.Second, Edges: []time.Duration{time.Minute, 2 * time.Minute}},
		&gpiostream.Program{Parts: []gpiostream.Stream{&gpiostream.BitStreamLSB{Res: time.Second, Bits: gpiostream.BitsLSB{0xCC}}}, Loops: 2},
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
