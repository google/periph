// Copyright 2016 The Periph Authors. All rights reserved.
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

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/host/cpu"
)

// SPI represents a SPI master implemented as bit-banging on 3 or 4 GPIO pins.
type SPI struct {
	sck gpio.PinOut // Clock
	sdi gpio.PinIn  // MISO
	sdo gpio.PinOut // MOSI
	csn gpio.PinOut // CS

	mu        sync.Mutex
	maxHzBus  int64
	maxHzDev  int64
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

// Duplex implements spi.Conn.
func (s *SPI) Duplex() conn.Duplex {
	// Maybe implement bitbanging SPI only in half mode?
	return conn.Full
}

// Speed implements spi.ConnCloser.
func (s *SPI) Speed(maxHz int64) error {
	if maxHz <= 0 {
		return errors.New("bitbang-spi: invalid maxHz")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxHzBus = maxHz
	if s.maxHzDev == 0 || s.maxHzBus < s.maxHzDev {
		s.halfCycle = time.Second / time.Duration(maxHz) / time.Duration(2)
	}
	return nil
}

// DevParams implements spi.Conn.
func (s *SPI) DevParams(maxHz int64, mode spi.Mode, bits int) error {
	if maxHz < 0 {
		return errors.New("bitbang-spi: invalid maxHz")
	}
	if mode != spi.Mode3 {
		return errors.New("bitbang-spi: not implemented")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.maxHzDev = maxHz
	if s.maxHzDev != 0 && (s.maxHzBus == 0 || s.maxHzDev < s.maxHzBus) {
		s.halfCycle = time.Second / time.Duration(maxHz) / time.Duration(2)
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
		return errors.New("bitbang-spi: write and read buffers must be the same length")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
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
		if err := miso.In(gpio.PullUp, gpio.NoEdge); err != nil {
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
	cpu.Nanospin(s.halfCycle)
}

var _ spi.Conn = &SPI{}
