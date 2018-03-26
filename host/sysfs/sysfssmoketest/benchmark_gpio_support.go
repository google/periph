// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file is expected to be copy-pasted in all benchmark smoke test. The
// only delta shall be the package name.

package sysfssmoketest

import (
	"fmt"
	"os"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
)

// runGPIOBenchmark runs the standardized GPIO benchmark for this specific
// implementation.
//
// Type Benchmark must have two members:
// - p: concrete type that implements gpio.PinIO.
// - pull: gpio.Pull value.
func (s *Benchmark) runGPIOBenchmark() {
	if !s.short {
		printBench("ReadNaive           ", testing.Benchmark(s.benchmarkReadNaive))
		printBench("ReadDiscard         ", testing.Benchmark(s.benchmarkReadDiscard))
		printBench("ReadSliceLevel      ", testing.Benchmark(s.benchmarkReadSliceLevel))
	}
	printBench("ReadBitsLSBLoop     ", testing.Benchmark(s.benchmarkReadBitsLSBLoop))
	if !s.short {
		printBench("ReadBitsMSBLoop     ", testing.Benchmark(s.benchmarkReadBitsMSBLoop))
	}
	printBench("ReadBitsLSBUnroll   ", testing.Benchmark(s.benchmarkReadBitsLSBUnroll))
	if !s.short {
		printBench("ReadBitsMSBUnroll   ", testing.Benchmark(s.benchmarkReadBitsMSBUnroll))
	}
	printBench("OutClock            ", testing.Benchmark(s.benchmarkOutClock))
	if !s.short {
		printBench("OutSliceLevel       ", testing.Benchmark(s.benchmarkOutSliceLevel))
	}
	printBench("OutBitsLSBLoop      ", testing.Benchmark(s.benchmarkOutBitsLSBLoop))
	if !s.short {
		printBench("OutBitsMSBLoop      ", testing.Benchmark(s.benchmarkOutBitsMSBLoop))
	}
	printBench("OutBitsLSBUnroll    ", testing.Benchmark(s.benchmarkOutBitsLSBUnroll))
	if !s.short {
		printBench("OutBitsMSBUnrool    ", testing.Benchmark(s.benchmarkOutBitsMSBUnroll))
	}
}

// Read

// benchmarkReadNaive reads but ignores the data.
//
// This is an intentionally naive benchmark.
func (s *Benchmark) benchmarkReadNaive(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Read()
	}
	b.StopTimer()
}

// benchmarkReadDiscard reads but discards the data except for the last value.
//
// It measures the maximum raw read speed, at least in theory.
func (s *Benchmark) benchmarkReadDiscard(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	l := gpio.Low
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l = p.Read()
	}
	b.StopTimer()
	b.Log(l)
}

// benchmarkReadSliceLevel reads into a []gpio.Level.
//
// This is 8x less space efficient that using bits packing, it measures if this
// has any performance impact versus bit packing.
func (s *Benchmark) benchmarkReadSliceLevel(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make([]gpio.Level, b.N)
	b.ResetTimer()
	for i := range buf {
		buf[i] = p.Read()
	}
	b.StopTimer()
}

// benchmarkReadBitsLSBLoop reads into a []byte using LSBF using a loop to
// iterate over the bits.
func (s *Benchmark) benchmarkReadBitsLSBLoop(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if p.Read() {
			mask := byte(1) << uint(i&7)
			buf[i/8] |= mask
		}
	}
	b.StopTimer()
}

// benchmarkReadBitsMSBLoop reads into a []byte using MSBF using a loop to
// iterate over the bits.
func (s *Benchmark) benchmarkReadBitsMSBLoop(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if p.Read() {
			mask := byte(1) << uint(7-(i&7))
			buf[i/8] |= mask
		}
	}
	b.StopTimer()
}

// benchmarkReadBitsLSBUnroll reads into a []byte using LSBF using an unrolled
// loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkReadBitsLSBLoop.
func (s *Benchmark) benchmarkReadBitsLSBUnroll(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	b.ResetTimer()
	for i := range buf {
		l := byte(0)
		if p.Read() {
			l |= 0x01
		}
		if p.Read() {
			l |= 0x02
		}
		if p.Read() {
			l |= 0x04
		}
		if p.Read() {
			l |= 0x08
		}
		if p.Read() {
			l |= 0x10
		}
		if p.Read() {
			l |= 0x20
		}
		if p.Read() {
			l |= 0x40
		}
		if p.Read() {
			l |= 0x80
		}
		buf[i] = l
	}
	b.StopTimer()
}

// benchmarkReadBitsMSBUnroll reads into a []byte using MSBF using an unrolled
// loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkReadBitsMSBLoop.
func (s *Benchmark) benchmarkReadBitsMSBUnroll(b *testing.B) {
	p := s.p
	if err := p.In(s.pull, gpio.NoEdge); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	b.ResetTimer()
	for i := range buf {
		l := byte(0)
		if p.Read() {
			l |= 0x80
		}
		if p.Read() {
			l |= 0x40
		}
		if p.Read() {
			l |= 0x20
		}
		if p.Read() {
			l |= 0x10
		}
		if p.Read() {
			l |= 0x08
		}
		if p.Read() {
			l |= 0x04
		}
		if p.Read() {
			l |= 0x02
		}
		if p.Read() {
			l |= 0x01
		}
		buf[i] = l
	}
	b.StopTimer()
}

