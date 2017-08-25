// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mmr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/onewire/onewirereg"
)

func ExampleDev8() {
	// Open a connection, using I²C as an example:
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()
	c := &i2c.Dev{Bus: bus, Addr: 0xD0}

	dev := Dev8{c, binary.BigEndian}
	v, err := dev.ReadUint8(0xD0)
	if err != nil {
		log.Fatal(err)
	}
	if v == 0x60 {
		fmt.Printf("Found bme280 on bus %s\n", bus)
	}
}

func ExampleDev8_ReadStruct() {
	// Open a connection, using I²C as an example:
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()
	c := &i2c.Dev{Bus: bus, Addr: 0xD0}

	dev := Dev8{c, binary.BigEndian}
	flags := struct {
		Flag16 uint16
		Flag8  [2]uint8
	}{}
	if err = dev.ReadStruct(0xD0, &flags); err != nil {
		log.Fatal(err)
	}
	// Use flags.Flag16 and flags.Flag8.
}

func ExampleDev8_WriteStruct() {
	// Open a connection, using 1-wire as an example:
	bus, err := onewirereg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()
	c := &onewire.Dev{Bus: bus, Addr: 0xD0}

	dev := Dev8{c, binary.LittleEndian}
	flags := struct {
		Flag16 uint16
		Flag8  [2]uint8
	}{
		0x1234,
		[2]uint8{1, 2},
	}
	if err = dev.WriteStruct(0xD0, &flags); err != nil {
		log.Fatal(err)
	}
}

//

func TestDev8_String(t *testing.T) {
	d := Dev8{Conn: &conntest.Discard{D: conn.Full}, Order: nil}
	if s := d.String(); s != "discard" {
		t.Fatal(s)
	}
}

func TestDev8_ReadUint_Duplex(t *testing.T) {
	d := Dev8{Conn: &conntest.Discard{D: conn.Full}, Order: nil}
	if v, err := d.ReadUint8(34); err == nil || v != 0 {
		t.Fatal(v, err)
	}
}

func TestDev8_ReadUint_Order_nil(t *testing.T) {
	d := Dev8{Conn: &conntest.Discard{D: conn.Half}, Order: nil}
	if v, err := d.ReadUint8(34); err == nil || v != 0 {
		t.Fatal(v, err)
	}
	if v, err := d.ReadUint16(34); err == nil || v != 0 {
		t.Fatal(v, err)
	}
	if v, err := d.ReadUint32(34); err == nil || v != 0 {
		t.Fatal(v, err)
	}
	if v, err := d.ReadUint64(34); err == nil || v != 0 {
		t.Fatal(v, err)
	}
}

func TestDev8_ReadUint(t *testing.T) {
	r := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: r[:1]}}, D: conn.Half}
	d := Dev8{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint8(34); err != nil || v != 0x01 {
		t.Fatal(v, err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: r[:2]}}, D: conn.Half}
	d = Dev8{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint16(34); err != nil || v != d.Order.Uint16(r) {
		t.Fatal(v, err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: r[:4]}}, D: conn.Half}
	d = Dev8{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint32(34); err != nil || v != d.Order.Uint32(r) {
		t.Fatal(v, err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: r[:8]}}, D: conn.Half}
	d = Dev8{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint64(34); err != nil || v != d.Order.Uint64(r) {
		t.Fatal(v, err)
	}
}

func TestDev8_ReadStruct_Order_nil(t *testing.T) {
	d := Dev8{Conn: &conntest.Discard{D: conn.Half}, Order: nil}
	if d.ReadStruct(34, &packed{}) == nil {
		t.Fatal("Order is nil")
	}
}

func TestDev8_ReadStruct_Precond_Fail(t *testing.T) {
	d := Dev8{Conn: &conntest.Playback{D: conn.Half}, Order: binary.LittleEndian}
	if d.ReadStruct(34, nil) == nil {
		t.Fatal("nil")
	}
	if d.ReadStruct(34, 1) == nil {
		t.Fatal("int")
	}
	x := [2]string{}
	if d.ReadStruct(34, &x) == nil {
		t.Fatal("pointer to array (not slice)")
	}
	y := struct {
		i *int
	}{}
	if d.ReadStruct(34, &y) == nil {
		t.Fatal("struct with int")
	}
}

func TestDev8_ReadStruct_Decode_fail(t *testing.T) {
	d := Dev8{Conn: &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: []byte{}}}, D: conn.Half}, Order: binary.LittleEndian}
	z := [0]int{}
	if err := d.ReadStruct(34, &z); err == nil {
		t.Fatal()
	}
	if err := d.ReadStruct(34, 1); err == nil {
		t.Fatal()
	}
}

