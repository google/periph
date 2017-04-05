// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"log"
	"testing"
)

func Example_NewI2C() {
	b, err := NewI2C(1)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	if err := b.Tx(23, []byte{0x10}, nil); err != nil {
		log.Fatal(err)
	}
}

//

func TestNewI2C(t *testing.T) {
	if b, err := NewI2C(-1); b != nil || err == nil {
		t.Fatal("invalid bus")
	}
}

func TestI2C_faked(t *testing.T) {
	// Create a fake I2C to test methods.
	bus := I2C{fc: closer(0), busNumber: 24}
	if s := bus.String(); s != "I2C24" {
		t.Fatal(s)
	}
	// These will all fail, need to mock ioctl.
	bus.Tx(0x401, nil, nil)
	bus.Tx(1, nil, nil)
	bus.Tx(1, []byte{0}, nil)
	bus.Tx(1, nil, []byte{0})
	bus.Tx(1, []byte{0}, []byte{0})
	bus.SetSpeed(0)
	bus.SCL()
	bus.SDA()
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2C_functionality(t *testing.T) {
	expected := "I2C|10BIT_ADDR|PROTOCOL_MANGLING|SMBUS_PEC|NOSTART|SMBUS_BLOCK_PROC_CALL|SMBUS_QUICK|SMBUS_READ_BYTE|SMBUS_WRITE_BYTE|SMBUS_READ_BYTE_DATA|SMBUS_WRITE_BYTE_DATA|SMBUS_READ_WORD_DATA|SMBUS_WRITE_WORD_DATA|SMBUS_PROC_CALL|SMBUS_READ_BLOCK_DATA|SMBUS_WRITE_BLOCK_DATA|SMBUS_READ_I2C_BLOCK|SMBUS_WRITE_I2C_BLOCK"
	if s := functionality(0xFFFFFFFF).String(); s != expected {
		t.Fatal(s)
	}
}

//

type closer int

func (c closer) Close() error {
	return nil
}
