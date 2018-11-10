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
	// Address pca9548 I²C Address.
	Address int
}

// Dev is handle to a pca9548 I²C Multiplexer.
type Dev struct {
	// Immutable.
	c       i2c.Bus
	address uint16

	// Mutable.
	mu   sync.Mutex
	port uint8
}

// Register creates a new handle to a pca9548 I²C multiplexer, and registers
// port names with the host. These ports can then be used as any other i2c.Bus.
// The registered port names are in the form: pca9548/mux-ADD-I where ADD is the
// multiplexer I²C address in hex and I is the port number.
// example: "pca9548-70-0" and "mux-70-0".
func Register(bus i2c.Bus, opts *Opts) (*Dev, error) {
	d := &Dev{
		c:       bus,
		port:    0xFF,
		address: uint16(opts.Address),
	}
	r := make([]byte, 1)
	err := bus.Tx(uint16(opts.Address), nil, r)
	if err != nil {
		return nil, errors.New("could not communicated with multiplexer: " + err.Error())
	}

	for i := uint8(0); i < 8; i++ {
		portID := strconv.FormatUint(uint64(i), 10)
		addrStr := strconv.FormatUint(uint64(opts.Address), 16)
		name := addrStr + "-" + portID
		opener := newOpener(d, i)
		if err := i2creg.Register("pca9548-"+name, []string{"mux-" + name}, int((opts.Address*10 + int(i))), opener); err != nil {
			return nil, err
		}
	}
	return d, nil
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
		err := d.c.Tx(d.address, []byte{uint8(1 << (port))}, nil)
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
	number uint8

	// Mutable.
	mu  sync.Mutex
	mux *Dev
}

// String gets the port number of the bus on the multiplexer.
func (p *port) String() string { return "Port:" + strconv.Itoa(int(p.number)) }

// SetSpeed is no implemented as the port slaves the master port clock.
func (p *port) SetSpeed(f physic.Frequency) error { return nil }

// Tx does a transaction on the multiplexer port it is register to.
func (p *port) Tx(addr uint16, w, r []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mux == nil {
		return errors.New("port" + strconv.FormatUint(uint64(p.number), 10) + "has been closed")
	}
	return p.mux.tx(p.number, addr, w, r)
}

func (p *port) Close() error {
	p.mux = nil
	return nil
}

// newOpener is a helper for creating an opener func.
func newOpener(d *Dev, portNumber uint8) i2creg.Opener {
	return func() (i2c.BusCloser, error) {
		return &port{
			mux:    d,
			number: portNumber,
		}, nil
	}
}

var _ conn.Resource = &Dev{}
var _ i2c.Bus = &port{}
