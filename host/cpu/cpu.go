// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cpu

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"periph.io/x/periph/host/fs"
)

// MaxSpeed returns the processor maximum speed in Hz.
//
// Returns 0 if it couldn't be calculated.
func MaxSpeed() int64 {
	if isLinux {
		return getMaxSpeedLinux()
	}
	return 0
}

// Nanospin spins for a short amount of time doing a busy loop.
//
// This function should be called with durations of 10µs or less.
func Nanospin(d time.Duration) {
	// TODO(maruel): Use runtime.LockOSThread()?
	if isLinux {
		nanospinLinux(d)
	} else {
		nanospinTime(d)
	}
}

//

var (
	mu       sync.Mutex
	maxSpeed int64 = -1
	openFile       = openFileOrig
)

func openFileOrig(path string, flag int) (io.ReadCloser, error) {
	f, err := fs.Open(path, flag)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func getMaxSpeedLinux() int64 {
	mu.Lock()
	defer mu.Unlock()
	if maxSpeed == -1 {
		maxSpeed = 0
		if f, err := openFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq", os.O_RDONLY); err == nil {
			defer f.Close()
			if b, err := ioutil.ReadAll(f); err == nil {
				s := strings.TrimSpace(string(b))
				if i, err := strconv.ParseInt(s, 10, 64); err == nil {
					// Weirdly, the speed is listed as khz. :(
					maxSpeed = i * 1000
				}
			}
		}
	}
	return maxSpeed
}

func nanospinTime(d time.Duration) {
	// TODO(maruel): That's not optimal; it's actually pretty bad.
	// time.Sleep() sleeps for really too long, calling it repeatedly with
	// minimal value will give the caller a wake rate of 5KHz or so, depending on
	// the host. This makes it useless for bitbanging protocol implementations.
	for start := time.Now(); time.Since(start) < d; {
	}
}
