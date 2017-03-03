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

// Write implements conn.Conn.
func (r *RecordRaw) Write(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()
	return r.W.Write(b)
}

// Tx implements conn.Conn.
func (r *RecordRaw) Tx(w, read []byte) error {
	if len(read) != 0 {
		return errors.New("conntest: not implemented")
	}
	_, err := r.Write(w)
	return err
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

// Write implements conn.Conn.
func (r *Record) Write(d []byte) (int, error) {
	if err := r.Tx(d, nil); err != nil {
		return 0, err
	}
	return len(d), nil
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

// Playback implements conn.Conn and plays back a recorded I/O flow.
//
// While "replay" type of unit tests are of limited value, they still present
// an easy way to do basic code coverage.
type Playback struct {
	sync.Mutex
	Ops []IO
}

func (p *Playback) String() string {
	return "playback"
}

// Write implements conn.Conn.
func (p *Playback) Write(d []byte) (int, error) {
	if err := p.Tx(d, nil); err != nil {
		return 0, err
	}
	return len(d), nil
}

// Tx implements conn.Conn.
func (p *Playback) Tx(w, r []byte) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) == 0 {
		return errors.New("conntest: unexpected Tx()")
	}
	if !bytes.Equal(p.Ops[0].Write, w) {
		return fmt.Errorf("conntest: unexpected write %#v != %#v", w, p.Ops[0].Write)
	}
	if len(p.Ops[0].Read) != len(r) {
		return fmt.Errorf("conntest: unexpected read buffer length %d != %d", len(r), len(p.Ops[0].Read))
	}
	copy(r, p.Ops[0].Read)
	p.Ops = p.Ops[1:]
	return nil
}

var _ conn.Conn = &RecordRaw{}
var _ conn.Conn = &Record{}
var _ conn.Conn = &Playback{}
