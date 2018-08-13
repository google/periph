// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package serial implements cross platform UART support exposed by the
// operating system.
//
// On POSIX, it is via devfs. On Windows, it is via Windows specific APIs.
package serial

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"periph.io/x/periph"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/conn/uart"
)

// Enumerate returns the available serial buses as exposed by the OS.
//
// TODO(maruel): Port number are likely not useful, we need port names.
func Enumerate() ([]int, error) {
	var out []int
	if !isWindows {
		// Do not use "/sys/class/tty/ttyS0/" as these are all owned by root.
		prefix := "/dev/ttyS"
		items, err := filepath.Glob(prefix + "*")
		if err != nil {
			return nil, err
		}
		out = make([]int, 0, len(items))
		for _, item := range items {
			i, err := strconv.Atoi(item[len(prefix):])
			if err != nil {
				continue
			}
			out = append(out, i)
		}
	}
	return out, nil
}

func newPortDevFs(portNumber int) (*Port, error) {
	// Use the devfs path for now.
	name := fmt.Sprintf("ttyS%d", portNumber)
	f, err := os.OpenFile("/dev/"+name, os.O_RDWR|syscall.O_NOCTTY, os.ModeExclusive)
	if err != nil {
		return nil, err
	}
	p := &Port{serialConn{name: name, f: f, portNumber: portNumber}}
	return p, nil
}

// Port is an open serial port.
type Port struct {
	conn serialConn
}

// Close implements uart.PortCloser.
func (p *Port) Close() error {
	err := p.conn.f.Close()
	p.conn.f = nil
	return err
}

// String implements uart.Port.
func (p *Port) String() string {
	return p.conn.String()
}

// Connect implements uart.Port.
func (p *Port) Connect(f physic.Frequency, stopBit uart.Stop, parity uart.Parity, flow uart.Flow, bits int) (conn.Conn, error) {
	if f > physic.GigaHertz {
		return nil, fmt.Errorf("sysfs-uart: invalid speed %s; maximum supported clock is 1GHz", f)
	}
	if f < 50*physic.Hertz {
		return nil, fmt.Errorf("sysfs-uart: invalid speed %s; minimum supported clock is 50Hz; did you forget to multiply by physic.KiloHertz?", f)
	}
	if bits < 5 || bits > 8 {
		return nil, fmt.Errorf("sysfs-uart: invalid bits %d; must be between 5 and 8", bits)
	}

	// Find the closest value in acceptedBauds.
	baud := uint32(f / physic.Hertz)
	//var op uint32
	for _, line := range acceptedBauds {
		if line[0] > baud {
			break
		}
		//op = line[1]
	}

	p.conn.mu.Lock()
	defer p.conn.mu.Unlock()
	if p.conn.f == nil {
		return nil, errors.New("sysfs-uart: already closed")
	}
	if p.conn.connected {
		return nil, errors.New("sysfs-uart: already connected")
	}
	p.conn.freqConn = f
	p.conn.bitsPerWord = uint8(bits)
	if flow != uart.RTSCTS {
		p.conn.muPins.Lock()
		p.conn.rts = gpio.INVALID
		p.conn.cts = gpio.INVALID
		p.conn.muPins.Unlock()
	}

	// TODO(maruel): ioctl with flags and op.

	return &p.conn, nil
}

// LimitSpeed implements uart.PortCloser.
func (p *Port) LimitSpeed(f physic.Frequency) error {
	if f > physic.GigaHertz {
		return fmt.Errorf("sysfs-uart: invalid speed %s; maximum supported clock is 1GHz", f)
	}
	if f < 50*physic.Hertz {
		return fmt.Errorf("sysfs-uart: invalid speed %s; minimum supported clock is 50Hz; did you forget to multiply by physic.KiloHertz?", f)
	}
	p.conn.mu.Lock()
	defer p.conn.mu.Unlock()
	p.conn.freqPort = f
	return nil
}

