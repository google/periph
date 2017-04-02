// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spitest is meant to be used to test drivers over a fake SPI bus.
package spitest

import (
	"errors"
	"io"
	"log"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
)

// RecordRaw implements spi.Conn. It sends everything written to it to W.
type RecordRaw struct {
	conntest.RecordRaw
}

// NewRecordRaw is a shortcut to create a RecordRaw
func NewRecordRaw(w io.Writer) *RecordRaw {
	return &RecordRaw{conntest.RecordRaw{W: w}}
}

// Close is a no-op.
func (r *RecordRaw) Close() error {
	r.Lock()
	defer r.Unlock()
	return nil
}

// Speed is a no-op.
func (r *RecordRaw) Speed(maxHz int64) error {
	return nil
}

// DevParams is a no-op.
func (r *RecordRaw) DevParams(maxHz int64, mode spi.Mode, bits int) error {
	return nil
}

// Record implements spi.Conn that records everything written to it.
//
// This can then be used to feed to Playback to do "replay" based unit tests.
type Record struct {
	sync.Mutex
	Conn spi.ConnCloser // Conn can be nil if only writes are being recorded.
	Ops  []conntest.IO
}

func (r *Record) String() string {
	return "record"
}

// Tx implements spi.Conn.
func (r *Record) Tx(w, read []byte) error {
	r.Lock()
	defer r.Unlock()
	if r.Conn == nil {
		if len(read) != 0 {
			return errors.New("spitest: read unsupported when no bus is connected")
		}
	} else {
		if err := r.Conn.Tx(w, read); err != nil {
			return err
		}
	}
	io := conntest.IO{Write: make([]byte, len(w))}
	if len(read) != 0 {
		io.Read = make([]byte, len(read))
	}
	copy(io.Write, w)
	copy(io.Read, read)
	r.Ops = append(r.Ops, io)
	return nil
}

// Duplex implements spi.Conn.
func (r *Record) Duplex() conn.Duplex {
	if r.Conn != nil {
		return r.Conn.Duplex()
	}
	return conn.DuplexUnknown
}

// Close implements spi.ConnCloser.
func (r *Record) Close() error {
	if r.Conn != nil {
		return r.Conn.Close()
	}
	return nil
}

// Speed implements spi.ConnCloser.
func (r *Record) Speed(maxHz int64) error {
	if r.Conn != nil {
		return r.Conn.Speed(maxHz)
	}
	return nil
}

// DevParams implements spi.ConnCloser.
func (r *Record) DevParams(maxHz int64, mode spi.Mode, bits int) error {
	if r.Conn != nil {
		return r.Conn.DevParams(maxHz, mode, bits)
	}
	return nil
}

// CLK implements spi.Pins.
func (r *Record) CLK() gpio.PinOut {
	if p, ok := r.Conn.(spi.Pins); ok {
		return p.CLK()
	}
	return gpio.INVALID
}

// MOSI implements spi.Pins.
func (r *Record) MOSI() gpio.PinOut {
	if p, ok := r.Conn.(spi.Pins); ok {
		return p.MOSI()
	}
	return gpio.INVALID
}

// MISO implements spi.Pins.
func (r *Record) MISO() gpio.PinIn {
	if p, ok := r.Conn.(spi.Pins); ok {
		return p.MISO()
	}
	return gpio.INVALID
}

// CS implements spi.Pins.
func (r *Record) CS() gpio.PinOut {
	if p, ok := r.Conn.(spi.Pins); ok {
		return p.CS()
	}
	return gpio.INVALID
}

// Playback implements spi.Conn and plays back a recorded I/O flow.
//
// While "replay" type of unit tests are of limited value, they still present
// an easy way to do basic code coverage.
type Playback struct {
	conntest.Playback
	CLKPin  gpio.PinIO
	MOSIPin gpio.PinIO
	MISOPin gpio.PinIO
	CSPin   gpio.PinIO
}

// Close implements i2c.BusCloser.
//
// Close() verifies that all the expected Ops have been consumed.
func (p *Playback) Close() error {
	return p.Playback.Close()
}

// Speed implements spi.ConnCloser.
func (p *Playback) Speed(maxHz int64) error {
	return nil
}

// DevParams implements spi.ConnCloser.
func (p *Playback) DevParams(maxHz int64, mode spi.Mode, bits int) error {
	return nil
}

// CLK implements spi.Pins.
func (p *Playback) CLK() gpio.PinOut {
	return p.CLKPin
}

// MOSI implements spi.Pins.
func (p *Playback) MOSI() gpio.PinOut {
	return p.MOSIPin
}

// MISO implements spi.Pins.
func (p *Playback) MISO() gpio.PinIn {
	return p.MISOPin
}

// CS implements spi.Pins.
func (p *Playback) CS() gpio.PinOut {
	return p.CSPin
}

// Log logs all operations done on an spi.ConnCloser.
type Log struct {
	Conn spi.ConnCloser
}

// Close implements spi.ConnCloser.
func (l *Log) Close() error {
	err := l.Conn.Close()
	log.Printf("%s.Close() = %v", l.Conn, err)
	return err
}

// Speed implements spi.ConnCloser.
func (l *Log) Speed(maxHz int64) error {
	err := l.Conn.Speed(maxHz)
	log.Printf("%s.Speed(%d) = %v", l.Conn, maxHz, err)
	return err
}

// DevParams implements spi.ConnCloser.
func (l *Log) DevParams(maxHz int64, mode spi.Mode, bits int) error {
	err := l.Conn.DevParams(maxHz, mode, bits)
	log.Printf("%s.DevParams(%d, %d, %d) = %v", l.Conn, maxHz, mode, bits, err)
	return err
}

// Tx implements spi.Conn.
func (l *Log) Tx(w, r []byte) error {
	err := l.Conn.Tx(w, r)
	log.Printf("%s.Tx(%#v, %#v) = %v", l.Conn, w, r, err)
	return err
}

// Duplex implements spi.Conn.
func (l *Log) Duplex() conn.Duplex {
	return l.Conn.Duplex()
}

//

var _ spi.Conn = &RecordRaw{}
var _ spi.Conn = &Record{}
var _ spi.Pins = &Record{}
var _ spi.Conn = &Playback{}
