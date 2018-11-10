// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9548

import (
	"testing"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/host"
)

func TestNew(t *testing.T) {
	bus := &i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x70, W: nil, R: []byte{0xFF}},
		},
		DontPanic: true,
	}
	d, err := Register(bus, &DefaultOpts)

	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}

	if d.address != uint16(DefaultOpts.Address) {
		t.Errorf("expeected address %d, but got %d", DefaultOpts.Address, d.address)
	}

	// Cleanup.
	i2creg.Unregister("mux-70-0")
	i2creg.Unregister("mux-70-1")
	i2creg.Unregister("mux-70-2")
	i2creg.Unregister("mux-70-3")
	i2creg.Unregister("mux-70-4")
	i2creg.Unregister("mux-70-5")
	i2creg.Unregister("mux-70-6")
	i2creg.Unregister("mux-70-7")
	i2creg.Unregister("pca9548-70-0")
	i2creg.Unregister("pca9548-70-1")
	i2creg.Unregister("pca9548-70-2")
	i2creg.Unregister("pca9548-70-3")
	i2creg.Unregister("pca9548-70-4")
	i2creg.Unregister("pca9548-70-5")
	i2creg.Unregister("pca9548-70-6")
	i2creg.Unregister("pca9548-70-7")
}

func TestDev_Tx(t *testing.T) {
	var tests = []struct {
		initial uint8
		port    string
		address uint16
		tx      []i2ctest.IO
		wantErr bool
	}{
		{
			port:    "mux-70-0",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x01}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-1",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x02}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-2",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x04}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-3",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x08}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-4",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x10}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-5",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x20}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-6",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x40}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-7",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x80}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-0",
			address: 0x70,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
			},
			wantErr: true,
		},
		{
			port:    "mux-70-0",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {

		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		host.Init()

		_, err := Register(bus, &DefaultOpts)
		if err != nil {
			t.Errorf("failed to open I²C: %v", err)
		}

		bus0, err := i2creg.Open(tt.port)
		if err != nil {
			t.Errorf("failed to open I²C: %v", err)
		}
		defer bus0.Close()
		err = bus0.Tx(tt.address, []byte{0xAA}, []byte{0xBB})

		if err != nil && !tt.wantErr {
			t.Errorf("expected no error but got: %v", err)
		}
		if err == nil && tt.wantErr {
			t.Errorf("expected error")
		}

		// Cleanup
		i2creg.Unregister("mux-70-0")
		i2creg.Unregister("mux-70-1")
		i2creg.Unregister("mux-70-2")
		i2creg.Unregister("mux-70-3")
		i2creg.Unregister("mux-70-4")
		i2creg.Unregister("mux-70-5")
		i2creg.Unregister("mux-70-6")
		i2creg.Unregister("mux-70-7")
		i2creg.Unregister("pca9548-70-0")
		i2creg.Unregister("pca9548-70-1")
		i2creg.Unregister("pca9548-70-2")
		i2creg.Unregister("pca9548-70-3")
		i2creg.Unregister("pca9548-70-4")
		i2creg.Unregister("pca9548-70-5")
		i2creg.Unregister("pca9548-70-6")
		i2creg.Unregister("pca9548-70-7")
	}
}

func Test_port_String(t *testing.T) {
	p := port{
		number: 6,
	}
	expected := "Port:6"

	got := p.String()
	if got != expected {
		t.Errorf("expected: \n%v but got: \n%v", expected, got)
	}
}
