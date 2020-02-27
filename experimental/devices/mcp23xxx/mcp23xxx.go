package mcp23xxx

import (
	"fmt"
	"strconv"

	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
)

type Dev struct {
	Pins [][]MCP23xxxPin
}

type Variant string

const (
	MCP23x08 Variant = "MCP23%s08"
	MCP23x09 Variant = "MCP23%s09"
	MCP23x16 Variant = "MCP23%s16"
	MCP23x17 Variant = "MCP23%s17"
	MCP23x18 Variant = "MCP23%s18"
)

func NewI2C(b i2c.Bus, variant Variant, addr uint16) (*Dev, error) {
	if addr&0xFFF8 != 0x20 {
		return nil, fmt.Errorf("%s: Supported address range is 0x20 - 0x27", variant)
	}
	devicename := fmt.Sprintf(string(variant), "0") + "_" + strconv.FormatInt(int64(addr), 16)
	ra := &i2cRegisterAccess{
		Dev: &i2c.Dev{Bus: b, Addr: addr},
	}
	return makeDev(ra, variant, devicename)
}

func NewSPI(b spi.Conn, variant Variant) (*Dev, error) {
	devicename := fmt.Sprintf(string(variant), "S")
	ra := &spiRegisterAccess{
		Conn: b,
	}
	return makeDev(ra, variant, devicename)
}

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
	case MCP23x08, MCP23x09:
		ports = mcp23x089port(devicename, ra)
	case MCP23x16:
		ports = mcp23x16ports(devicename, ra)
	case MCP23x17, MCP23x18:
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
		gppu: ra.define(0x0C),

		// interrupt handling registers
		gpinten: ra.define(0x04),
		intcon:  ra.define(0x08),
		intf:    ra.define(0x0E),
		intcap:  ra.define(0x10),
	}, {
		name: devicename + "_PORTB",
		// GPIO basic registers
		iodir: ra.define(0x01),
		gpio:  ra.define(0x13),
		olat:  ra.define(0x15),

		// polarity setting
		ipol: ra.define(0x03),

		// pull-up control register
		gppu: ra.define(0x0D),

		// interrupt handling registers
		gpinten: ra.define(0x05),
		intcon:  ra.define(0x0B),
		intf:    ra.define(0x0F),
		intcap:  ra.define(0x11),
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
		gppu: ra.define(0x06),

		// interrupt handling registers
		gpinten: ra.define(0x02),
		intcon:  ra.define(0x04),
		intf:    ra.define(0x07),
		intcap:  ra.define(0x08),
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
		gppu: nil,

		// interrupt handling registers
		gpinten: nil,
		intcon:  nil,
		intf:    nil,
		intcap:  ra.define(0x08),
	}, {
		name: devicename + "_PORT1",
		// GPIO basic registers
		iodir: ra.define(0x07),
		gpio:  ra.define(0x01),
		olat:  ra.define(0x03),

		// polarity setting
		ipol: ra.define(0x05),

		// pull-up control register
		gppu: nil,

		// interrupt handling registers
		gpinten: nil,
		intcon:  nil,
		intf:    nil,
		intcap:  ra.define(0x09),
	}}
}
