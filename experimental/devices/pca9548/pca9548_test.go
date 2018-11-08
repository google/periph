package pca9548

import (
	"log"
	"testing"

	"periph.io/x/periph/conn/i2c/i2ctest"
)

func TestNew(t *testing.T) {
	bus := &i2ctest.Playback{
		Ops:       nil,
		DontPanic: true,
	}
	d, err := New(bus, &DefaultOpts)

	if err != nil {
		t.Errorf("don't know how that happened, there was nothing that could fail")
	}

	if d.address != DefaultOpts.Address {
		t.Errorf("expeected address %d, but got %d", DefaultOpts.Address, d.address)
	}
}

func TestDev_Tx(t *testing.T) {
	var tests = []struct {
		initial uint8
		port    uint8
		address uint16
		tx      []i2ctest.IO
	}{
		{
			initial: 0xFF,
			port:    0,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x01}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			initial: 0,
			port:    1,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x02}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			initial: 0,
			port:    7,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x80}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			initial: 2,
			port:    2,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		d := &Dev{c: bus, address: 0x40, port: tt.initial}
		p := Port{d, tt.port}

		err := p.Tx(tt.address, []byte{0xAA}, []byte{0xBB})
		if err != nil {
			log.Println(err)
		}

	}
}
