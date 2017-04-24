// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package usb implements an USB peripheral registry.
package usb

import (
	"fmt"
	"io"
	"sync"

	"periph.io/x/periph/conn"
)

// Conn represents a connection to an USB peripheral.
type Conn interface {
	conn.Conn
	ID() *ID
}

// ConnCloser is an USB peripheral handle that can be closed.
//
// This interface is meant to be handled by the USB peripheral driver, not the
// application.
type ConnCloser interface {
	io.Closer
	Conn
}

// ID represents an USB peripheral by its ID.
type ID struct {
	VenID uint16
	DevID uint16
}

func (i ID) String() string {
	return fmt.Sprintf("%04x:%04x", i.VenID, i.DevID)
}

// Opener takes control of an already opened USB peripheral.
type Opener func(dev ConnCloser) error

// Register registers a driver for an USB peripheral.
//
// When this peripheral is found, the factory will be called with a peripheral
// handle.
func Register(id ID, opener Opener) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byID[id]; ok {
		return fmt.Errorf("usb: registering the same USB peripheral id %s twice", id)
	}

	byID[id] = opener
	for _, c := range buses {
		c <- Driver{id, opener}
	}
	return nil
}

// Driver is a registered driver.
type Driver struct {
	ID // TODO(maruel): Using only the ID as the selector is a bit naive. There's known conflicts.
	Opener
}

// RegisterBus is called by a bus that will send a notification everytime
// there's a new driver being registered.
func RegisterBus(c chan<- Driver) {
	mu.Lock()
	defer mu.Unlock()
	buses = append(buses, c)
	// First start by sending all the known drivers.
	for id, opener := range byID {
		c <- Driver{id, opener}
	}
}

//

var (
	mu    sync.Mutex
	byID  = map[ID]Opener{} //
	buses []chan<- Driver   // That's highly unlikely that the number of items is not exactly 1.
)
