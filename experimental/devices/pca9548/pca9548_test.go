// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9548

import (
	"strconv"
	"testing"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host"
)

func TestNew(t *testing.T) {
	tests := []struct {
		tx      []i2ctest.IO
		address int
		wantErr bool
	}{
		{
			tx:      []i2ctest.IO{{Addr: 0x70, W: nil, R: []byte{0xFF}}},
			address: 0x70,
			wantErr: false,
		},
		{
			tx:      []i2ctest.IO{{Addr: 0x70, W: nil, R: nil}},
			address: 0x70,
			wantErr: true,
		},
		{
			tx:      []i2ctest.IO{{Addr: 0x50, W: nil, R: nil}},
			address: 0x50,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		bus := &i2ctest.Playback{Ops: tt.tx, DontPanic: true}
		_, err := New(bus, &Opts{Addr: tt.address})

		if err != nil && !tt.wantErr {
			t.Errorf("got unexpected error %v", err)
		}

		if err == nil && tt.wantErr {
			t.Errorf("expected error but got none")
		}
	}
}

func TestRegisterPorts(t *testing.T) {
	tests := []struct {
		alias  string
		expect []string
	}{
		{
			"mux",
			[]string{
				"playback-pca9548-70-0",
				"playback-pca9548-70-1",
				"playback-pca9548-70-2",
				"playback-pca9548-70-3",
				"playback-pca9548-70-4",
				"playback-pca9548-70-5",
				"playback-pca9548-70-6",
				"playback-pca9548-70-7",
			},
		},
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

		portNames, err := mux.RegisterPorts(tt.alias)
		if err != nil {
			t.Fatalf("failed to create I²C mux: %v", err)
		}
		for i, port := range tt.expect {
			if portNames[i] != port {
				t.Errorf("expected port name %v but got %v", portNames[i], port)
			} else if err := i2creg.Unregister(port); err != nil {
				t.Errorf("failed to Unregister port %d: %v", i, err)
			}
		}
	}
	// Failing Case.
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

		if _, err := mux.RegisterPorts(tt.alias); err != nil {
			t.Fatalf("failed to create I²C mux: %v", err)
		}
		if _, err := mux.RegisterPorts(tt.alias); err == nil {
			t.Fatal("expected second registration to fail", err)
		}
		for i, port := range tt.expect {
			if err := i2creg.Unregister(port); err != nil {
				t.Errorf("failed to Unregister port %d: %v", i, err)
			}
		}
	}
}

func TestDev_Halt(t *testing.T) {
	d := &Dev{}
	err := d.Halt()
	if err != nil {
		t.Errorf("expected error but got none")
	}
}

func TestDev_String(t *testing.T) {
	tests := []struct {
		address uint16
		want    string
	}{
		{address: 0x70, want: "pca9548-70"},
		{address: 0x71, want: "pca9548-71"},
		{address: 0x72, want: "pca9548-72"},
		{address: 0x73, want: "pca9548-73"},
		{address: 0x74, want: "pca9548-74"},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops: []i2ctest.IO{
				{Addr: tt.address, W: nil, R: []byte{0xFF}},
			},
			DontPanic: false,
		}
		host.Init()

		mux, err := New(bus, &Opts{Addr: int(tt.address)})
		if err != nil {
			t.Fatalf("failed to create I²C mux: %v", err)
		}

		if got := mux.String(); got != tt.want {
			t.Errorf("Dev.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestDev_Tx(t *testing.T) {
	var tests = []struct {
		alias       string
		openPort    string
		initialPort int
		address     uint16
		tx          []i2ctest.IO
		wantErr     bool
	}{
		{
			alias:       "mux",
			openPort:    "mux0",
			initialPort: 0,
			address:     0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x01}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			alias:       "mux",
			openPort:    "mux0",
			initialPort: 0,
			address:     0x70,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: nil, R: []byte{0xFF}},
				{Addr: 0x70, W: []byte{0x01}, R: []byte{}},
				{Addr: 0x70, W: []byte{0xAA}, R: []byte{0xBB}},
			},
			wantErr: true,
		},
		{
			alias:       "mux",
			openPort:    "mux0",
			initialPort: 0,
			address:     0x30,
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

		_, err = mux.RegisterPorts(tt.alias)
		if err != nil {
			t.Fatalf("failed to open I²C: %v", err)
		}

		muxbus, err := i2creg.Open(tt.openPort)
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
		for i := 0; i < 8; i++ {
			if err := i2creg.Unregister("playback-" + mux.String() + "-" + strconv.Itoa(i)); err != nil {
				t.Errorf("failed to Unregister port %d: %v", i, err)
			}
		}
	}
}

func Test_port_Tx(t *testing.T) {
	tests := []struct {
		p       *port
		wantErr bool
	}{
		{
			p:       &port{number: 6},
			wantErr: true,
		},
		{
			p: &port{number: 0x6, mux: &Dev{address: 0x70, activePort: 6,
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

func Test_port_Close(t *testing.T) {

	p := &port{number: 6}

	p.Close()
	if err := p.Tx(0x01, nil, nil); err == nil {
		t.Errorf("expected error but got none")
	}

}

func Test_port_String(t *testing.T) {
	bus := &i2ctest.Playback{
		Ops:       []i2ctest.IO{{Addr: 0x70, W: nil, R: []byte{0xFF}}},
		DontPanic: true,
	}
	host.Init()

	mux, err := New(bus, &DefaultOpts)
	if err != nil {
		t.Fatalf("failed to open I²C: %v", err)
	}

	_, err = mux.RegisterPorts("mux")
	if err != nil {
		t.Fatalf("failed to open I²C: %v", err)
	}

	muxbus, err := i2creg.Open("mux0")
	if err != nil {
		t.Fatalf("failed to open I²C: %v", err)
	}

	expected := "Port:playback-pca9548-70-0(mux0)"
	got := muxbus.String()
	if got != expected {
		t.Errorf("expected: \n%v but got: \n%v", expected, got)
	}

	// Cleanup
	for i := 0; i < 8; i++ {
		if err := i2creg.Unregister("playback-" + mux.String() + "-" + strconv.Itoa(i)); err != nil {
			t.Errorf("failed to Unregister port %d: %v", i, err)
		}
	}
}

func Test_port_SetSpeed(t *testing.T) {
	p := &port{}
	err := p.SetSpeed(400 * physic.KiloHertz)
	if err == nil {
		t.Errorf("expected error but got none")
	}
}
