// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"log"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
)

func ExampleNewSPI() {
	b, err := NewSPI(0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	if err := b.Tx([]byte{0x10}, nil); err != nil {
		log.Fatal(err)
	}
}

//

func TestNewSPI(t *testing.T) {
	if b, err := NewSPI(-1, -1); b != nil || err == nil {
		t.Fatal("invalid bus")
	}
	if b, err := NewSPI(0, -1); b != nil || err == nil {
		t.Fatal("invalid bus")
	}
}

func TestSPI_IO(t *testing.T) {
	bus := SPI{f: ioctlClose(0), busNumber: 24}
	if err := bus.DevParams(1, spi.Mode3, 8); err != nil {
		t.Fatal(err)
	}
	if err := bus.Tx(nil, nil); err == nil {
		t.Fatal("nil values")
	}
	if err := bus.Tx([]byte{0}, nil); err != nil {
		t.Fatal(err)
	}
	if err := bus.Tx(nil, []byte{0}); err != nil {
		t.Fatal(err)
	}
	if err := bus.Tx([]byte{0}, []byte{0}); err != nil {
		t.Fatal(err)
	}
	if err := bus.Tx([]byte{0}, []byte{0, 1}); err == nil {
		t.Fatal("different lengths")
	}
	if err := bus.Tx(make([]byte, spiBufSize+1), nil); err == nil {
		t.Fatal("buffer too long")
	}
	if err := bus.TxPackets(nil); err == nil {
		t.Fatal("empty TxPackets")
	}
	p := []spi.Packet{
		{W: make([]byte, spiBufSize+1)},
	}
	if err := bus.TxPackets(p); err == nil {
		t.Fatal("buffer too long")
	}
	p = []spi.Packet{
		{W: []byte{0}, R: []byte{0, 1}},
	}
	if err := bus.TxPackets(p); err == nil {
		t.Fatal("different lengths")
	}
	p = []spi.Packet{
		{W: []byte{0}, R: []byte{0}},
	}
	if err := bus.TxPackets(p); err != nil {
		t.Fatal(err)
	}
	if n, err := bus.Read(nil); n != 0 || err == nil {
		t.Fatal(n, err)
	}
	if n, err := bus.Read([]byte{0}); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := bus.Write(nil); n != 0 || err == nil {
		t.Fatal(n, err)
	}
	if n, err := bus.Write([]byte{0}); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if d := bus.Duplex(); d != conn.Full {
		t.Fatal(d)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSPI_IO_not_initialized(t *testing.T) {
	bus := SPI{f: ioctlClose(0), busNumber: 24}
	if err := bus.Tx([]byte{0}, []byte{0}); err == nil {
		t.Fatal("not initialized")
	}
	if err := bus.TxPackets([]spi.Packet{{W: []byte{0}}}); err == nil {
		t.Fatal("not initialized")
	}
}

func TestSPI_pins(t *testing.T) {
	bus := SPI{f: ioctlClose(0), busNumber: 24}
	if p := bus.CLK(); p != gpio.INVALID {
		t.Fatal(p)
	}
	if p := bus.MOSI(); p != gpio.INVALID {
		t.Fatal(p)
	}
	if p := bus.MISO(); p != gpio.INVALID {
		t.Fatal(p)
	}
	if p := bus.CS(); p != gpio.INVALID {
		t.Fatal(p)
	}
}

func TestSPI_other(t *testing.T) {
	bus := SPI{f: ioctlClose(0), busNumber: 24}
	if s := bus.String(); s != "SPI24.0" {
		t.Fatal(s)
	}
	if err := bus.LimitSpeed(0); err == nil {
		t.Fatal("invalid speed")
	}
	if err := bus.LimitSpeed(1); err != nil {
		t.Fatal(err)
	}
	if v := bus.MaxTxSize(); v != spiBufSize {
		t.Fatal(v, spiBufSize)
	}
}

func TestSPI_DevParams(t *testing.T) {
	// Create a fake SPI to test methods.
	bus := SPI{f: ioctlClose(0), busNumber: 24}
	if err := bus.DevParams(-1, spi.Mode0, 8); err == nil {
		t.Fatal("invalid speed")
	}
	if err := bus.DevParams(1, -1, 8); err == nil {
		t.Fatal("invalid mode")
	}
	if err := bus.DevParams(1, spi.Mode0, 0); err == nil {
		t.Fatal("invalid bit")
	}
	if err := bus.DevParams(1, spi.Mode0|spi.HalfDuplex|spi.NoCS|spi.LSBFirst, 8); err != nil {
		t.Fatal(err)
	}
	if err := bus.DevParams(1, spi.Mode0, 8); err == nil {
		t.Fatal("double initialization")
	}
	if d := bus.Duplex(); d != conn.Half {
		t.Fatal(d)
	}
	if err := bus.Tx([]byte{0}, []byte{0}); err == nil {
		t.Fatal("half duplex")
	}
	p := []spi.Packet{
		{W: []byte{0}, R: []byte{0}},
	}
	if err := bus.TxPackets(p); err == nil {
		t.Fatal("half duplex")
	}
}

func TestSPIIOCTX(t *testing.T) {
	if v := spiIOCTx(1); v != 0x40206B00 {
		t.Fatalf("Expected 0x40206B00, got 0x%08X", v)
	}
	if v := spiIOCTx(9); v != 0x41206B00 {
		t.Fatalf("Expected 0x41206B00, got 0x%08X", v)
	}
}

//

func init() {
	spiBufSize = 4096
}