// Out

// benchmarkOutClock outputs an hardcoded clock.
//
// It measures maximum raw output performance when the bitstream is hardcoded.
func (s *Benchmark) benchmarkOutClock(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	n := (b.N + 1) / 2
	b.ResetTimer()
	for i := 0; i < n; i++ {
		_ = p.Out(gpio.High)
		_ = p.Out(gpio.Low)
	}
	b.StopTimer()
}

// benchmarkOutSliceLevel writes into a []gpio.Level.
//
// This is 8x less space efficient that using bits packing, it measures if this
// has any performance impact versus bit packing.
func (s *Benchmark) benchmarkOutSliceLevel(b *testing.B) {
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
		_ = p.Out(l)
	}
	b.StopTimer()
}

// benchmarkOutBitsLSBLoop writes into a []byte using LSBF using a loop to
// iterate over the bits.
func (s *Benchmark) benchmarkOutBitsLSBLoop(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0x55
	}
	b.ResetTimer()
	for _, l := range buf {
		for i := 0; i < 8; i++ {
			mask := byte(1) << uint(i)
			_ = p.Out(gpio.Level(l&mask != 0))
		}
	}
	b.StopTimer()
}

// benchmarkOutBitsMSBLoop writes into a []byte using MSBF using a loop to
// iterate over the bits.
func (s *Benchmark) benchmarkOutBitsMSBLoop(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0xAA
	}
	b.ResetTimer()
	for _, l := range buf {
		for i := 7; i >= 0; i-- {
			mask := byte(1) << uint(i)
			_ = p.Out(gpio.Level(l&mask != 0))
		}
	}
	b.StopTimer()
}

// benchmarkOutBitsLSBUnroll writes into a []byte using LSBF using an unrolled
// loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkOutBitsLSBLoop.
func (s *Benchmark) benchmarkOutBitsLSBUnroll(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0x55
	}
	b.ResetTimer()
	for _, l := range buf {
		_ = p.Out(gpio.Level(l&0x01 != 0))
		_ = p.Out(gpio.Level(l&0x02 != 0))
		_ = p.Out(gpio.Level(l&0x04 != 0))
		_ = p.Out(gpio.Level(l&0x08 != 0))
		_ = p.Out(gpio.Level(l&0x10 != 0))
		_ = p.Out(gpio.Level(l&0x20 != 0))
		_ = p.Out(gpio.Level(l&0x40 != 0))
		_ = p.Out(gpio.Level(l&0x80 != 0))
	}
	b.StopTimer()
}

// benchmarkOutBitsMSBUnroll writes into a []byte using MSBF using an unrolled
// loop to iterate over the bits.
//
// It is expected to be slightly faster than benchmarkOutBitsMSBLoop.
func (s *Benchmark) benchmarkOutBitsMSBUnroll(b *testing.B) {
	p := s.p
	if err := p.Out(gpio.Low); err != nil {
		b.Fatal(err)
	}
	buf := make([]byte, (b.N+7)/8)
	for i := 0; i < len(buf); i += 2 {
		buf[i] = 0xAA
	}
	b.ResetTimer()
	for _, l := range buf {
		_ = p.Out(gpio.Level(l&0x80 != 0))
		_ = p.Out(gpio.Level(l&0x40 != 0))
		_ = p.Out(gpio.Level(l&0x20 != 0))
		_ = p.Out(gpio.Level(l&0x10 != 0))
		_ = p.Out(gpio.Level(l&0x08 != 0))
		_ = p.Out(gpio.Level(l&0x04 != 0))
		_ = p.Out(gpio.Level(l&0x02 != 0))
		_ = p.Out(gpio.Level(l&0x01 != 0))
	}
	b.StopTimer()
}

//

func printBench(name string, r testing.BenchmarkResult) {
	if r.Bytes != 0 {
		fmt.Fprintf(os.Stderr, "unexpected %d bytes written\n", r.Bytes)
		return
	}
	if r.MemAllocs != 0 || r.MemBytes != 0 {
		fmt.Fprintf(os.Stderr, "unexpected %d bytes allocated as %d calls\n", r.MemBytes, r.MemAllocs)
		return
	}
	fmt.Printf("%s \t%s\t%s\n", name, r, toHz(r.N, r.T))
}

func toHz(n int, t time.Duration) string {
	// Periph has a ban on float64 on the library but it's not too bad on unit
	// and smoke tests.
	hz := float64(n) * float64(time.Second) / float64(t)
	switch {
	case hz >= 1000000:
		return fmt.Sprintf("%.1fMHz", hz*0.000001)
	case hz >= 1000:
		return fmt.Sprintf("%.1fkHz", hz*0.001)
	default:
		return fmt.Sprintf("%.1fHz", hz)
	}
}
