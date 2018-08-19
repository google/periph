package mpu9250

import (
	"fmt"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
)

// DebugF the debug function type.
type DebugF func(string, ...interface{})

// SpiTransport Encapsulates the SPI transport parameters.
type SpiTransport struct {
	device spi.Conn
	cs     gpio.PinOut
	debug  DebugF
}

// NewSpiTransport Creates the SPI transport using the provided device path and chip select pin reference.
func NewSpiTransport(path string, cs gpio.PinOut) (*SpiTransport, error) {
	dev, err := spireg.Open(path)
	if err != nil {
		return nil, wrapf("can't open SPI %v", err)
	}
	conn, err := dev.Connect(1*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return nil, wrapf("can't initialize SPI %v", err)
	}
	return &SpiTransport{device: conn, cs: cs, debug: noop}, nil
}

// EnableDebug Sets the debugging output using the local print function.
func (s *SpiTransport) EnableDebug(f DebugF) {
	s.debug = f
}

func (s *SpiTransport) writeByte(address byte, value byte) error {
	s.debug("write register %x value %x", address, value)
	var (
		buf = [...]byte{address, value}
		res [2]byte
	)
	if err := s.cs.Out(gpio.Low); err != nil {
		return err
	}
	if err := s.device.Tx(buf[:], res[:]); err != nil {
		return err
	}
	if err := s.cs.Out(gpio.High); err != nil {
		return err
	}
	return nil
}

func (s *SpiTransport) writeMagReg(address byte, value byte) error {
	return s.writeByte(address, value)
}

func (s *SpiTransport) writeMaskedReg(address byte, mask byte, value byte) error {
	s.debug("write masked %x, mask %x, value %x", address, mask, value)
	maskedValue := mask & value
	s.debug("masked value %x", maskedValue)
	regVal, err := s.readByte(address)
	if err != nil {
		return err
	}
	s.debug("current register %x", regVal)
	regVal = (regVal &^ maskedValue) | maskedValue
	s.debug("new value %x", regVal)
	return s.writeByte(address, regVal)
}

func (s *SpiTransport) readMaskedReg(address byte, mask byte) (byte, error) {
	s.debug("read masked %x, mask %x", address, mask)
	reg, err := s.readByte(address)
	if err != nil {
		return 0, err
	}
	s.debug("masked value %x", reg)
	return reg & mask, nil
}

func (s *SpiTransport) readByte(address byte) (byte, error) {
	s.debug("read register %x", address)
	var (
		buf = [...]byte{0x80 | address, 0}
		res [2]byte
	)
	if err := s.cs.Out(gpio.Low); err != nil {
		return 0, err
	}
	if err := s.device.Tx(buf[:], res[:]); err != nil {
		return 0, err
	}
	s.debug("register content %x:%x", res[0], res[1])
	if err := s.cs.Out(gpio.High); err != nil {
		return 0, err
	}
	return res[1], nil
}

func (s *SpiTransport) readUint16(address ...byte) (uint16, error) {
	if len(address) != 2 {
		return 0, fmt.Errorf("Only 2 bytes per read")
	}
	h, err := s.readByte(address[0])
	if err != nil {
		return 0, err
	}
	l, err := s.readByte(address[1])
	if err != nil {
		return 0, err
	}
	return uint16(h)<<8 | uint16(l), nil
}

func (s *SpiTransport) printFunc(msg string, args ...interface{}) {
	fmt.Printf("SPI: "+msg+"\n", args...)
}

func noop(string, ...interface{}) {}

var _ Proto = &SpiTransport{}
