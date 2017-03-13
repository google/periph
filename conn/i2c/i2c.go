// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package i2c defines an I²C bus.
//
// It includes an adapter to directly address an I²C device on a I²C bus
// without having to continuously specify the address when doing I/O. This
// enables the support of conn.Conn.
package i2c

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"

	"periph.io/x/periph/conn/gpio"
)

// Bus defines the interface a concrete I²C driver must implement.
//
// This interface is consummed by a device driver for a device sitting on a bus.
//
// This interface doesn't implement conn.Conn since a device address must be
// specified. Use i2cdev.Dev as an adapter to get a conn.Conn compatible
// object.
type Bus interface {
	fmt.Stringer
	Tx(addr uint16, w, r []byte) error
	// Speed changes the bus speed, if supported.
	Speed(hz int64) error
}

// BusCloser is an I²C bus that can be closed.
//
// This interface is meant to be handled by the application.
type BusCloser interface {
	io.Closer
	Bus
}

// Pins defines the pins that an I²C bus interconnect is using on the host.
//
// It is expected that a implementer of Bus also implement Pins but this is not
// a requirement.
type Pins interface {
	// SCL returns the CLK (clock) pin.
	SCL() gpio.PinIO
	// SDA returns the DATA pin.
	SDA() gpio.PinIO
}

// Dev is a device on a I²C bus.
//
// It implements conn.Conn.
//
// It saves from repeatedly specifying the device address.
type Dev struct {
	Bus  Bus
	Addr uint16
}

func (d *Dev) String() string {
	return fmt.Sprintf("%s(%d)", d.Bus, d.Addr)
}

// Tx does a transaction by adding the device's address to each command.
//
// It's a wrapper for Bus.Tx().
func (d *Dev) Tx(w, r []byte) error {
	return d.Bus.Tx(d.Addr, w, r)
}

// Write writes to the I²C bus without reading, implementing io.Writer.
//
// It's a wrapper for Tx()
func (d *Dev) Write(b []byte) (int, error) {
	if err := d.Tx(b, nil); err != nil {
		return 0, err
	}
	return len(b), nil
}

//

// DevReg8 is a Dev that exposes memory mapped registers in a 8bit address
// space.
type DevReg8 struct {
	Dev
	// Order specifies the binary encoding of words. It is expected to be either
	// binary.BigEndian or binary.LittleEndian or a specialized implemented if
	// necessary. A good example of such need is devices communicating 32bits
	// little endian values encoded over 16bits big endian words.
	Order binary.ByteOrder
}

// ReadRegUint8 reads a 8 bit register.
func (d *DevReg8) ReadRegUint8(reg uint8) (uint8, error) {
	var v [1]uint8
	err := d.Tx([]byte{reg}, v[:])
	return v[0], err
}

// ReadRegUint16 reads a 16 bit register.
func (d *DevReg8) ReadRegUint16(reg uint8) (uint16, error) {
	if d.Order == nil {
		return 0, errors.New("i2c: don't know if big or little endian")
	}
	var v [2]byte
	err := d.Tx([]byte{reg}, v[:])
	return d.Order.Uint16(v[:]), err
}

// ReadRegUint32 reads a 32 bit register.
func (d *DevReg8) ReadRegUint32(reg uint8) (uint32, error) {
	if d.Order == nil {
		return 0, errors.New("i2c: don't know if big or little endian")
	}
	var v [4]byte
	err := d.Tx([]byte{reg}, v[:])
	return d.Order.Uint32(v[:]), err
}

// ReadRegUint64 reads a 64 bit register.
func (d *DevReg8) ReadRegUint64(reg uint8) (uint64, error) {
	if d.Order == nil {
		return 0, errors.New("i2c: don't know if big or little endian")
	}
	var v [8]byte
	err := d.Tx([]byte{reg}, v[:])
	return d.Order.Uint64(v[:]), err
}

// ReadRegStruct writes the register number to the I²C bus, then reads the data
// into `b` and marshall it via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *DevReg8) ReadRegStruct(reg uint8, b interface{}) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	return readReg(&d.Dev, d.Order, []byte{reg}, b)
}

// WriteRegUint8 writes a 8 bit register.
func (d *DevReg8) WriteRegUint8(reg uint8, v uint8) error {
	return d.Tx([]byte{reg, v}, nil)
}

// WriteRegUint16 writes a 16 bit register.
func (d *DevReg8) WriteRegUint16(reg uint8, v uint16) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var a [3]byte
	a[0] = reg
	d.Order.PutUint16(a[1:], v)
	return d.Tx(a[:], nil)
}

