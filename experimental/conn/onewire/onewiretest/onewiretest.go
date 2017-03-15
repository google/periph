// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package onewiretest is meant to be used to test drivers over a fake 1-wire bus.
package onewiretest

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/experimental/conn/onewire"
)

// IO registers the I/O that happened on either a real or fake 1-wire bus.
type IO struct {
	Write []byte
	Read  []byte
	Pull  onewire.Pullup
}

// Record implements onewire.Bus that records everything written to it.
//
// This can then be used to feed to Playback to do "replay" based unit tests.
type Record struct {
	sync.Mutex
	Bus onewire.Bus // Bus can be nil if only writes are being recorded.
	Ops []IO
}

func (r *Record) String() string {
	return "record"
}

// Tx implements onewire.Bus.
func (r *Record) Tx(w, read []byte, pull onewire.Pullup) error {
	r.Lock()
	defer r.Unlock()
	if r.Bus == nil {
		if len(read) != 0 {
			return errors.New("onewiretest: read unsupported when no bus is connected")
		}
	} else {
		if err := r.Bus.Tx(w, read, pull); err != nil {
			return err
		}
	}
	io := IO{Write: make([]byte, len(w)), Pull: pull}
	if len(read) != 0 {
		io.Read = make([]byte, len(read))
		copy(io.Read, read)
	}
	copy(io.Write, w)
	r.Ops = append(r.Ops, io)
	return nil
}

// Q implements onewire.Pins.
func (r *Record) Q() gpio.PinIO {
	if p, ok := r.Bus.(onewire.Pins); ok {
		return p.Q()
	}
	return gpio.INVALID
}

// Search implements onewire.Bus
func (r *Record) Search(alarmOnly bool) ([]onewire.Address, error) {
	return nil, nil
}

// Playback implements onewire.Bus and plays back a recorded I/O flow.
//
// The bus' search function is special-cased. When a Tx operation has
// 0xf0 in w[0] the search state is reset and subsequent triplet operations
// respond according to the list of Devices.  In other words, Tx is
// replayed but the responses to SearchTriplet operations are simulated.
//
// While "replay" type of unit tests are of limited value, they still present
// an easy way to do basic code coverage.
type Playback struct {
	sync.Mutex
	Ops       []IO              // recorded operations
	Devices   []onewire.Address // devices that respond to a search operation
	inactive  []bool            // Devices that are no longer active in the search
	searchBit uint              // which bit is being searched next
}

func (p *Playback) String() string {
	return "playback"
}

// Close implements onewire.BusCloser.
func (p *Playback) Close() error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) != 0 {
		return fmt.Errorf("onewiretest: expected playback to be empty:\n%#v", p.Ops)
	}
	return nil
}

// Tx implements onewire.Bus.
func (p *Playback) Tx(w, r []byte, pull onewire.Pullup) error {
	p.Lock()
	defer p.Unlock()
	if len(p.Ops) == 0 {
		// log.Fatal() ?
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

// Search implements onewire.Bus using the Search function (which calls SearchTriplet).
func (p *Playback) Search(alarmOnly bool) ([]onewire.Address, error) {
	return onewire.Search(p, alarmOnly)
}

// SearchTriplet implements onewire.BusSearcher.
func (p *Playback) SearchTriplet(direction byte) (onewire.TripletResult, error) {
	tr := onewire.TripletResult{}
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

//

var _ onewire.Bus = &Record{}
var _ onewire.Pins = &Record{}
var _ onewire.Bus = &Playback{}
var _ onewire.BusSearcher = &Playback{}
