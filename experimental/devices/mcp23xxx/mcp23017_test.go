// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mcp23xxx

import (
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spitest"
)

func TestMCP23017_out(t *testing.T) {
	const address uint16 = 0x20
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// iodir is read on creation
			{Addr: address, W: []byte{0x00}, R: []byte{0xFF}},
			{Addr: address, W: []byte{0x01}, R: []byte{0xFF}},
			// iodira is set to output
			{Addr: address, W: []byte{0x00, 0xFE}, R: nil},
			// olata is read
			{Addr: address, W: []byte{0x14}, R: []byte{0x00}},
			// writing back unchanged value is omitted
			// writing high output
			{Addr: address, W: []byte{0x14, 0x01}, R: nil},
		},
	}

	dev, err := NewI2C(scenario, MCP23017, address)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	pA0 := gpioreg.ByName("MCP23017_20_PORTA_0")
	pA0.Out(gpio.Low)
	pA0.Out(gpio.High)
}

func TestMCP23S17_out(t *testing.T) {
	const address uint16 = 0x20
	scenario := &spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				// iodira is read
				{W: []byte{0x41, 0x00}, R: []byte{0xFF}},
				{W: []byte{0x41, 0x01}, R: []byte{0xFF}},
				// iodira is set to output
				{W: []byte{0x40, 0x00, 0xFE}, R: nil},
				// olata is read
				{W: []byte{0x41, 0x14}, R: []byte{0x00}},
				// writing back unchanged value is omitted
				// writing high output
				{W: []byte{0x40, 0x14, 0x01}, R: nil},
			},
		},
	}

	conn, err := scenario.Connect(1, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	dev, err := NewSPI(conn, MCP23S17)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	pA0 := gpioreg.ByName("MCP23S17_PORTA_0")

	pA0.Out(gpio.Low)
	pA0.Out(gpio.High)
}

func TestMCP23017_in(t *testing.T) {
	const address uint16 = 0x20
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// iodir is read on creation
			{Addr: address, W: []byte{0x00}, R: []byte{0xFF}},
			{Addr: address, W: []byte{0x01}, R: []byte{0xFF}},
			// not written, since it didn't change
			// gppua is read
			{Addr: address, W: []byte{0x0C}, R: []byte{0x00}},
			// not written, since it didn't change
			// gpio is read
			{Addr: address, W: []byte{0x12}, R: []byte{0x01}},
		},
	}

	dev, err := NewI2C(scenario, MCP23017, address)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	pA0 := gpioreg.ByName("MCP23017_20_PORTA_0")

	pA0.In(gpio.Float, gpio.NoEdge)
	l := pA0.Read()
	if l != gpio.High {
		t.Errorf("Input should be High")
	}
}

func TestMCP23017_inInverted(t *testing.T) {
	const address uint16 = 0x20
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// iodir is read on creation
			{Addr: address, W: []byte{0x00}, R: []byte{0xFF}},
			{Addr: address, W: []byte{0x01}, R: []byte{0xFF}},
			// not written, since it didn't change
			// gppua is read
			{Addr: address, W: []byte{0x0C}, R: []byte{0x00}},
			// not written, since it didn't change
			// polarity is set
			{Addr: address, W: []byte{0x02}, R: []byte{0x01}},
			// gpio is read
			{Addr: address, W: []byte{0x12}, R: []byte{0x01}},
		},
	}

	dev, err := NewI2C(scenario, MCP23017, address)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	pA0 := gpioreg.ByName("MCP23017_20_PORTA_0").(MCP23xxxPin)

	pA0.In(gpio.Float, gpio.NoEdge)
	pA0.SetPolarityInverted(true)
	l := pA0.Read()
	if l != gpio.High {
		t.Errorf("Input should be High")
	}
}

func TestMCP23017_inPullUp(t *testing.T) {
	const address uint16 = 0x20
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// iodir is read on creation
			{Addr: address, W: []byte{0x00}, R: []byte{0xFF}},
			{Addr: address, W: []byte{0x01}, R: []byte{0xFF}},
			// not written, since it didn't change
			// gppua is read and written
			{Addr: address, W: []byte{0x0C}, R: []byte{0x00}},
			{Addr: address, W: []byte{0x0C, 0x01}, R: nil},
			// not written, since it didn't change
			// gpio is read
			{Addr: address, W: []byte{0x12}, R: []byte{0x01}},
		},
	}

	dev, err := NewI2C(scenario, MCP23017, address)
	if err != nil {
		t.Fatal(err)
	}
	defer dev.Close()

	pA0 := gpioreg.ByName("MCP23017_20_PORTA_0")

	pA0.In(gpio.PullUp, gpio.NoEdge)
	l := pA0.Read()
	if l != gpio.High {
		t.Errorf("Input should be High")
	}
}
