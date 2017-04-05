// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"log"
	"testing"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/spi"
)

func Example_NewSPI() {
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

func TestSPI_faked(t *testing.T) {
	// Create a fake SPI to test methods.
	bus := SPI{frwc: readWriteCloser(0), busNumber: 24}
	if s := bus.String(); s != "SPI24.0" {
		t.Fatal(s)
	}
	// These will all fail, need to mock ioctl.
	bus.Tx(nil, nil)
	bus.Tx([]byte{0}, nil)
	bus.Tx(nil, []byte{0})
	bus.Tx([]byte{0}, []byte{0})
	bus.Speed(0)
	bus.Speed(1)
	bus.CLK()
	bus.MOSI()
	bus.MISO()
	bus.CS()
	if err := bus.DevParams(-1, spi.Mode0, 8); err == nil {
		t.Fatal("invalid speed")
	}
	if err := bus.DevParams(1, -1, 8); err == nil {
		t.Fatal("invalid mode")
	}
	if err := bus.DevParams(1, spi.Mode0, 0); err == nil {
		t.Fatal("invalid bit")
	}
	if err := bus.DevParams(1, spi.Mode0, 8); err == nil {
		t.Fatal("ioctl on invalid handle")
	}
	if d := bus.Duplex(); d != conn.Full {
		t.Fatal(d)
	}
	if err := bus.Close(); err != nil {
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

//

type readWriteCloser int

func (r readWriteCloser) Close() error {
	return nil
}

func (r readWriteCloser) Read(b []byte) (int, error) {
	return 0, nil
}

func (r readWriteCloser) Write(b []byte) (int, error) {
	return 0, nil
}
