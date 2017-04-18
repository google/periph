// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package distro

import (
	"errors"
	"reflect"
	"testing"

	"periph.io/x/periph/host/fs"
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
# foo = bar
FOO="aa""
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

	data = []byte("")
	expected = []string{}
	if actual := splitNull(data); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%# v != %# v", expected, actual)
	}
}

func TestIsArmbian(t *testing.T) {
	// At least ensure it doesn't crash.
	IsArmbian()
}

func TestIsDebian(t *testing.T) {
	// At least ensure it doesn't crash.
	IsDebian()
}

func TestIsRaspbian(t *testing.T) {
	// At least ensure it doesn't crash.
	IsRaspbian()
}

func TestIsUbuntu(t *testing.T) {
	// At least ensure it doesn't crash.
	IsUbuntu()
}

func TestCPUInfo_fail(t *testing.T) {
	defer reset()
	if c := CPUInfo(); len(c) != 0 {
		t.Fatal(c)
	}
}

func TestCPUInfo(t *testing.T) {
	defer reset()
	readFile = func(filename string) ([]byte, error) {
		if filename != "/proc/cpuinfo" {
			t.Fatal(filename)
		}
		return []byte("Processor	: AArch64\n"), nil
	}
	CPUInfo()
	expected := map[string]string{"Processor": "AArch64"}
	if c := makeCPUInfoLinux(); !reflect.DeepEqual(c, expected) {
		t.Fatal(c)
	}
}

func TestOSRelease_fail(t *testing.T) {
	defer reset()
	if c := OSRelease(); len(c) != 0 {
		t.Fatal(c)
	}
}

func TestOSRelease(t *testing.T) {
	defer reset()
	readFile = func(filename string) ([]byte, error) {
		if filename != "/etc/os-release" {
			t.Fatal(filename)
		}
		return []byte("VERSION_ID=8\n"), nil
	}
	OSRelease()
	expected := map[string]string{"VERSION_ID": "8"}
	if c := makeOSReleaseLinux(); !reflect.DeepEqual(c, expected) {
		t.Fatal(c)
	}
}

//

func init() {
	fs.Inhibit()
	reset()
}

func reset() {
	cpuInfo = nil
	dtCompatible = nil
	dtModel = ""
	osRelease = nil
	readFile = func(filename string) ([]byte, error) {
		return nil, errors.New("no file can be opened in unit tests")
	}
}
