// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"io"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
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

func TestNewSPIinternal(t *testing.T) {
	defer reset()
	ioctlOpen = func(path string, flag int) (ioctlCloser, error) {
		return &ioctlClose{}, nil
	}
	s, err := newSPI(65535, 255)
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal(err)
	}
	if v := s.String(); v != "SPI65535.255" {
		t.Fatal(v)
	}
}

func TestNewSPIinternal_Err(t *testing.T) {
	if _, err := newSPI(65536, 255); err == nil {
		t.Fatal("bad bus number")
	}
	if _, err := newSPI(65535, 256); err == nil {
		t.Fatal("bad bus number")
	}
	defer reset()
	ioctlOpen = func(path string, flag int) (ioctlCloser, error) {
		return nil, errors.New("foo")
	}
	if _, err := newSPI(65535, 255); err.Error() != "sysfs-spi: foo" {
		t.Fatal(err)
	}
}

func TestSPI_Tx(t *testing.T) {
	f := ioctlClose{}
	p := SPI{spiConn{f: &f, busNumber: 24}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode3, 8)
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
	if err := c.Tx(make([]byte, drvSPI.bufSize+1), nil); err.Error() != "sysfs-spi: maximum Tx length is 4096, got 4097 bytes" {
		t.Fatal("buffer too long")
	}
	// Inject error.
	f.ioctlErr = errors.New("foo")
	if err := c.Tx([]byte{0}, nil); err.Error() != "sysfs-spi: Tx() failed: foo" {
		t.Fatal(err)
	}
}

