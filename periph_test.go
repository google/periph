// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package periph

import (
	"errors"
	"testing"
)

func TestInitSimple(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	})
	if len(byName) != 1 {
		t.Fatal(byName)
	}
	state, err := Init()
	if err != nil || len(state.Loaded) != 1 {
		t.Fatal(state, err)
	}

	// Call a second time, should return the same data.
	state2, err2 := Init()
	if err2 != nil || len(state2.Loaded) != len(state.Loaded) || state2.Loaded[0] != state.Loaded[0] {
		t.Fatal(state2, err2)
	}
}

func TestInitSkip(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      false,
			err:     nil,
		},
	})
	state, err := Init()
	if err != nil || len(state.Skipped) != 1 {
		t.Fatal(state, err)
	}
	if s := state.Skipped[0].String(); s != "CPU: <nil>" {
		t.Fatal(s)
	}
}

func TestInitErr(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     errors.New("oops"),
		},
	})
	state, err := Init()
	if err != nil || len(state.Loaded) != 0 || len(state.Failed) != 1 {
		t.Fatal(state, err)
	}
	if s := state.Failed[0].String(); s != "CPU: oops" {
		t.Fatal(s)
	}
}

func TestInitPrerequisitesCircular(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]Driver{
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
	state, err := Init()
	if err == nil || len(state.Loaded) != 0 {
		t.Fatal(state, err)
	}
	if err.Error() != "periph: found cycle(s) in drivers dependencies:\nBoard: CPU\nCPU: Board" {
		t.Fatal(err)
	}
}

func TestInitPrerequisitesMissing(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]Driver{
		&driver{
			name:    "CPU",
			prereqs: []string{"Board"},
			ok:      true,
			err:     nil,
		},
	})
	state, err := Init()
	if err == nil || len(state.Loaded) != 0 {
		t.Fatal(state, err)
	}
}

func TestInitAfterMissing(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]Driver{
		&driver{
			name:  "CPU",
			after: []string{"Board"},
			ok:    true,
			err:   nil,
		},
	})
	state, err := Init()
	if err != nil || len(state.Loaded) != 1 {
		t.Fatal(state, err)
	}
}

func TestDependencySkipped(t *testing.T) {
	defer reset()
	reset()
	registerDrivers([]Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      false,
			err:     errors.New("skipped"),
		},
		&driver{
			name:    "Board",
			prereqs: []string{"CPU"},
			ok:      true,
			err:     nil,
		},
	})
	state, err := Init()
	if err != nil || len(state.Skipped) != 2 {
		t.Fatal(state, err)
	}
}

func TestRegisterLate(t *testing.T) {
	defer reset()
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
	if Register(d) == nil {
		t.Fatal("can't register after Init()")
	}
}

func TestRegisterTwice(t *testing.T) {
	defer reset()
	reset()
	d := &driver{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if err := Register(d); err != nil {
		t.Fatal(err)
	}
	if Register(d) == nil {
		t.Fatal("can't register twice")
	}
}

func TestMustRegisterPanic(t *testing.T) {
	defer reset()
	reset()
	d := &driver{
		name:    "CPU",
		prereqs: nil,
		ok:      true,
		err:     nil,
	}
	if err := Register(d); err != nil {
		t.Fatal(err)
	}
	panicked := false
	defer func() {
		if err := recover(); err != nil {
			panicked = true
		}
	}()
	MustRegister(d)
	if !panicked {
		t.Fatal("MustRegister() should have panicked on driver registration failure")
	}
}

func TestPrerequisitesExplodeStagesSimple(t *testing.T) {
	defer reset()
	reset()
	d := []Driver{
		&driver{
			name:    "CPU",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
	}
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 1 || len(actual[0].drvs) != 1 {
		t.Fatal(actual)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrerequisitesExplodeStages1Dep(t *testing.T) {
	defer reset()
	reset()
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
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 2 || len(actual[0].drvs) != 1 || actual[0].drvs["CPU-generic"] != d[1] || len(actual[1].drvs) != 1 || actual[1].drvs["CPU-specialized"] != d[0] || err != nil {
		t.Fatal(actual, err)
	}
}

func TestPrerequisitesExplodeStagesCycle(t *testing.T) {
	defer reset()
	reset()
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
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 0 {
		t.Fatal(actual)
	}
	if err == nil {
		t.Fatal("cycle should have been detected")
	}
}

func TestPrerequisitesExplodeStages3Dep(t *testing.T) {
	defer reset()
	reset()
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
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 3 || len(actual[0].drvs) != 1 || len(actual[1].drvs) != 2 || len(actual[2].drvs) != 1 {
		t.Fatal(actual)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestAfterExplodeStages3Dep(t *testing.T) {
	defer reset()
	reset()
	// This explodes the stage into 3 due to diamond shaped DAG.
	d := []Driver{
		&driver{
			name:  "base2",
			after: []string{"root"},
			ok:    true,
			err:   nil,
		},
		&driver{
			name:  "base1",
			after: []string{"root"},
			ok:    true,
			err:   nil,
		},
		&driver{
			name:    "root",
			prereqs: nil,
			ok:      true,
			err:     nil,
		},
		&driver{
			name:  "super",
			after: []string{"base1", "base2"},
			ok:    true,
			err:   nil,
		},
	}
	registerDrivers(d)
	actual, err := explodeStages()
	if len(actual) != 3 || len(actual[0].drvs) != 1 || len(actual[1].drvs) != 2 || len(actual[2].drvs) != 1 {
		t.Fatal(actual)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestDrivers(t *testing.T) {
	var d []Driver
	d = insertDriver(d, &driver{name: "b"})
	d = insertDriver(d, &driver{name: "d"})
	d = insertDriver(d, &driver{name: "c"})
	d = insertDriver(d, &driver{name: "a"})
	for i, l := range []string{"a", "b", "c", "d"} {
		if d[i].String() != l {
			t.Fatal(d)
		}
	}
}

func TestFailures(t *testing.T) {
	var d []DriverFailure
	d = insertDriverFailure(d, DriverFailure{D: &driver{name: "b"}})
	d = insertDriverFailure(d, DriverFailure{D: &driver{name: "d"}})
	d = insertDriverFailure(d, DriverFailure{D: &driver{name: "c"}})
	d = insertDriverFailure(d, DriverFailure{D: &driver{name: "a"}})
	for i, l := range []string{"a", "b", "c", "d"} {
		if d[i].D.String() != l {
			t.Fatal(d)
		}
	}
}

//

func reset() {
	byName = map[string]Driver{}
	state = nil
}

func registerDrivers(drivers []Driver) {
	for _, d := range drivers {
		MustRegister(d)
	}
}

type driver struct {
	name    string
	prereqs []string
	after   []string
	ok      bool
	err     error
}

func (d *driver) String() string {
	return d.name
}

func (d *driver) Prerequisites() []string {
	return d.prereqs
}

func (d *driver) After() []string {
	return d.after
}

func (d *driver) Init() (bool, error) {
	return d.ok, d.err
}
