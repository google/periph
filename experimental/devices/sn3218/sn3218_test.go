// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sn3218

import (
	"bytes"
	"testing"

	"periph.io/x/periph/conn/i2c/i2ctest"
)

func setup() *i2ctest.Record {
	bus := i2ctest.Record{
		Bus: nil,
		Ops: []i2ctest.IO{},
	}
	return &bus
}

func TestNew(t *testing.T) {
	bus := i2ctest.Record{
		Bus: nil,
		Ops: []i2ctest.IO{},
	}
	_, err := New(&bus)
	if err != nil {
		t.Fatal("New should not return error", err)
	}
	if len(bus.Ops) > 1 {
		t.Fatal("Expected 0 operation to I2CBus, got ", len(bus.Ops))
	}

	if !bytes.Equal(bus.Ops[0].W, []byte{0x17, 0xFF}) {
		t.Fatal("Expected: 0x17, 0xFF (reset), got: ", bus.Ops[0].W)
	}
}

func TestHalt(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	err := dev.Halt()
	if err != nil {
		t.Fatal("Halt should not return error, but did", err)
	}
	if len(bus.Ops) != 2 {
		t.Fatal("Expected 2 operations, got", len(bus.Ops))
	}
	if !bytes.Equal(bus.Ops[1].W, []byte{0x17, 0xFF}) {
		t.Fatal("I2C write different than expected (reset)")
	}
}

func TestWakeUp(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	dev.WakeUp()
	if len(bus.Ops) != 2 {
		t.Fatal("Expected 2 operations, got", len(bus.Ops))
	}
	if bus.Ops[1].Addr != 0x54 {
		t.Fatal("Expected: Write to address 0x54, got: ", bus.Ops[1].Addr)
	}
	if !bytes.Equal(bus.Ops[1].W, []byte{0x00, 0x01}) {
		t.Fatal("Expected: 0x00, 0x01, got: ", bus.Ops[1].W)
	}
}

func TestSleep(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	dev.Sleep()
	if !bytes.Equal(bus.Ops[1].W, []byte{0x00, 0x00}) {
		t.Fatal("Expected: 0x00, 0x00, got: ", bus.Ops[1].W)
	}
}

func TestGetState(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	state, brightness, err := dev.GetState(0)
	if state != false || brightness != 0 || err != nil {
		t.Fatal("Expected: false, 0, nil, got: ", state, brightness, err)
	}
	if _, _, err := dev.GetState(-1); err == nil {
		t.Fatal("Expected error, but error is nil")
	}
	if _, _, err := dev.GetState(18); err == nil {
		t.Fatal("Expected error, but error is nil")
	}
}

func TestSwitch(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	err := dev.Switch(7, true)
	if err != nil {
		t.Fatal("Expected: err == nil, got:", err)
	}
	if state, _, _ := dev.GetState(7); !state {
		t.Fatal("Expected: LED on, but was off")
	}
	dev.Switch(7, false)
	if state, _, _ := dev.GetState(7); state {
		t.Fatal("Expected: LED off, but was on")
	}
	if len(bus.Ops) != 5 {
		t.Fatal("Expected 5 i2c writes, got: ", len(bus.Ops))
	}
	if !bytes.Equal(bus.Ops[1].W, []byte{0x13, 0x00, 0x02, 0x00}) {
		t.Fatal("Expected 0x13, 0x00, 0x02, 0x00, got:", bus.Ops[1].W)
	}
	if !bytes.Equal(bus.Ops[2].W, []byte{0x16, 0xFF}) {
		t.Fatal("Expected 0x16, 0xFF got:", bus.Ops[2].W)
	}
	if !bytes.Equal(bus.Ops[3].W, []byte{0x13, 0x00, 0x00, 0x00}) {
		t.Fatal("Expected 0x13, 0x00, 0x00, 0x00, got: ", bus.Ops[3].W)
	}
	if !bytes.Equal(bus.Ops[4].W, []byte{0x16, 0xFF}) {
		t.Fatal("Expected 0x16, 0xFF got:", bus.Ops[4].W)
	}
	if err = dev.Switch(19, true); err == nil {
		t.Fatal("Tried to switch LED out of range and expected error, but error is nil...")
	}

}

func TestSetGlobalBrightness(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	dev.BrightnessAll(100)
	for i := 0; i < 17; i++ {
		if dev.brightness[i] != 100 {
			t.Fatal("Brightness of LED", i, " should be 100, but is", dev.brightness[i])
		}
	}

	if len(bus.Ops) != 3 {
		t.Fatal("Expected 3 operations on I2C, got", len(bus.Ops))
	}

	if !bytes.Equal(bus.Ops[1].W, []byte{0x01, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64, 0x64}) {
		t.Fatal("Write operation to I2C different than expected")
	}

	if !bytes.Equal(bus.Ops[2].W, []byte{0x16, 0xFF}) {
		t.Fatal("Expected update command, but got something else")
	}
}

func TestSetBrightness(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	if _, brightness, _ := dev.GetState(9); brightness != 0 {
		t.Fatal("Brightness should be 0, but it's not")
	}
	if err := dev.Brightness(9, 8); err != nil {
		t.Fatal("There should be no error, but it is", err)
	}
	if _, brightness, _ := dev.GetState(9); brightness != 8 {
		t.Fatal("Brightness should be 8, but it's not")
	}
	if len(bus.Ops) != 3 {
		t.Fatal("Expected 3 i2c operations, got", len(bus.Ops))
	}
	if !bytes.Equal(bus.Ops[1].W, []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 0}) {
		t.Fatal("Write operation to I2C different than expected")
	}
	if err := dev.Brightness(42, 100); err == nil {
		t.Fatal("Expected error because channel out of range, but error was nil")
	}
}

func TestSwitchAll(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)
	dev.SwitchAll(true)
	for i := 0; i < 17; i++ {
		if state, _, _ := dev.GetState(i); !state {
			t.Fatal("LED should be on, but is off: ", i)
		}
	}
	if len(bus.Ops) != 3 {
		t.Fatal("Expected 3 operations on I2C, got", len(bus.Ops))
	}
	if !bytes.Equal(bus.Ops[1].W, []byte{19, 63, 63, 63}) {
		t.Fatal("Data written to bus different than expected")
	}

	dev.SwitchAll(false)
	for i := 0; i < 17; i++ {
		if state, _, _ := dev.GetState(i); state {
			t.Fatal("LED should be off, but is on: ", i)
		}
	}
	if len(bus.Ops) != 5 {
		t.Fatal("Expected 4 operations on I2C, got", len(bus.Ops))
	}
	if !bytes.Equal(bus.Ops[3].W, []byte{19, 0, 0, 0}) {
		t.Fatal("Data written to bus different than expected")
	}
}

func TestBoolArrayToInt(t *testing.T) {
	bus := setup()
	dev, _ := New(bus)

	result := dev.stateArrayToInt()
	if result != 0 {
		t.Error("Expected: 0, got: ", result)
	}

	dev.on[0] = true
	result = dev.stateArrayToInt()
	if result != 1 {
		t.Error("Expected: 1, got: ", result)
	}

	dev.on[1] = true
	result = dev.stateArrayToInt()
	if result != 3 {
		t.Error("Expected: 3, got ", result)
	}

	for i := 0; i < 18; i++ {
		dev.on[i] = true
	}
	result = dev.stateArrayToInt()
	if result != 262143 {
		t.Error("Expected: 262143, got ", result)
	}
}
