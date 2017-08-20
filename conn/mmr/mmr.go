// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package mmr defines helpers to interact with devices exposing Memory
// Mapped Registers protocol.
//
// The protocol is defined two supported commands:
//  - Write Address, Read Value
//  - Write Address, Write Value
package mmr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"

	"periph.io/x/periph/conn"
)

// Dev8 is a connection that exposes memory mapped registers in a 8bit address
// space.
type Dev8 struct {
	Conn conn.Conn
	// Order specifies the binary encoding of words. It is expected to be either
	// binary.BigEndian or binary.LittleEndian or a specialized implemented if
	// necessary. A good example of such need is devices communicating 32bits
	// little endian values encoded over 16bits big endian words.
	Order binary.ByteOrder
}

func (d *Dev8) String() string {
	return fmt.Sprintf("%s", d.Conn)
}

// ReadUint8 reads a 8 bit register.
func (d *Dev8) ReadUint8(reg uint8) (uint8, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [1]uint8
	err := d.Conn.Tx([]byte{reg}, v[:])
	return v[0], err
}

// ReadUint16 reads a 16 bit register.
func (d *Dev8) ReadUint16(reg uint8) (uint16, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [2]byte
	err := d.Conn.Tx([]byte{reg}, v[:])
	return d.Order.Uint16(v[:]), err
}

// ReadUint32 reads a 32 bit register.
func (d *Dev8) ReadUint32(reg uint8) (uint32, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [4]byte
	err := d.Conn.Tx([]byte{reg}, v[:])
	return d.Order.Uint32(v[:]), err
}

// ReadUint64 reads a 64 bit register.
func (d *Dev8) ReadUint64(reg uint8) (uint64, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [8]byte
	err := d.Conn.Tx([]byte{reg}, v[:])
	return d.Order.Uint64(v[:]), err
}

// ReadStruct writes the register number to the connection, then reads the data
// into `b` and marshall it via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *Dev8) ReadStruct(reg uint8, b interface{}) error {
	if err := d.check(); err != nil {
		return err
	}
	return readReg(d.Conn, d.Order, []byte{reg}, b)
}

// WriteUint8 writes a 8 bit register.
func (d *Dev8) WriteUint8(reg uint8, v uint8) error {
	if err := d.check(); err != nil {
		return err
	}
	return d.Conn.Tx([]byte{reg, v}, nil)
}

// WriteUint16 writes a 16 bit register.
func (d *Dev8) WriteUint16(reg uint8, v uint16) error {
	if err := d.check(); err != nil {
		return err
	}
	var a [3]byte
	a[0] = reg
	d.Order.PutUint16(a[1:], v)
	return d.Conn.Tx(a[:], nil)
}

// WriteUint32 writes a 32 bit register.
func (d *Dev8) WriteUint32(reg uint8, v uint32) error {
	if err := d.check(); err != nil {
		return err
	}
	var a [5]byte
	a[0] = reg
	d.Order.PutUint32(a[1:], v)
	return d.Conn.Tx(a[:], nil)
}

// WriteUint64 writes a 64 bit register.
func (d *Dev8) WriteUint64(reg uint8, v uint64) error {
	if err := d.check(); err != nil {
		return err
	}
	var a [9]byte
	a[0] = reg
	d.Order.PutUint64(a[1:], v)
	return d.Conn.Tx(a[:], nil)
}

// WriteStruct writes the register number to the connection, then the data
// `b` marshalled via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *Dev8) WriteStruct(reg uint8, b interface{}) error {
	if err := d.check(); err != nil {
		return err
	}
	return writeReg(d.Conn, d.Order, []byte{reg}, b)
}

func (d *Dev8) check() error {
	if d.Conn == nil {
		return errors.New("reg: missing connection")
	}
	if d.Conn.Duplex() != conn.Half {
		return errors.New("reg: connection must be half-duplex")
	}
	if d.Order == nil {
		return errors.New("reg: don't know if big or little endian")
	}
	return nil
}

//

// Dev16 is a Dev that exposes memory mapped registers in a 16bits address
// space.
type Dev16 struct {
	Conn conn.Conn
	// Order specifies the binary encoding of words. It is expected to be either
	// binary.BigEndian or binary.LittleEndian or a specialized implemented if
	// necessary. A good example of such need is devices communicating 32bits
	// little endian values encoded over 16bits big endian words.
	Order binary.ByteOrder
}

func (d *Dev16) String() string {
	return fmt.Sprintf("%s", d.Conn)
}

// ReadUint8 reads a 8 bit register.
func (d *Dev16) ReadUint8(reg uint16) (uint8, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [1]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Conn.Tx(r[:], v[:])
	return v[0], err
}

// ReadUint16 reads a 16 bit register.
func (d *Dev16) ReadUint16(reg uint16) (uint16, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [2]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Conn.Tx(r[:], v[:])
	return d.Order.Uint16(v[:]), err
}

// ReadUint32 reads a 32 bit register.
func (d *Dev16) ReadUint32(reg uint16) (uint32, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [4]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Conn.Tx(r[:], v[:])
	return d.Order.Uint32(v[:]), err
}

