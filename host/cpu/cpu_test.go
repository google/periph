// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cpu

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"periph.io/x/periph/host/fs"
)

func TestMaxSpeed_fail(t *testing.T) {
	defer reset()
	if s := MaxSpeed(); s != 0 {
		t.Fatal(s)
	}
}

func TestMaxSpeed(t *testing.T) {
	defer reset()
	openFile = func(path string, flag int) (io.ReadCloser, error) {
		if path != "/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq" {
			t.Fatal(path)
		}
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		return ioutil.NopCloser(bytes.NewBufferString("1001\n")), nil
	}
	MaxSpeed()
	if s := getMaxSpeedLinux(); s != 1001000 {
		t.Fatal(s)
	}
}

func TestNanospin(t *testing.T) {
	Nanospin(time.Microsecond)
	nanospinTime(time.Microsecond)
}

//

func init() {
	fs.Inhibit()
}

func reset() {
	openFile = openFileOrig
	maxSpeed = -1
}
