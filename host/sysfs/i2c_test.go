// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"log"
	"testing"

	"periph.io/x/periph/conn/i2c/i2creg"
)

func ExampleNewI2C() {
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
	bus := I2C{f: ioctlClose(0), busNumber: 24}
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
	bus.SetSpeed(1)
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

func TestDriver_Init(t *testing.T) {
	d := driverI2C{}
	if _, err := d.Init(); err == nil {
		// It will fail on non-linux.
		defer func() {
			for _, name := range d.buses {
				if err := i2creg.Unregister(name); err != nil {
					t.Fatal(err)
				}
			}
		}()
		if len(d.buses) != 0 {
			// It may fail due to ACL.
			b, _ := i2creg.Open("")
			if b != nil {
				b.Close()
			}
		}
	}
	if d.Prerequisites() != nil {
		t.Fatal("unexpected prerequisite")
	}
	i2cMu.Lock()
	i2cMu.Unlock()
	if setSpeed != nil {
		t.Fatal("unexpected setSpeed")
	}
	defer func() {
		setSpeed = nil
	}()
	if SetSpeedHook(nil) == nil {
		t.Fatal("must fail on nil hook")
	}
	if err := SetSpeedHook(func(hz int64) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if SetSpeedHook(func(hz int64) error { return nil }) == nil {
		t.Fatal("second SetSpeedHook must fail")
	}
}
