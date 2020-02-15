package mcp23xxx

import (
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spitest"
)

func TestMCP23017_out(t *testing.T) {
	const address uint16 = 0x20
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// iodira is read
			i2ctest.IO{Addr: address, W: []byte{0x00}, R: []byte{0xFF}},
			// iodira is set to output
			i2ctest.IO{Addr: address, W: []byte{0x00, 0xFE}, R: nil},
			// olata is read
			i2ctest.IO{Addr: address, W: []byte{0x14}, R: []byte{0x00}},
			// writing back unchanged value is omitted
			// writing high output
			i2ctest.IO{Addr: address, W: []byte{0x14, 0x01}, R: nil},
		},
	}

	dev, err := NewI2C(scenario, MCP23x17, address)
	if err != nil {
		t.Fatal(err)
	}

	pA0 := dev.Pins[0][0]

	pA0.Out(gpio.Low)
	pA0.Out(gpio.High)
}

func TestMCP23S17_out(t *testing.T) {
	const address uint16 = 0x20
	scenario := &spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				// iodira is read
				conntest.IO{W: []byte{0x41, 0x00}, R: []byte{0xFF}},
				// iodira is set to output
				conntest.IO{W: []byte{0x40, 0x00, 0xFE}, R: nil},
				// olata is read
				conntest.IO{W: []byte{0x41, 0x14}, R: []byte{0x00}},
				// writing back unchanged value is omitted
				// writing high output
				conntest.IO{W: []byte{0x40, 0x14, 0x01}, R: nil},
			},
		},
	}

	conn, err := scenario.Connect(1, spi.Mode0, 8)
	if err != nil {
		t.Fatal(err)
	}
	dev, err := NewSPI(conn, MCP23x17)
	if err != nil {
		t.Fatal(err)
	}

	pA0 := dev.Pins[0][0]

	pA0.Out(gpio.Low)
	pA0.Out(gpio.High)
}

func TestMCP23017_in(t *testing.T) {
	const address uint16 = 0x20
	scenario := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			// iodira is read
			i2ctest.IO{Addr: address, W: []byte{0x00}, R: []byte{0xFF}},
			// not written, since it didn't change
			// gppua is read
			i2ctest.IO{Addr: address, W: []byte{0x0C}, R: []byte{0x00}},
			// not written, since it didn't change
			// gpio is read
			i2ctest.IO{Addr: address, W: []byte{0x12}, R: []byte{0x01}},
		},
	}

	dev, err := NewI2C(scenario, MCP23x17, address)
	if err != nil {
		t.Fatal(err)
	}

	pA0 := dev.Pins[0][0]

	pA0.In(gpio.Float, gpio.NoEdge)
	l := pA0.Read()
	if l != gpio.High {
		t.Errorf("Input should be High")
	}
}
