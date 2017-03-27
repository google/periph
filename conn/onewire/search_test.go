// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
)

// TestSearch tests the onewire.Search function using the Playback bus preloaded
// with a synthetic set of devices.
func TestSearch(t *testing.T) {
	p := playback{
		Devices: []Address{
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
		crc := CalcCRC(buf[:7])
		p.Devices[i] = (Address(crc) << 56) | (p.Devices[i] & 0x00ffffffffffffff)
	}

	// We're doing one search operation per device, plus a last one.
	p.Ops = make([]IO, len(p.Devices)+1)
	for i := 0; i < len(p.Ops); i++ {
		p.Ops[i] = IO{Write: []byte{0xf0}, Pull: WeakPullup}
	}

	// Start search.
	if err := p.Tx([]byte{0xf0}, nil, WeakPullup); err != nil {
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
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSearch_tx_err(t *testing.T) {
	p := playback{}
	if addrs, err := p.Search(true); len(addrs) != 0 || err == nil {
		t.Fatal("expected Tx() error")
	}
}

//

type IO struct {
	Write []byte
	Read  []byte
	Pull  Pullup
}

// playback is stripped down a copy of onewiretest.Playback.
type playback struct {
	Ops       []IO
	Devices   []Address
	inactive  []bool
	searchBit uint
}

func (p *playback) String() string {
	return "playback"
}

func (p *playback) Close() error {
	if len(p.Ops) != 0 {
		return fmt.Errorf("onewiretest: expected playback to be empty:\n%#v", p.Ops)
	}
	return nil
}

func (p *playback) Tx(w, r []byte, pull Pullup) error {
	if len(p.Ops) == 0 {
		return errors.New("onewiretest: unexpected Tx()")
	}
	if !bytes.Equal(p.Ops[0].Write, w) {
		return fmt.Errorf("onewiretest: unexpected write %#v != %#v", w, p.Ops[0].Write)
	}
	if len(p.Ops[0].Read) != len(r) {
		return fmt.Errorf("onewiretest: unexpected read buffer length %d != %d", len(r), len(p.Ops[0].Read))
	}
	if pull != p.Ops[0].Pull {
		return fmt.Errorf("onewiretest: unexpected pullup %s != %s", pull, p.Ops[0].Pull)
	}
	// Determine whether this starts a search and reset search state.
	if len(w) > 0 && w[0] == 0xf0 {
		p.searchBit = 0
		p.inactive = make([]bool, len(p.Devices))
	}
	// Concoct response.
	copy(r, p.Ops[0].Read)
	p.Ops = p.Ops[1:]
	return nil
}

func (p *playback) Search(alarmOnly bool) ([]Address, error) {
	return Search(p, alarmOnly)
}

func (p *playback) SearchTriplet(direction byte) (TripletResult, error) {
	tr := TripletResult{}
	if p.searchBit > 63 {
		return tr, errors.New("onewiretest: search performs more than 64 triplet operations")
	}
	if len(p.inactive) != len(p.Devices) {
		return tr, errors.New("onewiretest: Devices must be initialized before starting search")
	}
	// Figure out the devices' response.
	for i := range p.Devices {
		if p.inactive[i] {
			continue
		}
		if (p.Devices[i]>>p.searchBit)&1 == 0 {
			tr.GotZero = true
		} else {
			tr.GotOne = true
		}
	}
	// Decide in which direction to take the search.
	switch {
	case tr.GotZero && !tr.GotOne:
		tr.Taken = 0
	case !tr.GotZero && tr.GotOne:
		tr.Taken = 1
	default:
		tr.Taken = direction
	}
	// Inactivate devices in the direction not taken.
	for i := range p.Devices {
		if uint8((p.Devices[i]>>p.searchBit)&1) != tr.Taken {
			p.inactive[i] = true
		}
	}

	p.searchBit++
	return tr, nil
}
