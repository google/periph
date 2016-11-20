// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewiretest

import (
	"encoding/binary"
	"testing"

	"github.com/google/periph/experimental/conn/onewire"
)

// TestSearch tests the onewire.Search function using the Playback bus preloaded
// with a synthetic set of devices.
func TestSearch(t *testing.T) {
	p := Playback{
		Devices: []onewire.Address{
			0x0000000000000000,
			0x0000000000000001,
			0x0010000000000000,
			0x0000100000000000,
			0xffffffffffffffff,
			0xfc0000013199a928,
			0xf100000131856328,
		},
	}
	// Fix-up the CRC byte for each device.
	var buf [8]byte
	for i := range p.Devices {
		binary.LittleEndian.PutUint64(buf[:], uint64(p.Devices[i]))
		crc := onewire.CalcCRC(buf[:7])
		p.Devices[i] = (onewire.Address(crc) << 56) | (p.Devices[i] & 0x00ffffffffffffff)
	}

	// We're doing one search operation per device, plus a last one.
	p.Ops = make([]IO, len(p.Devices)+1)
	for i := 0; i < len(p.Ops); i++ {
		p.Ops[i] = IO{Write: []byte{0xf0}, Pull: onewire.WeakPullup}
	}

	// Start search.
	if err := p.Tx([]byte{0xf0}, nil, onewire.WeakPullup); err != nil {
		t.Fatal(err)
	}

	// Perform search.
	addrs, err := p.Search(false)
	if err != nil {
		t.Fatal(err)
	}

	// Verify we got all devices.
	if len(addrs) != len(p.Devices) {
		t.Fatalf("expected %d devices, got %d", len(p.Devices), len(addrs))
	}
match:
	for _, ai := range p.Devices {
		for _, aj := range addrs {
			if ai == aj {
				continue match
			}
		}
		t.Errorf("expected to find %#x but didn't", ai)
	}
}
