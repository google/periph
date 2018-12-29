// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/physic"
)

// ThermalSensors is all the sensors discovered on this host via sysfs.  It
// includes 'thermal' devices as well as temperature 'hwmon' devices, so
// pre-configured onewire temperature sensors will be discovered automatically.
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
	name           string
	root           string
	sensorFilename string
	typeFilename   string

	mu        sync.Mutex
	nameType  string
	f         fileIO
	precision physic.Temperature

	done chan struct{}
}

func (t *ThermalSensor) String() string {
	return t.name
}

// Halt stops a continuous sense that was started with SenseContinuous.
func (t *ThermalSensor) Halt() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done != nil {
		close(t.done)
		t.done = nil
	}
	return nil
}

// Type returns the type of sensor as exported by sysfs.
func (t *ThermalSensor) Type() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.nameType == "" {
		nameType, err := t.readType()
		if err != nil {
			return err.Error()
		}
		t.nameType = nameType
	}
	return t.nameType
}

func (t *ThermalSensor) readType() (string, error) {
	f, err := fileIOOpen(t.root+t.typeFilename, os.O_RDONLY)
	if os.IsNotExist(err) {
		return "<unknown>", nil
	}
	if err != nil {
		return "", fmt.Errorf("sysfs-thermal: %v", err)
	}
	defer f.Close()
	var buf [256]byte
	n, err := f.Read(buf[:])
	if err != nil {
		return "", fmt.Errorf("sysfs-thermal: %v", err)
	}
	if n < 2 {
		return "<unknown>", nil
	}
	return string(buf[:n-1]), nil
}

// Sense implements physic.SenseEnv.
func (t *ThermalSensor) Sense(e *physic.Env) error {
	if err := t.open(); err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	var buf [24]byte
	n, err := seekRead(t.f, buf[:])
	if err != nil {
		return fmt.Errorf("sysfs-thermal: %v", err)
	}
	if n < 2 {
		return errors.New("sysfs-thermal: failed to read temperature")
	}
	i, err := strconv.Atoi(string(buf[:n-1]))
	if err != nil {
		return fmt.Errorf("sysfs-thermal: %v", err)
	}
	if t.precision == 0 {
		t.precision = physic.MilliKelvin
		if i < 100 {
			t.precision *= 1000
		}
	}
	e.Temperature = physic.Temperature(i)*t.precision + physic.ZeroCelsius
	return nil
}

// SenseContinuous implements physic.SenseEnv.
func (t *ThermalSensor) SenseContinuous(interval time.Duration) (<-chan physic.Env, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.done != nil {
		return nil, nil
	}
	done := make(chan struct{})
	ret := make(chan physic.Env)
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-done:
				close(ret)
				return
			case <-ticker.C:
				var e physic.Env
				if err := t.Sense(&e); err == nil {
					ret <- e
				}
			}
		}
	}()

	t.done = done
	return ret, nil
}

// Precision implements physic.SenseEnv.
func (t *ThermalSensor) Precision(e *physic.Env) {
	if t.precision == 0 {
		dummy := physic.Env{}
		// Ignore the error.
		_ = t.Sense(&dummy)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e.Temperature = t.precision
}

//

func (t *ThermalSensor) open() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.f != nil {
		return nil
	}
	f, err := fileIOOpen(t.root+t.sensorFilename, os.O_RDONLY)
	if err != nil {
		return fmt.Errorf("sysfs-thermal: %v", err)
	}
	t.f = f
	return nil
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

func (d *driverThermalSensor) After() []string {
	return nil
}

// Init initializes thermal sysfs handling code.
//
// Uses sysfs as described* at
// https://www.kernel.org/doc/Documentation/thermal/sysfs-api.txt
//
// * for the most minimalistic meaning of 'described'.
func (d *driverThermalSensor) Init() (bool, error) {
	if err := d.discoverDevices("/sys/class/thermal/*/temp", "type"); err != nil {
		return true, err
	}
	if err := d.discoverDevices("/sys/class/hwmon/*/temp*_input", "device/name"); err != nil {
		return true, err
	}
	if len(ThermalSensors) == 0 {
		return false, errors.New("sysfs-thermal: no sensor found")
	}
	return true, nil
}

func (d *driverThermalSensor) discoverDevices(glob, typeFilename string) error {
	// This driver is only registered on linux, so there is no legitimate time to
	// skip it.
	items, err := filepath.Glob(glob)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	sort.Strings(items)
	for _, item := range items {
		base := filepath.Dir(item)
		ThermalSensors = append(ThermalSensors, &ThermalSensor{
			name:           filepath.Base(base),
			root:           base + "/",
			sensorFilename: filepath.Base(item),
			typeFilename:   typeFilename,
		})
	}
	return nil
}

func init() {
	if isLinux {
		periph.MustRegister(&drvThermalSensor)
	}
}

var drvThermalSensor driverThermalSensor

var _ conn.Resource = &ThermalSensor{}
var _ physic.SenseEnv = &ThermalSensor{}
