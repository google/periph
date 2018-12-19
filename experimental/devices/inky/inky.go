// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package inky

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

const (
	cols = 104
	rows = 212
	rotation = -90

	speed = 488 * physic.KiloHertz
	spiBits = 8
	chunkSize = 4096
	spiCommand = gpio.Low
	spiData = gpio.High
)

const (
	Black = 0x00
	Red = 0x33
	Yello = 0x33
	White = 0xff
)

// FIXME: Expose public symbols as relevant. Do not export more than needed!
// See https://periph.io/project/#requirements
// for the expectations.
//
// Use the following layout for drivers:
//  - exported support symbols
//  - Opts struct
//  - New func
//  - Dev struct and methods
//  - Private support code

// New opens a handle to the device. FIXME.
func New(p spi.Port, dc gpio.PinOut, reset gpio.PinOut, busy gpio.PinIn) (*Dev, error) {
	c, err := p.Connect(speed, spi.Mode0, spiBits)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inky over spi: %v", err)
	}

	d := &Dev{
		c: c,
		dc: dc,
		r: reset,
		busy: busy,
	}

	return d, nil
}

// Dev is a handle to the device. FIXME.
type Dev struct {
	c conn.Conn
	// Data or command SPI message.
	dc gpio.PinOut
	r gpio.PinOut
	busy gpio.PinIn
}

func (d *Dev) reset() {
	d.r.Out(gpio.Low)
	time.Sleep(100 * time.Millisecond)
	d.r.Out(gpio.High)
	time.Sleep(100 * time.Millisecond)

	d.busy.In(gpio.PullUp, gpio.FallingEdge)
	defer d.busy.In(gpio.PullUp, gpio.NoEdge)
	d.sendCommand(0x12, nil)  // Soft Reset
	log.Println("Waiting for soft reset")
	d.busy.WaitForEdge(-1)
}

func (d *Dev) Update(border byte) error {
	log.Printf("Resetting")
	d.reset()

	log.Printf("Getting ready for update")
	d.sendCommand(0x74, []byte{0x54})  // Set Analog Block Control.
	d.sendCommand(0x7e, []byte{0x3b})  // Set Digital Block Control.

	r := make([]byte, 3)
	binary.LittleEndian.PutUint16(r, rows)
	d.sendCommand(0x01, r)  // Gate setting

	d.sendCommand(0x03, []byte{0x10, 0x01})  // Gate Driving Voltage.

	d.sendCommand(0x3a, []byte{0x07})  // Dummy line period
	d.sendCommand(0x3b, []byte{0x04})  // Gate line width
	d.sendCommand(0x11, []byte{0x03})  // Data entry mode setting 0x03 = X/Y increment

	d.sendCommand(0x04, nil)  // Power on
	d.sendCommand(0x2c, []byte{0x3c})  // VCOM Register, 0x3c = -1.5v?

	log.Printf("Setting border colour")
	d.sendCommand(0x3c, []byte{0x00})
	d.sendCommand(0x3c, []byte{border})  // Border colour.

	// TODO(hatstand): Support Yellow.

	log.Printf("Sending LUT")
	d.sendCommand(0x32, redLUT)  // Set LUTs

	d.sendCommand(0x44, []byte{0x00, cols / 8 - 1})  // Set RAM X Start/End
	h := make([]byte, 4)
	binary.LittleEndian.PutUint16(h[2:], rows)
	d.sendCommand(0x45, h)  // Set RAM Y Start/End

	// Pure white.
	log.Printf("Writing B/W")
	d.sendCommand(0x4e, []byte{0x00})
	d.sendCommand(0x4f, []byte{0x00, 0x00})
	black, _ := pack(makeBlank())
	d.sendCommand(0x24, black)

	// Pure red.
	log.Printf("Writing red")
	d.sendCommand(0x43, []byte{0x00})
	d.sendCommand(0x4f, []byte{0x00, 0x00})
	red, _ := pack(makeBlank())
	d.sendCommand(0x26, red)

	d.sendCommand(0x22, []byte{0xc7})
	d.busy.In(gpio.PullUp, gpio.FallingEdge)
	defer d.busy.In(gpio.PullUp, gpio.NoEdge)
	d.sendCommand(0x20, nil)

	log.Printf("Waiting for update to finish")
	d.busy.WaitForEdge(-1)

	log.Printf("Going back to sleep")
	d.sendCommand(0x10, []byte{0x01})  // Enter deep sleep.
	return nil
}

func (d *Dev) sendCommand(command byte, data []byte) error {
	d.dc.Out(spiCommand)
	err := d.c.Tx([]byte{command}, nil)
	if err != nil {
		panic("halp")
		return fmt.Errorf("failed to send command %x to inky: %v", command, err)
	}
	if data != nil {
		err = d.sendData(data)
		if err != nil {
			panic("halp")
			return fmt.Errorf("failed to send data for command %x to inky: %v", command, err)
		}
	}
	return nil
}

func (d *Dev) sendData(data []byte) error {
	if len(data) > 4096 {
		log.Fatalf("Sending more data than chunk size")
	}
	d.dc.Out(spiData)
	err := d.c.Tx(data, nil)
	if err != nil {
		panic("halp")
		return fmt.Errorf("failed to send data to inky: %v", err)
	}
	return nil
}

func pack(bits []bool) ([]byte, error) {
	if len(bits) % 8 != 0 {
		return nil, fmt.Errorf("len(bits) must be multiple of 8 but is %d", len(bits))
	}

	ret := make([]byte, len(bits) / 8)
	for i, b := range bits {
		index := i / 8
		shift := uint(i) % 8
		ret[index] |= (boolToByte(b) << shift)
	}
	return ret, nil
}

func makeBlank() []bool {
	return make([]bool, rows * cols)
}

func makeFilled() []bool {
	ret := makeBlank()
	for i, _ := range ret {
		ret[i] = true
	}
	return ret
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}
