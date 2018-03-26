// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"os"
	"testing"
	"time"

	"periph.io/x/periph/conn/physic"
)

func TestThermalSensorByName_not_present(t *testing.T) {
	defer resetThermal()
	ThermalSensors = []*ThermalSensor{
		{name: "cpu", root: "//\000/"},
	}
	if _, err := ThermalSensorByName("missing"); err == nil || err.Error() != "sysfs-thermal: invalid sensor name" {
		t.Fatal("missing")
	}
}

func TestThermalSensorByName_cant_open(t *testing.T) {
	defer resetThermal()
	ThermalSensors = []*ThermalSensor{{name: "cpu", root: "//\000/"}}
	if _, err := ThermalSensorByName("cpu"); err == nil || err.Error() != "sysfs-thermal: file I/O is inhibited" {
		t.Fatal("expected failure")
	}
}

func TestThermalSensorByName_success(t *testing.T) {
	defer resetThermal()
	ThermalSensors = []*ThermalSensor{{name: "cpu", root: "//\000/", f: &file{}}}
	if _, err := ThermalSensorByName("cpu"); err != nil {
		t.Fatal(err)
	}
}

func TestThermalSensor_fail(t *testing.T) {
	defer resetThermal()
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	if s := d.String(); s != "cpu" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
	if s := d.Type(); s != "sysfs-thermal: file I/O is inhibited" {
		t.Fatal(s)
	}
	e := physic.Env{}
	if err := d.Sense(&e); err == nil || err.Error() != "sysfs-thermal: file I/O is inhibited" {
		t.Fatal("should have failed")
	}
	if _, err := d.SenseContinuous(time.Second); err == nil || err.Error() != "sysfs-thermal: not implemented" {
		t.Fatal(err)
	}
}

func TestThermalSensor_Type_success(t *testing.T) {
	defer resetThermal()
	fileIOOpen = func(path string, flag int) (fileIO, error) {
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		switch path {
		case "//\x00/type":
			return &fileRead{t: t, ops: [][]byte{[]byte("dummy\n")}}, nil
		default:
			t.Fatalf("unknown %q", path)
			return nil, errors.New("unknown file")
		}
	}
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	if s := d.Type(); s != "dummy" {
		t.Fatal(s)
	}
}

func TestThermalSensor_Type_fail_1(t *testing.T) {
	defer resetThermal()
	fileIOOpen = func(path string, flag int) (fileIO, error) {
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		switch path {
		case "//\x00/type":
			return &file{}, nil
		default:
			t.Fatalf("unknown %q", path)
			return nil, errors.New("unknown file")
		}
	}
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	if s := d.Type(); s != "sysfs-thermal: not implemented" {
		t.Fatal(s)
	}
}

func TestThermalSensor_Type_fail_2(t *testing.T) {
	defer resetThermal()
	fileIOOpen = func(path string, flag int) (fileIO, error) {
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		switch path {
		case "//\x00/type":
			return &fileRead{t: t, ops: [][]byte{[]byte("\n")}}, nil
		default:
			t.Fatalf("unknown %q", path)
			return nil, errors.New("unknown file")
		}
	}
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	if s := d.Type(); s != "<unknown>" {
		t.Fatal(s)
	}
}

func TestThermalSensor_Sense_success(t *testing.T) {
	defer resetThermal()
	fileIOOpen = func(path string, flag int) (fileIO, error) {
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		switch path {
		case "//\x00/temp":
			return &fileRead{t: t, ops: [][]byte{[]byte("42\n")}}, nil
		default:
			t.Fatalf("unknown %q", path)
			return nil, errors.New("unknown file")
		}
	}
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	e := physic.Env{}
	if err := d.Sense(&e); err != nil {
		t.Fatal(err)
	}
	if e.Temperature != 42*physic.Celsius+physic.ZeroCelsius {
		t.Fatal(e.Temperature)
	}
}

func TestThermalSensor_Sense_fail_1(t *testing.T) {
	defer resetThermal()
	fileIOOpen = func(path string, flag int) (fileIO, error) {
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		switch path {
		case "//\x00/temp":
			return &file{}, nil
		default:
			t.Fatalf("unknown %q", path)
			return nil, errors.New("unknown file")
		}
	}
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	e := physic.Env{}
	if err := d.Sense(&e); err == nil || err.Error() != "sysfs-thermal: not implemented" {
		t.Fatal(err)
	}
}

func TestThermalSensor_Sense_fail_2(t *testing.T) {
	defer resetThermal()
	fileIOOpen = func(path string, flag int) (fileIO, error) {
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		switch path {
		case "//\x00/temp":
			return &fileRead{t: t, ops: [][]byte{[]byte("\n")}}, nil
		default:
			t.Fatalf("unknown %q", path)
			return nil, errors.New("unknown file")
		}
	}
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	e := physic.Env{}
	if err := d.Sense(&e); err == nil || err.Error() != "sysfs-thermal: failed to read temperature" {
		t.Fatal(err)
	}
}

func TestThermalSensor_Sense_fail_3(t *testing.T) {
	defer resetThermal()
	fileIOOpen = func(path string, flag int) (fileIO, error) {
		if flag != os.O_RDONLY {
			t.Fatal(flag)
		}
		switch path {
		case "//\x00/temp":
			return &fileRead{t: t, ops: [][]byte{[]byte("aa\n")}}, nil
		default:
			t.Fatalf("unknown %q", path)
			return nil, errors.New("unknown file")
		}
	}
	d := ThermalSensor{name: "cpu", root: "//\000/"}
	e := physic.Env{}
	if err := d.Sense(&e); err == nil || err.Error() != "sysfs-thermal: strconv.Atoi: parsing \"aa\": invalid syntax" {
		t.Fatal(err)
	}
}

func TestThermalSensorDriver(t *testing.T) {
	defer resetThermal()

	d := &driverThermalSensor{}
	if len(d.Prerequisites()) != 0 {
		t.Fatal("unexpected ThermalSensor prerequisites")
	}
	// It may pass or fail, as long as it doesn't panic.
	_, _ = d.Init()
}

//

func resetThermal() {
	ThermalSensors = nil
	reset()
}

type fileRead struct {
	file
	t   *testing.T
	ops [][]byte
}

func (f *fileRead) Read(p []byte) (int, error) {
	if len(f.ops) == 0 {
		f.t.Fatal("no more content")
		return 0, errors.New("no more content")
	}
	l := len(f.ops[0])
	copy(p, f.ops[0])
	f.ops = f.ops[1:]
	return l, nil
}
