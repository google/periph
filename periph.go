// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package periph is a peripherals I/O library. It contains host, devices, and
// test packages to emulate the hardware.
//
// periph acts as a registry of drivers.
//
// Every device driver should register itself in their package init() function
// by calling periph.Register().
//
// The user call periph.Init() on startup to initialize all the registered drivers
// in the correct order all at once.
//
//   - cmd/ contains executables to communicate directly with the devices or the
//     buses using raw protocols.
//   - devices/ contains devices drivers that are connected to a bus (i.e I²C,
//     SPI, GPIO) that can be controlled by the host, i.e. ssd1306 (display
//     controller), bm280 (environmental sensor), etc. 'devices' contains the
//     interfaces and subpackages contain contain concrete types.
//   - experimental/ contains the drivers that are in the experimental area,
//     not yet considered stable. See DESIGN.md for the process to move drivers
//     out of this area.
//   - host/ contains all the implementations relating to the host itself, the
//     CPU and buses that are exposed by the host onto which devices can be
//     connected, i.e. I²C, SPI, GPIO, etc. 'host' contains the interfaces
//     and subpackages contain contain concrete types.
//   - conn/ contains interfaces for all the supported protocols and
//     connections (I²C, SPI, GPIO, etc).
//   - tests/ contains smoke tests.
package periph

import (
	"errors"
	"fmt"
	"sync"
)

// Driver is an implementation for a protocol.
type Driver interface {
	// String returns the name of the driver, as to be presented to the user.
	//
	// It must be unique in the list of registered drivers.
	String() string
	// Prerequisites returns a list of drivers that must be successfully loaded
	// first before attempting to load this driver.
	//
	// A driver listing a prerequisite not registered is a fatal failure at
	// initialization time.
	Prerequisites() []string
	// Init initializes the driver.
	//
	// A driver may enter one of the three following state: loaded successfully,
	// was skipped as irrelevant on this host, failed to load.
	//
	// On success, it must return true, nil.
	//
	// When irrelevant (skipped), it must return false, errors.New(<reason>).
	//
	// On failure, it must return true, errors.New(<reason>). The failure must
	// state why it failed, for example an expected OS provided driver couldn't
	// be opened, e.g. /dev/gpiomem on Raspbian.
	Init() (bool, error)
}

// DriverFailure is a driver that failed loaded.
type DriverFailure struct {
	D   Driver
	Err error
}

func (d DriverFailure) String() string {
	return fmt.Sprintf("%s: %v", d.D, d.Err)
}

// State is the state of loaded device drivers.
type State struct {
	Loaded  []Driver
	Skipped []DriverFailure
	Failed  []DriverFailure
}

// Init initially all the relevant drivers.
//
// Drivers are started concurrently.
//
// It returns the list of all drivers loaded and errors on the first call, if
// any. They are ordered by the time they finished initializing.
//
// Second call is ignored and errors are discarded.
//
// Users will want to use host.Init(), which guarantees a baseline of included
// drivers.
func Init() (*State, error) {
	mu.Lock()
	defer mu.Unlock()
	if state != nil {
		return state, nil
	}
	state = &State{}
	cD := make(chan Driver)
	cS := make(chan DriverFailure)
	cE := make(chan DriverFailure)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range cD {
			state.Loaded = append(state.Loaded, d)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range cS {
			state.Skipped = append(state.Skipped, d)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for f := range cE {
			state.Failed = append(state.Failed, f)
		}
	}()

	stages, err := explodeStages(allDrivers)
	if err != nil {
		return state, err
	}
	loaded := map[string]struct{}{}
	for _, drivers := range stages {
		loadStage(drivers, loaded, cD, cS, cE)
	}
	close(cD)
	close(cS)
	close(cE)
	wg.Wait()
	return state, nil
}

// Register registers a driver to be initialized automatically on Init().
//
// The d.String() value must be unique across all registered drivers.
//
// It is an error to call Register() after Init() was called.
func Register(d Driver) error {
	mu.Lock()
	defer mu.Unlock()
	if state != nil {
		return errors.New("drivers: can't call Register() after Init()")
	}

	n := d.String()
	if _, ok := byName[n]; ok {
		return fmt.Errorf("drivers.Register(%q): driver with same name was already registered", d)
	}
	byName[n] = d
	allDrivers = append(allDrivers, d)
	return nil
}

// MustRegister calls Register and panics if registration fails.
func MustRegister(d Driver) {
	if err := Register(d); err != nil {
		panic(err)
	}
}

//

var (
	mu         sync.Mutex
	allDrivers []Driver
	byName     = map[string]Driver{}
	state      *State
)

// explodeStages creates multiple stages if needed.
//
// It searches if there's any driver than has dependency on another driver from
// this stage and creates intermediate stage if so.
func explodeStages(drivers []Driver) ([][]Driver, error) {
	dependencies := map[string]map[string]struct{}{}
	for _, d := range drivers {
		dependencies[d.String()] = map[string]struct{}{}
	}
	// TODO(maruel): Lower number of stages by merging parallel dependencies.
	for _, d := range drivers {
		name := d.String()
		for _, depName := range d.Prerequisites() {
			if _, ok := byName[depName]; !ok {
				return nil, fmt.Errorf("drivers: unsatisfied dependency %q->%q; it is missing; skipping", name, depName)
			}
			// Dependency between two drivers of the same type. This can happen
			// when there's a process class driver and a processor specialization
			// driver. As an example, allwinner->R8, allwinner->A64, etc.
			dependencies[name][depName] = struct{}{}
		}
	}

	var stages [][]Driver
	for len(dependencies) != 0 {
		// Create a stage.
		var stage []string
		var l []Driver
		for name, deps := range dependencies {
			if len(deps) == 0 {
				stage = append(stage, name)
				l = append(l, byName[name])
				delete(dependencies, name)
			}
		}
		if len(stage) == 0 {
			return nil, fmt.Errorf("drivers: found cycle(s) in drivers dependencies; %v", dependencies)
		}
		stages = append(stages, l)

		// Trim off.
		for _, passed := range stage {
			for name := range dependencies {
				delete(dependencies[name], passed)
			}
		}
	}
	return stages, nil
}

// loadStage loads all the drivers in this stage concurrently.
func loadStage(drivers []Driver, loaded map[string]struct{}, cD chan<- Driver, cS chan<- DriverFailure, cE chan<- DriverFailure) {
	var wg sync.WaitGroup
	// Use int for concurrent access.
	skip := make([]error, len(drivers))
	for i, driver := range drivers {
		// Load only the driver if prerequisites were loaded. They are
		// guaranteed to be in a previous stage by explodeStages().
		for _, dep := range driver.Prerequisites() {
			if _, ok := loaded[dep]; !ok {
				skip[i] = fmt.Errorf("dependency not loaded: %q", dep)
				break
			}
		}
	}

	for i, driver := range drivers {
		if err := skip[i]; err != nil {
			cS <- DriverFailure{driver, err}
			continue
		}
		wg.Add(1)
		go func(d Driver, j int) {
			defer wg.Done()
			if ok, err := d.Init(); ok {
				if err == nil {
					cD <- d
					return
				}
				cE <- DriverFailure{d, err}
			} else {
				// Do not assert that err != nil, as this is hard to test thoroughly.
				cS <- DriverFailure{d, err}
				if err != nil {
					err = errors.New("no reason was given")
				}
				skip[j] = err
			}
		}(driver, i)
	}
	wg.Wait()

	for i, driver := range drivers {
		if skip[i] != nil {
			continue
		}
		loaded[driver.String()] = struct{}{}
	}
}