// WriteRegUint32 writes a 32 bit register.
func (d *DevReg8) WriteRegUint32(reg uint8, v uint32) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var a [5]byte
	a[0] = reg
	d.Order.PutUint32(a[1:], v)
	return d.Tx(a[:], nil)
}

// WriteRegUint64 writes a 64 bit register.
func (d *DevReg8) WriteRegUint64(reg uint8, v uint64) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var a [9]byte
	a[0] = reg
	d.Order.PutUint64(a[1:], v)
	return d.Tx(a[:], nil)
}

// WriteRegStruct writes the register number to the I²C bus, then the data
// `b` marshalled via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *DevReg8) WriteRegStruct(reg uint8, b interface{}) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	return writeReg(&d.Dev, d.Order, []byte{reg}, b)
}

//

// DevReg16 is a Dev that exposes memory mapped registers in a 16bits address
// space.
type DevReg16 struct {
	Dev
	// Order specifies the binary encoding of words. It is expected to be either
	// binary.BigEndian or binary.LittleEndian or a specialized implemented if
	// necessary. A good example of such need is devices communicating 32bits
	// little endian values encoded over 16bits big endian words.
	Order binary.ByteOrder
}

// ReadRegUint8 reads a 8 bit register.
func (d *DevReg16) ReadRegUint8(reg uint16) (uint8, error) {
	if d.Order == nil {
		return 0, errors.New("i2c: don't know if big or little endian")
	}
	var v [1]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Tx(r[:], v[:])
	return v[0], err
}

// ReadRegUint16 reads a 16 bit register.
func (d *DevReg16) ReadRegUint16(reg uint16) (uint16, error) {
	if d.Order == nil {
		return 0, errors.New("i2c: don't know if big or little endian")
	}
	var v [2]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Tx(r[:], v[:])
	return d.Order.Uint16(v[:]), err
}

// ReadRegUint32 reads a 32 bit register.
func (d *DevReg16) ReadRegUint32(reg uint16) (uint32, error) {
	if d.Order == nil {
		return 0, errors.New("i2c: don't know if big or little endian")
	}
	var v [4]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Tx(r[:], v[:])
	return d.Order.Uint32(v[:]), err
}

// ReadRegUint64 reads a 64 bit register.
func (d *DevReg16) ReadRegUint64(reg uint16) (uint64, error) {
	if d.Order == nil {
		return 0, errors.New("i2c: don't know if big or little endian")
	}
	var v [8]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Tx(r[:], v[:])
	return d.Order.Uint64(v[:]), err
}

// ReadRegStruct writes the register number to the I²C bus, then reads the data
// into `b` and marshall it via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *DevReg16) ReadRegStruct(reg uint16, b interface{}) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	return readReg(&d.Dev, d.Order, r[:], b)
}

// WriteRegUint8 writes a 8 bit register.
func (d *DevReg16) WriteRegUint8(reg uint16, v uint8) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var r [3]byte
	d.Order.PutUint16(r[:], reg)
	r[2] = v
	return d.Tx(r[:], nil)
}

// WriteRegUint16 writes a 16 bit register.
func (d *DevReg16) WriteRegUint16(reg uint16, v uint16) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var r [4]byte
	d.Order.PutUint16(r[:], reg)
	d.Order.PutUint16(r[2:], v)
	return d.Tx(r[:], nil)
}

// WriteRegUint32 writes a 32 bit register.
func (d *DevReg16) WriteRegUint32(reg uint16, v uint32) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var r [6]byte
	d.Order.PutUint16(r[:], reg)
	d.Order.PutUint32(r[2:], v)
	return d.Tx(r[:], nil)
}

// WriteRegUint64 writes a 64 bit register.
func (d *DevReg16) WriteRegUint64(reg uint16, v uint64) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var r [10]byte
	d.Order.PutUint16(r[:], reg)
	d.Order.PutUint64(r[2:], v)
	return d.Tx(r[:], nil)
}

// WriteRegStruct writes the register number to the I²C bus, then the data
// `b` marshalled via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *DevReg16) WriteRegStruct(reg uint16, b interface{}) error {
	if d.Order == nil {
		return errors.New("i2c: don't know if big or little endian")
	}
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	return writeReg(&d.Dev, d.Order, r[:], b)
}

//

// All returns all the I²C buses available on this host.
func All() map[string]Opener {
	mu.Lock()
	defer mu.Unlock()
	out := make(map[string]Opener, len(byName))
	for k, v := range byName {
		out[k] = v
	}
	return out
}

