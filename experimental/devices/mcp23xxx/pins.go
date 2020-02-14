package mcp23xxx

import (
	"errors"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
)

// MCP23xxxPin extends gpio.PinIO interface with features supported by MCP23xxx devices
type MCP23xxxPin interface {
	gpio.PinIO
	// SetPolarity if set to Inverted, GPIO register bit reflects the same logic state of the input pin
	SetPolarity(p Polarity) error
	Polarity() Polarity
}

type Polarity bool

const (
	Normal   Polarity = false
	Inverted Polarity = true
)

type port struct {
	name string

	// GPIO basic registers
	iodir *registerCache
	gpio  *registerCache
	olat  *registerCache

	// polarity setting
	ipol *registerCache

	// pull-up control register
	// Not present in device
	gppu *registerCache

	// interrupt handling registers
	gpinten *registerCache
	intcon  *registerCache
	intf    *registerCache
	intcap  *registerCache
}

type portpin struct {
	port   *port
	pinbit uint8
}

func (p *port) pins() []MCP23xxxPin {
	result := make([]MCP23xxxPin, 8)
	var i uint8
	for i = 0; i < 8; i++ {
		result[i] = &portpin{
			port:   p,
			pinbit: i,
		}
	}
	return result
}

func (p *portpin) String() string {
	return p.Name()
}

func (p *portpin) Halt() error {
	// To halt all drive, set to high-impedance input
	return p.In(gpio.Float, gpio.NoEdge)
}

func (p *portpin) Name() string {
	return p.port.name + "_" + strconv.Itoa(int(p.pinbit))
}

func (p *portpin) Number() int {
	return int(p.pinbit)
}

func (p *portpin) Function() string {
	return string(p.Func())
}

func (p *portpin) In(pull gpio.Pull, edge gpio.Edge) error {
	// Set pin to input
	err := p.port.iodir.getAndSetBit(p.pinbit, true, true)
	if err != nil {
		return err
	}
	// Set pullup
	switch pull {
	case gpio.PullNoChange:
		// don't check, don't change
	case gpio.PullDown:
		// pull down is not supported by any device
		return errors.New("MCP23xxx: PullDown is not supported")
	case gpio.PullUp:
		if p.port.gppu == nil {
			return errors.New("MCP23xxx: PullUp is not supported by this device")
		}
		err = p.port.gppu.getAndSetBit(p.pinbit, true, true)
		if err != nil {
			return err
		}
	case gpio.Float:
		if p.port.gppu != nil {
			err = p.port.gppu.getAndSetBit(p.pinbit, false, true)
			if err != nil {
				return err
			}
		}
	}
	// check edge detection
	// TODO interrupt support
	return nil
}

func (p *portpin) Read() gpio.Level {
	v, _ := p.port.gpio.getBit(p.pinbit, false)
	if v {
		return gpio.High
	}
	return gpio.Low
}

func (p *portpin) WaitForEdge(timeout time.Duration) bool {
	// TODO interrupt handling
	return false
}

func (p *portpin) Pull() gpio.Pull {
	if p.port.gppu == nil {
		return gpio.Float
	}
	v, err := p.port.gppu.getBit(p.pinbit, true)
	if err != nil {
		return gpio.PullNoChange
	}
	if v {
		return gpio.PullUp
	}
	return gpio.Float
}

func (p *portpin) DefaultPull() gpio.Pull {
	return gpio.Float
}

func (p *portpin) Out(l gpio.Level) error {
	err := p.port.iodir.getAndSetBit(p.pinbit, false, true)
	if err != nil {
		return err
	}
	return p.port.olat.getAndSetBit(p.pinbit, l == gpio.High, true)
}

func (p *portpin) PWM(duty gpio.Duty, f physic.Frequency) error {
	return errors.New("MCP23xxx: PWM is not supported")
}

func (p *portpin) Func() pin.Func {
	v, _ := p.port.iodir.getBit(p.pinbit, true)
	if v {
		return gpio.IN
	}
	return gpio.OUT
}

func (p *portpin) SupportedFuncs() []pin.Func {
	return supportedFuncs[:]
}

func (p *portpin) SetFunc(f pin.Func) error {
	var v bool
	switch f {
	case gpio.IN:
		v = true
	case gpio.OUT:
		v = false
	default:
		return errors.New("MCP23xxx: Function not supported: " + string(f))
	}
	return p.port.iodir.getAndSetBit(p.pinbit, v, true)
}

func (_ *portpin) SetPolarity(p Polarity) error {
	// TODO support polarity
	return nil
}
func (_ *portpin) Polarity() Polarity {
	// TODO support polarity
	return Normal
}

var supportedFuncs = [...]pin.Func{gpio.IN, gpio.OUT}
