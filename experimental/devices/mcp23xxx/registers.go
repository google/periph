package mcp23xxx

import (
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
)

type registerAccess interface {
	define(address uint8) *registerCache
	readRegister(address uint8) (uint8, error)
	writeRegister(address uint8, value uint8) error
}

type i2cRegisterAccess struct {
	*i2c.Dev
}

func (ra *i2cRegisterAccess) readRegister(address uint8) (uint8, error) {
	r := make([]byte, 1)
	err := ra.Tx([]byte{address}, r)
	return r[0], err
}

func (ra *i2cRegisterAccess) writeRegister(address uint8, value uint8) error {
	return ra.Tx([]byte{address, value}, nil)
}

func (ra *i2cRegisterAccess) define(address uint8) *registerCache {
	return newRegister(ra, address)
}

type spiRegisterAccess struct {
	spi.Conn
}

func (ra *spiRegisterAccess) readRegister(address uint8) (uint8, error) {
	r := make([]byte, 1)
	err := ra.Tx([]byte{0x41, address}, r)
	return r[0], err
}

func (ra *spiRegisterAccess) writeRegister(address uint8, value uint8) error {
	return ra.Tx([]byte{0x40, address, value}, nil)
}

func (ra *spiRegisterAccess) define(address uint8) *registerCache {
	return newRegister(ra, address)
}

type registerCache struct {
	registerAccess
	address uint8
	got     bool
	cache   uint8
}

func newRegister(ra registerAccess, address uint8) *registerCache {
	return &registerCache{
		registerAccess: ra,
		address:        address,
		got:            false,
	}
}

func (r *registerCache) readValue(cached bool) (uint8, error) {
	if cached && r.got {
		return r.cache, nil
	}
	v, err := r.readRegister(r.address)
	if err == nil {
		r.got = true
		r.cache = v
	}
	return v, err
}

func (r *registerCache) writeValue(value uint8, cached bool) error {
	if cached && r.got && value == r.cache {
		return nil
	}

	err := r.writeRegister(r.address, value)
	if err != nil {
		return err
	}
	r.got = true
	r.cache = value
	return nil
}

func (r *registerCache) getAndSetBit(bit uint8, value bool, cached bool) error {
	v, err := r.readValue(cached)
	if err != nil {
		return err
	}
	if value {
		v |= 1 << bit
	} else {
		v &= ^(1 << bit)
	}
	return r.writeValue(v, cached)
}

func (r *registerCache) getBit(bit uint8, cached bool) (bool, error) {
	v, err := r.readValue(cached)
	return 0 != (v & (1 << bit)), err
}