// New returns an open handle to the first available I²C bus.
//
// Specify busNumber -1 to get the first available bus. This is the recommended
// value.
func New(busNumber int) (BusCloser, error) {
	opener, err := find(busNumber)
	if err != nil {
		return nil, err
	}
	return opener()
}

// Opener opens an I²C bus.
type Opener func() (BusCloser, error)

// Register registers an I²C bus.
//
// Registering the same bus name twice is an error.
func Register(name string, busNumber int, opener Opener) error {
	if opener == nil {
		return errors.New("i2c: nil opener")
	}
	if len(name) == 0 {
		return errors.New("i2c: empty name")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; ok {
		return fmt.Errorf("i2c: registering the same bus %s twice", name)
	}
	if busNumber != -1 {
		if _, ok := byNumber[busNumber]; ok {
			return fmt.Errorf("i2c: registering the same bus %d twice", busNumber)
		}
	}

	byName[name] = opener
	if busNumber != -1 {
		byNumber[busNumber] = opener
	}
	return nil
}

// Unregister removes a previously registered I²C bus.
//
// This can happen when an I²C bus is exposed via an USB device and the device
// is unplugged.
func Unregister(name string, busNumber int) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := byName[name]; !ok {
		return fmt.Errorf("i2c: unknown bus name %q", name)
	}
	if _, ok := byNumber[busNumber]; !ok {
		return fmt.Errorf("i2c: unknown bus number %d", busNumber)
	}

	delete(byName, name)
	delete(byNumber, busNumber)
	return nil
}

//

func readReg(d *Dev, order binary.ByteOrder, reg []byte, b interface{}) error {
	if b == nil {
		return errors.New("i2c: ReadRegStruct() requires a pointer or slice to an int or struct, got nil")
	}
	v := reflect.ValueOf(b)
	if !isAcceptable(v.Type()) {
		return fmt.Errorf("i2c: ReadRegStruct() requires a slice or a pointer to a int or struct, got %s", v.Kind())
	}
	buf := make([]byte, int(getSize(v)))
	if err := d.Tx(reg, buf); err != nil {
		return err
	}
	if err := binary.Read(bytes.NewReader(buf), order, b); err != nil {
		return fmt.Errorf("i2c: decoding failed: %s", err)
	}
	return nil
}

// writeReg writes an object `b` to register `reg`.
//
// Warning: reg is modified.
func writeReg(d *Dev, order binary.ByteOrder, reg []byte, b interface{}) error {
	if b == nil {
		return errors.New("i2c: ReadRegStruct() requires a pointer or slice to an int or struct, got nil")
	}
	t := reflect.TypeOf(b)
	if !isAcceptable(t) {
		return fmt.Errorf("i2c: ReadRegStruct() requires a slice or a pointer to a int or struct, got %s", t.Kind())
	}
	buf := bytes.NewBuffer(reg)
	if err := binary.Write(buf, order, b); err != nil {
		return fmt.Errorf("i2c: encoding failed: %s", err)
	}
	return d.Tx(buf.Bytes(), nil)
}

// isAcceptable returns true if the struct can be safely serialized.
// TODO(maruel): Keep a cache because this is somewhat expensive.
// TODO(maruel): Run perf tests.
func isAcceptable(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice:
		return isAcceptableInner(t.Elem())
	default:
		return false
	}
}

func getSize(v reflect.Value) int {
	switch v.Kind() {
	case reflect.Ptr:
		return int(v.Type().Elem().Size())
	case reflect.Slice:
		// TODO(maruel): Misaligned items.
		return int(v.Type().Elem().Size()) * v.Len()
	default:
		return 0
	}
}
func isAcceptableInner(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Array:
		return isAcceptableInner(t.Elem())
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if f := t.Field(i); !isAcceptableInner(f.Type) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func find(busNumber int) (Opener, error) {
	mu.Lock()
	defer mu.Unlock()
	if len(byNumber) == 0 {
		return nil, errors.New("i2c: no bus found; did you forget to call Init()?")
	}
	if busNumber == -1 {
		busNumber = int((^uint(0)) >> 1)
		for n := range byNumber {
			if busNumber > n {
				busNumber = n
			}
		}
	}
	bus, ok := byNumber[busNumber]
	if !ok {
		return nil, fmt.Errorf("i2c: no bus %d", busNumber)
	}
	return bus, nil
}

var (
	mu       sync.Mutex
	byName   = map[string]Opener{}
	byNumber = map[int]Opener{}
)
