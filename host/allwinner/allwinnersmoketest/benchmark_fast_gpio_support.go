// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file is expected to be copy-pasted in all GPIO benchmark smoke test that
// support FastOut(). The only delta shall be the package name.

package allwinnersmoketest

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiostream"
)

// runFastGPIOBenchmark runs the standardized GPIO benchmark for this specific
// implementation plus the FastOut variants.
func (s *Benchmark) runFastGPIOBenchmark() {
	s.runGPIOBenchmark()
	if !s.short {
		printBench("FastReadNaive       ", testing.Benchmark(s.benchmarkFastReadNaive))
		printBench("FastReadDiscard     ", testing.Benchmark(s.benchmarkFastReadDiscard))
		printBench("FastReadSliceLevel  ", testing.Benchmark(s.benchmarkFastReadSliceLevel))
	}
	printBench("FastReadBitsLSBLoop ", testing.Benchmark(s.benchmarkFastReadBitsLSBLoop))
	if !s.short {
		printBench("FastReadBitsMSBLoop ", testing.Benchmark(s.benchmarkFastReadBitsMSBLoop))
	}
	printBench("FastReadBitsLSBUnrol", testing.Benchmark(s.benchmarkFastReadBitsLSBUnroll))
	if !s.short {
		printBench("FastReadBitsMSBUnrol", testing.Benchmark(s.benchmarkFastReadBitsMSBUnroll))
	}
	printBench("FastOutClock        ", testing.Benchmark(s.benchmarkFastOutClock))
	if !s.short {
		printBench("FastOutSliceLevel   ", testing.Benchmark(s.benchmarkFastOutSliceLevel))
	}
	printBench("FastOutBitsLSBLoop  ", testing.Benchmark(s.benchmarkFastOutBitsLSBLoop))
	if !s.short {
		printBench("FastOutBitsMSBLoop  ", testing.Benchmark(s.benchmarkFastOutBitsMSBLoop))
	}
	printBench("FastOutBitsLSBUnroll", testing.Benchmark(s.benchmarkFastOutBitsLSBUnroll))
	if !s.short {
		printBench("FastOutBitsMSBUnroll", testing.Benchmark(s.benchmarkFastOutBitsMSBUnroll))
		printBench("FastOutInterface    ", testing.Benchmark(s.benchmarkFastOutInterface))
		printBench("FastOutMemberVariabl", testing.Benchmark(s.benchmarkFastOutMemberVariabl))
	}
}

// FastRead

// benchmarkFastInNaive reads but ignores the data.
//
// This is an intentionally naive benchmark.
func (s *Benchmark) benchmarkFastReadNaive(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.FastRead()
	}
	b.StopTimer()
}

// benchmarkFastReadDiscard reads but discards the data except for the last
// value.
//
// It measures the maximum raw read speed, at least in theory.
func (s *Benchmark) benchmarkFastReadDiscard(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	l := gpio.Low
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l = p.FastRead()
	}
	b.StopTimer()
	b.Log(l)
}

// benchmarkFastReadSliceLevel reads into a []gpio.Level.
//
// This is 8x less space efficient that using bits packing, it measures if this
// has any performance impact versus bit packing.
func (s *Benchmark) benchmarkFastReadSliceLevel(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make([]gpio.Level, b.N)
	b.ResetTimer()
	for i := range buf {
		buf[i] = p.FastRead()
	}
	b.StopTimer()
}

// benchmarkFastReadBitsLSBLoop reads into a []gpiostream.BitsLSB using a loop
// to iterate over the bits.
func (s *Benchmark) benchmarkFastReadBitsLSBLoop(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsLSB, (b.N+7)/8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if p.FastRead() {
			mask := byte(1) << uint(i&7)
			buf[i/8] |= mask
		}
	}
	b.StopTimer()
}

// benchmarkFastReadBitsMSBLoop reads into a []gpiostream.BitsMSB using a loop
// to iterate over the bits.
func (s *Benchmark) benchmarkFastReadBitsMSBLoop(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsMSB, (b.N+7)/8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if p.FastRead() {
			mask := byte(1) << uint(7-(i&7))
			buf[i/8] |= mask
		}
	}
	b.StopTimer()
}

// benchmarkFastReadBitsLSBUnroll reads into a []gpiostream.BitsLSB using an
// unrolled loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkFastReadBitsLSBLoop.
func (s *Benchmark) benchmarkFastReadBitsLSBUnroll(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsLSB, (b.N+7)/8)
	b.ResetTimer()
	for i := range buf {
		l := byte(0)
		if p.FastRead() {
			l |= 0x01
		}
		if p.FastRead() {
			l |= 0x02
		}
		if p.FastRead() {
			l |= 0x04
		}
		if p.FastRead() {
			l |= 0x08
		}
		if p.FastRead() {
			l |= 0x10
		}
		if p.FastRead() {
			l |= 0x20
		}
		if p.FastRead() {
			l |= 0x40
		}
		if p.FastRead() {
			l |= 0x80
		}
		buf[i] = l
	}
	b.StopTimer()
}

// benchmarkFastReadBitsMSBUnroll reads into a []gpiostream.BitsMSB using an
// unrolled loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkFastReadBitsMSBLoop.
func (s *Benchmark) benchmarkFastReadBitsMSBUnroll(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsMSB, (b.N+7)/8)
	b.ResetTimer()
	for i := range buf {
		l := byte(0)
		if p.FastRead() {
			l |= 0x80
		}
		if p.FastRead() {
			l |= 0x40
		}
		if p.FastRead() {
			l |= 0x20
		}
		if p.FastRead() {
			l |= 0x10
		}
		if p.FastRead() {
			l |= 0x08
		}
		if p.FastRead() {
			l |= 0x04
		}
		if p.FastRead() {
			l |= 0x02
		}
		if p.FastRead() {
			l |= 0x01
		}
		buf[i] = l
	}
	b.StopTimer()
}