func TestDev8_ReadStruct_struct(t *testing.T) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}); err != nil {
		t.Fatal(err)
	}
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: buf.Bytes()}}, D: conn.Half}
	d := Dev8{Conn: c, Order: binary.LittleEndian}
	p := &packed{}
	if err := d.ReadStruct(34, p); err != nil {
		t.Fatal(err)
	}
	if p.U64 != 0x123456789abcdef0 {
		t.Fatalf("u64: %v", p.U64)
	}
	if p.U32 != 0x12345678 {
		t.Fatalf("u32: %v", p.U32)
	}
	if p.U16 != 0x1234 {
		t.Fatalf("u16: %v", p.U16)
	}
	if p.U8[0] != 0x12 || p.U8[1] != 0x01 {
		t.Fatalf("u8: %v", p.U8)
	}
}

func TestDev8_ReadStruct_slice(t *testing.T) {
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: []byte{1, 2}}}, D: conn.Half}
	d := Dev8{Conn: c, Order: binary.LittleEndian}
	p := make([]uint8, 2)
	if err := d.ReadStruct(34, p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p, []uint8{1, 2}) {
		t.Fatal(p)
	}
}

func TestDev8_WriteUint_nil(t *testing.T) {
	d := Dev8{Conn: nil, Order: binary.BigEndian}
	if d.WriteUint8(34, 1) == nil {
		t.Fatal("Conn is nil")
	}
	d = Dev8{Conn: &conntest.Discard{D: conn.Half}, Order: nil}
	if d.WriteUint8(34, 1) == nil {
		t.Fatal("Order is nil")
	}
	if d.WriteUint16(34, 1) == nil {
		t.Fatal("Order is nil")
	}
	if d.WriteUint32(34, 1) == nil {
		t.Fatal("Order is nil")
	}
	if d.WriteUint64(34, 1) == nil {
		t.Fatal("Order is nil")
	}
}

func TestDev8_WriteUint(t *testing.T) {
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{34, 56}}}, D: conn.Half}
	d := Dev8{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint8(34, 56); err != nil {
		t.Fatal(err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{34, 0x78, 0x56}}}, D: conn.Half}
	d = Dev8{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint16(34, 0x5678); err != nil {
		t.Fatal(err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{34, 0xbc, 0x9a, 0x78, 0x56}}}, D: conn.Half}
	d = Dev8{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint32(34, 0x56789abc); err != nil {
		t.Fatal(err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{34, 0x34, 0x12, 0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56}}}, D: conn.Half}
	d = Dev8{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint64(34, 0x56789abcdef01234); err != nil {
		t.Fatal(err)
	}
}

func TestDev8_WriteStruct_Order_nil(t *testing.T) {
	d := Dev8{Conn: &conntest.Discard{D: conn.Half}, Order: nil}
	if err := d.WriteStruct(34, &packed{}); err == nil {
		t.Fatal()
	}
}

func TestDev8_WriteStruct_Precond_Fail(t *testing.T) {
	d := Dev8{Conn: &conntest.Playback{D: conn.Half}, Order: binary.LittleEndian}
	if err := d.WriteStruct(34, nil); err == nil {
		t.Fatal()
	}
	if err := d.WriteStruct(34, 1); err == nil {
		t.Fatal()
	}
	// TODO(maruel): Pointer to arrays could be supported.
	x := [2]string{}
	if err := d.WriteStruct(34, &x); err == nil {
		t.Fatal()
	}
	y := struct {
		i *int
	}{}
	if err := d.WriteStruct(34, &y); err == nil {
		t.Fatal()
	}
	z := [0]int{}
	if err := d.WriteStruct(34, &z); err == nil {
		t.Fatal()
	}
}

func TestDev8_WriteStruct(t *testing.T) {
	c := &conntest.Playback{
		Ops: []conntest.IO{
			{
				W: []byte{
					34,
					0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
					0x12, 0x34, 0x56, 0x78,
					0x12, 0x34,
					0x12, 0x01,
				},
			},
		},
		D: conn.Half,
	}
	d := Dev8{Conn: c, Order: binary.BigEndian}
	p := &packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}
	if err := d.WriteStruct(34, p); err != nil {
		t.Fatal(err)
	}
}

func TestDev8_WriteStruct_uint16(t *testing.T) {
	c := &conntest.Playback{
		Ops: []conntest.IO{
			{
				W: []byte{
					34,
					0x12, 0x34,
				},
			},
		},
		D: conn.Half,
	}
	d := Dev8{Conn: c, Order: binary.BigEndian}
	if err := d.WriteStruct(34, uint16(0x1234)); err != nil {
		t.Fatal(err)
	}
}

//

