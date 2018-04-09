// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spireg defines the SPI registry for SPI ports discovered on the host.
//
// SPI ports discovered on the host are automatically registered in the SPI
// registry by host.Init().
package spireg

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	"periph.io/x/periph/conn/spi"
)

// Opener opens an handle to a port.
//
// It is provided by the actual port driver.
type Opener func() (spi.PortCloser, error)

// Ref references a SPI port.
//
// It is returned by All() to enumerate all registered ports.
type Ref struct {
	// Name of the port.
	//
	// It must not be a sole number. It must be unique across the host.
	Name string
	// Aliases are the alternative names that can be used to reference this port.
	Aliases []string
	// Number of the port or -1 if the port doesn't have any "native" number.
	//
	// Buses provided by the CPU normally have a 0 based number. Buses provided
	// via an addon (like over USB) generally are not numbered.
	//
	// The port is a bus number plus a CS line.
	Number int
	// Open is the factory to open an handle to this SPI port.
	Open Opener
}

// Open opens a SPI port by its name, an alias or its number and returns an
// handle to it.
//
// Specify the empty string "" to get the first available port. This is the
// recommended default value unless an application knows the exact port to use.
//
// Each port can register multiple aliases, each leading to the same port
// handle.
//
// "Bus number" is a generic concept that is highly dependent on the platform
// and OS. On some platform, the first port may have the number 0, 1 or as high
// as 32766. Bus numbers are not necessarily continuous and may not start at 0.
// It was observed that the bus number as reported by the OS may change across
// OS revisions.
//
// A SPI port is constructed of the bus number and the chip select (CS) number.
//
// When the SPI port is provided by an off board plug and play bus like USB via
// a FT232H USB device, there can be no associated number.
func Open(name string) (spi.PortCloser, error) {
	var r *Ref
	var err error
	func() {
		mu.Lock()
		defer mu.Unlock()
		if len(byName) == 0 {
			err = errors.New("spireg: no port found; did you forget to call Init()?")
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
		return nil, errors.New("spireg: can't open unknown port: " + strconv.Quote(name))
	}
	return r.Open()
}

// All returns a copy of all the registered references to all know SPI ports
// available on this host.
//
// The list is sorted by the port name.
func All() []*Ref {
	mu.Lock()
	defer mu.Unlock()
	out := make([]*Ref, 0, len(byName))
	for _, v := range byName {
		r := &Ref{Name: v.Name, Aliases: make([]string, len(v.Aliases)), Number: v.Number, Open: v.Open}
		copy(r.Aliases, v.Aliases)
		out = insertRef(out, r)
	}
	return out
}

// Register registers a SPI port.
//
// Registering the same port name twice is an error, e.g. o.Name(). o.Number()
// can be -1 to signify that the port doesn't have an inherent "bus number". A
// good example is a port provided over a FT232H device connected on an USB bus.
// In this case, the port name should be created from the serial number of the
// device for unique identification.
//
// Only ports with the CS #0 are registered with their number.
func Register(name string, aliases []string, number int, o Opener) error {
	if len(name) == 0 {
		return errors.New("spireg: can't register a port with no name")
	}
	if o == nil {
		return errors.New("spireg: can't register port " + strconv.Quote(name) + " with nil Opener")
	}
	if number < -1 {
		return errors.New("spireg: can't register port " + strconv.Quote(name) + " with invalid port number " + strconv.Itoa(number))
	}
	if _, err := strconv.Atoi(name); err == nil {
		return errors.New("spireg: can't register port " + strconv.Quote(name) + " with name being only a number")
	}
	if strings.Contains(name, ":") {
		return errors.New("spireg: can't register port " + strconv.Quote(name) + " with name containing ':'")
	}
	for _, alias := range aliases {
		if len(alias) == 0 {
			return errors.New("spireg: can't register port " + strconv.Quote(name) + " with an empty alias")
		}
		if name == alias {
			return errors.New("spireg: can't register port " + strconv.Quote(name) + " with an alias the same as the port name")
		}
		if _, err := strconv.Atoi(alias); err == nil {
			return errors.New("spireg: can't register port " + strconv.Quote(name) + " with an alias that is a number: " + strconv.Quote(alias))
		}
		if strings.Contains(alias, ":") {
			return errors.New("spireg: can't register port " + strconv.Quote(name) + " with an alias containing ':': " + strconv.Quote(alias))
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return errors.New("spireg: can't register port " + strconv.Quote(name) + " twice")
	}
	if _, ok := byAlias[name]; ok {
		return errors.New("spireg: can't register port " + strconv.Quote(name) + " twice; it is already an alias")
	}
	if number != -1 {
		if _, ok := byNumber[number]; ok {
			return errors.New("spireg: can't register port " + strconv.Quote(name) + "; port number " + strconv.Itoa(number) + " is already registered")
		}
	}
	for _, alias := range aliases {
		if _, ok := byName[alias]; ok {
			return errors.New("spireg: can't register port " + strconv.Quote(name) + " twice; alias " + strconv.Quote(alias) + " is already a port")
		}
		if _, ok := byAlias[alias]; ok {
			return errors.New("spireg: can't register port " + strconv.Quote(name) + " twice; alias " + strconv.Quote(alias) + " is already an alias")
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

// Unregister removes a previously registered SPI port.
//
// This can happen when a SPI port is exposed via an USB device and the device
// is unplugged.
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()
	r := byName[name]
	if r == nil {
		return errors.New("spireg: can't unregister unknown port name " + strconv.Quote(name))
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

// getDefault returns the Ref that should be used as the default port.
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

func insertRef(l []*Ref, r *Ref) []*Ref {
	n := r.Name
	i := search(len(l), func(i int) bool { return l[i].Name > n })
	l = append(l, nil)
	copy(l[i+1:], l[i:])
	l[i] = r
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
