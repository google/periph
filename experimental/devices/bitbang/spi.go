// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Specification
//
// Motorola never published a proper specification.
// http://electronics.stackexchange.com/questions/30096/spi-specifications
// http://www.nxp.com/files/microcontrollers/doc/data_sheet/M68HC11E.pdf page 120
// http://www.st.com/content/ccc/resource/technical/document/technical_note/58/17/ad/50/fa/c9/48/07/DM00054618.pdf/files/DM00054618.pdf/jcr:content/translations/en.DM00054618.pdf

package bitbang

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/conn/spi"
	"github.com/google/pio/host"
)

// SPI represents a SPI master implemented as bit-banging on 3 or 4 GPIO pins.
type SPI struct {
	sck       gpio.PinOut // Clock
	sdi       gpio.PinIn  // MISO
	sdo       gpio.PinOut // MOSI
	csn       gpio.PinOut // CS
	lock      sync.Mutex
	mode      spi.Mode
	bits      int
	halfCycle time.Duration
}

func (s *SPI) String() string {
	return fmt.Sprintf("bitbang/spi(%s, %s, %s, %s)", s.sck, s.sdi, s.sdo, s.csn)
}

// Close implements spi.ConnCloser.
func (s *SPI) Close() error {
	return nil
}

// Speed implements spi.Conn.
func (s *SPI) Speed(hz int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.halfCycle = time.Second / time.Duration(hz) / time.Duration(2)
	return nil
}

// Configure implements spi.Conn.
func (s *SPI) Configure(mode spi.Mode, bits int) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if mode != spi.Mode3 {
		return errors.New("not implemented")
	}
	s.mode = mode
	s.bits = bits
	return nil
}

// Tx implements spi.Conn.
//
// BUG(maruel): Implement mode.
// BUG(maruel): Implement bits.
// BUG(maruel): Test if read works.
func (s *SPI) Tx(w, r []byte) error {
	if len(r) != 0 && len(w) != len(r) {
		return errors.New("write and read buffers must be the same length")
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.csn != nil {
		s.csn.Out(gpio.Low)
		s.sleepHalfCycle()
	}
	for i := uint(0); i < uint(len(w)*8); i++ {
		s.sdo.Out(w[i/8]&(1<<(i%8)) != 0)
		s.sleepHalfCycle()
		s.sck.Out(gpio.Low)
		s.sleepHalfCycle()
		if len(r) != 0 {
			if s.sdi.Read() == gpio.High {
				r[i/8] |= 1 << (i % 8)
			}
		}
		s.sck.Out(gpio.Low)
	}
	if s.csn != nil {
		s.csn.Out(gpio.High)
	}
	return nil
}

// Write implements spi.Conn.
func (s *SPI) Write(d []byte) (int, error) {
	if err := s.Tx(d, nil); err != nil {
		return 0, err
	}
	return len(d), nil
}

// CLK implements spi.Pins.
func (s *SPI) CLK() gpio.PinOut {
	return s.sck
}

// MOSI implements spi.Pins.
func (s *SPI) MOSI() gpio.PinOut {
	return s.sdo
}

// MISO implements spi.Pins.
func (s *SPI) MISO() gpio.PinIn {
	return s.sdi
}

// CS implements spi.Pins.
func (s *SPI) CS() gpio.PinOut {
	return s.csn
}

// NewSPI returns an object that communicates SPI over 3 or 4 pins.
//
// BUG(maruel): Completely untested.
//
// cs can be nil.
func NewSPI(clk, mosi gpio.PinOut, miso gpio.PinIn, cs gpio.PinOut, speedHz int64) (*SPI, error) {
	if err := clk.Out(gpio.High); err != nil {
		return nil, err
	}
	if err := mosi.Out(gpio.High); err != nil {
		return nil, err
	}
	if miso != nil {
		if err := miso.In(gpio.Up, gpio.None); err != nil {
			return nil, err
		}
	}
	if cs != nil {
		// Low means active.
		if err := cs.Out(gpio.High); err != nil {
			return nil, err
		}
	}
	s := &SPI{
		sck:       clk,
		sdi:       miso,
		sdo:       mosi,
		csn:       cs,
		mode:      spi.Mode3,
		bits:      8,
		halfCycle: time.Second / time.Duration(speedHz) / time.Duration(2),
	}
	return s, nil
}

//

// sleep does a busy loop to act as fast as possible.
func (s *SPI) sleepHalfCycle() {
	host.Nanospin(s.halfCycle)
}

var _ spi.Conn = &SPI{}
