// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"os"
	"testing"
)

func TestAlloc_fail(t *testing.T) {
	defer reset()
	if m, err := Alloc(0); m != nil || err == nil {
		t.Fatal("0 bytes")
	}
	if m, err := Alloc(1); m != nil || err == nil {
		t.Fatal("not 4096 bytes")
	}
	if m, err := Alloc(4096); m != nil || err == nil {
		t.Fatal("ReadPageMap() fails")
	}
}

func TestAlloLinuxc_fail(t *testing.T) {
	if m, err := allocLinux(8192); m != nil || err == nil {
		t.Fatal("only 4096 is supported")
	}
}

func TestAlloc_no_high_bit(t *testing.T) {
	defer reset()
	openFile = func(path string, flag int) (fileIO, error) {
		if path != "/proc/self/pagemap" {
			t.Fatal(path)
		}
		if flag != os.O_RDONLY|os.O_SYNC {
			t.Fatal(flag)
		}
		return &simpleFile{data: []byte{1, 2, 3, 4, 5, 6, 7, 8}}, nil
	}
	if m, err := allocLinux(4096); m != nil || err == nil {
		t.Fatal("page not physically mapped")
	}
}

func TestAlloc(t *testing.T) {
	defer reset()
	openFile = func(path string, flag int) (fileIO, error) {
		if path != "/proc/self/pagemap" {
			t.Fatal(path)
		}
		if flag != os.O_RDONLY|os.O_SYNC {
			t.Fatal(flag)
		}
		return &simpleFile{data: []byte{1, 2, 3, 4, 5, 6, 7, 0x80}}, nil
	}
	_, err := allocLinux(4096)
	if isLinux && !isWSL() {
		if err != nil {
			t.Fatal(err)
		}
	} else {
		if err == nil {
			t.Fatal("syscall.Mlock() is not implemented")
		}
	}
}
