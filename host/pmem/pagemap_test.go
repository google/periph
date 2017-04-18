// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pmem

import (
	"os"
	"testing"
)

func TestReadPageMap_fail(t *testing.T) {
	defer reset()
	if u, err := ReadPageMap(8192); u != 0 || err == nil {
		t.Fatal("can't open file")
	}
}

func TestReadPageMap(t *testing.T) {
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
	u, err := readPageMapLinux(8192)
	if err != nil {
		t.Fatal(err)
	}
	if u != 0x807060504030201 {
		t.Fatal(u)
	}
}

func TestReadPageMap_short(t *testing.T) {
	defer reset()
	openFile = func(path string, flag int) (fileIO, error) {
		if path != "/proc/self/pagemap" {
			t.Fatal(path)
		}
		if flag != os.O_RDONLY|os.O_SYNC {
			t.Fatal(flag)
		}
		return &simpleFile{data: []byte{1, 2}}, nil
	}
	if u, err := readPageMapLinux(8192); u != 0 || err == nil {
		t.Fatal("didn't read 8 bytes")
	}
}

func TestReadPageMap_read_fail(t *testing.T) {
	defer reset()
	openFile = func(path string, flag int) (fileIO, error) {
		if path != "/proc/self/pagemap" {
			t.Fatal(path)
		}
		if flag != os.O_RDONLY|os.O_SYNC {
			t.Fatal(flag)
		}
		return &simpleFile{}, nil
	}
	if u, err := readPageMapLinux(8192); u != 0 || err == nil {
		t.Fatal("Read() failed")
	}
}

func TestReadPageMap_seek_fail(t *testing.T) {
	defer reset()
	openFile = func(path string, flag int) (fileIO, error) {
		if path != "/proc/self/pagemap" {
			t.Fatal(path)
		}
		if flag != os.O_RDONLY|os.O_SYNC {
			t.Fatal(flag)
		}
		return &failFile{}, nil
	}
	if u, err := readPageMapLinux(8192); u != 0 || err == nil {
		t.Fatal("Seek() failed")
	}
}
