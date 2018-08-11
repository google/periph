// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package sysfs implements experimental sysfs support not yet in mainline.
package sysfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/experimental/conn/uart"
)

// EnumerateUART returns the available serial buses.
func EnumerateUART() ([]int, error) {
	// Do not use "/sys/class/tty/ttyS0/" as these are all owned by root.
	prefix := "/dev/ttyS"
	items, err := filepath.Glob(prefix + "*")
	if err != nil {
		return nil, err
	}
	out := make([]int, 0, len(items))
	for _, item := range items {
		i, err := strconv.Atoi(item[len(prefix):])
		if err != nil {
			continue
		}
		out = append(out, i)
	}
	return out, nil
}

// UART is an open serial bus via sysfs.
type UART struct {
	conn uartConn
}

func newUART(portNumber int) (*UART, error) {
	// Use the devfs path for now.
	name := fmt.Sprintf("ttyS%d", portNumber)
	f, err := os.OpenFile("/dev/"+name, os.O_RDWR|syscall.O_NOCTTY, os.ModeExclusive)
	if err != nil {
		return nil, err
	}
	u := &UART{uartConn{name: name, f: f, portNumber: portNumber}}
	return u, nil
}

// Close implements uart.PortCloser.
func (u *UART) Close() error {
	err := u.conn.f.Close()
	u.conn.f = nil
	return err
}

// String implements uart.Port.
func (u *UART) String() string {
	return u.conn.String()
}

// Connect implements uart.Port.
func (u *UART) Connect(f physic.Frequency, stopBit uart.Stop, parity uart.Parity, flow uart.Flow, bits int) (conn.Conn, error) {
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

	u.conn.mu.Lock()
	defer u.conn.mu.Unlock()
	if u.conn.f == nil {
		return nil, errors.New("sysfs-uart: already closed")
	}
	if u.conn.connected {
		return nil, errors.New("sysfs-uart: already connected")
	}
	u.conn.freqConn = f
	u.conn.bitsPerWord = uint8(bits)
	if flow != uart.RTSCTS {
		u.conn.muPins.Lock()
		u.conn.rts = gpio.INVALID
		u.conn.cts = gpio.INVALID
		u.conn.muPins.Unlock()
	}

	// TODO(maruel): ioctl with flags and op.

	return &u.conn, nil
}

// LimitSpeed implements uart.PortCloser.
func (u *UART) LimitSpeed(f physic.Frequency) error {
	if f > physic.GigaHertz {
		return fmt.Errorf("sysfs-uart: invalid speed %s; maximum supported clock is 1GHz", f)
	}
	if f < 50*physic.Hertz {
		return fmt.Errorf("sysfs-uart: invalid speed %s; minimum supported clock is 50Hz; did you forget to multiply by physic.KiloHertz?", f)
	}
	u.conn.mu.Lock()
	defer u.conn.mu.Unlock()
	u.conn.freqPort = f
	return nil
}

// RX implements uart.Pins.
func (u *UART) RX() gpio.PinIn {
	return u.conn.RX()
}

// TX implements uart.Pins.
func (u *UART) TX() gpio.PinOut {
	return u.conn.TX()
}

// RTS implements uart.Pins.
func (u *UART) RTS() gpio.PinOut {
	return u.conn.RTS()
}

// CTS implements uart.Pins.
func (u *UART) CTS() gpio.PinIn {
	return u.conn.CTS()
}

type uartConn struct {
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
func (u *uartConn) String() string {
	return u.name
}

// Duplex implements conn.Conn.
func (u *uartConn) Duplex() conn.Duplex {
	return conn.Full
}

// Read implements io.Reader.
func (u *uartConn) Read(b []byte) (int, error) {
	return u.f.Read(b)
}

// Write implements io.Writer.
func (u *uartConn) Write(b []byte) (int, error) {
	return u.f.Write(b)
}

// Tx implements conn.Conn.
func (u *uartConn) Tx(w, r []byte) error {
	if len(w) != 0 {
		if _, err := u.f.Write(w); err != nil {
			return err
		}
	}
	if len(r) != 0 {
		_, err := u.f.Read(r)
		return err
	}
	return nil
}

// RX implements uart.Pins.
func (u *uartConn) RX() gpio.PinIn {
	u.initPins()
	return u.rx
}

// TX implements uart.Pins.
func (u *uartConn) TX() gpio.PinOut {
	u.initPins()
	return u.tx
}

// RTS implements uart.Pins.
func (u *uartConn) RTS() gpio.PinOut {
	u.initPins()
	return u.rts
}

// CTS implements uart.Pins.
func (u *uartConn) CTS() gpio.PinIn {
	u.initPins()
	return u.cts
}

func (u *uartConn) initPins() {
	u.muPins.Lock()
	defer u.muPins.Unlock()
	if u.rx != nil {
		return
	}
	if u.rx = gpioreg.ByName(fmt.Sprintf("UART%d_RX", u.portNumber)); u.rx == nil {
		u.rx = gpio.INVALID
	}
	if u.tx = gpioreg.ByName(fmt.Sprintf("UART%d_TX", u.portNumber)); u.tx == nil {
		u.tx = gpio.INVALID
	}
	// u.rts is set to INVALID if no hardware RTS/CTS flow control is used.
	if u.rts == nil {
		if u.rts = gpioreg.ByName(fmt.Sprintf("UART%d_RTS", u.portNumber)); u.rts == nil {
			u.rts = gpio.INVALID
		}
		if u.cts = gpioreg.ByName(fmt.Sprintf("UART%d_CTS", u.portNumber)); u.cts == nil {
			u.cts = gpio.INVALID
		}
	}
}

//

// driverUART implements periph.Driver.
type driverUART struct {
}

func (d *driverUART) String() string {
	return "sysfs-uart"
}

func (d *driverUART) Init() (bool, error) {
	return true, nil
}

var _ uart.PortCloser = &UART{}
var _ uart.Pins = &UART{}
var _ conn.Conn = &uartConn{}
var _ uart.Pins = &uartConn{}
