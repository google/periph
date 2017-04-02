// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package conntest implements fakes for package conn.
package conntest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"periph.io/x/periph/conn"
)

// RecordRaw implements conn.Conn. It sends everything written to it to W.
type RecordRaw struct {
	sync.Mutex
	W io.Writer
}

func (r *RecordRaw) String() string {
	return "recordraw"
}

// Tx implements conn.Conn.
func (r *RecordRaw) Tx(w, read []byte) error {
	if len(read) != 0 {
		return errors.New("conntest: not implemented")
	}
	_, err := r.W.Write(w)
	return err
}

// Duplex implements conn.Conn.
func (r *RecordRaw) Duplex() conn.Duplex {
	return conn.Half
}

// IO registers the I/O that happened on either a real or fake connection.
type IO struct {
	Write []byte
	Read  []byte
}

// Record implements conn.Conn that records everything written to it.
//
// This can then be used to feed to Playback to do "replay" based unit tests.
type Record struct {
	sync.Mutex
	Conn conn.Conn // Conn can be nil if only writes are being recorded.
	Ops  []IO
}

func (r *Record) String() string {
	return "record"
}

// Tx implements conn.Conn.
func (r *Record) Tx(w, read []byte) error {
	r.Lock()
	defer r.Unlock()
	if r.Conn == nil {
		if len(read) != 0 {
			return errors.New("conntest: read unsupported when no bus is connected")
		}
	} else {
		if err := r.Conn.Tx(w, read); err != nil {
			return err
		}
	}
	io := IO{Write: make([]byte, len(w))}
	if len(read) != 0 {
		io.Read = make([]byte, len(read))
	}
	copy(io.Write, w)
	copy(io.Read, read)
	r.Ops = append(r.Ops, io)
	return nil
}

// Duplex implements conn.Conn.
func (r *Record) Duplex() conn.Duplex {
	if r.Conn != nil {
		return r.Conn.Duplex()
	}
	return conn.DuplexUnknown
}

// Playback implements conn.Conn and plays back a recorded I/O flow.
//
// While "replay" type of unit tests are of limited value, they still present
// an easy way to do basic code coverage.
type Playback struct {
	sync.Mutex
	Ops   []IO
	D     conn.Duplex
	Count int
}

func (p *Playback) String() string {
	return "playback"
}

// Close verifies that all the expected Ops have been consumed.
func (p *Playback) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != p.Count {
		return fmt.Errorf("conntest: expected playback to be empty: I/O count %d; expected %d", p.Count, len(p.Ops))
	}
	return nil
}

// Tx implements conn.Conn.
func (p *Playback) Tx(w, r []byte) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) <= p.Count {
		return fmt.Errorf("conntest: unexpected Tx() (count #%d)", p.Count)
	}
	if !bytes.Equal(p.Ops[p.Count].Write, w) {
		return fmt.Errorf("conntest: unexpected write (count #%d) %#v != %#v", p.Count, w, p.Ops[p.Count].Write)
	}
	if len(p.Ops[p.Count].Read) != len(r) {
		return fmt.Errorf("conntest: unexpected read buffer length (count #%d) %d != %d", p.Count, len(r), len(p.Ops[p.Count].Read))
	}
	copy(r, p.Ops[p.Count].Read)
	p.Count++
	return nil
}

// Duplex implements conn.Conn.
func (p *Playback) Duplex() conn.Duplex {
	p.Lock()
	defer p.Unlock()
	return p.D
}

// Discard implements conn.Conn and discards all writes and reads zeros. It
// never fails.
type Discard struct {
	D conn.Duplex
}

func (d *Discard) String() string {
	return "discard"
}

// Tx implements conn.Conn.
func (d *Discard) Tx(w, r []byte) error {
	for i := range r {
		r[i] = 0
	}
	return nil
}

// Duplex implements conn.Conn.
func (d *Discard) Duplex() conn.Duplex {
	return d.D
}

//

var _ conn.Conn = &RecordRaw{}
var _ conn.Conn = &Record{}
var _ conn.Conn = &Playback{}
