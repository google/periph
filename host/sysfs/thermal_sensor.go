// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"periph.io/x/periph"
	"periph.io/x/periph/devices"
)

// ThermalSensors is all the sensors discovered on this host via sysfs.
var ThermalSensors []*ThermalSensor

// ThermalSensorByName returns a *ThermalSensor for the sensor name, if any.
func ThermalSensorByName(name string) (*ThermalSensor, error) {
	// TODO(maruel): Use a bisect or a map. For now we don't expect more than a
	// handful of thermal sensors so it doesn't matter.
	for _, t := range ThermalSensors {
		if t.name == name {
			if err := t.open(); err != nil {
				return nil, err
			}
			return t, nil
		}
	}
	return nil, errors.New("sysfs-thermal: invalid sensor name")
}

// ThermalSensor represents one thermal sensor on the system.
type ThermalSensor struct {
	name string
	root string

	mu       sync.Mutex
	nameType string
	fTemp    *os.File
}

func (t *ThermalSensor) String() string {
	return t.name
}

// Type returns the type of sensor as exported by sysfs.
func (t *ThermalSensor) Type() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.nameType == "" {
		f, err := os.OpenFile(t.root+"type", os.O_RDONLY, 0600)
		if err != nil {
			return err.Error()
		}
		defer f.Close()
		var buf [256]byte
		n, err := f.Read(buf[:])
		if err != nil {
			return err.Error()
		}
		if n < 2 {
			t.nameType = "<unknown>"
		} else {
			t.nameType = string(buf[:n-1])
		}
	}
	return t.nameType
}

// Sense implements devices.Environmental.
func (t *ThermalSensor) Sense(env *devices.Environment) error {
	if err := t.open(); err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, err := t.fTemp.Seek(0, 0); err != nil {
		return err
	}
	var buf [64]byte
	n, err := t.fTemp.Read(buf[:])
	if err != nil {
		return err
	}
	if n < 2 {
		return errors.New("sysfs-thermal: failed to read temperature")
	}
	i, err := strconv.Atoi(string(buf[:n-1]))
	if err != nil {
		return err
	}
	if i < 100 {
		i *= 1000
	}
	env.Temperature = devices.Celsius(i)
	return nil
}

//

func (t *ThermalSensor) open() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	var err error
	if t.fTemp == nil {
		t.fTemp, err = os.OpenFile(t.root+"temp", os.O_RDONLY, 0600)
	}
	return err
}

// driverThermalSensor implements periph.Driver.
type driverThermalSensor struct {
}

func (d *driverThermalSensor) String() string {
	return "sysfs-thermal"
}

func (d *driverThermalSensor) Prerequisites() []string {
	return nil
}

// Init initializes thermal sysfs handling code.
//
// Uses sysfs as described* at
// https://www.kernel.org/doc/Documentation/thermal/sysfs-api.txt
//
// * for the most minimalistic meaning of 'described'.
func (d *driverThermalSensor) Init() (bool, error) {
	// This driver is only registered on linux, so there is no legitimate time to
	// skip it.
	items, err := filepath.Glob("/sys/class/thermal/*/temp")
	if err != nil {
		return true, err
	}
	if len(items) == 0 {
		return false, errors.New("sysfs-thermal: no sensor found")
	}
	sort.Strings(items)
	for _, item := range items {
		base := filepath.Dir(item)
		ThermalSensors = append(ThermalSensors, &ThermalSensor{
			name: filepath.Base(base),
			root: base + "/",
		})
	}
	return true, nil
}

func init() {
	if isLinux {
		periph.MustRegister(&driverThermalSensor{})
	}
}

var _ devices.Environmental = &ThermalSensor{}
