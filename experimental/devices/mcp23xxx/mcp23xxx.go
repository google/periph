// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mcp23xxx

import (
	"fmt"
	"strconv"
	"strings"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
)

// Dev his a handle for a configured MCP23xxx device.
type Dev struct {
	// Pins provide access to extender pins.
	Pins [][]MCP23xxxPin
}

// Variant is the type denoting a specific variant of the family.
type Variant string

const (
	// MCP23008 8-bit I2C extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23008
	MCP23008 Variant = "MCP23008"

	// MCP23S08 8-bit SPI extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23S08
	MCP23S08 Variant = "MCP23S08"

	// MCP23009 8-bit I2C extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23009
	MCP23009 Variant = "MCP23009"

	// MCP23S09 8-bit SPI extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23S09
	MCP23S09 Variant = "MCP23S09"

	// MCP23016 16-bit I2C extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23016
	MCP23016 Variant = "MCP23016"

	// MCP23017 8-bit I2C extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23017
	MCP23017 Variant = "MCP23017"

	// MCP23S17 8-bit SPI extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23S17
	MCP23S17 Variant = "MCP23S17"

	// MCP23018 8-bit I2C extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23018
	MCP23018 Variant = "MCP23018"

	// MCP23S18 8-bit SPI extender. Datasheet: https://www.microchip.com/wwwproducts/en/MCP23S18
	MCP23S18 Variant = "MCP23S18"
)

// NewI2C initializes an IO extender through I2C connection.
func NewI2C(b i2c.Bus, variant Variant, addr uint16) (*Dev, error) {
	if addr&0xFFF8 != 0x20 {
		return nil, fmt.Errorf("%s: Supported address range is 0x20 - 0x27", variant)
	}
	devicename := strings.ReplaceAll(string(variant), "x", "0") + "_" + strconv.FormatInt(int64(addr), 16)
	ra := &i2cRegisterAccess{
		Dev: &i2c.Dev{Bus: b, Addr: addr},
	}
	return makeDev(ra, variant, devicename)
}

// NewSPI initializes an IO extender through I2C connection.
func NewSPI(b spi.Conn, variant Variant) (*Dev, error) {
	devicename := strings.ReplaceAll(string(variant), "x", "S")
	ra := &spiRegisterAccess{
		Conn: b,
	}
	return makeDev(ra, variant, devicename)
}

// Close removes any registration to the device.
func (d *Dev) Close() error {
	for _, port := range d.Pins {
		for _, pin := range port {
			err := gpioreg.Unregister(pin.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func makeDev(ra registerAccess, variant Variant, devicename string) (*Dev, error) {
	var ports []*port
	switch variant {
	case MCP23008, MCP23009, MCP23S08, MCP23S09:
		ports = mcp23x089port(devicename, ra)
	case MCP23016:
		ports = mcp23x16ports(devicename, ra)
	case MCP23017, MCP23S17, MCP23018, MCP23S18:
		ports = mcp23x178ports(devicename, ra)
	default:
		return nil, fmt.Errorf("%s: Unsupported variant", devicename)
	}

	pins := make([][]MCP23xxxPin, len(ports))
	for i, port := range ports {
		// pre-cache iodir
		_, err := port.iodir.readValue(false)
		if err != nil {
			return nil, err
		}
		pins[i] = port.pins()
		for _, pin := range pins[i] {
			gpioreg.Register(pin)
		}
	}
	return &Dev{
		Pins: pins,
	}, nil
}

func mcp23x178ports(devicename string, ra registerAccess) []*port {
	return []*port{{
		name: devicename + "_PORTA",
		// GPIO basic registers
		iodir: ra.define(0x00),
		gpio:  ra.define(0x12),
		olat:  ra.define(0x14),

		// polarity setting
		ipol: ra.define(0x02),

		// pull-up control register
		gppu:          ra.define(0x0C),
		supportPullup: true,

		// interrupt handling registers
		gpinten:          ra.define(0x04),
		intcon:           ra.define(0x08),
		intf:             ra.define(0x0E),
		intcap:           ra.define(0x10),
		supportInterrupt: true,
	}, {
		name: devicename + "_PORTB",
		// GPIO basic registers
		iodir: ra.define(0x01),
		gpio:  ra.define(0x13),
		olat:  ra.define(0x15),

		// polarity setting
		ipol:          ra.define(0x03),
		supportPullup: true,

		// pull-up control register
		gppu: ra.define(0x0D),

		// interrupt handling registers
		gpinten:          ra.define(0x05),
		intcon:           ra.define(0x0B),
		intf:             ra.define(0x0F),
		intcap:           ra.define(0x11),
		supportInterrupt: true,
	}}
}

func mcp23x089port(devicename string, ra registerAccess) []*port {
	return []*port{{
		name: devicename,
		// GPIO basic registers
		iodir: ra.define(0x00),
		gpio:  ra.define(0x09),
		olat:  ra.define(0x0A),

		// polarity setting
		ipol: ra.define(0x01),

		// pull-up control register
		gppu:          ra.define(0x06),
		supportPullup: true,

		// interrupt handling registers
		gpinten:          ra.define(0x02),
		intcon:           ra.define(0x04),
		intf:             ra.define(0x07),
		intcap:           ra.define(0x08),
		supportInterrupt: true,
	}}
}

func mcp23x16ports(devicename string, ra registerAccess) []*port {
	return []*port{{
		name: devicename + "_PORT0",
		// GPIO basic registers
		iodir: ra.define(0x06),
		gpio:  ra.define(0x00),
		olat:  ra.define(0x02),

		// polarity setting
		ipol: ra.define(0x04),

		// pull-up control register
		supportPullup: false,

		// interrupt handling registers
		supportInterrupt: false,
		intcap:           ra.define(0x08),
	}, {
		name: devicename + "_PORT1",
		// GPIO basic registers
		iodir: ra.define(0x07),
		gpio:  ra.define(0x01),
		olat:  ra.define(0x03),

		// polarity setting
		ipol: ra.define(0x05),

		// pull-up control register
		supportPullup: false,

		// interrupt handling registers
		supportInterrupt: false,
		intcap:           ra.define(0x09),
	}}
}