// FastOut

// benchmarkFastOutClock outputs an hardcoded clock.
//
// It measures maximum raw output performance when the bitstream is hardcoded.
func (s *Benchmark) benchmarkFastOutClock(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	n := (b.N + 1) / 2
	b.ResetTimer()
	for i := 0; i < n; i++ {
		p.FastOut(gpio.High)
		p.FastOut(gpio.Low)
	}
	b.StopTimer()
}

// benchmarkFastOutSliceLevel writes into a []gpio.Level.
//
// This is 8x less space efficient that using bits packing, it measures if this
// has any performance impact versus bit packing.
func (s *Benchmark) benchmarkFastOutSliceLevel(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make([]gpio.Level, b.N)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = gpio.High
	}
	b.ResetTimer()
	for _, l := range buf {
		p.FastOut(l)
	}
	b.StopTimer()
}

// benchmarkFastOutBitsLSBLoop writes into a []gpiostream.BitsLSB using a loop
// to iterate over the bits.
func (s *Benchmark) benchmarkFastOutBitsLSBLoop(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsLSB, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0x55
	}
	b.ResetTimer()
	for _, l := range buf {
		for i := 0; i < 8; i++ {
			mask := byte(1) << uint(i)
			p.FastOut(gpio.Level(l&mask != 0))
		}
	}
	b.StopTimer()
}

// benchmarkFastOutBitsMSBLoop writes into a []gpiostream.BitsMSB using a loop
// to iterate over the bits.
func (s *Benchmark) benchmarkFastOutBitsMSBLoop(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsMSB, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0xAA
	}
	b.ResetTimer()
	for _, l := range buf {
		for i := 7; i >= 0; i-- {
			mask := byte(1) << uint(i)
			p.FastOut(gpio.Level(l&mask != 0))
		}
	}
	b.StopTimer()
}

// benchmarkFastOutBitsLSBUnroll writes into a []gpiostream.BitsLSB using an
// unrolled loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkFastOutBitsLSBLoop.
func (s *Benchmark) benchmarkFastOutBitsLSBUnroll(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsLSB, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0x55
	}
	b.ResetTimer()
	for _, l := range buf {
		p.FastOut(gpio.Level(l&0x01 != 0))
		p.FastOut(gpio.Level(l&0x02 != 0))
		p.FastOut(gpio.Level(l&0x04 != 0))
		p.FastOut(gpio.Level(l&0x08 != 0))
		p.FastOut(gpio.Level(l&0x10 != 0))
		p.FastOut(gpio.Level(l&0x20 != 0))
		p.FastOut(gpio.Level(l&0x40 != 0))
		p.FastOut(gpio.Level(l&0x80 != 0))
	}
	b.StopTimer()
}

// benchmarkFastOutBitsMSBUnroll writes into a []gpiostream.BitsMSB using an
// unrolled loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkFastOutBitsMSBLoop.
func (s *Benchmark) benchmarkFastOutBitsMSBUnroll(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsMSB, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0xAA
	}
	b.ResetTimer()
	for _, l := range buf {
		p.FastOut(gpio.Level(l&0x80 != 0))
		p.FastOut(gpio.Level(l&0x40 != 0))
		p.FastOut(gpio.Level(l&0x20 != 0))
		p.FastOut(gpio.Level(l&0x10 != 0))
		p.FastOut(gpio.Level(l&0x08 != 0))
		p.FastOut(gpio.Level(l&0x04 != 0))
		p.FastOut(gpio.Level(l&0x02 != 0))
		p.FastOut(gpio.Level(l&0x01 != 0))
	}
	b.StopTimer()
}

// benchmarkFastOutInterface is an anti-pattern where an interface is used.
//
// It is otherwise the same as benchmarkFastOutBitsMSBUnroll.
func (s *Benchmark) benchmarkFastOutInterface(b *testing.B) {
	type fastOuter interface {
		Out(l gpio.Level) error
		FastOut(l gpio.Level)
	}
	var p fastOuter = s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsMSB, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0xAA
	}
	b.ResetTimer()
	for _, l := range buf {
		p.FastOut(gpio.Level(l&0x80 != 0))
		p.FastOut(gpio.Level(l&0x40 != 0))
		p.FastOut(gpio.Level(l&0x20 != 0))
		p.FastOut(gpio.Level(l&0x10 != 0))
		p.FastOut(gpio.Level(l&0x08 != 0))
		p.FastOut(gpio.Level(l&0x04 != 0))
		p.FastOut(gpio.Level(l&0x02 != 0))
		p.FastOut(gpio.Level(l&0x01 != 0))
	}
	b.StopTimer()
}

// benchmarkFastOutMemberVariabl is an anti-pattern where the struct member
// variable is used.
//
// It is otherwise the same as benchmarkFastOutBitsMSBUnroll.
func (s *Benchmark) benchmarkFastOutMemberVariabl(b *testing.B) {
	if err := s.p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make(gpiostream.BitsMSB, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0xAA
	}
	b.ResetTimer()
	for _, l := range buf {
		s.p.FastOut(gpio.Level(l&0x80 != 0))
		s.p.FastOut(gpio.Level(l&0x40 != 0))
		s.p.FastOut(gpio.Level(l&0x20 != 0))
		s.p.FastOut(gpio.Level(l&0x10 != 0))
		s.p.FastOut(gpio.Level(l&0x08 != 0))
		s.p.FastOut(gpio.Level(l&0x04 != 0))
		s.p.FastOut(gpio.Level(l&0x02 != 0))
		s.p.FastOut(gpio.Level(l&0x01 != 0))
	}
	b.StopTimer()
}
