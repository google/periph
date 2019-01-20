// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fs

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"
)

func TestEpollEvent_String(t *testing.T) {
	if s := (epollIN | epollOUT).String(); s != "IN|OUT" {
		t.Fatal(s)
	}
	if s := (epollERR | epollEvent(0x1000)).String(); s != "ERR|0x1000" {
		t.Fatal(s)
	}
	if s := epollEvent(0).String(); s != "0" {
		t.Fatal(s)
	}
}

func TestAddFd_Zero(t *testing.T) {
	// We assume this is a bad file descriptor.
	ev := getListener(t)

	const flags = epollET | epollPRI
	if err := ev.addFd(0xFFFFFFFF, make(chan time.Time), flags); err == nil || err.Error() != "bad file descriptor" {
		t.Fatal("expected failure", err)
	}
}

func TestAddFd_File(t *testing.T) {
	// listen cannot listen to a file.
	ev := getListener(t)

	f, err := ioutil.TempFile("", "periph_fs")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Fatal(err)
		}
	}()

	const flags = epollET | epollPRI
	if err := ev.addFd(f.Fd(), make(chan time.Time), flags); err == nil || err.Error() != "operation not permitted" {
		t.Fatal("expected failure", err)
	}
}

func TestListen_Pipe(t *testing.T) {
	start := time.Now()
	ev := getListener(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	c := make(chan time.Time)
	// Pipes do not support epollPRI, so use epollIN instead.
	const flags = epollET | epollIN
	if err := ev.addFd(r.Fd(), c, flags); err != nil {
		t.Fatal(err)
	}

	// Produce a single event.
	if _, err := w.Write([]byte("foo")); err != nil {
		t.Fatal(err)
	}
	expectChan(t, c, start)
	notExpectChan(t, c, "should have produced a single event")

	// Produce one or two events.
	if _, err := w.Write([]byte("bar")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("baz")); err != nil {
		t.Fatal(err)
	}
	expectChan(t, c, start)
	// It's a race condition between EpollWait() and reading back from the
	// channel.
	select {
	case <-c:
	default:
	}

	if err := ev.removeFd(r.Fd()); err != nil {
		t.Fatal(err)
	}
}

func TestListen_Socket(t *testing.T) {
	start := time.Now()
	ev := getListener(t)

	ln, err := net.ListenTCP("tcp4", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ln.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	conn, err := net.DialTCP("tcp4", nil, ln.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	recv, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := recv.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	f, err := recv.(*net.TCPConn).File()
	if err != nil {
		t.Fatal(err)
	}

	// This channel needs to be buffered since there's going to be an even
	// immediately triggered.
	c := make(chan time.Time, 1)
	// TODO(maruel): Sockets do support epollPRI on out-of-band data. This would
	// make this test a bit more similar to testing a GPIO sysfs file descriptor.
	const flags = epollET | epollIN
	if err := ev.addFd(f.Fd(), c, flags); err != nil {
		t.Fatal(err)
	}
	notExpectChan(t, c, "starting should not produce an event")

	// Produce one or two events.
	// It's a race condition between EpollWait() and reading back from the
	// channel.
	if _, err := conn.Write([]byte("bar\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Write([]byte("baz\n")); err != nil {
		t.Fatal(err)
	}
	expectChan(t, c, start)
	// It's a race condition between EpollWait() and reading back from the
	// channel.
	select {
	case <-c:
	default:
	}

	// Empty the buffer.
	var buf [16]byte
	expected := "bar\nbaz\n"
	if n, err := recv.Read(buf[:]); n != len(expected) || err != nil {
		t.Fatal(n, err)
	}
	if s := string(buf[:len(expected)]); s != expected {
		t.Fatal(s)
	}

	// Produce one event.
	if _, err := conn.Write([]byte("foo\n")); err != nil {
		t.Fatal(err)
	}
	expectChan(t, c, start)
	// This is part of https://github.com/google/periph/issues/323
	//notExpectChan(t, c, "should have produced a single event")
	// Instead consume any extraneous event.
	select {
	case <-c:
	default:
	}

	if err := ev.removeFd(f.Fd()); err != nil {
		t.Fatal(err)
	}
}

//

// getListener returns a preinitialized eventsListener.
//
// Note: This object creates a goroutine once initialized that will leak.
func getListener(t *testing.T) *eventsListener {
	ev := &eventsListener{}
	if err := ev.init(); err != nil {
		t.Fatal(err)
	}
	return ev
}

func expectChan(t *testing.T, c <-chan time.Time, start time.Time) {
	select {
	case v := <-c:
		if v.Before(start) {
			t.Fatal("received an timestamp that was too early", v, start)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out after 5 seconds, waiting for an event")
	}
}

func notExpectChan(t *testing.T, c <-chan time.Time, errmsg string) {
	select {
	case <-c:
		t.Fatal(errmsg)
	default:
	}
}
