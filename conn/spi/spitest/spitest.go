// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spitest is meant to be used to test drivers over a fake SPI port.
package spitest

import (
	"io"
	"log"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
)

// RecordRaw implements spi.PortCloser.
//
// It sends everything written to it to W.
type RecordRaw struct {
	conntest.RecordRaw
	Initialized bool
}

// NewRecordRaw is a shortcut to create a RecordRaw
func NewRecordRaw(w io.Writer) *RecordRaw {
	return &RecordRaw{RecordRaw: conntest.RecordRaw{W: w}}
}

// Close is a no-op.
func (r *RecordRaw) Close() error {
	r.Lock()
	defer r.Unlock()
	return nil
}

// LimitSpeed is a no-op.
func (r *RecordRaw) LimitSpeed(maxHz int64) error {
	return nil
}

// Connect is a no-op.
func (r *RecordRaw) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	r.Lock()
	defer r.Unlock()
	if r.Initialized {
		return nil, conntest.Errorf("spitest: Connect cannot be called twice")
	}
	r.Initialized = true
	return &recordRawConn{r}, nil
}

type recordRawConn struct {
	r *RecordRaw
}

func (r *recordRawConn) String() string {
	return r.r.String()
}

func (r *recordRawConn) Tx(w, read []byte) error {
	return r.r.Tx(w, read)
}

func (r *recordRawConn) Duplex() conn.Duplex {
	return r.r.Duplex()
}

func (r *recordRawConn) TxPackets(p []spi.Packet) error {
	return conntest.Errorf("spitest: TxPackets is not implemented")
}

//

// Record implements spi.PortCloser that records everything written to it.
//
// This can then be used to feed to Playback to do "replay" based unit tests.
type Record struct {
	sync.Mutex
	Port        spi.PortCloser // Port can be nil if only writes are being recorded.
	Ops         []conntest.IO
	Initialized bool
}

func (r *Record) String() string {
	return "record"
}

// Close implements spi.PortCloser.
func (r *Record) Close() error {
	if r.Port != nil {
		return r.Port.Close()
	}
	return nil
}

// LimitSpeed implements spi.PortCloser.
func (r *Record) LimitSpeed(maxHz int64) error {
	if r.Port != nil {
		return r.Port.LimitSpeed(maxHz)
	}
	return nil
}

// Connect implements spi.PortCloser.
func (r *Record) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	r.Lock()
	defer r.Unlock()
	if r.Initialized {
		return nil, conntest.Errorf("spitest: Connect cannot be called twice")
	}
	r.Initialized = true
	if r.Port != nil {
		c, err := r.Port.Connect(maxHz, mode, bits)
		if err != nil {
			return nil, err
		}
		return &recordConn{r, c}, nil
	}
	return &recordConn{r, nil}, nil
}

// CLK implements spi.Pins.
func (r *Record) CLK() gpio.PinOut {
	if p, ok := r.Port.(spi.Pins); ok {
		return p.CLK()
	}
	return gpio.INVALID
}

// MOSI implements spi.Pins.
func (r *Record) MOSI() gpio.PinOut {
	if p, ok := r.Port.(spi.Pins); ok {
		return p.MOSI()
	}
	return gpio.INVALID
}

// MISO implements spi.Pins.
func (r *Record) MISO() gpio.PinIn {
	if p, ok := r.Port.(spi.Pins); ok {
		return p.MISO()
	}
	return gpio.INVALID
}

// CS implements spi.Pins.
func (r *Record) CS() gpio.PinOut {
	if p, ok := r.Port.(spi.Pins); ok {
		return p.CS()
	}
	return gpio.INVALID
}

func (r *Record) txInternal(c spi.Conn, w, read []byte) error {
	io := conntest.IO{}
	if len(w) != 0 {
		io.W = make([]byte, len(w))
		copy(io.W, w)
	}
	r.Lock()
	defer r.Unlock()
	if r.Port == nil {
		if len(read) != 0 {
			return conntest.Errorf("spitest: read unsupported when no port is connected")
		}
	} else {
		if err := c.Tx(w, read); err != nil {
			return err
		}
	}
	if len(read) != 0 {
		io.R = make([]byte, len(read))
		copy(io.R, read)
	}
	r.Ops = append(r.Ops, io)
	return nil
}

//

type recordConn struct {
	r *Record
	c spi.Conn
}

func (r *recordConn) String() string {
	return r.r.String()
}

