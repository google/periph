// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package i2creg defines I²C bus registry to list buses present on the host.
package i2creg

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"periph.io/x/periph/conn/i2c"
)

// Opener opens an handle to a bus.
//
// It is provided by the actual bus driver.
type Opener func() (i2c.BusCloser, error)

// Ref references an I²C bus.
//
// It is returned by All() to enumerate all registered buses.
type Ref struct {
	// Name of the bus.
	//
	// It must not be a sole number. It must be unique across the host.
	Name string
	// Aliases are the alternative names that can be used to reference this bus.
	Aliases []string
	// Number of the bus or -1 if the bus doesn't have any "native" number.
	//
	// Buses provided by the CPU normally have a 0 based number. Buses provided
	// via an addon (like over USB) generally are not numbered.
	Number int
	// Open is the factory to open an handle to this I²C bus.
	Open Opener
}

// Open opens an I²C bus by its name, an alias or its number and returns an
// handle to it.
//
// Specify the empty string "" to get the first available bus. This is the
// recommended default value unless an application knows the exact bus to use.
//
// Each bus can register multiple aliases, each leading to the same bus handle.
//
// "Bus number" is a generic concept that is highly dependent on the platform
// and OS. On some platform, the first bus may have the number 0, 1 or higher.
// Bus numbers are not necessarily continuous and may not start at 0. It was
// observed that the bus number as reported by the OS may change across OS
// revisions.
//
// When the I²C bus is provided by an off board plug and play bus like USB via
// a FT232H USB device, there can be no associated number.
func Open(name string) (i2c.BusCloser, error) {
	var r *Ref
	var err error
	func() {
		mu.Lock()
		defer mu.Unlock()
		if len(byName) == 0 {
			err = wrapf("no bus found; did you forget to call Init()?")
			return
		}
		if len(name) == 0 {
			r = getDefault()
			return
		}
		// Try by name, by alias, by number.
		if r = byName[name]; r == nil {
			if r = byAlias[name]; r == nil {
				if i, err2 := strconv.Atoi(name); err2 == nil {
					r = byNumber[i]
				}
			}
		}
	}()
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, wrapf("can't open unknown bus: %q", name)
	}
	return r.Open()
}

// All returns a copy of all the registered references to all know I²C buses
// available on this host.
//
// The list is sorted by the bus name.
func All() []*Ref {
	var out refList
	func() {
		mu.Lock()
		defer mu.Unlock()
		out = make(refList, 0, len(byName))
		for _, v := range byName {
			r := &Ref{Name: v.Name, Aliases: make([]string, len(v.Aliases)), Number: v.Number, Open: v.Open}
			copy(r.Aliases, v.Aliases)
			out = append(out, r)
		}
	}()
	sort.Sort(out)
	return out
}

// Register registers an I²C bus.
//
// Registering the same bus name twice is an error, e.g. o.Name(). o.Number()
// can be -1 to signify that the bus doesn't have an inherent "bus number". A
// good example is a bus provided over a FT232H device connected on an USB bus.
// In this case, the bus name should be created from the serial number of the
// device for unique identification.
func Register(name string, aliases []string, number int, o Opener) error {
	if len(name) == 0 {
		return wrapf("can't register a bus with no name")
	}
	if o == nil {
		return wrapf("can't register bus %q with nil Opener", name)
	}
	if number < -1 {
		return wrapf("can't register bus %q with invalid bus number %d", name, number)
	}
	if _, err := strconv.Atoi(name); err == nil {
		return wrapf("can't register bus %q with name being only a number", name)
	}
	if strings.Contains(name, ":") {
		return wrapf("can't register bus %q with name containing ':'", name)
	}
	for _, alias := range aliases {
		if len(alias) == 0 {
			return wrapf("can't register bus %q with an empty alias", name)
		}
		if name == alias {
			return wrapf("can't register bus %q with an alias the same as the bus name", name)
		}
		if _, err := strconv.Atoi(alias); err == nil {
			return wrapf("can't register bus %q with an alias that is a number: %q", name, alias)
		}
		if strings.Contains(alias, ":") {
			return wrapf("can't register bus %q with an alias containing ':': %q", name, alias)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return wrapf("can't register bus %q twice", name)
	}
	if _, ok := byAlias[name]; ok {
		return wrapf("can't register bus %q twice; it is already an alias", name)
	}
	if number != -1 {
		if _, ok := byNumber[number]; ok {
			return wrapf("can't register bus %q; bus number %d is already registered", name, number)
		}
	}
	for _, alias := range aliases {
		if _, ok := byName[alias]; ok {
			return wrapf("can't register bus %q twice; alias %q is already a bus", name, alias)
		}
		if _, ok := byAlias[alias]; ok {
			return wrapf("can't register bus %q twice; alias %q is already an alias", name, alias)
		}
	}

	r := &Ref{Name: name, Aliases: make([]string, len(aliases)), Number: number, Open: o}
	copy(r.Aliases, aliases)
	byName[name] = r
	if number != -1 {
		byNumber[number] = r
	}
	for _, alias := range aliases {
		byAlias[alias] = r
	}
	return nil
}

// Unregister removes a previously registered I²C bus.
//
// This can happen when an I²C bus is exposed via an USB device and the device
// is unplugged.
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()
	r := byName[name]
	if r == nil {
		return wrapf("can't unregister unknown bus name %q", name)
	}
	delete(byName, name)
	delete(byNumber, r.Number)
	for _, alias := range r.Aliases {
		delete(byAlias, alias)
	}
	return nil
}

//

var (
	mu     sync.Mutex
	byName = map[string]*Ref{}
	// Caches
	byNumber = map[int]*Ref{}
	byAlias  = map[string]*Ref{}
)

// getDefault returns the Ref that should be used as the default bus.
func getDefault() *Ref {
	var o *Ref
	if len(byNumber) == 0 {
		// Fallback to use byName using a lexical sort.
		name := ""
		for n, o2 := range byName {
			if len(name) == 0 || n < name {
				o = o2
				name = n
			}
		}
		return o
	}
	number := int((^uint(0)) >> 1)
	for n, o2 := range byNumber {
		if number > n {
			number = n
			o = o2
		}
	}
	return o
}

// wrapf returns an error that is wrapped with the package name.
func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("i2creg: "+format, a...)
}

type refList []*Ref

func (r refList) Len() int           { return len(r) }
func (r refList) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r refList) Less(i, j int) bool { return r[i].Name < r[j].Name }