func TestDev16_String(t *testing.T) {
	d := Dev16{Conn: &conntest.Discard{D: conn.Full}, Order: nil}
	if s := d.String(); s != "discard" {
		t.Fatal(s)
	}
}

func TestDev16_ReadUint_Duplex(t *testing.T) {
	d := Dev16{Conn: &conntest.Discard{D: conn.Full}, Order: nil}
	if v, err := d.ReadUint8(34); err == nil || v != 0 {
		t.Fatal(v, err)
	}
}

func TestDev16_ReadUint_Order_nil(t *testing.T) {
	d := Dev16{Conn: &conntest.Playback{D: conn.Half}, Order: nil}
	if v, err := d.ReadUint8(0x1234); err == nil || v != 0 {
		t.Fatal(v, err)
	}
	if v, err := d.ReadUint16(0x1234); err == nil || v != 0 {
		t.Fatal(v, err)
	}
	if v, err := d.ReadUint32(0x1234); err == nil || v != 0 {
		t.Fatal(v, err)
	}
	if v, err := d.ReadUint64(0x1234); err == nil || v != 0 {
		t.Fatal(v, err)
	}
}

func TestDev16_ReadUint(t *testing.T) {
	r := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{0x12, 0x34}, R: r[:1]}}, D: conn.Half}
	d := Dev16{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint8(0x1234); err != nil || v != 0x01 {
		t.Fatal(v, err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{0, 34}, R: r[:2]}}, D: conn.Half}
	d = Dev16{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint16(34); err != nil || v != d.Order.Uint16(r) {
		t.Fatal(v, err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{0, 34}, R: r[:4]}}, D: conn.Half}
	d = Dev16{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint32(34); err != nil || v != d.Order.Uint32(r) {
		t.Fatal(v, err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{0, 34}, R: r[:8]}}, D: conn.Half}
	d = Dev16{Conn: c, Order: binary.BigEndian}
	if v, err := d.ReadUint64(34); err != nil || v != d.Order.Uint64(r) {
		t.Fatal(v, err)
	}
}

func TestDev16_ReadStruct_Order_nil(t *testing.T) {
	d := Dev16{Conn: &conntest.Discard{D: conn.Half}, Order: nil}
	if err := d.ReadStruct(0x1234, &packed{}); err == nil {
		t.Fatal()
	}
}

func TestDev16_ReadStruct_Precond_Fail(t *testing.T) {
	d := Dev16{
		Conn:  &conntest.Playback{D: conn.Half, DontPanic: true},
		Order: binary.LittleEndian,
	}
	if err := d.ReadStruct(0x1234, nil); err == nil {
		t.Fatal()
	}
	if err := d.ReadStruct(0x1234, 1); err == nil {
		t.Fatal()
	}
	// TODO(maruel): Pointer to arrays could be supported.
	x := [2]string{}
	if err := d.ReadStruct(0x1234, &x); err == nil {
		t.Fatal()
	}
	y := struct {
		i *int
	}{}
	if err := d.ReadStruct(0x1234, &y); err == nil {
		t.Fatal()
	}
	z := [0]int{}
	if err := d.ReadStruct(0x1234, &z); err == nil {
		t.Fatal()
	}
}

func TestDev16_ReadStruct_Decode_fail(t *testing.T) {
	d := Dev16{
		Conn:  &conntest.Playback{Ops: []conntest.IO{{W: []byte{34}, R: []byte{}}}, D: conn.Half, DontPanic: true},
		Order: binary.LittleEndian,
	}
	z := [0]int{}
	if err := d.ReadStruct(34, &z); err == nil {
		t.Fatal()
	}
}

func TestDev16_ReadStruct_struct(t *testing.T) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}); err != nil {
		t.Fatal(err)
	}
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{0x34, 0x12}, R: buf.Bytes()}}, D: conn.Half}
	d := Dev16{Conn: c, Order: binary.LittleEndian}
	p := &packed{}
	if err := d.ReadStruct(0x1234, p); err != nil {
		t.Fatal(err)
	}
	if p.U64 != 0x123456789abcdef0 {
		t.Fatalf("u64: %v", p.U64)
	}
	if p.U32 != 0x12345678 {
		t.Fatalf("u32: %v", p.U32)
	}
	if p.U16 != 0x1234 {
		t.Fatalf("u16: %v", p.U16)
	}
	if p.U8[0] != 0x12 || p.U8[1] != 0x01 {
		t.Fatalf("u8: %v", p.U8)
	}
}

func TestDev16_ReadStruct_fail(t *testing.T) {
	d := Dev16{Conn: &conntest.RecordRaw{W: writeFail{}}, Order: binary.LittleEndian}
	if d.ReadStruct(34, &packed{}) == nil {
		t.Fatal()
	}
}