func (r *recordConn) Duplex() conn.Duplex {
	if r.c != nil {
		return r.c.Duplex()
	}
	return conn.DuplexUnknown
}

func (r *recordConn) Tx(w, read []byte) error {
	return r.r.txInternal(r.c, w, read)
}

// TxPackets is not yet implemented.
func (r *recordConn) TxPackets(p []spi.Packet) error {
	return conntest.Errorf("spitest: TxPackets is not implemented")
}

// CLK implements spi.Pins.
func (r *recordConn) CLK() gpio.PinOut {
	return r.r.CLK()
}

// MOSI implements spi.Pins.
func (r *recordConn) MOSI() gpio.PinOut {
	return r.r.MOSI()
}

// MISO implements spi.Pins.
func (r *recordConn) MISO() gpio.PinIn {
	return r.r.MISO()
}

// CS implements spi.Pins.
func (r *recordConn) CS() gpio.PinOut {
	return r.r.CS()
}

//

// Playback implements spi.PortCloser and plays back a recorded I/O flow.
//
// While "replay" type of unit tests are of limited value, they still present
// an easy way to do basic code coverage.
type Playback struct {
	conntest.Playback
	CLKPin      gpio.PinIO
	MOSIPin     gpio.PinIO
	MISOPin     gpio.PinIO
	CSPin       gpio.PinIO
	Initialized bool
}

// Close implements spi.PortCloser.
//
// Close() verifies that all the expected Ops have been consumed.
func (p *Playback) Close() error {
	return p.Playback.Close()
}

// LimitSpeed implements spi.PortCloser.
func (p *Playback) LimitSpeed(maxHz int64) error {
	return nil
}

// Connect implements spi.PortCloser.
func (p *Playback) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	if p.Initialized {
		return nil, conntest.Errorf("spitest: Connect cannot be called twice")
	}
	p.Initialized = true
	return &playbackConn{p}, nil
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

type playbackConn struct {
	p *Playback
}

func (p *playbackConn) String() string {
	return p.p.String()
}

func (p *playbackConn) Duplex() conn.Duplex {
	return p.p.Duplex()
}

func (p *playbackConn) Tx(w, r []byte) error {
	return p.p.Tx(w, r)
}

func (p *playbackConn) TxPackets(packets []spi.Packet) error {
	return conntest.Errorf("spitest: TxPackets is not implemented")
}

func (p *playbackConn) CLK() gpio.PinOut {
	return p.p.CLK()
}

func (p *playbackConn) MOSI() gpio.PinOut {
	return p.p.MOSI()
}

func (p *playbackConn) MISO() gpio.PinIn {
	return p.p.MISO()
}

func (p *playbackConn) CS() gpio.PinOut {
	return p.p.CS()
}

//

// Log logs all operations done on an spi.PortCloser.
type Log struct {
	spi.PortCloser
}

// Close implements spi.PortCloser.
func (l *Log) Close() error {
	err := l.PortCloser.Close()
	log.Printf("%s.Close() = %v", l.PortCloser, err)
	return err
}

// LimitSpeed implements spi.PortCloser.
func (l *Log) LimitSpeed(maxHz int64) error {
	err := l.PortCloser.LimitSpeed(maxHz)
	log.Printf("%s.LimitSpeed(%d) = %v", l.PortCloser, maxHz, err)
	return err
}

// Connect implements spi.PortCloser.
func (l *Log) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	c, err := l.PortCloser.Connect(maxHz, mode, bits)
	log.Printf("%s.Connect(%d, %d, %d) = %v", l.PortCloser, maxHz, mode, bits, err)
	return &LogConn{c}, err
}

//

// LogConn logs all operations done on an spi.Conn.
type LogConn struct {
	spi.Conn
}

// Tx implements spi.Conn.
func (l *LogConn) Tx(w, r []byte) error {
	err := l.Conn.Tx(w, r)
	log.Printf("%s.Tx(%#v, %#v) = %v", l.Conn, w, r, err)
	return err
}

// TxPackets is not yet implemented.
func (l *LogConn) TxPackets(p []spi.Packet) error {
	return conntest.Errorf("spitest: TxPackets is not implemented")
}

//

var _ spi.PortCloser = &RecordRaw{}
var _ spi.PortCloser = &Record{}
var _ spi.PortCloser = &Playback{}
var _ spi.PortCloser = &Log{}
var _ spi.Pins = &Record{}
var _ spi.Pins = &Playback{}
