// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package periph is a peripheral I/O library.
//
// It contains host and device drivers, and test packages to emulate the
// hardware.
//
// You will find API documentation in godoc, to learn more about the goals and
// design, visit https://periph.io/
//
// Package periph acts as a registry of drivers. It is focused on providing
// high quality host drivers that provide high-speed access to the hardware on
// the host computer itself.
//
// It is less concerned about implementing all possible device drivers that may
// be attached to the host's I²C, SPI, or other buses and pio pins.
//
// Every device driver should register itself in its package init() function by
// calling periph.MustRegister().
//
// The user must call periph.Init() on startup to initialize all the registered
// drivers in the correct order all at once.
//
// → cmd/ contains executables to communicate directly with the devices or the
// buses using raw protocols.
//
// → conn/ contains interfaces and registries for all the supported protocols
// and connections (I²C, SPI, GPIO, etc).
//
// → devices/ contains devices drivers that are connected to a bus (i.e I²C,
// SPI, GPIO) that can be controlled by the host, i.e. ssd1306 (display
// controller), bm280 (environmental sensor), etc. 'devices' contains the
// interfaces and subpackages contain contain concrete types.
//
// → experimental/ contains the drivers that are in the experimental area, not
// yet considered stable. See
// https://periph.io/project/#driver-lifetime-management for the process to
// move drivers out of this area.
//
// → host/ contains all the implementations relating to the host itself, the
// CPU and buses that are exposed by the host onto which devices can be
// connected, i.e. I²C, SPI, GPIO, etc. 'host' contains the interfaces and
// subpackages contain contain concrete types.
package periph // import "periph.io/x/periph"

import (
	"errors"
	"fmt"
	"sort"
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

// DriverFailure is a driver that wasn't loaded, either because it was skipped
// or because it failed to load.
type DriverFailure struct {
	D   Driver
	Err error
}

func (d DriverFailure) String() string {
	return fmt.Sprintf("%s: %v", d.D, d.Err)
}

// State is the state of loaded device drivers.
//
// Each list is sorted by the driver name.
type State struct {
	Loaded  []Driver
	Skipped []DriverFailure
	Failed  []DriverFailure
}

// Init initialises all the relevant drivers.
//
// Drivers are started concurrently.
//
// It is safe to call this function multiple times, the previous state is
// returned on later calls.
//
// Users will want to use host.Init(), which guarantees a baseline of included
// host drivers.
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
	for _, drvs := range stages {
		loadStage(drvs, loaded, cD, cS, cE)
	}
	close(cD)
	close(cS)
	close(cE)
	wg.Wait()
	d := drivers(state.Loaded)
	sort.Sort(d)
	state.Loaded = d
	f := failures(state.Skipped)
	sort.Sort(f)
	state.Skipped = f
	f = failures(state.Failed)
	sort.Sort(f)
	state.Failed = f
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
		return errors.New("periph: can't call Register() after Init()")
	}

	n := d.String()
	if _, ok := byName[n]; ok {
		return fmt.Errorf("periph: driver with same name %q was already registered", d)
	}
	byName[n] = d
	allDrivers = append(allDrivers, d)
	return nil
}

// MustRegister calls Register() and panics if registration fails.
//
// This is the function to call in a driver's package init() function.
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
func explodeStages(drvs []Driver) ([][]Driver, error) {
	dependencies := map[string]map[string]struct{}{}
	for _, d := range drvs {
		dependencies[d.String()] = map[string]struct{}{}
	}
	// TODO(maruel): Lower number of stages by merging parallel dependencies.
	for _, d := range drvs {
		name := d.String()
		for _, depName := range d.Prerequisites() {
			if _, ok := byName[depName]; !ok {
				return nil, fmt.Errorf("periph: unsatisfied dependency %q->%q; it is missing; skipping", name, depName)
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
			return nil, fmt.Errorf("periph: found cycle(s) in drivers dependencies; %v", dependencies)
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
func loadStage(drvs []Driver, loaded map[string]struct{}, cD chan<- Driver, cS chan<- DriverFailure, cE chan<- DriverFailure) {
	var wg sync.WaitGroup
	// Use int for concurrent access.
	skip := make([]error, len(drvs))
	for i, d := range drvs {
		// Load only the driver if prerequisites were loaded. They are
		// guaranteed to be in a previous stage by explodeStages().
		for _, dep := range d.Prerequisites() {
			if _, ok := loaded[dep]; !ok {
				skip[i] = fmt.Errorf("dependency not loaded: %q", dep)
				break
			}
		}
	}

	for i, drv := range drvs {
		if err := skip[i]; err != nil {
			cS <- DriverFailure{drv, err}
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
		}(drv, i)
	}
	wg.Wait()

	for i, d := range drvs {
		if skip[i] != nil {
			continue
		}
		loaded[d.String()] = struct{}{}
	}
}

type drivers []Driver

func (d drivers) Len() int           { return len(d) }
func (d drivers) Less(i, j int) bool { return d[i].String() < d[j].String() }
func (d drivers) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

type failures []DriverFailure

func (f failures) Len() int           { return len(f) }
func (f failures) Less(i, j int) bool { return f[i].D.String() < f[j].D.String() }
func (f failures) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