// ReadUint64 reads a 64 bit register.
func (d *Dev16) ReadUint64(reg uint16) (uint64, error) {
	if err := d.check(); err != nil {
		return 0, err
	}
	var v [8]byte
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	err := d.Conn.Tx(r[:], v[:])
	return d.Order.Uint64(v[:]), err
}

// ReadStruct writes the register number to the connection, then reads the data
// into `b` and marshall it via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *Dev16) ReadStruct(reg uint16, b interface{}) error {
	if err := d.check(); err != nil {
		return err
	}
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	return readReg(d.Conn, d.Order, r[:], b)
}

// WriteUint8 writes a 8 bit register.
func (d *Dev16) WriteUint8(reg uint16, v uint8) error {
	if err := d.check(); err != nil {
		return err
	}
	var r [3]byte
	d.Order.PutUint16(r[:], reg)
	r[2] = v
	return d.Conn.Tx(r[:], nil)
}

// WriteUint16 writes a 16 bit register.
func (d *Dev16) WriteUint16(reg uint16, v uint16) error {
	if err := d.check(); err != nil {
		return err
	}
	var r [4]byte
	d.Order.PutUint16(r[:], reg)
	d.Order.PutUint16(r[2:], v)
	return d.Conn.Tx(r[:], nil)
}

// WriteUint32 writes a 32 bit register.
func (d *Dev16) WriteUint32(reg uint16, v uint32) error {
	if err := d.check(); err != nil {
		return err
	}
	var r [6]byte
	d.Order.PutUint16(r[:], reg)
	d.Order.PutUint32(r[2:], v)
	return d.Conn.Tx(r[:], nil)
}

// WriteUint64 writes a 64 bit register.
func (d *Dev16) WriteUint64(reg uint16, v uint64) error {
	if err := d.check(); err != nil {
		return err
	}
	var r [10]byte
	d.Order.PutUint16(r[:], reg)
	d.Order.PutUint64(r[2:], v)
	return d.Conn.Tx(r[:], nil)
}

// WriteStruct writes the register number to the connection, then the data
// `b` marshalled via `.Order` as appropriate.
//
// It is expected to be called with a slice of integers, slice of structs,
// pointer to an integer or to a struct.
func (d *Dev16) WriteStruct(reg uint16, b interface{}) error {
	if err := d.check(); err != nil {
		return err
	}
	var r [2]byte
	d.Order.PutUint16(r[:], reg)
	return writeReg(d.Conn, d.Order, r[:], b)
}

func (d *Dev16) check() error {
	if d.Conn == nil {
		return errors.New("reg: missing connection")
	}
	if d.Conn.Duplex() != conn.Half {
		return errors.New("reg: connection must be half-duplex")
	}
	if d.Order == nil {
		return errors.New("reg: don't know if big or little endian")
	}
	return nil
}

//

func readReg(c conn.Conn, order binary.ByteOrder, reg []byte, b interface{}) error {
	if b == nil {
		return errors.New("reg: ReadRegStruct() requires a pointer or slice to an int or struct, got nil")
	}
	v := reflect.ValueOf(b)
	if !isAcceptableRead(v.Type()) {
		return fmt.Errorf("reg: ReadRegStruct() requires a slice or a pointer to a int or struct, got %s as %T", v.Kind(), b)
	}
	buf := make([]byte, int(getSize(v)))
	if err := c.Tx(reg, buf); err != nil {
		return err
	}
	if err := binary.Read(bytes.NewReader(buf), order, b); err != nil {
		return fmt.Errorf("reg: decoding failed: %s", err)
	}
	return nil
}

// writeReg writes an object `b` to register `reg`.
//
// Warning: reg is modified.
func writeReg(c conn.Conn, order binary.ByteOrder, reg []byte, b interface{}) error {
	if b == nil {
		return errors.New("reg: WriteRegStruct() requires a pointer or slice to an int or struct, got nil")
	}
	t := reflect.TypeOf(b)
	if !isAcceptableWrite(t) {
		return fmt.Errorf("reg: WriteRegStruct() requires a slice or a pointer to a int or struct, got %s as %T", t.Kind(), b)
	}
	buf := bytes.NewBuffer(reg)
	if err := binary.Write(buf, order, b); err != nil {
		return fmt.Errorf("reg: encoding failed: %s", err)
	}
	return c.Tx(buf.Bytes(), nil)
}

// isAcceptableRead returns true if the struct can be safely serialized for
// reads.
//
// TODO(maruel): Run perf tests.
func isAcceptableRead(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice:
		return isAcceptableInner(t.Elem())
	default:
		return false
	}
}

// isAcceptableWrite returns true if the struct can be safely serialized for
// writes.
//
// TODO(maruel): Run perf tests.
func isAcceptableWrite(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice:
		return isAcceptableInner(t.Elem())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
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

var _ fmt.Stringer = &Dev8{}
var _ fmt.Stringer = &Dev16{}
