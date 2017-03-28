// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spi defines the SPI protocol.
package spi

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
)

// Interfaces

// Mode determines how communication is done. The bits can be OR'ed to change
// the polarity and phase used for communication.
//
// CPOL means the clock polarity. Idle is High when set.
//
// CPHA is the clock phase, sample on trailing edge when set.
type Mode int

// Valid SPI clock and phase.
const (
	Mode0 Mode = 0x0 // CPOL=0, CPHA=0
	Mode1 Mode = 0x1 // CPOL=0, CPHA=1
	Mode2 Mode = 0x2 // CPOL=1, CPHA=0
	Mode3 Mode = 0x3 // CPOL=1, CPHA=1
)

// Conn defines the interface a concrete SPI driver must implement.
//
// It is expected to implement fmt.Stringer and optionally io.Writer and
// io.Reader.
type Conn interface {
	conn.Conn
	// Speed changes the bus speed.
	Speed(hz int64) error
	// Configure changes the communication parameters of the bus.
	Configure(mode Mode, bits int) error
}

// ConnCloser is a SPI bus that can be closed.
//
// This interface is meant to be handled by the application.
type ConnCloser interface {
	io.Closer
	Conn
}

// Pins defines the pins that a SPI bus interconnect is using on the host.
//
// It is expected that a implementer of Conn also implement Pins but this is
// not a requirement.
type Pins interface {
	// CLK returns the SCK (clock) pin.
	CLK() gpio.PinOut
	// MOSI returns the SDO (master out, slave in) pin.
	MOSI() gpio.PinOut
	// MISO returns the SDI (master in, slave out) pin.
	MISO() gpio.PinIn
	// CS returns the CSN (chip select) pin.
	CS() gpio.PinOut
}

// Registry

// Opener opens an handle to a bus.
//
// It is provided by the actual bus driver.
type Opener func() (ConnCloser, error)

// Ref references an SPI bus.
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
	// Open is the factory to open an handle to this SPI bus.
	Open Opener
}

// Open opens an SPI bus by its name, an alias or its number and returns an
// handle to it.
//
// Specify the empty string "" to get the first available bus. This is the
// recommended default value unless an application knows the exact bus to use.
//
// Each bus can register multiple aliases, each leading to the same bus handle.
//
// "Bus number" is a generic concept that is highly dependent on the platform
// and OS. On some platform, the first bus may have the number 0, 1 or as high
// as 32766. Bus numbers are not necessarily continuous and may not start at 0.
// It was observed that the bus number as reported by the OS may change across
// OS revisions.
//
// When the SPI bus is provided by an off board plug and play bus like USB via
// an FT232H USB device, there can be no associated number.
func Open(name string) (ConnCloser, error) {
	var r *Ref
	var err error
	func() {
		mu.Lock()
		defer mu.Unlock()
		if len(byName) == 0 {
			err = errors.New("spi: no bus found; did you forget to call Init()?")
			return
		}
		if len(name) == 0 {
			// Asking for the default bus.
			r = defaultBus()
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
		return nil, fmt.Errorf("spi: unknown bus %q", name)
	}
	return r.Open()
}

// All returns a copy of all the registered references to all know SPI buses
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

// Register registers an SPI bus.
//
// Registering the same bus name twice is an error, e.g. o.Name(). o.Number()
// can be -1 to signify that the bus doesn't have an inherent "bus number". A
// good example is a bus provided over a FT232H device connected on an USB bus.
// In this case, the bus name should be created from the serial number of the
// device for unique identification.
func Register(name string, aliases []string, number int, o Opener) error {
	if o == nil {
		return errors.New("spi: nil Opener")
	}
	if number < -1 {
		return errors.New("spi: invalid bus number")
	}
	if len(name) == 0 {
		return errors.New("spi: empty name")
	}
	if _, err := strconv.Atoi(name); err == nil {
		return fmt.Errorf("spi: can't register an alias being only a number %q", name)
	}
	for _, alias := range aliases {
		if len(alias) == 0 {
			return errors.New("spi: empty alias")
		}
		if name == alias {
			return errors.New("spi: alias of the same name than the bus itself")
		}
		if _, err := strconv.Atoi(alias); err == nil {
			return fmt.Errorf("spi: can't register an alias being only a number %q", alias)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return fmt.Errorf("spi: registering the same bus %q twice", name)
	}
	if _, ok := byAlias[name]; ok {
		return fmt.Errorf("spi: registering the same bus %q twice", name)
	}
	if number != -1 {
		if _, ok := byNumber[number]; ok {
			return fmt.Errorf("spi: registering the same bus %d twice", number)
		}
	}
	for _, alias := range aliases {
		if _, ok := byName[alias]; ok {
			return fmt.Errorf("spi: registering the same bus %q twice", alias)
		}
		if _, ok := byAlias[alias]; ok {
			return fmt.Errorf("spi: registering the same bus %q twice", alias)
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

// Unregister removes a previously registered SPI bus.
//
// This can happen when an SPI bus is exposed via an USB device and the device
// is unplugged.
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()
	r := byName[name]
	if r == nil {
		return fmt.Errorf("spi: unknown bus name %q", name)
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

func defaultBus() *Ref {
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
	busNumber := int((^uint(0)) >> 1)
	for n, o2 := range byNumber {
		if busNumber > n {
			busNumber = n
			o = o2
		}
	}
	return o
}

type refList []*Ref

func (r refList) Len() int           { return len(r) }
func (r refList) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r refList) Less(i, j int) bool { return r[i].Name < r[j].Name }

var _ conn.Conn = Conn(nil)
