// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package i2c

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"
	"testing"
)

func ExampleAll() {
	fmt.Print("I²C buses available:\n")
	for name := range All() {
		fmt.Printf("- %s\n", name)
	}
}

func ExampleDevReg8() {
	// Find a specific device on all available I²C buses:
	for _, opener := range All() {
		bus, err := opener()
		if err != nil {
			log.Fatal(err)
		}
		dev := DevReg8{Dev{bus, 0x76}, binary.BigEndian}
		v, err := dev.ReadRegUint8(0xD0)
		if err != nil {
			log.Fatal(err)
		}
		if v == 0x60 {
			fmt.Printf("Found bme280 on bus %s\n", bus)
		}
		bus.Close()
	}
}

func ExampleDevReg8_ReadRegStruct() {
	bus, err := New(-1)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()
	dev := DevReg8{Dev{bus, 0x76}, binary.BigEndian}
	flags := struct {
		Flag16 uint16
		Flag8  [2]uint8
	}{}
	if err = dev.ReadRegStruct(0xD0, &flags); err != nil {
		log.Fatal(err)
	}
	// Use flags.Flag16 and flags.Flag8.
}

func ExampleDevReg8_WriteRegStruct() {
	bus, err := New(-1)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()
	dev := DevReg8{Dev{bus, 0x76}, binary.BigEndian}
	flags := struct {
		Flag16 uint16
		Flag8  [2]uint8
	}{
		0x1234,
		[2]uint8{1, 2},
	}
	if err = dev.WriteRegStruct(0xD0, &flags); err != nil {
		log.Fatal(err)
	}
}

//

func TestDevString(t *testing.T) {
	d := Dev{&fakeBus{}, 12}
	if s := d.String(); s != "fake(12)" {
		t.Fatalf("got %s", s)
	}
}