func TestSPI_TxPackets(t *testing.T) {
	f := ioctlClose{}
	p := SPI{spiConn{f: &f, busNumber: 24}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode3, 8)
	if err != nil {
		t.Fatal(err)
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
	// Inject error.
	f.ioctlErr = errors.New("foo")
	if err := c.TxPackets(pkt); err.Error() != "sysfs-spi: TxPackets() failed: foo" {
		t.Fatal(err)
	}
}

func TestSPI_Read(t *testing.T) {
	f := ioctlClose{}
	p := SPI{spiConn{f: &f, busNumber: 24}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode3, 8)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := c.(io.Reader).Read(nil); n != 0 || err == nil {
		t.Fatal(n, err)
	}
	if n, err := c.(io.Reader).Read([]byte{0}); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := c.(io.Reader).Read(make([]byte, drvSPI.bufSize+1)); n != 0 || err.Error() != "sysfs-spi: maximum Read length is 4096, got 4097 bytes" {
		t.Fatal(n, err)
	}
	// Inject error.
	f.ioctlErr = errors.New("foo")
	if n, err := c.(io.Reader).Read([]byte{0}); n != 0 || err.Error() != "sysfs-spi: Read() failed: foo" {
		t.Fatal(n, err)
	}
}

func TestSPI_Write(t *testing.T) {
	f := ioctlClose{}
	p := SPI{spiConn{f: &f, busNumber: 24}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode3, 8)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := c.(io.Writer).Write(nil); n != 0 || err == nil {
		t.Fatal(n, err)
	}
	if n, err := c.(io.Writer).Write([]byte{0}); n != 1 || err != nil {
		t.Fatal(n, err)
	}
	if n, err := c.(io.Writer).Write(make([]byte, drvSPI.bufSize+1)); n != 0 || err.Error() != "sysfs-spi: maximum Write length is 4096, got 4097 bytes" {
		t.Fatal(n, err)
	}
	// Inject error.
	f.ioctlErr = errors.New("foo")
	if n, err := c.(io.Writer).Write([]byte{0}); n != 0 || err.Error() != "sysfs-spi: Write() failed: foo" {
		t.Fatal(n, err)
	}
}

func TestSPI_Pins(t *testing.T) {
	p := SPI{spiConn{f: &ioctlClose{}, busNumber: 24}}
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
	p := SPI{spiConn{name: "SPI24.0", f: &ioctlClose{}, busNumber: 24}}
	if s := p.String(); s != "SPI24.0" {
		t.Fatal(s)
	}
	if err := p.LimitSpeed(physic.GigaHertz + 1); err == nil {
		t.Fatal("invalid speed")
	}
	if err := p.LimitSpeed(100*physic.Hertz - 1); err == nil {
		t.Fatal("invalid speed")
	}
	if err := p.LimitSpeed(physic.KiloHertz); err != nil {
		t.Fatal(err)
	}
	if v := p.MaxTxSize(); v != drvSPI.bufSize {
		t.Fatal(v, drvSPI.bufSize)
	}
}

func TestSPI_Connect_Err(t *testing.T) {
	p := SPI{spiConn{f: &ioctlClose{}, busNumber: 24}}
	if _, err := p.Connect(physic.GigaHertz+1, spi.Mode0, 8); err == nil {
		t.Fatal("invalid speed")
	}
	if _, err := p.Connect(100*physic.Hertz-1, spi.Mode0, 8); err == nil {
		t.Fatal("invalid speed")
	}
	if _, err := p.Connect(100*physic.Hertz, -1, 8); err == nil {
		t.Fatal("invalid mode")
	}
	if _, err := p.Connect(100*physic.Hertz, spi.Mode0, 0); err == nil {
		t.Fatal("invalid bit")
	}
	_, err := p.Connect(100*physic.Hertz, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Connect(100*physic.Hertz, spi.Mode0, 8); err == nil {
		t.Fatal("double initialization")
	}
}

func TestSPI_Connect_Err2(t *testing.T) {
	p := SPI{spiConn{f: &ioctlClose{ioctlErr: errors.New("foo")}}}
	if _, err := p.Connect(100*physic.Hertz, spi.Mode0, 8); err.Error() != "sysfs-spi: setting mode Mode0 failed: foo" {
		t.Fatal(err)
	}
}

func TestSPI_Connect_Half(t *testing.T) {
	p := SPI{spiConn{f: &ioctlClose{}, busNumber: 24}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode0|spi.HalfDuplex|spi.NoCS|spi.LSBFirst, 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Connect(100*physic.Hertz, spi.Mode0, 8); err == nil {
		t.Fatal("double initialization")
	}
	if d := c.Duplex(); d != conn.Half {
		t.Fatal(d)
	}
	if err := c.Tx([]byte{0}, []byte{0}); err != nil {
		t.Fatal(err)
	}
	pkt := []spi.Packet{
		{W: []byte{0}, R: []byte{0}},
	}
	if err := c.TxPackets(pkt); err == nil {
		t.Fatal("half duplex")
	}
	// Confirm memory allocation for large number of packets.
	pkt = make([]spi.Packet, len(p.conn.io)+1)
	for i := range pkt {
		pkt[i].R = []byte{0}
	}
	if err := c.TxPackets(pkt); err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
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
	p := SPI{spiConn{f: &ioctlClose{}}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	if v := c.(conn.Limits).MaxTxSize(); v != 4096 {
		t.Fatal(v)
	}
	if d := c.Duplex(); d != conn.Full {
		t.Fatal(d)
	}
	if err = p.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err = p.Connect(100*physic.Hertz, spi.Mode0, 8); err == nil {
		t.Fatal("an spi object cannot be reused for now")
	}
}

func TestSPI_Close_Err(t *testing.T) {
	p := SPI{spiConn{f: &ioctlClose{closeErr: errors.New("foo")}}}
	if err := p.Close(); err.Error() != "sysfs-spi: foo" {
		t.Fatal(err)
	}
}

func BenchmarkSPI_Tx(b *testing.B) {
	b.ReportAllocs()
	f := ioctlClose{}
	p := SPI{spiConn{f: &f}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode0, 8)
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
	f := ioctlClose{}
	p := SPI{spiConn{f: &f}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode0, 8)
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
	f := ioctlClose{}
	p := SPI{spiConn{f: &f}}
	c, err := p.Connect(100*physic.Hertz, spi.Mode0, 8)
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
