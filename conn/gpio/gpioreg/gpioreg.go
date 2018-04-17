// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpioreg defines a registry for the known digital pins.
package gpioreg

import (
	"errors"
	"strconv"
	"sync"

	"periph.io/x/periph/conn/gpio"
)

// ByName returns a GPIO pin from its name, gpio number or one of its aliases.
//
// For example on a Raspberry Pi, the following values will return the same
// GPIO: the gpio as a number "2", the chipset name "GPIO2", the board pin
// position "P1_3", it's function name "I2C1_SDA".
//
// Returns nil if the gpio pin is not present.
func ByName(name string) gpio.PinIO {
	mu.Lock()
	defer mu.Unlock()
	if p, ok := byName[name]; ok {
		return p
	}
	if dest, ok := byAlias[name]; ok {
		if p := getByNameDeep(dest); p != nil {
			// Wraps the destination in an alias, so the name makes sense to the user.
			// The main drawback is that casting into other gpio interfaces like
			// gpio.PinPWM requires going through gpio.RealPin first.
			return &pinAlias{p, name}
		}
	}
	return nil
}

// All returns all the GPIO pins available on this host.
//
// The list is guaranteed to be in order of name using 'natural sorting'.
//
// This list excludes aliases.
//
// This list excludes non-GPIO pins like GROUND, V3_3, etc, since they are not
// GPIO.
func All() []gpio.PinIO {
	mu.Lock()
	defer mu.Unlock()
	out := make([]gpio.PinIO, 0, len(byName))
	for _, p := range byName {
		out = insertPinByName(out, p)
	}
	return out
}

// Aliases returns all pin aliases.
//
// The list is guaranteed to be in order of aliase name.
func Aliases() []gpio.PinIO {
	mu.Lock()
	defer mu.Unlock()
	out := make([]gpio.PinIO, 0, len(byAlias))
	for name, dest := range byAlias {
		// Skip aliases that were not resolved.
		if p := getByNameDeep(dest); p != nil {
			out = insertPinByName(out, &pinAlias{p, name})
		}
	}
	return out
}

// Register registers a GPIO pin.
//
// Registering the same pin number or name twice is an error.
//
// Deprecated: `preferred` is now ignored.
//
// The pin registered cannot implement the interface RealPin.
func Register(p gpio.PinIO, preferred bool) error {
	// TODO(maruel): Remove preferred in v3.
	name := p.Name()
	if len(name) == 0 {
		return errors.New("gpioreg: can't register a pin with no name")
	}
	if r, ok := p.(gpio.RealPin); ok {
		return errors.New("gpioreg: can't register pin " + strconv.Quote(name) + ", it is already an alias to " + strconv.Quote(r.Real().String()))
	}

	mu.Lock()
	defer mu.Unlock()
	if orig, ok := byName[name]; ok {
		return errors.New("gpioreg: can't register pin " + strconv.Quote(name) + " twice; already registered as " + strconv.Quote(orig.String()))
	}
	if dest, ok := byAlias[name]; ok {
		return errors.New("gpioreg: can't register pin " + strconv.Quote(name) + "; an alias already exist to: " + strconv.Quote(dest))
	}
	byName[name] = p
	return nil
}

// RegisterAlias registers an alias for a GPIO pin.
//
// It is possible to register an alias for a pin that itself has not been
// registered yet. It is valid to register an alias to another alias. It is
// valid to register the same alias multiple times, overriding the previous
// alias.
func RegisterAlias(alias string, dest string) error {
	if len(alias) == 0 {
		return errors.New("gpioreg: can't register an alias with no name")
	}
	if len(dest) == 0 {
		return errors.New("gpioreg: can't register alias " + strconv.Quote(alias) + " with no dest")
	}

	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[alias]; ok {
		return errors.New("gpioreg: can't register alias " + strconv.Quote(alias) + " for a pin that exists")
	}
	byAlias[alias] = dest
	return nil
}

// Unregister removes a previously registered GPIO pin or alias from the GPIO
// pin registry.
//
// This can happen when a GPIO pin is exposed via an USB device and the device
// is unplugged, or when a generic OS provided pin is superseded by a CPU
// specific implementation.
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		delete(byName, name)
		return nil
	}
	if _, ok := byAlias[name]; ok {
		delete(byAlias, name)
		return nil
	}
	return errors.New("gpioreg: can't unregister unknown pin name " + strconv.Quote(name))
}

//

var (
	mu      sync.Mutex
	byName  = map[string]gpio.PinIO{}
	byAlias = map[string]string{}
)

// pinAlias implements an alias for a PinIO.
//
// pinAlias implements the RealPin interface, which allows querying for the
// real pin under the alias.
type pinAlias struct {
	gpio.PinIO
	name string
}

// String returns the alias name along the real pin's Name() in parenthesis, if
// known, else the real pin's number.
func (a *pinAlias) String() string {
	return a.name + "(" + a.PinIO.Name() + ")"
}

// Name returns the pinAlias's name.
func (a *pinAlias) Name() string {
	return a.name
}

// Real returns the real pin behind the alias
func (a *pinAlias) Real() gpio.PinIO {
	return a.PinIO
}

// getByNameDeep recursively resolves the aliases to get the pin.
func getByNameDeep(name string) gpio.PinIO {
	if p, ok := byName[name]; ok {
		return p
	}
	if dest, ok := byAlias[name]; ok {
		if p := getByNameDeep(dest); p != nil {
			// Return the deep pin directly, bypassing the aliases.
			return p
		}
	}
	return nil
}

// insertPinByName inserts pin p into list l while keeping l ordered by name.
func insertPinByName(l []gpio.PinIO, p gpio.PinIO) []gpio.PinIO {
	n := p.Name()
	i := search(len(l), func(i int) bool { return lessNatural(n, l[i].Name()) })
	l = append(l, nil)
	copy(l[i+1:], l[i:])
	l[i] = p
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
