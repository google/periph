// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package pca9548

import (
	"fmt"
	"testing"

	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/host"
)

func TestNew(t *testing.T) {
	bus := &i2ctest.Playback{
		Ops:       nil,
		DontPanic: true,
	}
	d, err := Register(bus, &DefaultOpts)
	defer func() {
		ports := d.ListPortNames()
		for _, port := range ports {
			i2creg.Unregister(port)
		}
	}()
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
		port    string
		address uint16
		tx      []i2ctest.IO
		wantErr bool
	}{
		{
			port:    "mux-70-0",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x01}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-1",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x02}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-2",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x04}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-3",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x08}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-4",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x10}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-5",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x20}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-6",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x40}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-7",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{0x80}, R: []byte{}},
				{Addr: 0x30, W: []byte{0xAA}, R: []byte{0xBB}},
			},
		},
		{
			port:    "mux-70-0",
			address: 0x70,
			tx:      []i2ctest.IO{},
			wantErr: true,
		},
		{
			port:    "mux-70-0",
			address: 0x30,
			tx: []i2ctest.IO{
				{Addr: 0x70, W: []byte{}, R: []byte{}},
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

		d, err := Register(bus, &DefaultOpts)
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
		ports := d.ListPortNames()
		for _, port := range ports {
			i2creg.Unregister(port)
		}
	}
}

func TestScanList_String(t *testing.T) {
	var sl ScanList = make(map[string][]uint16)
	sl["mux-70-1"] = []uint16{0x49}

	expected := "Scan Results:\nPort[mux-70-1] 1 found\n" +
		"\tDevice at 0x49"

	got := sl.String()
	if got != expected {
		t.Errorf("expected: \n%v but got: \n%v", expected, got)
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

func TestDev_Scan(t *testing.T) {
	tx := []i2ctest.IO{
		{Addr: 0x70, W: []byte{0x01}, R: []byte{}}, // select port
		{Addr: 0x01, W: []byte{}, R: []byte{0x00}}, // ack
		{Addr: 0x02, W: []byte{}, R: []byte{0x00}},
		{Addr: 0x03, W: []byte{}, R: []byte{0x00}},
		{Addr: 0x04, W: []byte{}, R: []byte{0x00}},
	}
	bus := &i2ctest.Playback{
		Ops:       tx,
		DontPanic: true,
	}
	host.Init()

	d, err := Register(bus, &DefaultOpts)
	if err != nil {
		t.Errorf("failed to open I²C: %v", err)
	}

	sl := d.Scan()
	fmt.Printf("%+#v", sl)

	if len(sl["mux-70-0"]) == 0 {
		t.Errorf("expected list not to be empty")
	}

	ports := d.ListPortNames()
	for _, port := range ports {
		i2creg.Unregister(port)
	}

}
