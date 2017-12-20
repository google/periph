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

// PinInLSBF

func TestPinInLSBF(t *testing.T) {
	p := &PinInLSBF{
		N:   "Yo",
		Ops: []InOpLSBF{{BitStreamLSBF: gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}, Pull: gpio.PullNoChange}},
	}
	b := gpiostream.BitStreamLSBF{Res: time.Second, Bits: make(gpiostream.BitsLSBF, 1)}
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

func TestPinInLSBF_fail_res(t *testing.T) {
	p := &PinInLSBF{
		Ops:       []InOpLSBF{{BitStreamLSBF: gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSBF{Res: time.Minute, Bits: make(gpiostream.BitsLSBF, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinInLSBF_fail_len(t *testing.T) {
	p := &PinInLSBF{
		Ops:       []InOpLSBF{{BitStreamLSBF: gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSBF{Res: time.Second, Bits: make(gpiostream.BitsLSBF, 2)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different len")
	}
}

func TestPinInLSBF_fail_pull(t *testing.T) {
	p := &PinInLSBF{
		Ops:       []InOpLSBF{{BitStreamLSBF: gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSBF{Res: time.Second, Bits: make(gpiostream.BitsLSBF, 1)}
	if p.StreamIn(gpio.PullDown, &b) == nil {
		t.Fatal("different pull")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinInLSBF_fail_count(t *testing.T) {
	p := &PinInLSBF{
		Ops:       []InOpLSBF{{BitStreamLSBF: gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}, Pull: gpio.PullNoChange}},
		Count:     1,
		DontPanic: true,
	}
	b := gpiostream.BitStreamLSBF{Res: time.Second, Bits: make(gpiostream.BitsLSBF, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("count too large")
	}
}

func TestPinInLSBF_panic_res(t *testing.T) {
	p := &PinInLSBF{
		Ops: []InOpLSBF{{BitStreamLSBF: gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}, Pull: gpio.PullNoChange}},
	}
	defer func() {
		if err, ok := recover().(error); !ok {
			t.Fatal("expected conntest error, got nothing")
		} else if !conntest.IsErr(err) {
			t.Fatalf("expected conntest error, got %v", err)
		}
	}()
	b := gpiostream.BitStreamLSBF{Res: time.Minute, Bits: make(gpiostream.BitsLSBF, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
}

// PinInMSBF

func TestPinInMSBF(t *testing.T) {
	p := &PinInMSBF{
		N:   "Yo",
		Ops: []InOpMSBF{{BitStreamMSBF: gpiostream.BitStreamMSBF{Res: time.Second, Bits: gpiostream.BitsMSBF{0xCC}}, Pull: gpio.PullNoChange}},
	}
	b := gpiostream.BitStreamMSBF{Res: time.Second, Bits: make(gpiostream.BitsMSBF, 1)}
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

func TestPinInMSBF_fail_res(t *testing.T) {
	p := &PinInMSBF{
		Ops:       []InOpMSBF{{BitStreamMSBF: gpiostream.BitStreamMSBF{Res: time.Second, Bits: gpiostream.BitsMSBF{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSBF{Res: time.Minute, Bits: make(gpiostream.BitsMSBF, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinInMSBF_fail_len(t *testing.T) {
	p := &PinInMSBF{
		Ops:       []InOpMSBF{{BitStreamMSBF: gpiostream.BitStreamMSBF{Res: time.Second, Bits: gpiostream.BitsMSBF{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSBF{Res: time.Second, Bits: make(gpiostream.BitsMSBF, 2)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different len")
	}
}

func TestPinInMSBF_fail_pull(t *testing.T) {
	p := &PinInMSBF{
		Ops:       []InOpMSBF{{BitStreamMSBF: gpiostream.BitStreamMSBF{Res: time.Second, Bits: gpiostream.BitsMSBF{0xCC}}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSBF{Res: time.Second, Bits: make(gpiostream.BitsMSBF, 1)}
	if p.StreamIn(gpio.PullDown, &b) == nil {
		t.Fatal("different pull")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinInMSBF_fail_count(t *testing.T) {
	p := &PinInMSBF{
		Ops:       []InOpMSBF{{BitStreamMSBF: gpiostream.BitStreamMSBF{Res: time.Second, Bits: gpiostream.BitsMSBF{0xCC}}, Pull: gpio.PullNoChange}},
		Count:     1,
		DontPanic: true,
	}
	b := gpiostream.BitStreamMSBF{Res: time.Second, Bits: make(gpiostream.BitsMSBF, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("count too large")
	}
}

func TestPinInMSBF_panic_res(t *testing.T) {
	p := &PinInMSBF{
		Ops: []InOpMSBF{{BitStreamMSBF: gpiostream.BitStreamMSBF{Res: time.Second, Bits: gpiostream.BitsMSBF{0xCC}}, Pull: gpio.PullNoChange}},
	}
	defer func() {
		if err, ok := recover().(error); !ok {
			t.Fatal("expected conntest error, got nothing")
		} else if !conntest.IsErr(err) {
			t.Fatalf("expected conntest error, got %v", err)
		}
	}()
	b := gpiostream.BitStreamMSBF{Res: time.Minute, Bits: make(gpiostream.BitsMSBF, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
}

// PinOutPlayback

func TestPinOutPlayback(t *testing.T) {
	p := &PinOutPlayback{N: "Yo", Ops: []gpiostream.Stream{&gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}}}
	if err := p.StreamOut(&gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}); err != nil {
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
	if p.StreamOut(&gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}) == nil {
		t.Fatal("expected failure")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}}}
	if p.StreamOut(&gpiostream.BitStreamLSBF{Res: time.Minute, Bits: gpiostream.BitsLSBF{0xCC}}) == nil {
		t.Fatal("different Res")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}}}
	if p.Close() == nil {
		t.Fatal("expected failure")
	}
}

// PinOutRecord

func TestPinOutRecord(t *testing.T) {
	p := &PinOutRecord{N: "Yo"}
	data := []gpiostream.Stream{
		&gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}},
		&gpiostream.BitStreamMSBF{Res: time.Second, Bits: gpiostream.BitsMSBF{0xCC}},
		&gpiostream.EdgeStream{Res: time.Second, Edges: []uint16{60, 120}},
		&gpiostream.Program{Parts: []gpiostream.Stream{&gpiostream.BitStreamLSBF{Res: time.Second, Bits: gpiostream.BitsLSBF{0xCC}}}, Loops: 2},
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
