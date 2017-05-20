// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"io/ioutil"
	"path"
	"reflect"
	"strconv"
	"testing"
)

func TestNewOneWire(t *testing.T) {
	// Test no error on Linux
	_, err := NewOneWire()
	if isLinux {
		if err != nil {
			t.Fatal("Error not expected, Linux is supported")
		}
	}
}

func TestNewOneWireReturn(t *testing.T) {
	expected := "*sysfs.oneWire"
	ow, _ := NewOneWire()
	if reflect.TypeOf(ow).String() != expected {
		t.Fatal("Error. Expected a pointer to sysfs.oneWire")
	}
}

func TestScanWithoutLoadDriver(t *testing.T) {
	ow, err := newOneWire()
	_, err = ow.Scan("")
	if err == nil {
		t.Fatal("Error expected but not returned. Scan cannot proceed without kernel drivers loaded")
	}
}

func TestScanWithLoadDriver(t *testing.T) {
	if isLinux && busCheck() {
		ow, err := newOneWire()
		if err != nil {
			t.Error(err)
		}
		err = ow.LoadDrivers()
		if err != nil {
			t.Error(err)
		}
		_, err = ow.Scan("")
		if err != nil {
			t.Error(err)
		}
	}
}

func TestRead(t *testing.T) {
	if isLinux && busCheck() {
		devices := deviceCheck()
		if devices >= 1 {
			// Check existing of w1devices then read
			expected := "map[string]*sysfs.OneWireDevice"
			ow, err := newOneWire()
			if err != nil {
				t.Error(err)
			}
			err = ow.LoadDrivers()
			if err != nil {
				t.Error(err)
			}
			dev, err := ow.Scan("")
			if err == nil {
				if reflect.TypeOf(dev).String() != expected {
					t.Fatal("Error. Expected return a map with key string and val *sysfs.OneWireDevice{}")
				}
			} else {
				t.Fatal(err)
			}

			// Correct number of devices
			if len(dev) == devices {
				t.Fatalf("Error. %v devices expected but %v found", devices, len(dev))
			}
		}
	}
}

// busCheck is a helper to identify if w1 bus available for further tests
func busCheck() bool {
	pth := "/sys/bus/w1/devices/"
	_, err := ioutil.ReadDir(pth)
	if err != nil {
		return false
	}
	return true
}

// deviceCheck is a helper to identify if hardware is attached for further tests
func deviceCheck() int {
	pth := "/sys/bus/w1/devices/w1_bus_master1"
	cmd := "w1_master_slave_count"
	_, err := ioutil.ReadDir(pth)
	if err != nil {
		return 0
	}
	// read w1_master_slave_count
	f, err := ioutil.ReadFile(path.Join(pth, cmd))
	if err != nil {
		return 0
	}
	i, err := strconv.ParseInt(string(f[0]), 10, 32)
	if err != nil {
		return 0
	}
	// min of 1 device
	if i >= 1 {
		return int(i)
	}
	return 0
}