func TestDevTx(t *testing.T) {
	exErr := errors.New("yes")
	b := &fakeBus{err: exErr, r: []byte{1, 2, 3}}
	d := Dev{b, 12}
	r := make([]byte, 3)
	w := []byte{3, 4, 5}
	if err := d.Tx(w, r); exErr != err {
		t.Fatalf("got %s", err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal("w")
	}
	expected := []byte{1, 2, 3}
	if !bytes.Equal(r, expected) {
		t.Fatalf("r: %v != %v", b.r, expected)
	}
	if b.addr != 12 {
		t.Fatalf("got %d", b.addr)
	}
}

func TestDevWrite(t *testing.T) {
	b := &fakeBus{}
	d := Dev{b, 12}
	w := []byte{3, 4, 5}
	if n, err := d.Write(w); err != nil || n != 3 {
		t.Fatalf("got %s", err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal("w")
	}
	if b.addr != 12 {
		t.Fatalf("got %d", b.addr)
	}
}

func TestDevWriteErr(t *testing.T) {
	exErr := errors.New("yes")
	b := &fakeBus{err: exErr}
	d := Dev{b, 12}
	w := []byte{3, 4, 5}
	if n, err := d.Write(w); err != exErr || n != 0 {
		t.Fatalf("got %s", err)
	}
	if !bytes.Equal(b.w, w) {
		t.Fatal("w")
	}
	if b.addr != 12 {
		t.Fatalf("got %d", b.addr)
	}
}

//

func TestDevReg8_ReadRegUint_nil(t *testing.T) {
	d := DevReg8{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if v, err := d.ReadRegUint16(34); err == nil || v != 0 {
		t.Fatalf("%v, %v", v, err)
	}
	if v, err := d.ReadRegUint32(34); err == nil || v != 0 {
		t.Fatalf("%v, %v", v, err)
	}
	if v, err := d.ReadRegUint64(34); err == nil || v != 0 {
		t.Fatalf("%v, %v", v, err)
	}
}

func TestDevReg8_ReadRegUint(t *testing.T) {
	r := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	b := &fakeBus{r: r}
	d := DevReg8{Dev: Dev{b, 12}, Order: binary.BigEndian}
	if v, err := d.ReadRegUint8(34); err != nil || v != 0x01 {
		t.Fatalf("%v, %v", v, err)
	}
	if b.addr != 12 {
		t.Fatal(b.addr)
	}
	if !bytes.Equal(b.w, []byte{34}) {
		t.Fatal(b.w)
	}
	b.r = r
	if v, err := d.ReadRegUint16(34); err != nil || v != d.Order.Uint16(r) {
		t.Fatalf("%v, %v", v, err)
	}
	b.r = r
	if v, err := d.ReadRegUint32(34); err != nil || v != d.Order.Uint32(r) {
		t.Fatalf("%v, %v", v, err)
	}
	b.r = r
	if v, err := d.ReadRegUint64(34); err != nil || v != d.Order.Uint64(r) {
		t.Fatalf("%v, %v", v, err)
	}
}

func TestDevReg8_ReadRegStruct_nil(t *testing.T) {
	d := DevReg8{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if err := d.ReadRegStruct(34, &packed{}); err == nil {
		t.Fatal()
	}
}

func TestDevReg8_ReadRegStruct_Precond_Fail(t *testing.T) {
	b := &fakeBus{}
	d := DevReg8{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	if err := d.ReadRegStruct(34, nil); err == nil {
		t.Fatal()
	}
	if err := d.ReadRegStruct(34, 1); err == nil {
		t.Fatal()
	}
	x := [2]string{}
	if err := d.ReadRegStruct(34, &x); err == nil {
		t.Fatal()
	}
	y := struct {
		i *int
	}{}
	if err := d.ReadRegStruct(34, &y); err == nil {
		t.Fatal()
	}
	z := [0]int{}
	if err := d.ReadRegStruct(34, &z); err == nil {
		t.Fatal()
	}
}

func TestDevReg8_ReadRegStruct_struct(t *testing.T) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}); err != nil {
		t.Fatal(err)
	}
	b := &fakeBus{r: buf.Bytes()}
	d := DevReg8{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	p := &packed{}
	if err := d.ReadRegStruct(34, p); err != nil {
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

	b.err = errors.New("fail")
	b.r = buf.Bytes()
	if d.ReadRegStruct(34, p) != b.err {
		t.Fatal()
	}
}

func TestDevReg8_ReadRegStruct_slice(t *testing.T) {
	b := &fakeBus{r: []uint8{1, 2}}
	d := DevReg8{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	p := make([]uint8, 2)
	if err := d.ReadRegStruct(34, p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p, []uint8{1, 2}) {
		t.Fatal(p)
	}
}

func TestDevReg8_WriteRegUint_nil(t *testing.T) {
	d := DevReg8{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if err := d.WriteRegUint16(34, 1); err == nil {
		t.Fatalf("%v", err)
	}
	if err := d.WriteRegUint32(34, 1); err == nil {
		t.Fatalf("%v", err)
	}
	if err := d.WriteRegUint64(34, 1); err == nil {
		t.Fatalf("%v", err)
	}
}

func TestDevReg8_WriteRegUint(t *testing.T) {
	b := &fakeBus{}
	d := DevReg8{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	if err := d.WriteRegUint8(34, 56); err != nil {
		t.Fatalf("%v", err)
	}
	if !bytes.Equal(b.w, []byte{34, 56}) {
		t.Fatal(b.w)
	}
	b.w = nil
	if err := d.WriteRegUint16(34, 0x5678); err != nil {
		t.Fatalf("%v", err)
	}
	if !bytes.Equal(b.w, []byte{34, 0x78, 0x56}) {
		t.Fatal(b.w)
	}
	b.w = nil
	if err := d.WriteRegUint32(34, 0x56789abc); err != nil {
		t.Fatalf("%v", err)
	}
	if !bytes.Equal(b.w, []byte{34, 0xbc, 0x9a, 0x78, 0x56}) {
		t.Fatal(b.w)
	}
	b.w = nil
	if err := d.WriteRegUint64(34, 0x56789abcdef01234); err != nil {
		t.Fatalf("%v", err)
	}
	if !bytes.Equal(b.w, []byte{34, 0x34, 0x12, 0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56}) {
		t.Fatal(b.w)
	}
}

func TestDevReg8_WriteRegStruct_nil(t *testing.T) {
	d := DevReg8{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if err := d.WriteRegStruct(34, &packed{}); err == nil {
		t.Fatal()
	}
}

func TestDevReg8_WriteRegStruct_Precond_Fail(t *testing.T) {
	b := &fakeBus{}
	d := DevReg8{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	if err := d.WriteRegStruct(34, nil); err == nil {
		t.Fatal()
	}
	if err := d.WriteRegStruct(34, 1); err == nil {
		t.Fatal()
	}
	x := [2]string{}
	if err := d.WriteRegStruct(34, &x); err == nil {
		t.Fatal()
	}
	y := struct {
		i *int
	}{}
	if err := d.WriteRegStruct(34, &y); err == nil {
		t.Fatal()
	}
	z := [0]int{}
	if err := d.WriteRegStruct(34, &z); err == nil {
		t.Fatal()
	}
}

func TestDevReg8_WriteRegStruct(t *testing.T) {
	b := &fakeBus{}
	d := DevReg8{Dev: Dev{b, 12}, Order: binary.BigEndian}
	p := &packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}
	if err := d.WriteRegStruct(34, p); err != nil {
		t.Fatal(err)
	}
	expected := []byte{
		34,
		0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
		0x12, 0x34, 0x56, 0x78,
		0x12, 0x34,
		0x12, 0x01,
	}
	if !bytes.Equal(b.w, expected) {
		t.Fatalf("%v != %v", b.w, expected)
	}
}

//

func TestDevReg16_ReadRegUint_nil(t *testing.T) {
	d := DevReg16{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if v, err := d.ReadRegUint8(0x1234); err == nil || v != 0 {
		t.Fatalf("%v, %v", v, err)
	}
	if v, err := d.ReadRegUint16(0x1234); err == nil || v != 0 {
		t.Fatalf("%v, %v", v, err)
	}
	if v, err := d.ReadRegUint32(0x1234); err == nil || v != 0 {
		t.Fatalf("%v, %v", v, err)
	}
	if v, err := d.ReadRegUint64(0x1234); err == nil || v != 0 {
		t.Fatalf("%v, %v", v, err)
	}
}

func TestDevReg16_ReadRegUint(t *testing.T) {
	r := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	b := &fakeBus{r: r}
	d := DevReg16{Dev: Dev{b, 12}, Order: binary.BigEndian}
	if v, err := d.ReadRegUint8(0x1234); err != nil || v != 0x01 {
		t.Fatalf("%v, %v", v, err)
	}
	if b.addr != 12 {
		t.Fatal(b.addr)
	}
	expected := []byte{0x12, 0x34}
	if !bytes.Equal(b.w, expected) {
		t.Fatalf("%v != %v", b.w, expected)
	}
	b.r = r
	if v, err := d.ReadRegUint16(0x1234); err != nil || v != d.Order.Uint16(r) {
		t.Fatalf("%v, %v", v, err)
	}
	b.r = r
	if v, err := d.ReadRegUint32(0x1234); err != nil || v != d.Order.Uint32(r) {
		t.Fatalf("%v, %v", v, err)
	}
	b.r = r
	if v, err := d.ReadRegUint64(0x1234); err != nil || v != d.Order.Uint64(r) {
		t.Fatalf("%v, %v", v, err)
	}
}

func TestDevReg16_ReadRegStruct_nil(t *testing.T) {
	d := DevReg16{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if err := d.ReadRegStruct(0x1234, &packed{}); err == nil {
		t.Fatal()
	}
}

func TestDevReg16_ReadRegStruct_Precond_Fail(t *testing.T) {
	b := &fakeBus{}
	d := DevReg16{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	if err := d.ReadRegStruct(0x1234, nil); err == nil {
		t.Fatal()
	}
	if err := d.ReadRegStruct(0x1234, 1); err == nil {
		t.Fatal()
	}
	x := [2]string{}
	if err := d.ReadRegStruct(0x1234, &x); err == nil {
		t.Fatal()
	}
	y := struct {
		i *int
	}{}
	if err := d.ReadRegStruct(0x1234, &y); err == nil {
		t.Fatal()
	}
	z := [0]int{}
	if err := d.ReadRegStruct(0x1234, &z); err == nil {
		t.Fatal()
	}
}

func TestDevReg16_ReadRegStruct_struct(t *testing.T) {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}); err != nil {
		t.Fatal(err)
	}
	b := &fakeBus{r: buf.Bytes()}
	d := DevReg16{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	p := &packed{}
	if err := d.ReadRegStruct(0x1234, p); err != nil {
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

	b.err = errors.New("fail")
	b.r = buf.Bytes()
	if d.ReadRegStruct(0x1234, p) != b.err {
		t.Fatal()
	}
}

func TestDevReg16_ReadRegStruct_slice(t *testing.T) {
	b := &fakeBus{r: []uint8{1, 2}}
	d := DevReg16{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	p := make([]uint8, 2)
	if err := d.ReadRegStruct(0x1234, p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(p, []uint8{1, 2}) {
		t.Fatal(p)
	}
}

func TestDevReg16_WriteRegUint_nil(t *testing.T) {
	d := DevReg16{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if err := d.WriteRegUint8(0x1234, 1); err == nil {
		t.Fatalf("%v", err)
	}
	if err := d.WriteRegUint16(0x1234, 1); err == nil {
		t.Fatalf("%v", err)
	}
	if err := d.WriteRegUint32(0x1234, 1); err == nil {
		t.Fatalf("%v", err)
	}
	if err := d.WriteRegUint64(0x1234, 1); err == nil {
		t.Fatalf("%v", err)
	}
}

func TestDevReg16_WriteRegUint(t *testing.T) {
	b := &fakeBus{}
	d := DevReg16{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	if err := d.WriteRegUint8(0x1234, 56); err != nil {
		t.Fatalf("%v", err)
	}
	expected := []byte{0x34, 0x12, 56}
	if !bytes.Equal(b.w, expected) {
		t.Fatal("%v != %v", b.w, expected)
	}
	b.w = nil
	if err := d.WriteRegUint16(0x1234, 0x5678); err != nil {
		t.Fatalf("%v", err)
	}
	expected = []byte{0x34, 0x12, 0x78, 0x56}
	if !bytes.Equal(b.w, expected) {
		t.Fatal("%v != %v", b.w, expected)
	}
	b.w = nil
	if err := d.WriteRegUint32(0x1234, 0x56789abc); err != nil {
		t.Fatalf("%v", err)
	}
	expected = []byte{0x34, 0x12, 0xbc, 0x9a, 0x78, 0x56}
	if !bytes.Equal(b.w, expected) {
		t.Fatal("%v != %v", b.w, expected)
	}
	b.w = nil
	if err := d.WriteRegUint64(0x1234, 0x56789abcdef01234); err != nil {
		t.Fatalf("%v", err)
	}
	expected = []byte{0x34, 0x12, 0x34, 0x12, 0xf0, 0xde, 0xbc, 0x9a, 0x78, 0x56}
	if !bytes.Equal(b.w, expected) {
		t.Fatal("%v != %v", b.w, expected)
	}
}

func TestDevReg16_WriteRegStruct_nil(t *testing.T) {
	d := DevReg16{Dev: Dev{&fakeBus{}, 12}, Order: nil}
	if err := d.WriteRegStruct(0x1234, &packed{}); err == nil {
		t.Fatal()
	}
}

func TestDevReg16_WriteRegStruct_Precond_Fail(t *testing.T) {
	b := &fakeBus{}
	d := DevReg16{Dev: Dev{b, 12}, Order: binary.LittleEndian}
	if err := d.WriteRegStruct(0x1234, nil); err == nil {
		t.Fatal()
	}
	if err := d.WriteRegStruct(0x1234, 1); err == nil {
		t.Fatal()
	}
	x := [2]string{}
	if err := d.WriteRegStruct(0x1234, &x); err == nil {
		t.Fatal()
	}
	y := struct {
		i *int
	}{}
	if err := d.WriteRegStruct(0x1234, &y); err == nil {
		t.Fatal()
	}
	z := [0]int{}
	if err := d.WriteRegStruct(0x1234, &z); err == nil {
		t.Fatal()
	}
}

func TestDevReg16_WriteRegStruct(t *testing.T) {
	b := &fakeBus{}
	d := DevReg16{Dev: Dev{b, 12}, Order: binary.BigEndian}
	p := &packed{0x123456789abcdef0, 0x12345678, 0x1234, [2]uint8{0x12, 0x01}}
	if err := d.WriteRegStruct(0x1234, p); err != nil {
		t.Fatal(err)
	}
	expected := []byte{
		0x12, 0x34,
		0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
		0x12, 0x34, 0x56, 0x78,
		0x12, 0x34,
		0x12, 0x01,
	}
	if !bytes.Equal(b.w, expected) {
		t.Fatalf("%v != %v", b.w, expected)
	}
}

//

func TestEdgeCases(t *testing.T) {
	if getSize(reflect.ValueOf(nil)) != 0 {
		t.FailNow()
	}
}

//
func TestAll(t *testing.T) {
	defer reset()
	byName = map[string]Opener{"foo": nil}
	actual := All()
	if len(actual) != 1 {
		t.Fatalf("%v", actual)
	}
	if _, ok := actual["foo"]; !ok {
		t.FailNow()
	}
}

func TestNew(t *testing.T) {
	defer reset()
	if _, err := New(-1); err == nil {
		t.FailNow()
	}

	byNumber = map[int]Opener{42: fakeBusOpener}
	if v, err := New(-1); err != nil || v == nil {
		t.FailNow()
	}
	if v, err := New(42); err != nil || v == nil {
		t.FailNow()
	}
	if v, err := New(1); err == nil || v != nil {
		t.FailNow()
	}
}

func TestRegister(t *testing.T) {
	defer reset()
	if Unregister("", 42) == nil {
		t.FailNow()
	}
	if Register("a", 42, nil) == nil {
		t.FailNow()
	}
	if Register("", 42, fakeBusOpener) == nil {
		t.FailNow()
	}
	if err := Register("a", 42, fakeBusOpener); err != nil {
		t.Fatal(err)
	}
	if Register("a", 42, fakeBusOpener) == nil {
		t.FailNow()
	}
	if Register("b", 42, fakeBusOpener) == nil {
		t.FailNow()
	}
	if Unregister("", 42) == nil {
		t.FailNow()
	}
	if Unregister("a", 0) == nil {
		t.FailNow()
	}
	if err := Unregister("a", 42); err != nil {
		t.Fatal(err)
	}
}

//

type packed struct {
	U64 uint64
	U32 uint32
	U16 uint16
	U8  [2]uint8
}

func fakeBusOpener() (BusCloser, error) {
	return &fakeBus{}, nil
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byName = map[string]Opener{}
	byNumber = map[int]Opener{}
}

type fakeBus struct {
	speed int64
	err   error
	addr  uint16
	w, r  []byte
}

func (f *fakeBus) Close() error {
	return nil
}

func (f *fakeBus) String() string {
	return "fake"
}

func (f *fakeBus) Tx(addr uint16, w, r []byte) error {
	f.addr = addr
	f.w = append(f.w, w...)
	copy(r, f.r)
	f.r = f.r[len(r):]
	return f.err
}

func (f *fakeBus) Speed(hz int64) error {
	f.speed = hz
	return f.err
}
