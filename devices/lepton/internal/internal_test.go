// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"bytes"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	if d := DurationMS(100).ToD(); d != 100*time.Millisecond {
		t.Fatal(d)
	}
}

func TestCentiK(t *testing.T) {
	if c := CentiK(100).ToC(); c != -272150 {
		t.Fatal(c)
	}
}

func TestCRC16(t *testing.T) {
	d := []byte{0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39}
	if c := CRC16(d); c != 0x31c3 {
		t.Fatal(c)
	}
}

func TestBig16(t *testing.T) {
	if s := Big16.String(); s != "big16" {
		t.Fatal(s)
	}
	d := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if v := Big16.Uint16(d); v != 0x0102 {
		t.Fatal(v)
	}
	if v := Big16.Uint32(d); v != 0x03040102 {
		t.Fatal(v)
	}
	if v := Big16.Uint64(d); v != 0x0708050603040102 {
		t.Fatal(v)
	}

	d = make([]byte, 2)
	Big16.PutUint16(d, 0x0102)
	if !bytes.Equal(d, []byte{0x01, 0x02}) {
		t.Fatal(d)
	}
	d = make([]byte, 4)
	Big16.PutUint32(d, 0x01020304)
	if !bytes.Equal(d, []byte{0x03, 0x04, 0x01, 0x02}) {
		t.Fatal(d)
	}
	d = make([]byte, 8)
	Big16.PutUint64(d, 0x0102030405060708)
	if !bytes.Equal(d, []byte{0x07, 0x08, 0x05, 0x06, 0x03, 0x04, 0x01, 0x02}) {
		t.Fatal(d)
	}
}
