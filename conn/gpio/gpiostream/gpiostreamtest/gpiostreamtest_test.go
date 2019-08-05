// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpiostreamtest

import (
	"reflect"
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
)

// PinIn

func TestPinIn(t *testing.T) {
	p := &PinIn{
		N:   "Yo",
		Ops: []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
	}
	b := gpiostream.BitStream{Freq: physic.Hertz, Bits: make([]byte, 1), LSBF: true}
	if err := p.StreamIn(gpio.PullNoChange, &b); err != nil {
		t.Fatal(err)
	}
	if s := p.String(); s != "Yo" {
		t.Fatal(s)
	}
	if s := p.Name(); s != "Yo" {
		t.Fatal(s)
	}
	if n := p.Number(); n != -1 {
		t.Fatal(n)
	}
	if s := p.Function(); s != "IN" {
		t.Fatal(s)
	}
	if f := p.Func(); f != gpio.IN {
		t.Fatal(f)
	}
	if v := p.SupportedFuncs(); !reflect.DeepEqual(v, []pin.Func{gpio.IN}) {
		t.Fatal(v)
	}
	if err := p.SetFunc(gpio.IN); err != nil {
		t.Fatal(err)
	}
	if err := p.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPinIn_fail_type(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.EdgeStream{Freq: physic.MilliHertz, Edges: make([]uint16, 1)}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("unsupported EdgeStream")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
	if err := p.SetFunc(pin.FuncNone); err == nil {
		t.Fatal("expected failure")
	}
}

func TestPinIn_fail_res(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Freq: physic.MilliHertz, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinIn_fail_len(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Freq: physic.Hertz, Bits: make([]byte, 2), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different len")
	}
}

func TestPinIn_fail_LSBF(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Freq: physic.Hertz, Bits: make([]byte, 1), LSBF: false}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different LSBF")
	}
}

func TestPinIn_fail_pull(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		DontPanic: true,
	}
	b := gpiostream.BitStream{Freq: physic.Hertz, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullDown, &b) == nil {
		t.Fatal("different pull")
	}
	if p.Close() == nil {
		t.Fatal("Count doesn't match Ops")
	}
}

func TestPinIn_fail_count(t *testing.T) {
	p := &PinIn{
		Ops:       []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
		Count:     1,
		DontPanic: true,
	}
	b := gpiostream.BitStream{Freq: physic.Hertz, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("count too large")
	}
}

func TestPinIn_panic_res(t *testing.T) {
	p := &PinIn{
		Ops: []InOp{{BitStream: gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}, Pull: gpio.PullNoChange}},
	}
	defer func() {
		if err, ok := recover().(error); !ok {
			t.Fatal("expected conntest error, got nothing")
		} else if !conntest.IsErr(err) {
			t.Fatalf("expected conntest error, got %v", err)
		}
	}()
	b := gpiostream.BitStream{Freq: physic.MilliHertz, Bits: make([]byte, 1), LSBF: true}
	if p.StreamIn(gpio.PullNoChange, &b) == nil {
		t.Fatal("different res")
	}
}

// PinOutPlayback

func TestPinOutPlayback(t *testing.T) {
	p := &PinOutPlayback{N: "Yo", Ops: []gpiostream.Stream{&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}}}
	if err := p.StreamOut(&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}); err != nil {
		t.Fatal(err)
	}
	if s := p.String(); s != "Yo" {
		t.Fatal(s)
	}
	if s := p.Name(); s != "Yo" {
		t.Fatal(s)
	}
	if n := p.Number(); n != -1 {
		t.Fatal(n)
	}
	if s := p.Function(); s != "OUT" {
		t.Fatal(s)
	}
	if f := p.Func(); f != gpio.OUT {
		t.Fatal(f)
	}
	if v := p.SupportedFuncs(); !reflect.DeepEqual(v, []pin.Func{gpio.OUT}) {
		t.Fatal(v)
	}
	if err := p.SetFunc(gpio.OUT); err != nil {
		t.Fatal(err)
	}
	if err := p.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPinOutPlayback_fail(t *testing.T) {
	p := &PinOutPlayback{DontPanic: true}
	if p.StreamOut(&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}) == nil {
		t.Fatal("expected failure")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}}}
	if p.StreamOut(&gpiostream.BitStream{Freq: physic.MilliHertz, Bits: []byte{0xCC}, LSBF: true}) == nil {
		t.Fatal("different Freq")
	}
	p = &PinOutPlayback{DontPanic: true, Ops: []gpiostream.Stream{&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}}}
	if p.Close() == nil {
		t.Fatal("expected failure")
	}
	if err := p.SetFunc(pin.FuncNone); err == nil {
		t.Fatal("expected failure")
	}
}

// PinOutRecord

func TestPinOutRecord(t *testing.T) {
	p := &PinOutRecord{N: "Yo"}
	data := []gpiostream.Stream{
		&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true},
		&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: false},
		&gpiostream.EdgeStream{Freq: physic.Hertz, Edges: []uint16{60, 120}},
		&gpiostream.Program{Parts: []gpiostream.Stream{&gpiostream.BitStream{Freq: physic.Hertz, Bits: []byte{0xCC}, LSBF: true}}, Loops: 2},
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
	if s := p.Name(); s != "Yo" {
		t.Fatal(s)
	}
	if n := p.Number(); n != -1 {
		t.Fatal(n)
	}
	if s := p.Function(); s != "OUT" {
		t.Fatal(s)
	}
	if f := p.Func(); f != gpio.OUT {
		t.Fatal(f)
	}
	if v := p.SupportedFuncs(); !reflect.DeepEqual(v, []pin.Func{gpio.OUT}) {
		t.Fatal(v)
	}
	if err := p.SetFunc(gpio.OUT); err != nil {
		t.Fatal(err)
	}
	if err := p.Halt(); err != nil {
		t.Fatal(err)
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
	if err := p.SetFunc(pin.FuncNone); err == nil {
		t.Fatal("expected failure")
	}
}
