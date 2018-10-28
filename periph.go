// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package periph is a peripheral I/O library.
//
// Package periph acts as a registry of drivers. It is focused on providing
// high quality host drivers that provide high-speed access to the hardware on
// the host computer itself.
//
// To learn more about the goals and design, visit https://periph.io/
//
// Every device driver should register itself in its package init() function by
// calling periph.MustRegister().
//
// User shall call either host.Init() or hostextra.Init() on startup to
// initialize all the registered drivers.
//
// Cmd
//
// cmd/ contains executable tools to communicate directly with the devices or
// the buses.
//
// cmd/ is allowed to import from conn/, devices/ and host/.
//
// Conn
//
// conn/ contains interfaces and registries for all the supported protocols and
// connections (I²C, SPI, GPIO, etc).
//
// conn/ is not allowed to import from any other package.
//
// Devices
//
// devices/ contains devices drivers that are connected to bus, port or
// connection (i.e I²C, SPI, 1-wire, GPIO) that can be controlled by the host,
// i.e. ssd1306 (display controller), bm280 (environmental sensor), etc.
//
// devices/ is allowed to import from conn/ and host/.
//
// Experimental
//
// experimental/ contains the drivers that are in the experimental area, not
// yet considered stable. See
// https://periph.io/project/#driver-lifetime-management for the process to
// move drivers out of this area.
//
// experimental/ is allowed to import from conn/, devices/ and host/.
//
// Host
//
// host/ contains all the implementations relating to the host itself, the CPU
// and buses that are exposed by the host onto which devices can be connected,
// i.e. I²C, SPI, GPIO, etc.
//
// host/ is allowed to import from conn/ only.
package periph // import "periph.io/x/periph"

import (
	"errors"
	"strconv"
	"strings"
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
	// After returns a list of drivers that must be loaded first before
	// attempting to load this driver.
	//
	// Unlike Prerequisites(), this driver will still be attempted even if the
	// listed driver is missing or failed to load.
	//
	// This permits serialization without hard requirement.
	After() []string
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
	out := d.D.String() + ": "
	if d.Err != nil {
		out += d.Err.Error()
	} else {
		out += "<nil>"
	}
	return out
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
	return initImpl()
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
		return errors.New("periph: driver with same name " + strconv.Quote(n) + " was already registered")
	}
	byName[n] = d
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
	// mu guards byName and state.
	// - byName is only mutated by Register().
	// - state is only mutated by Init().
	//
	// Once Init() is called, Register() refuses registering more drivers, thus
	// byName is immutable once Init() started.
	mu     sync.Mutex
	byName = map[string]Driver{}
	state  *State
)

// stage is a set of drivers that can be loaded in parallel.
type stage struct {
	// Subset of byName drivers, for the ones in this stage.
	drvs map[string]Driver
}

// explodeStages creates one or multiple stages by processing byName.
//
// It searches if there's any driver than has dependency on another driver and
// create stages from this DAG.
//
// It also verifies that there is not cycle in the DAG.
//
// When this function starts, allDriver and byName are guaranteed to be
// immutable. state must not be touched by this function.
func explodeStages() ([]*stage, error) {
	// First, create the DAG.
	dag := map[string]map[string]struct{}{}
	for name, d := range byName {
		m := map[string]struct{}{}
		for _, p := range d.Prerequisites() {
			if _, ok := byName[p]; !ok {
				return nil, errors.New("periph: unsatisfied dependency " + strconv.Quote(name) + "->" + strconv.Quote(p) + "; it is missing; skipping")
			}
			m[p] = struct{}{}
		}
		for _, p := range d.After() {
			// Skip undefined drivers silently, unlike Prerequisites().
			if _, ok := byName[p]; ok {
				m[p] = struct{}{}
			}
		}
		dag[name] = m
	}

	// Create stages.
	var stages []*stage
	for len(dag) != 0 {
		s := &stage{drvs: map[string]Driver{}}
		for name, deps := range dag {
			// This driver has no dependency, add it to the current stage.
			if len(deps) == 0 {
				s.drvs[name] = byName[name]
				delete(dag, name)
			}
		}
		if len(s.drvs) == 0 {
			// Print out the remaining DAG so users can diagnose.
			// It'd probably be nicer if it were done in Register()?
			s := make([]string, 0, len(dag))
			for name, deps := range dag {
				x := make([]string, 0, len(deps))
				for d := range deps {
					x = insertString(x, d)
				}
				s = insertString(s, name+": "+strings.Join(x, ", "))
			}
			return nil, errors.New("periph: found cycle(s) in drivers dependencies:\n" + strings.Join(s, "\n"))
		}
		stages = append(stages, s)

		// Trim the dependencies for the items remaining in the dag.
		for passed := range s.drvs {
			for name := range dag {
				delete(dag[name], passed)
			}
		}
	}
	return stages, nil
}

func insertDriver(l []Driver, d Driver) []Driver {
	n := d.String()
	i := search(len(l), func(i int) bool { return l[i].String() > n })
	l = append(l, nil)
	copy(l[i+1:], l[i:])
	l[i] = d
	return l
}

func insertDriverFailure(l []DriverFailure, f DriverFailure) []DriverFailure {
	n := f.String()
	i := search(len(l), func(i int) bool { return l[i].String() > n })
	l = append(l, DriverFailure{})
	copy(l[i+1:], l[i:])
	l[i] = f
	return l
}

func insertString(l []string, s string) []string {
	i := search(len(l), func(i int) bool { return l[i] > s })
	l = append(l, "")
	copy(l[i+1:], l[i:])
	l[i] = s
	return l
}

// search implements the same algorithm as sort.Search().
//
// It was extracted to to not depend on sort, which depends on reflect.
func search(n int, f func(int) bool) int {
	lo := 0
	for hi := n; lo < hi; {
		if i := int(uint(lo+hi) >> 1); !f(i) {
			lo = i + 1
		} else {
			hi = i
		}
	}
	return lo
}
