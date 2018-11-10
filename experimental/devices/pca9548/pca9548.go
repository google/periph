// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9548

import (
	"errors"
	"strconv"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
)

// DefaultOpts is the recommended default options.
var DefaultOpts = Opts{Address: 0x70}

// Opts is the pca9548 configuration.
type Opts struct {
	// Address pca9548 I²C Address. Valid addresses for the NXP pca9548 are 0x70
	// to 0x77.
	Address int
}

// Dev is handle to a pca9548 I²C Multiplexer.
type Dev struct {
	// Immutable.
	c       i2c.Bus
	address uint16

	// Mutable.
	mu       sync.Mutex
	port     uint8
	numPorts uint8
}

// New creates a new handle to a pca9548 I²C multiplexer.
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	d := &Dev{
		c:       bus,
		port:    0xFF,
		address: uint16(opts.Address),
		// TODO(NeuralSpaz): Make number of ports safely settable after New()
		numPorts: 8,
	}
	r := make([]byte, 1)
	err := bus.Tx(uint16(opts.Address), nil, r)
	if err != nil {
		return nil, errors.New("could not communicate with multiplexer: " + err.Error())
	}
	return d, nil
}

// Register registers port number and port name with the host. These ports can
// then be used as any other i2c.Bus. Alias bus name that needs to be unique.
func (d *Dev) Register(port int, alias string) error {
	if port >= int(d.numPorts) || port < 0 {
		return errors.New("port number must be between 0 and 7")
	}

	portID := strconv.Itoa(int(port))
	addrStr := strconv.FormatUint(uint64(d.address), 16)
	name := d.c.String() + "-pca9548-" + addrStr + "-" + portID

	opener := newOpener(d, uint8(port), alias)
	return i2creg.Register(name, []string{alias}, int(d.address)*10+port, opener)
}

// Halt does nothing.
func (d *Dev) Halt() error {
	// TODO(NeuralSpaz): Find a good way to halt that also can resume.
	return nil
}

// String returns the bus base name for multiplexer ports.
func (d *Dev) String() string {
	return "pca9548-" + strconv.FormatUint(uint64(d.address), 16)
}

// tx wraps the master bus tx, maintains which port that each bus is registered
// on so that communication from the master is always on the right port.
func (d *Dev) tx(port uint8, address uint16, w, r []byte) error {
	if address == d.address {
		return errors.New("device address conflicts with multiplexer address")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	// Change active port if needed.
	if port != d.port {
		err := d.c.Tx(d.address, []byte{uint8(1 << port)}, nil)
		if err != nil {
			return errors.New("failed to change active port on multiplexer: " + err.Error())
		}
		d.port = port
	}
	return d.c.Tx(address, w, r)
}

// Port is a i2c.Bus on the multiplexer.
type port struct {
	// Immutable.
	name   string
	number uint8

	// Mutable.
	mu  sync.Mutex
	mux *Dev
}

// String gets the port number of the bus on the multiplexer.
func (p *port) String() string { return "Port:" + p.name + " " + strconv.Itoa(int(p.number)) }

// SetSpeed is no implemented as the port slaves the master port clock.
func (p *port) SetSpeed(f physic.Frequency) error { return nil }

// Tx does a transaction on the multiplexer port it is register to.
func (p *port) Tx(addr uint16, w, r []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mux == nil {
		return errors.New("port " + strconv.Itoa(int(p.number)) + " has been closed")
	}
	return p.mux.tx(p.number, addr, w, r)
}

// Close closes a port, no futher communication can take place on that port.
func (p *port) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mux = nil
	return nil
}

// newOpener is a helper for creating an opener func.
func newOpener(d *Dev, portNumber uint8, alias string) i2creg.Opener {
	return func() (i2c.BusCloser, error) {
		return &port{
			name:   alias,
			mux:    d,
			number: portNumber,
		}, nil
	}
}

var _ conn.Resource = &Dev{}
var _ i2c.Bus = &port{}
