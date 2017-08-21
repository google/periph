// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package gpioreg defines a registry for the known digital pins.
package gpioreg

import (
	"fmt"
	"sort"
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
	return getByName(name)
}

// All returns all the GPIO pins available on this host.
//
// The list is guaranteed to be in order of number.
//
// This list excludes aliases.
//
// This list excludes non-GPIO pins like GROUND, V3_3, etc, since they are not
// GPIO.
func All() []gpio.PinIO {
	mu.Lock()
	defer mu.Unlock()
	out := make(pinList, 0, len(byNumber))
	seen := make(map[int]struct{}, len(byNumber[0]))
	// Memory-mapped pins have highest priority, include all of them.
	for _, p := range byNumber[0] {
		out = append(out, p)
		seen[p.Number()] = struct{}{}
	}
	// Add in OS accessible pins that cannot be accessed via memory-map.
	for _, p := range byNumber[1] {
		if _, ok := seen[p.Number()]; !ok {
			out = append(out, p)
		}
	}
	sort.Sort(out)
	return out
}

// Aliases returns all pin aliases.
//
// The list is guaranteed to be in order of aliase name.
func Aliases() []gpio.PinIO {
	mu.Lock()
	defer mu.Unlock()
	out := make(pinList, 0, len(byAlias))
	for _, p := range byAlias {
		// Skip aliases that were not resolved. This requires resolving all aliases.
		if p.PinIO == nil {
			if p.PinIO = getByName(p.dest); p.PinIO == nil {
				continue
			}
		}
		out = append(out, p)
	}
	sort.Sort(out)
	return out
}

// Register registers a GPIO pin.
//
// Registering the same pin number or name twice is an error.
//
// `preferred` should be true when the pin being registered is exposing as much
// functionality as possible via the underlying hardware. This is normally done
// by accessing the CPU memory mapped registers directly.
//
// `preferred` should be false when the functionality is provided by the OS and
// is limited or slower.
//
// The pin registered cannot implement the interface RealPin.
func Register(p gpio.PinIO, preferred bool) error {
	name := p.Name()
	if len(name) == 0 {
		return wrapf("can't register a pin with no name")
	}
	if _, err := strconv.Atoi(name); err == nil {
		return wrapf("can't register pin %q with name being only a number", name)
	}
	number := p.Number()
	if number < 0 {
		return wrapf("can't register pin %q with invalid pin number %d", name, number)
	}
	i := 0
	other := 1
	if !preferred {
		i = 1
		other = 0
	}

	mu.Lock()
	defer mu.Unlock()
	if orig, ok := byNumber[i][number]; ok {
		return wrapf("can't register pin %q twice with the same number %d; already registered as %s", name, number, orig)
	}
	if orig, ok := byName[i][name]; ok {
		return wrapf("can't register pin %q twice; already registered as %s", name, orig)
	}
	if r, ok := p.(gpio.RealPin); ok {
		return wrapf("can't register pin %q, it is already an alias: %s; use RegisterAlias() instead", name, r)
	}
	if alias, ok := byAlias[name]; ok {
		return wrapf("can't register pin %q; an alias already exist: %s", name, alias)
	}
	if orig, ok := byName[other][name]; ok && number != orig.Number() {
		return wrapf("can't register pin %q twice with different number; already registered as %s", name, orig)
	}
	byNumber[i][number] = p
	byName[i][name] = p
	return nil
}

// RegisterAlias registers an alias for a GPIO pin.
//
// It is possible to register an alias for a pin that itself has not been
// registered yet. It is valid to register an alias to another alias or to a
// number.
func RegisterAlias(alias string, dest string) error {
	if len(alias) == 0 {
		return wrapf("can't register an alias with no name")
	}
	if len(dest) == 0 {
		return wrapf("can't register alias %q with no dest", alias)
	}
	if _, err := strconv.Atoi(alias); err == nil {
		return wrapf("can't register alias %q with name being only a number", alias)
	}

	mu.Lock()
	defer mu.Unlock()
	if orig := byAlias[alias]; orig != nil {
		return wrapf("can't register alias %q twice; it is already an alias: %v", alias, orig)
	}
	byAlias[alias] = &pinAlias{name: alias, dest: dest}
	return nil
}

//

var (
	mu sync.Mutex
	// The first map is preferred pins, the second is for more limited pins,
	// usually going through OS-provided abstraction layer.
	byNumber = [2]map[int]gpio.PinIO{{}, {}}
	byName   = [2]map[string]gpio.PinIO{{}, {}}
	byAlias  = map[string]*pinAlias{}
)

// pinAlias implements an alias for a PinIO.
//
// pinAlias implements the RealPin interface, which allows querying for the
// real pin under the alias.
type pinAlias struct {
	gpio.PinIO
	name string
	dest string
}

// String returns the alias name along the real pin's Name() in parenthesis, if
// known, else the real pin's number.
func (a *pinAlias) String() string {
	if a.PinIO == nil {
		return fmt.Sprintf("%s(%s)", a.name, a.dest)
	}
	return fmt.Sprintf("%s(%s)", a.name, a.PinIO.Name())
}

// Name returns the pinAlias's name.
func (a *pinAlias) Name() string {
	return a.name
}

// Real returns the real pin behind the alias
func (a *pinAlias) Real() gpio.PinIO {
	return a.PinIO
}

func getByNumber(number int) gpio.PinIO {
	if p, ok := byNumber[0][number]; ok {
		return p
	}
	if p, ok := byNumber[1][number]; ok {
		return p
	}
	return nil
}

// getByName recursively resolves the aliases to get the pin.
func getByName(name string) gpio.PinIO {
	if p, ok := byName[0][name]; ok {
		return p
	}
	if p, ok := byName[1][name]; ok {
		return p
	}
	if p, ok := byAlias[name]; ok {
		if p.PinIO == nil {
			if p.PinIO = getByName(p.dest); p.PinIO == nil {
				return nil
			}
		}
		return p
	}
	if i, err := strconv.Atoi(name); err == nil {
		return getByNumber(i)
	}
	return nil
}

// wrapf returns an error that is wrapped with the package name.
func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("gpioreg: "+format, a...)
}

type pinList []gpio.PinIO

func (p pinList) Len() int           { return len(p) }
func (p pinList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p pinList) Less(i, j int) bool { return p[i].Number() < p[j].Number() }

func init() {
	Register(gpio.INVALID, true)
}
