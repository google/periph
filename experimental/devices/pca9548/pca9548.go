// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9548

import (
	"errors"
	"strconv"
	"sync"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
)

// Dev is handle to a pca9548 I²C Multiplexer.
type Dev struct {
	c       i2c.Bus
	address uint16
	names   []string
	ports   uint8
	// mu guards port
	mu   sync.Mutex
	port uint8
}

// DefaultOpts is the recommended default options.
var DefaultOpts = Opts{Address: 0x70, Ports: 8}

// Opts is the pca9548 configuration.
type Opts struct {
	// Address pca9548 I²C Address.
	Address uint16
	// Ports number of physical ports on I²C Multiplexer.
	Ports uint8
}

// Register creates a new handel to a pca9548 I²C multiplexer, and registers
// port names with the host. These ports can then be used as any other i2c.Bus.
// The registered port names are in the form: i2cmux/mux-ADD-I where ADD is the
// multiplexer I²C address in hex and I is the port number.
// example: "i2cmux-70-0" and "mux-70-0".
func Register(bus i2c.Bus, opts *Opts) (*Dev, error) {
	d := &Dev{
		c:       bus,
		port:    0xFF,
		address: opts.Address,
		ports:   opts.Ports,
	}

	for i := uint8(0); i < opts.Ports; i++ {
		portID := strconv.FormatUint(uint64(i), 10)
		addrStr := strconv.FormatUint(uint64(opts.Address), 16)
		name := addrStr + "-" + portID
		opener := newOpener(d, i)
		if err := i2creg.Register("i2cmux-"+name, []string{"mux-" + name}, int((opts.Address*10)+uint16(i)), opener); err != nil {
			return nil, err
		}
		d.names = append(d.names, "i2cmux-"+name)
		d.names = append(d.names, "mux-"+name)
	}
	return d, nil
}

// ListPortNames lists all the port names registered for the multiplexer.
func (d *Dev) ListPortNames() []string {
	return d.names
}

// Scan scans every port of the multiplexer and every I²C address for devices,
// returns a map of I²C multiplexer ports names to slice of device I²C address.
func (d *Dev) Scan() ScanList {
	devices := make(map[string][]uint16)
	rx := []byte{0x00}
	addrStr := strconv.FormatUint(uint64(d.address), 16)
	for port := uint8(0); port < d.ports; port++ {
		portID := strconv.FormatUint(uint64(port), 10)
		for address := uint16(1); address < 0x77; address++ {
			err := d.tx(port, address, nil, rx)
			if err == nil {
				portName := "mux-" + addrStr + "-" + portID
				devices[portName] = append(devices[portName], address)
			}
		}
	}
	return devices
}

// ScanList is a map of I²C port names and I²C address of discovered devices.
type ScanList map[string][]uint16

func (l ScanList) String() string {
	s := "Scan Results:"
	for port, addresses := range l {
		s += "\nPort[" + port + "] " + strconv.FormatInt(int64(len(addresses)), 10) + " found"
		for _, addr := range addresses {
			s += "\n\tDevice at 0x" + strconv.FormatUint(uint64(addr), 16)
		}
	}
	return s
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

// tx wraps the master bus tx, maintains which port that each bus is registered
// on so that communication from the master is always on the right port.
func (d *Dev) tx(port uint8, address uint16, w, r []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if address == d.address {
		return errors.New("device address conflicts with multiplexer address")
	}
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

// String gets the port number of the bus on the multiplexer
func (p *port) String() string { return "Port:" + strconv.Itoa(int(p.number)) }

// SetSpeed is no implemented as the port slaves the master port clock
func (p *port) SetSpeed(f physic.Frequency) error { return nil }

// Tx does a transaction on the multiplexer port it is register to.
func (p *port) Tx(addr uint16, w, r []byte) error { return p.mux.tx(p.number, addr, w, r) }

// Port is a i2c.Bus on the multiplexer
type port struct {
	number uint8
	// mu guards mux
	mu  sync.Mutex
	mux *Dev
}

func (p *port) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mux = nil
	return nil
}
