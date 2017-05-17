// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

// OneWire is an open OneWire bus
type oneWire struct {
	path         string
	modProbeCmd  string
	thermMod     string
	gpioMod      string
	masterPrefix string
}

// OneWireDevice represents a single OneWire device
type OneWireDevice struct {
	mtx sync.Mutex
	f   *os.File
}

// NewOneWire provides access to OneWire bus on linux devices
func NewOneWire() (*oneWire, error) {
	if isLinux {
		return newOneWire(config)
	}
	return nil, errors.New("sysfs-onewire: not implemented on non-linux OSes")
}

func newOneWire() (*oneWire, error) {
	ow := oneWire{
		path:         "/sys/bus/w1/devices/",
		modProbeCmd:  "/sbin/modprobe",
		thermMod:     "w1-therm",
		gpioMod:      "w1-gpio",
		masterPrefix: "w1_bus_master",
	}

	// Check system requirements satisfied
	err := ow.check()
	if err != nil {
		return &ow, err
	}
	return &ow, nil
}

// Read returns the contents of a OneWire device file as a Reader
// Assumption is that specific device abstractions will do what they
// need to with the data
func (owd *OneWireDevice) Read() (*bufio.Reader, error) {
	var reading *bufio.Reader
	owd.mtx.Lock()
	defer owd.mtx.Unlock()
	if owd.f == nil {
		return reading, errors.New("sysfs-onewire: device file handle closed")
	}
	if _, err := owd.f.Seek(0, 0); err != nil {
		return reading, fmt.Errorf("sysfs-onewire: %v", err)
	}
	reading = bufio.NewReader(owd.f)
	return reading, nil
}

// Checks system requirements are satisfied
func (ow *oneWire) check() error {
	// Check modules available
	mod := exec.Command(ow.modProbeCmd, ow.gpioMod)
	if err := mod.Run(); err != nil {
		return fmt.Errorf("sysfs-onewire: %v module not found", ow.gpioMod)
	}
	mod = exec.Command(ow.modProbeCmd, ow.thermMod)
	if err := mod.Run(); err != nil {
		return fmt.Errorf("sysfs-onewire: %v module not found", ow.thermMod)
	}

	// Check for OneWire bus
	bus, err := ioutil.ReadDir(ow.path)
	if err != nil {
		return fmt.Errorf("sysfs-onewire: %v path not found", ow.path)
	}
	master := false
	for i := range bus {
		if strings.HasPrefix(bus[i].Name(), ow.masterPrefix) {
			master = true
			break
		}
	}
	if !master {
		return errors.New("sysfs-onewire: onewire master bus not found")
	}
	return nil
}

// Scan returns map of discovered OneWire devices, filtered by prefix if required
func (ow *oneWire) Scan(prefix string) (map[string]*OneWireDevice, error) {
	var devices map[string]*OneWireDevice
	devices = make(map[string]*OneWireDevice)

	files, err := ioutil.ReadDir(ow.path)
	if err != nil {
		return devices, fmt.Errorf("sysfs-onewire: %v", err)
	}
	if prefix != "" {
		for i := range files {
			if strings.HasPrefix(files[i].Name(), prefix) {
				if (files[i].Mode() & os.ModeSymlink) == os.ModeSymlink {
					f, err := os.Open(path.Join(path.Join(ow.path, files[i].Name()), "w1_slave"))
					if err != nil {
						return nil, fmt.Errorf("sysfs-onewire: %v", err)
					}
					device := &OneWireDevice{
						mtx: sync.Mutex{},
						f:   f,
					}
					devices[files[i].Name()] = device
				}
			}
		}
	} else {
		for i := range files {
			if strings.Index(files[i].Name(), "-") == 2 {
				if (files[i].Mode() & os.ModeSymlink) == os.ModeSymlink {
					f, err := os.Open(path.Join(path.Join(ow.path, files[i].Name()), "w1_slave"))
					if err != nil {
						return nil, fmt.Errorf("sysfs-onewire: %v", err)
					}
					device := &OneWireDevice{
						mtx: sync.Mutex{},
						f:   f,
					}
					devices[files[i].Name()] = device
				}
			}
		}
	}
	// Check devices
	if len(devices) == 0 {
		return devices, errors.New("sysfs-onewire: no onewire devices found")
	}
	return devices, nil
}
