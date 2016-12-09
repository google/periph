// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package periph

import (
	"errors"
	"fmt"
	"log"
	"testing"
)

func ExampleInit() {
	// You probably want host.Init() instead as it registers all the
	// periph-provided host drivers automatically.
	state, err := Init()
	if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}
	fmt.Printf("Using drivers:\n")
	for _, driver := range state.Loaded {
		fmt.Printf("- %s\n", driver)
	}
	fmt.Printf("Drivers skipped:\n")
	for _, failure := range state.Skipped {
		fmt.Printf("- %s: %s\n", failure.D, failure.Err)
	}
	// Having drivers failing to load may not require process termination. It
	// is possible to continue to run in partial failure mode.
	fmt.Printf("Drivers failed to load:\n")
	for _, failure := range state.Failed {
		fmt.Printf("- %s: %v\n", failure.D, failure.Err)
	}

	// Use pins, buses, devices, etc.
}

func TestInitSimple(t *testing.T) {
	initTest([]Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	})
	if len(allDrivers) != 1 {
		t.Fatalf("%v", allDrivers)
	}
	if len(byName) != 1 {
		t.Fatalf("%v", byName)
	}
	state, err := Init()
	if err != nil || len(state.Loaded) != 1 {
		t.Fatalf("%v, %v", state, err)
	}

	// Call a second time, should return the same data.
	state2, err2 := Init()
	if err2 != nil || len(state2.Loaded) != len(state.Loaded) || state2.Loaded[0] != state.Loaded[0] {
		t.Fatalf("%v, %v", state2, err2)
	}
}

func TestInitSkip(t *testing.T) {
	initTest([]Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      false,
			err:     nil,
		},
	})
	if state, err := Init(); err != nil || len(state.Loaded) != 0 {
		t.Fatalf("%v, %v", state, err)
	}
}

func TestInitErr(t *testing.T) {
	initTest([]Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     errors.New("oops"),
		},
	})
	if state, err := Init(); err != nil || len(state.Loaded) != 0 || len(state.Failed) != 1 {
		t.Fatalf("%v, %v", state, err)
	}
}

func TestInitCircular(t *testing.T) {
	initTest([]Driver{
		&driver{
			name:    "CPU",
			prereqs: []string{"Board"},
			ok:      true,
			err:     nil,
		},
		&driver{
			name:    "Board",
			prereqs: []string{"CPU"},
			ok:      true,
			err:     nil,
		},
	})
	if state, err := Init(); err == nil || len(state.Loaded) != 0 {
		t.Fatalf("%v, %v", state, err)
	}
}

func TestInitMissing(t *testing.T) {
	initTest([]Driver{
		&driver{
			name:    "CPU",
			prereqs: []string{"Board"},
			ok:      true,
			err:     nil,
		},
	})
	if state, err := Init(); err == nil || len(state.Loaded) != 0 {
		t.Fatalf("%v, %v", state, err)
	}
}

func TestRegisterLate(t *testing.T) {
	reset()
	if _, err := Init(); err != nil {
		t.Fatal(err)
	}
	d := &driver{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if err := Register(d); err == nil {
		t.Fail()
	}
}

func TestRegisterTwice(t *testing.T) {
	reset()
	d := &driver{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if err := Register(d); err != nil {
		t.Fail()
	}
	if err := Register(d); err == nil {
		t.Fail()
	}
}

func TestMustRegisterPanic(t *testing.T) {
	reset()
	d := &driver{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if err := Register(d); err != nil {
		t.Fail()
	}
	panicked := false
	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()
	MustRegister(d)
	if !panicked {
		t.Fail()
	}
}

func TestExplodeStagesSimple(t *testing.T) {
	d := []Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	}
	initTest(d)
	actual, err := explodeStages(d)
	if len(actual) != 1 || len(actual[0]) != 1 {
		t.Fatalf("%v", actual)
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestExplodeStages1Dep(t *testing.T) {
	// This explodes the stage into two.
	d := []Driver{
		&driver{
			name:    "CPU-specialized",
			prereqs: []string{"CPU-generic"},
			ok:      true,
			err:     nil,
		},
		&driver{
			name:    "CPU-generic",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	}
	initTest(d)
	actual, err := explodeStages(d)
	if len(actual) != 2 || len(actual[0]) != 1 || actual[0][0] != d[1] || len(actual[1]) != 1 || actual[1][0] != d[0] || err != nil {
		t.Fatalf("%v, %v", actual, err)
	}
}

func TestExplodeStagesCycle(t *testing.T) {
	d := []Driver{
		&driver{
			name:    "A",
			prereqs: []string{"B"},
			ok:      true,
			err:     nil,
		},
		&driver{
			name:    "B",
			prereqs: []string{"C"},
			ok:      true,
			err:     nil,
		},
		&driver{
			name:    "C",
			prereqs: []string{"A"},
			ok:      true,
			err:     nil,
		},
	}
	initTest(d)
	actual, err := explodeStages(d)
	if len(actual) != 0 {
		t.Fatalf("%v", actual)
	}
	if err == nil {
		t.Fail()
	}
}

func TestExplodeStages3Dep(t *testing.T) {
	// This explodes the stage into 3 due to diamond shaped DAG.
	d := []Driver{
		&driver{
			name:    "base2",
			prereqs: []string{"root"},
			ok:      true,
			err:     nil,
		},
		&driver{
			name:    "base1",
			prereqs: []string{"root"},
			ok:      true,
			err:     nil,
		},
		&driver{
			name:    "root",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
		&driver{
			name:    "super",
			prereqs: []string{"base1", "base2"},
			ok:      true,
			err:     nil,
		},
	}
	initTest(d)
	actual, err := explodeStages(d)
	if len(actual) != 3 || len(actual[0]) != 1 || len(actual[1]) != 2 || len(actual[2]) != 1 {
		t.Fatalf("%v", actual)
	}
	if err != nil {
		t.Fatalf("%v", err)
	}
}

//

func reset() {
	allDrivers = []Driver{}
	byName = map[string]Driver{}
	state = nil
}

func initTest(drivers []Driver) {
	reset()
	for _, d := range drivers {
		MustRegister(d)
	}
}

type driver struct {
	name    string
	prereqs []string
	ok      bool
	err     error
}

func (d *driver) String() string {
	return d.name
}

func (d *driver) Prerequisites() []string {
	return d.prereqs
}

func (d *driver) Init() (bool, error) {
	return d.ok, d.err
}
