// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

import (
	"reflect"
	"testing"
)

func TestSplitSemiColon(t *testing.T) {
	data := `/proc/cpuinfo
Processor	: AArch64 Processor rev 4 (aarch64)
processor	: 0
processor	: 1
Features	: fp asimd aes pmull sha1 sha2 crc32 
CPU architecture: AArch64
CPU part	: 0xd03

Hardware	: sun50iw1p1
# foo : bar
`
	expected := map[string]string{
		"CPU architecture": "AArch64",
		"CPU part":         "0xd03",
		"Features":         "fp asimd aes pmull sha1 sha2 crc32",
		"Hardware":         "sun50iw1p1",
		"Processor":        "AArch64 Processor rev 4 (aarch64)",
		"processor":        "0",
	}
	if actual := splitSemiColon(data); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%# v != %# v", expected, actual)
	}
}

func TestSplitStrict(t *testing.T) {
	data := `PRETTY_NAME="Raspbian GNU/Linux 8 (jessie)"
VERSION_ID="8"
VERSION="8 (jessie)"
ID_LIKE=debian
HOME_URL="http://www.raspbian.org/"
# foo : bar
`
	expected := map[string]string{
		"HOME_URL":    "http://www.raspbian.org/",
		"ID_LIKE":     "debian",
		"PRETTY_NAME": "Raspbian GNU/Linux 8 (jessie)",
		"VERSION":     "8 (jessie)",
		"VERSION_ID":  "8",
	}
	if actual := splitStrict(data); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%# v != %# v", expected, actual)
	}
}

func TestSplitNull(t *testing.T) {
	data := []byte("line 1\x00line 2\x00line 3\x00")
	expected := []string{"line 1", "line 2", "line 3"}
	if actual := splitNull(data); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%# v != %# v", expected, actual)
	}
}
