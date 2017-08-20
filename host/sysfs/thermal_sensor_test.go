// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"testing"
	"time"
)

func TestThermalSensorByName(t *testing.T) {
	if d, err := ThermalSensorByName(""); d != nil || err == nil {
		t.Fatal("invalid bus")
	}
}

func TestThermalSensor(t *testing.T) {
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	if s := d.String(); s != "cpu" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
	if _, err := d.SenseContinuous(time.Second); err == nil {
		t.Fatal("not yet implemented")
	}
}
