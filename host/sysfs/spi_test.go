// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"io"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
)

func TestNewSPI(t *testing.T) {
	if p, err := NewSPI(-1, 0); p != nil || err == nil {
		t.Fatal("invalid bus number")
	}
	if p, err := NewSPI(0, -1); p != nil || err == nil {
		t.Fatal("invalid CS")
	}
}

func TestSPI_IO(t *testing.T) {
	p := SPI{f: &ioctlClose{}, busNumber: 24}
	c, err := p.Connect(1, spi.Mode3, 8)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Tx(nil, nil); err == nil {
		t.Fatal("nil values")
	}
	if err := c.Tx([]byte{0}, nil); err != nil {
		t.Fatal(err)
	}
	if err := c.Tx(nil, []byte{0}); err != nil {
		t.Fatal(err)
	}
	if err := c.Tx([]byte{0}, []byte{0}); err != nil {
		t.Fatal(err)
	}
	if err := c.Tx([]byte{0}, []byte{0, 1}); err == nil {
		t.Fatal("different lengths")
	}
	// This assumes bufSize was initialized.
	if err := c.Tx(make([]byte, drvSPI.bufSize+1), nil); err == nil {
		t.Fatal("buffer too long")
	}
	if err := c.TxPackets(nil); err == nil {
		t.Fatal("empty TxPackets")
	}
	pkt := []spi.Packet{
		{W: make([]byte, drvSPI.bufSize+1)},
	}
	if err := c.TxPackets(pkt); err == nil {
		t.Fatal("buffer too long")
	}
	pkt = []spi.Packet{
		{W: []byte{0}, R: []byte{0, 1}},
	}
	if err := c.TxPackets(pkt); err == nil {
		t.Fatal("different lengths")
	}
	pkt = []spi.Packet{
		{W: []byte{0}, R: []byte{0}},
	}
	if err := c.TxPackets(pkt); err != nil {
		t.Fatal(err)
	}
	if n, err := c.(io.Reader).Read(nil); n != 0 || err == nil {
		t.Fatal(n, err)
	}
	if n, err := c.(io.Reader).Read([]byte{0}); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := c.(io.Writer).Write(nil); n != 0 || err == nil {
		t.Fatal(n, err)
	}
	if n, err := c.(io.Writer).Write([]byte{0}); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if d := c.Duplex(); d != conn.Full {
		t.Fatal(d)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSPI_IO_not_initialized(t *testing.T) {
	p := SPI{f: &ioctlClose{}, busNumber: 24}
	if _, err := p.txInternal([]byte{0}, []byte{0}); err == nil {
		t.Fatal("not initialized")
	}
	if p.txPackets([]spi.Packet{{W: []byte{0}}}) == nil {
		t.Fatal("not initialized")
	}
}

func TestSPI_pins(t *testing.T) {
	p := SPI{f: &ioctlClose{}, busNumber: 24}
	if c := p.CLK(); c != gpio.INVALID {
		t.Fatal(c)
	}
	if m := p.MOSI(); m != gpio.INVALID {
		t.Fatal(m)
	}
	if m := p.MISO(); m != gpio.INVALID {
		t.Fatal(m)
	}
	if c := p.CS(); c != gpio.INVALID {
		t.Fatal(c)
	}
}

func TestSPI_other(t *testing.T) {
	p := SPI{f: &ioctlClose{}, busNumber: 24}
	if s := p.String(); s != "SPI24.0" {
		t.Fatal(s)
	}
	if err := p.LimitSpeed(0); err == nil {
		t.Fatal("invalid speed")
	}
	if err := p.LimitSpeed(1); err != nil {
		t.Fatal(err)
	}
	if v := p.MaxTxSize(); v != drvSPI.bufSize {
		t.Fatal(v, drvSPI.bufSize)
	}
}

func TestSPI_Connect(t *testing.T) {
	// Create a fake SPI to test methods.
	p := SPI{f: &ioctlClose{}, busNumber: 24}
	if _, err := p.Connect(-1, spi.Mode0, 8); err == nil {
		t.Fatal("invalid speed")
	}
	if _, err := p.Connect(1, -1, 8); err == nil {
		t.Fatal("invalid mode")
	}
	if _, err := p.Connect(1, spi.Mode0, 0); err == nil {
		t.Fatal("invalid bit")
	}
	c, err := p.Connect(1, spi.Mode0|spi.HalfDuplex|spi.NoCS|spi.LSBFirst, 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Connect(1, spi.Mode0, 8); err == nil {
		t.Fatal("double initialization")
	}
	if d := c.Duplex(); d != conn.Half {
		t.Fatal(d)
	}
	if err := c.Tx([]byte{0}, []byte{0}); err == nil {
		t.Fatal("half duplex")
	}
	pkt := []spi.Packet{
		{W: []byte{0}, R: []byte{0}},
	}
	if err := c.TxPackets(pkt); err == nil {
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

func TestSPIDriver(t *testing.T) {
	if len((&driverSPI{}).Prerequisites()) != 0 {
		t.Fatal("unexpected SPI prerequisites")
	}
}

func TestSPI_OpenClose(t *testing.T) {
	p := SPI{f: &ioctlClose{}}
	_, err := p.Connect(10, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err = p.Connect(10, spi.Mode0, 8); err == nil {
		t.Fatal("an spi object cannot be reused for now")
	}
}

func BenchmarkSPI_Tx(b *testing.B) {
	b.ReportAllocs()
	i := ioctlClose{}
	p := SPI{f: &i}
	c, err := p.Connect(10, spi.Mode0, 8)
	if err != nil {
		b.Fatal(err)
	}
	var w [16]byte
	var r [16]byte
	for i := 0; i < b.N; i++ {
		if err := c.Tx(w[:], r[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSPI_TxPackets2(b *testing.B) {
	b.ReportAllocs()
	i := ioctlClose{}
	p := SPI{f: &i}
	c, err := p.Connect(10, spi.Mode0, 8)
	if err != nil {
		b.Fatal(err)
	}
	var w [16]byte
	var r [16]byte
	tx := [2]spi.Packet{
		{W: w[:], KeepCS: true},
		{R: r[:]},
	}
	for i := 0; i < b.N; i++ {
		if err := c.TxPackets(tx[:]); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSPI_TxPackets5(b *testing.B) {
	b.ReportAllocs()
	i := ioctlClose{}
	p := SPI{f: &i}
	c, err := p.Connect(10, spi.Mode0, 8)
	if err != nil {
		b.Fatal(err)
	}
	var w [16]byte
	var r [16]byte
	tx := [5]spi.Packet{
		{W: w[:], KeepCS: true},
		{R: r[:], BitsPerWord: 16},
		{W: w[:], R: r[:], KeepCS: true},
		{R: r[:]},
		{W: w[:], R: r[:]},
	}
	for i := 0; i < b.N; i++ {
		if err := c.TxPackets(tx[:]); err != nil {
			b.Fatal(err)
		}
	}
}

//

func init() {
	drvSPI.bufSize = 4096
}
