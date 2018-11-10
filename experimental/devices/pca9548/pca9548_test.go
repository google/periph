// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9548

import (
	"testing"

	"periph.io/x/periph/conn/physic"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/host"
)

func TestNew(t *testing.T) {
	tests := []struct {
		tx      []i2ctest.IO
		wantErr bool
	}{
		{
			tx:      []i2ctest.IO{{Addr: 0x70, W: nil, R: []byte{0xFF}}},
			wantErr: false,
		},
		{
			tx:      []i2ctest.IO{{Addr: 0x70, W: nil, R: nil}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		bus := &i2ctest.Playback{Ops: tt.tx, DontPanic: true}
		_, err := New(bus, &DefaultOpts)

		if err != nil && !tt.wantErr {
			t.Errorf("got unexpected error %v", err)
		}

		if err == nil && tt.wantErr {
			t.Errorf("expected error but got none")
		}
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		alias   string
		port    int
		wantErr bool
	}{
		{"mux0", 0, false},
		{"mux0", -1, true},
	}

	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
			},
			DontPanic: true,
		}
		host.Init()

		mux, err := New(bus, &DefaultOpts)
		if err != nil {
			t.Fatalf("failed to create I²C mux: %v", err)
		}

		err = mux.Register(tt.port, tt.alias)
		if err != nil && !tt.wantErr {
			t.Errorf("got unexpected error %v", err)
		}

		if err == nil && tt.wantErr {
			t.Errorf("expected error but got none")
		}

		// Cleanup
		i2creg.Unregister("playback-pca9548-70-0")
		i2creg.Unregister(tt.alias)
	}
}

func TestDev_Halt(t *testing.T) {
	// TODO(neuralSpaz)
	d := &Dev{}
	err := d.Halt()
	if err != nil {
		t.Errorf("expected error but got none")
	}
}

func TestDev_String(t *testing.T) {
	tests := []struct {
		d    *Dev
		want string
	}{
		{d: &Dev{address: 0x70}, want: "pca9548-70"},
	}

	for _, tt := range tests {

		if got := tt.d.String(); got != tt.want {
			t.Errorf("Dev.String() = %v, want %v", got, tt.want)
		}
	}
}
func TestDev_Tx(t *testing.T) {
	var tests = []struct {
		alias   string
		port    int
		address uint16
		tx      []i2ctest.IO
		wantErr bool
	}{
		{
			alias:   "mux0",
			port:    0,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x01}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux1",
			port:    1,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x02}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux2",
			port:    2,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x04}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux3",
			port:    3,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x08}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux4",
			port:    4,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x10}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux5",
			port:    5,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x20}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux6",
			port:    6,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x40}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux7",
			port:    7,
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x80}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:   "mux0",
			port:    0,
			address: 0x70,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
			},
			wantErr: true,
		},
		{
			alias:   "mux0",
			port:    0,
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

		mux, err := New(bus, &DefaultOpts)
		if err != nil {
			t.Fatalf("failed to open I²C: %v", err)
		}

		err = mux.Register(tt.port, tt.alias)
		if err != nil {
			t.Fatalf("failed to open I²C: %v", err)
		}
		muxbus, err := i2creg.Open(tt.alias)
		if err != nil {
			t.Fatalf("failed to open I²C: %v", err)
		}
		defer muxbus.Close()
		err = muxbus.Tx(tt.address, []byte{0xAA}, []byte{0xBB})

		if err != nil && !tt.wantErr {
			t.Errorf("expected no error but got: %v", err)
		}
		if err == nil && tt.wantErr {
			t.Errorf("expected error")
		}

		// Cleanup
		i2creg.Unregister("mux0")
		i2creg.Unregister("mux1")
		i2creg.Unregister("mux2")
		i2creg.Unregister("mux3")
		i2creg.Unregister("mux4")
		i2creg.Unregister("mux5")
		i2creg.Unregister("mux6")
		i2creg.Unregister("mux7")
		i2creg.Unregister("playback-pca9548-70-0")
		i2creg.Unregister("playback-pca9548-70-1")
		i2creg.Unregister("playback-pca9548-70-2")
		i2creg.Unregister("playback-pca9548-70-3")
		i2creg.Unregister("playback-pca9548-70-4")
		i2creg.Unregister("playback-pca9548-70-5")
		i2creg.Unregister("playback-pca9548-70-6")
		i2creg.Unregister("playback-pca9548-70-7")
	}
}

func Test_port_Tx(t *testing.T) {
	tests := []struct {
		p       port
		wantErr bool
	}{
		{
			p:       port{number: 6},
			wantErr: true,
		},
		{
			p: port{number: 0x6, mux: &Dev{address: 0x70, port: 6,
				c: &i2ctest.Playback{
					Ops: []i2ctest.IO{{Addr: 0x1, W: nil, R: nil}},
				},
			},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := tt.p.Tx(0x01, nil, nil)

		if err != nil && !tt.wantErr {
			t.Errorf("got unexpected error %v", err)
		}

		if err == nil && tt.wantErr {
			t.Errorf("expected error but got none")
		}
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

func Test_port_SetSpeed(t *testing.T) {
	// TODO(neuralSpaz)
	p := &port{}
	err := p.SetSpeed(400 * physic.KiloHertz)
	if err != nil {
		t.Errorf("expected error but got none")
	}
}
