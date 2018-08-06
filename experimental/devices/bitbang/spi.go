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
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/host/cpu"
)

// NewSPI returns an spi.PortCloser that communicates SPI over 3 or 4 pins.
//
// BUG(maruel): Completely untested.
//
// cs can be nil.
func NewSPI(clk, mosi gpio.PinOut, miso gpio.PinIn, cs gpio.PinOut) (*SPI, error) {
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
	return &SPI{spiConn: spiConn{sck: clk, sdi: miso, sdo: mosi, csn: cs}}, nil
}

// SPI represents a SPI master port implemented as bit-banging on 3 or 4 GPIO
// pins.
type SPI struct {
	spiConn spiConn
}

func (s *SPI) String() string {
	return fmt.Sprintf("bitbang/spi(%s, %s, %s, %s)", s.spiConn.sck, s.spiConn.sdi, s.spiConn.sdo, s.spiConn.csn)
}

// Close implements spi.PortCloser.
func (s *SPI) Close() error {
	return nil
}

// Connect implements spi.PortCloser.
func (s *SPI) Connect(f physic.Frequency, mode spi.Mode, bits int) (spi.Conn, error) {
	if f < 0 {
		return nil, errors.New("bitbang-spi: invalid frequency")
	}
	s.spiConn.mu.Lock()
	defer s.spiConn.mu.Unlock()
	s.spiConn.freqDev = f
	if s.spiConn.freqDev != 0 && (s.spiConn.freqPort == 0 || s.spiConn.freqDev < s.spiConn.freqPort) {
		s.spiConn.halfCycle = f.Duration() / 2
	}
	s.spiConn.mode = mode
	s.spiConn.bits = bits
	return &s.spiConn, nil
}

// LimitSpeed implements spi.PortCloser.
func (s *SPI) LimitSpeed(f physic.Frequency) error {
	if f <= 0 {
		return errors.New("bitbang-spi: invalid frequency")
	}
	s.spiConn.mu.Lock()
	defer s.spiConn.mu.Unlock()
	s.spiConn.freqPort = f
	if s.spiConn.freqDev == 0 || s.spiConn.freqPort < s.spiConn.freqDev {
		s.spiConn.halfCycle = f.Duration() / 2
	}
	return nil
}

// CLK implements spi.Pins.
func (s *SPI) CLK() gpio.PinOut {
	return s.spiConn.sck
}

// MOSI implements spi.Pins.
func (s *SPI) MOSI() gpio.PinOut {
	return s.spiConn.sdo
}

// MISO implements spi.Pins.
func (s *SPI) MISO() gpio.PinIn {
	return s.spiConn.sdi
}

// CS implements spi.Pins.
func (s *SPI) CS() gpio.PinOut {
	return s.spiConn.csn
}

//

// spiConn implements spi.Conn.
type spiConn struct {
	// Immutable.
	sck gpio.PinOut // Clock
	sdi gpio.PinIn  // MISO
	sdo gpio.PinOut // MOSI
	csn gpio.PinOut // CS

	// Mutable.
	mu        sync.Mutex
	freqPort  physic.Frequency
	freqDev   physic.Frequency
	mode      spi.Mode
	bits      int
	halfCycle time.Duration
}

func (s *spiConn) String() string {
	return fmt.Sprintf("bitbang/spi(%s, %s, %s, %s)", s.sck, s.sdi, s.sdo, s.csn)
}

// Duplex implements spi.Conn.
func (s *spiConn) Duplex() conn.Duplex {
	// Maybe implement bitbanging SPI only in half mode?
	return conn.Full
}

func (s *spiConn) clockOn() error {
	if s.mode&spi.Mode2 == spi.Mode2 {
		return s.sck.Out(gpio.Low)
	}
	return s.sck.Out(gpio.High)
}

func (s *spiConn) clockOff() error {
	if s.mode&spi.Mode2 == spi.Mode2 {
		return s.sck.Out(gpio.High)
	}
	return s.sck.Out(gpio.Low)
}

func (s *spiConn) readAfterClockPulse() bool {
	return s.mode&spi.Mode1 == spi.Mode1
}

// Tx implements spi.Conn.
//
// BUG(maruel): Implement mode.
// BUG(maruel): Implement bits.
// BUG(maruel): Test if read works.
func (s *spiConn) Tx(w, r []byte) error {
	if len(r) != 0 && len(w) != len(r) {
		return errors.New("bitbang-spi: write and read buffers must be the same length")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.csn != nil {
		_ = s.csn.Out(gpio.Low)
		s.sleepHalfCycle()
	}
	for i := uint(0); i < uint(len(w)*8); i++ {
		_ = s.sdo.Out(w[i/8]&(1<<(i%8)) != 0)
		s.sleepHalfCycle()
		_ = s.clockOn()
		s.sleepHalfCycle()

		if s.readAfterClockPulse() {
			_ = s.clockOff()
		}

		if len(r) != 0 {
			if s.sdi.Read() == gpio.High {
				r[i/8] |= 1 << (i % 8)
			}
		}

		if !s.readAfterClockPulse() {
			_ = s.clockOff()
		}
	}
	if s.csn != nil {
		_ = s.csn.Out(gpio.High)
	}
	return nil
}

// TxPackets implements spi.Conn.
func (s *spiConn) TxPackets(p []spi.Packet) error {
	return errors.New("bitbang-spi: not implemented")
}

// Write implements io.Writer.
func (s *spiConn) Write(d []byte) (int, error) {
	if err := s.Tx(d, nil); err != nil {
		return 0, err
	}
	return len(d), nil
}

// CLK implements spi.Pins.
func (s *spiConn) CLK() gpio.PinOut {
	return s.sck
}

// MOSI implements spi.Pins.
func (s *spiConn) MOSI() gpio.PinOut {
	return s.sdo
}

// MISO implements spi.Pins.
func (s *spiConn) MISO() gpio.PinIn {
	return s.sdi
}

// CS implements spi.Pins.
func (s *spiConn) CS() gpio.PinOut {
	return s.csn
}

//

// sleep does a busy loop to act as fast as possible.
func (s *spiConn) sleepHalfCycle() {
	cpu.Nanospin(s.halfCycle)
}

var _ spi.PortCloser = &SPI{}
var _ fmt.Stringer = &SPI{}
