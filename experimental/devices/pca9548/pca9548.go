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
	// to 0x77. The address is set by pulling A0~A2 low or high. Please refer
	// to the datasheet.
	Address int
}

// Dev is handle to a pca9548 I²C Multiplexer.
type Dev struct {
	// Immutable.
	c        i2c.Bus
	address  uint16
	name     string
	numPorts uint8

	// Mutable.
	mu         sync.Mutex
	activePort uint8
}

// New creates a new handle to a pca9548 I²C multiplexer.
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	if opts.Address < 0x70 || opts.Address > 0x77 {
		return nil, errors.New("Address outside valid range of 0x70-0x77")
	}
	d := &Dev{
		c:          bus,
		activePort: 0xFF,
		address:    uint16(opts.Address),
		numPorts:   8,
		name:       "pca9548-" + strconv.FormatUint(uint64(opts.Address), 16),
	}
	r := make([]byte, 1)
	err := bus.Tx(uint16(opts.Address), nil, r)
	if err != nil {
		return nil, errors.New("could not establish communicate with multiplexer: " + err.Error())
	}
	return d, nil
}

// RegisterPorts registers multiplexer ports with the host. These ports can
// then be used as any other i2c.Bus. Busses will be named "alias0", "alias1"
// etc. If using more than one multiplexer note that the alias must be unique.
// Returns slice of ports names registered and error.
func (d *Dev) RegisterPorts(alias string) ([]string, error) {
	var portNames []string
	for i := uint8(0); i < d.numPorts; i++ {
		portStr := strconv.Itoa(int(i))
		addrStr := strconv.FormatUint(uint64(d.address), 16)
		portName := d.c.String() + "-pca9548-" + addrStr + "-" + portStr
		opener := newOpener(d, i, alias+portStr, portName)
		if err := i2creg.Register(portName, []string{alias + portStr}, -1, opener); err != nil {
			return portNames, err
		}
		portNames = append(portNames, portName)
	}
	return portNames, nil
}

// Halt does nothing.
func (d *Dev) Halt() error {
	return nil
}

// String returns the bus base name for multiplexer ports.
func (d *Dev) String() string {
	return d.name
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
	if port != d.activePort {
		if err := d.c.Tx(d.address, []byte{1 << port}, nil); err != nil {
			return errors.New("failed to change active port on multiplexer: " + err.Error())
		}
		d.activePort = port
	}
	return d.c.Tx(address, w, r)
}

// newOpener is a helper for creating an opener func.
func newOpener(d *Dev, portNumber uint8, alias string, name string) i2creg.Opener {
	return func() (i2c.BusCloser, error) {
		return &port{
			name:   name + "(" + alias + ")",
			mux:    d,
			number: portNumber,
		}, nil
	}
}

// port is a i2c.BusCloser.
type port struct {
	// Immutable.
	name   string
	number uint8

	// Mutable.
	mu  sync.Mutex
	mux *Dev
}

// String gets the port number of the bus on the multiplexer.
func (p *port) String() string { return "Port:" + p.name }

// SetSpeed is no implemented as the port slaves the master port clock.
func (p *port) SetSpeed(f physic.Frequency) error {
	return errors.New("SetSpeed is not impelmented on a port by port basis")
}

// Tx does a transaction on the multiplexer port it is register to.
func (p *port) Tx(addr uint16, w, r []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mux == nil {
		return errors.New(p.String() + " has been closed")
	}
	return p.mux.tx(p.number, addr, w, r)
}

// Close closes a port.
func (p *port) Close() error {
	p.mu.Lock()
	p.mux = nil
	p.mu.Unlock()
	return nil
}

var _ conn.Resource = &Dev{}
var _ i2c.Bus = &port{}
