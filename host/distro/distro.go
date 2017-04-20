// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package distro implements common functionality to auto-detect features on
// the host; generally about linux distributions.
//
// Most of the functions exported as in the form IsFoo() where Foo is a linux
// distribution.
package distro

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// IsArmbian returns true if running on a Armbian distribution.
//
// http://www.armbian.com/
func IsArmbian() bool {
	if isArm && isLinux {
		// Armbian presents itself as debian in /etc/os-release so OSRelease()
		// cannot be used..
		_, err := os.Stat("/etc/armbian.txt")
		return err == nil
	}
	return false
}

// IsDebian returns true if running on an Debian derived distribution.
//
// This function returns true on both Armbian, Raspbian and Ubuntu.
//
// https://debian.org/
func IsDebian() bool {
	if isLinux {
		// http://0pointer.de/public/systemd-man/os-release.html#ID_LIKE=
		if OSRelease()["ID"] == "debian" {
			return true
		}
		for _, part := range strings.Split(OSRelease()["ID_LIKE"], " ") {
			if part == "debian" {
				return true
			}
		}
	}
	return false
}

// IsRaspbian returns true if running on a Raspbian distribution.
//
// https://raspbian.org/
func IsRaspbian() bool {
	if isArm && isLinux {
		return OSRelease()["ID"] == "raspbian"
	}
	return false
}

// IsUbuntu returns true if running on an Ubuntu derived distribution.
//
// https://ubuntu.com/
func IsUbuntu() bool {
	if isLinux {
		return OSRelease()["ID"] == "ubuntu"
	}
	return false
}

// OSRelease returns parsed data from /etc/os-release.
//
// For more information, see
// http://0pointer.de/public/systemd-man/os-release.html
func OSRelease() map[string]string {
	if isLinux {
		return makeOSReleaseLinux()
	}
	return osRelease
}

// CPU

// CPUInfo returns parsed data from /proc/cpuinfo.
func CPUInfo() map[string]string {
	if isLinux {
		return makeCPUInfoLinux()
	}
	return cpuInfo
}

//

var (
	mu        sync.Mutex
	cpuInfo   map[string]string
	osRelease map[string]string
	readFile  = ioutil.ReadFile
)

func splitSemiColon(content string) map[string]string {
	// Strictly speaking this format isn't ok, there can be multiple group.
	out := map[string]string{}
	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		// This format may have space around the ':'.
		key := strings.TrimRightFunc(parts[0], unicode.IsSpace)
		if len(key) == 0 || key[0] == '#' {
			continue
		}
		// Ignore duplicate keys.
		// TODO(maruel): Keep them all.
		if _, ok := out[key]; !ok {
			// Trim on both side, trailing space was observed on "Features" value.
			out[key] = strings.TrimFunc(parts[1], unicode.IsSpace)
		}
	}
	return out
}

func splitStrict(content string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		if len(key) == 0 || key[0] == '#' {
			continue
		}
		// Overwrite previous key.
		value := parts[1]
		if len(value) > 2 && value[0] == '"' && value[len(value)-1] == '"' {
			// Not exactly 100% right but #closeenough. See for more details
			// https://www.freedesktop.org/software/systemd/man/os-release.html
			var err error
			value, err = strconv.Unquote(value)
			if err != nil {
				continue
			}
		}
		out[key] = value
	}
	return out
}

// splitNull returns the null-terminated strings in the data
func splitNull(data []byte) []string {
	ss := strings.Split(string(data), "\x00")
	// The last string is typically null-terminated, so remove empty string
	// from end of array.
	if len(ss) > 0 && len(ss[len(ss)-1]) == 0 {
		ss = ss[:len(ss)-1]
	}
	return ss
}

func makeCPUInfoLinux() map[string]string {
	mu.Lock()
	defer mu.Unlock()
	if cpuInfo == nil {
		cpuInfo = map[string]string{}
		if bytes, err := readFile("/proc/cpuinfo"); err == nil {
			cpuInfo = splitSemiColon(string(bytes))
		}
	}
	return cpuInfo
}

func makeOSReleaseLinux() map[string]string {
	mu.Lock()
	defer mu.Unlock()
	if osRelease == nil {
		// This file may not exist on older distros. Send a PR if you want to have
		// a specific fallback.
		osRelease = map[string]string{}
		if bytes, err := readFile("/etc/os-release"); err == nil {
			osRelease = splitStrict(string(bytes))
		}
	}
	return osRelease
}