func TestDev16_ReadStruct_slice(t *testing.T) {
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{0x34, 0x12}, R: []byte{1, 2}}}, D: conn.Half}
	d := Dev16{Conn: c, Order: binary.LittleEndian}
	p := make([]uint8, 2)
	if err := d.ReadStruct(0x1234, p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p, []uint8{1, 2}) {
		t.Fatal(p)
	}
}

func TestDev16_WriteUint_nil(t *testing.T) {
	d := Dev16{Conn: nil, Order: binary.BigEndian}
	if d.WriteUint8(34, 1) == nil {
		t.Fatal("Conn is nil")
	}
	d = Dev16{Conn: &conntest.Discard{D: conn.Half}, Order: nil}
	if d.WriteUint8(0x1234, 1) == nil {
		t.Fatal("Order is nil")
	}
	if d.WriteUint16(0x1234, 1) == nil {
		t.Fatal("Order is nil")
	}
	if d.WriteUint32(0x1234, 1) == nil {
		t.Fatal("Order is nil")
	}
	if d.WriteUint64(0x1234, 1) == nil {
		t.Fatal("Order is nil")
	}
}

func TestDev16_WriteUint(t *testing.T) {
	c := &conntest.Playback{Ops: []conntest.IO{{W: []byte{0x34, 0x12, 56}}}, D: conn.Half}
	d := Dev16{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint8(0x1234, 56); err != nil {
		t.Fatal(err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{0x34, 0x12, 0x78, 0x56}}}, D: conn.Half}
	d = Dev16{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint16(0x1234, 0x5678); err != nil {
		t.Fatal(err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{0x34, 0x12, 0xbc, 0x9a, 0x78, 0x56}}}, D: conn.Half}
	d = Dev16{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint32(0x1234, 0x56789abc); err != nil {
		t.Fatal(err)
	}
	c = &conntest.Playback{Ops: []conntest.IO{{W: []byte{0x34, 0x12, 0x34, 0x12, 0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56}}}, D: conn.Half}
	d = Dev16{Conn: c, Order: binary.LittleEndian}
	if err := d.WriteUint64(0x1234, 0x56789abcdef01234); err != nil {
		t.Fatal(err)
	}
}

func TestDev16_WriteStruct_Order_nil(t *testing.T) {
	d := Dev16{Conn: &conntest.Discard{D: conn.Half}, Order: nil}
	if err := d.WriteStruct(0x1234, &packed{}); err == nil {
		t.Fatal()
	}
}

func TestDev16_WriteStruct_Precond_Fail(t *testing.T) {
	d := Dev16{Conn: &conntest.Playback{D: conn.Half}, Order: binary.LittleEndian}
	if err := d.WriteStruct(0x1234, nil); err == nil {
		t.Fatal()
	}
	if err := d.WriteStruct(0x1234, 1); err == nil {
		t.Fatal()
	}
	x := [2]string{}
	if err := d.WriteStruct(0x1234, &x); err == nil {
		t.Fatal()
	}
	y := struct {
		i *int
	}{}
	if err := d.WriteStruct(0x1234, &y); err == nil {
		t.Fatal()
	}
	z := [0]int{}
	if err := d.WriteStruct(0x1234, &z); err == nil {
		t.Fatal()
	}
}

func TestDev16_WriteStruct(t *testing.T) {
	c := &conntest.Playback{
		Ops: []conntest.IO{
			{
				W: []byte{
					0x12, 0x34,
					0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
					0x12, 0x34, 0x56, 0x78,
					0x12, 0x34,
					0x12, 0x01,
				},
			},
		},
		D: conn.Half,
	}
	d := Dev16{Conn: c, Order: binary.BigEndian}
	p := &packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}
	if err := d.WriteStruct(0x1234, p); err != nil {
		t.Fatal(err)
	}
}

func TestDev16_WriteStruct_uint16(t *testing.T) {
	c := &conntest.Playback{
		Ops: []conntest.IO{
			{
				W: []byte{
					0x12, 0x34,
					0x56, 0x78,
				},
			},
		},
		D: conn.Half,
	}
	d := Dev16{Conn: c, Order: binary.BigEndian}
	if err := d.WriteStruct(0x1234, uint16(0x5678)); err != nil {
		t.Fatal(err)
	}
}

//

func TestEdgeCases(t *testing.T) {
	if getSize(reflect.ValueOf(nil)) != 0 {
		t.FailNow()
	}
}

//

type packed struct {
	U64 uint64
	U32 uint32
	U16 uint16
	U8  [2]uint8
}

type writeFail struct{}

func (w writeFail) Write(p []byte) (int, error) {
	return 0, errors.New("simulating failure")
}