// RX implements uart.Pins.
func (p *Port) RX() gpio.PinIn {
	return p.conn.RX()
}

// TX implements uart.Pins.
func (p *Port) TX() gpio.PinOut {
	return p.conn.TX()
}

// RTS implements uart.Pins.
func (p *Port) RTS() gpio.PinOut {
	return p.conn.RTS()
}

// CTS implements uart.Pins.
func (p *Port) CTS() gpio.PinIn {
	return p.conn.CTS()
}

type serialConn struct {
	// Immutable
	name       string
	f          *os.File
	portNumber int

	mu          sync.Mutex
	freqPort    physic.Frequency // Frequency specified at LimitSpeed()
	freqConn    physic.Frequency // Frequency specified at Connect()
	bitsPerWord uint8
	connected   bool

	// Use a separate lock for the pins, so that they can be queried while a
	// transaction is happening.
	muPins sync.Mutex
	rx     gpio.PinIn
	tx     gpio.PinOut
	rts    gpio.PinOut
	cts    gpio.PinIn
}

// String implements conn.Conn.
func (s *serialConn) String() string {
	return s.name
}

// Duplex implements conn.Conn.
func (s *serialConn) Duplex() conn.Duplex {
	return conn.Full
}

// Read implements io.Reader.
func (s *serialConn) Read(b []byte) (int, error) {
	return s.f.Read(b)
}

// Write implements io.Writer.
func (s *serialConn) Write(b []byte) (int, error) {
	return s.f.Write(b)
}

// Tx implements conn.Conn.
func (s *serialConn) Tx(w, r []byte) error {
	if len(w) != 0 {
		if _, err := s.f.Write(w); err != nil {
			return err
		}
	}
	if len(r) != 0 {
		_, err := s.f.Read(r)
		return err
	}
	return nil
}

// RX implements uart.Pins.
func (s *serialConn) RX() gpio.PinIn {
	s.initPins()
	return s.rx
}

// TX implements uart.Pins.
func (s *serialConn) TX() gpio.PinOut {
	s.initPins()
	return s.tx
}

// RTS implements uart.Pins.
func (s *serialConn) RTS() gpio.PinOut {
	s.initPins()
	return s.rts
}

// CTS implements uart.Pins.
func (s *serialConn) CTS() gpio.PinIn {
	s.initPins()
	return s.cts
}

func (s *serialConn) initPins() {
	s.muPins.Lock()
	defer s.muPins.Unlock()
	if s.rx != nil {
		return
	}
	if s.rx = gpioreg.ByName(fmt.Sprintf("UART%d_RX", s.portNumber)); s.rx == nil {
		s.rx = gpio.INVALID
	}
	if s.tx = gpioreg.ByName(fmt.Sprintf("UART%d_TX", s.portNumber)); s.tx == nil {
		s.tx = gpio.INVALID
	}
	// s.rts is set to INVALID if no hardware RTS/CTS flow control is used.
	if s.rts == nil {
		if s.rts = gpioreg.ByName(fmt.Sprintf("UART%d_RTS", s.portNumber)); s.rts == nil {
			s.rts = gpio.INVALID
		}
		if s.cts = gpioreg.ByName(fmt.Sprintf("UART%d_CTS", s.portNumber)); s.cts == nil {
			s.cts = gpio.INVALID
		}
	}
}

//

// driverSerial implements periph.Driver.
type driverSerial struct {
}

func (d *driverSerial) String() string {
	return "serial"
}

func (d *driverSerial) Prerequisites() []string {
	return nil
}

func (d *driverSerial) After() []string {
	return nil
}

func (d *driverSerial) Init() (bool, error) {
	return true, nil
}

func init() {
	periph.MustRegister(&drv)
}

var drv driverSerial

var _ uart.PortCloser = &Port{}
var _ uart.Pins = &Port{}
var _ conn.Conn = &serialConn{}
var _ uart.Pins = &serialConn{}
