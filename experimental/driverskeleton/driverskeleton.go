// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package driverskeleton

import (
	"errors"

	"periph.io/x/periph"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
)

// FIXME: Expose public symbols as relevant. Do not export more than needed!
// See https://periph.io/x/periph/tree/master/doc/drivers#requirements
// for the expectations.
//
// Use the following layout for drivers:
//  - exported support symbols
//  - Opts struct
//  - New func
//  - Dev struct and methods
//  - Private support code

// New opens a handle to the device. FIXME.
func New(i i2c.Bus) (*Dev, error) {
	d := &Dev{&i2c.Dev{Bus: i, Addr: 42}}
	// FIXME: Simulate a setup dance.
	var b [2]byte
	if err := d.c.Tx([]byte("in"), b[:]); err != nil {
		return nil, err
	}
	if b[0] != 'I' || b[1] != 'N' {
		return nil, errors.New("driverskeleton: unexpected reply")
	}
	return d, nil
}

// Dev is a handle to the device. FIXME.
type Dev struct {
	c conn.Conn
}

// Read is a method on your device. FIXME.
func (d *Dev) Read() string {
	var b [12]byte
	if err := d.c.Tx([]byte("what"), b[:]); err != nil {
		return err.Error()
	}
	return string(b[:])
}

//

// FIXME: A driver is generally only needed for host drivers. If you implement
// a device driver, delete the remainder of this file.

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	// FIXME: Change this string to be unique. It must match the directory name.
	return "driverskeleton"
}

func (d *driver) Prerequisites() []string {
	// FIXME: Declare prerequisites drivers if relevant.
	return nil
}

func (d *driver) After() []string {
	// FIXME: Declare drivers that should be loaded before, if present, but are
	// not required.
	return nil
}

func (d *driver) Init() (bool, error) {
	// FIXME: If the driver is not needed, do the following:
	// return false, errors.New("not running on a skeleton")

	// FIXME: Add implementation.

	return true, errors.New("driverskeleton: not implemented")
}

func init() {
	// Since isArm is a compile time constant, the compile can strip the
	// unnecessary code and unused private symbols.
	//if isArm {
	periph.MustRegister(&driver{})
	//}
}

// FIXME: This verifies that the driver implements all the required methods.
var _ periph.Driver = &driver{}
