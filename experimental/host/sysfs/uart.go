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

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
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
//
// TODO(maruel): It's not yet implemented. Should probably defer to an already
// working library like https://github.com/tarm/serial
type UART struct {
	conn uartConn
}

func newUART(busNumber int) (*UART, error) {
	// Use the devfs path for now.
	name := fmt.Sprintf("/dev/ttyS%d", busNumber)
	f, err := os.OpenFile(name, os.O_RDWR, os.ModeExclusive)
	if err != nil {
		return nil, err
	}
	u := &UART{uartConn{name: name, f: f, busNumber: busNumber}}
	return u, nil
}

// Close implements uart.PortCloser.
func (u *UART) Close() error {
	err := u.conn.f.Close()
	u.conn.f = nil
	return err
}

func (u *UART) String() string {
	return u.conn.String()
}

func (u *UART) Connect(f physic.Frequency, stopBit uart.Stop, parity uart.Parity, flow uart.Flow, bits int) (conn.Conn, error) {
	u.conn.mu.Lock()
	defer u.conn.mu.Unlock()
	if u.conn.f == nil {
		return nil, errors.New("already closed")
	}
	if u.conn.opened {
		return nil, errors.New("already connected")
	}
	return &u.conn, nil
}

// LimitSpeed implements uart.PortCloser.
func (u *UART) LimitSpeed(f physic.Frequency) error {
	return errors.New("sysfs-uart: not implemented")
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
	name      string
	f         *os.File
	busNumber int

	mu     sync.Mutex
	opened bool
}

func (u *uartConn) String() string {
	return u.name
}

// Duplex implements conn.Conn.
func (u *uartConn) Duplex() conn.Duplex {
	return conn.Full
}

// Write implements io.Writer.
func (u *uartConn) Write(b []byte) (int, error) {
	return 0, errors.New("sysfs-uart: not implemented")
}

// Tx implements conn.Conn.
func (u *uartConn) Tx(w, r []byte) error {
	return errors.New("sysfs-uart: not implemented")
}

// RX implements uart.Pins.
func (u *uartConn) RX() gpio.PinIn {
	return gpio.INVALID
}

// TX implements uart.Pins.
func (u *uartConn) TX() gpio.PinOut {
	return gpio.INVALID
}

// RTS implements uart.Pins.
func (u *uartConn) RTS() gpio.PinOut {
	return gpio.INVALID
}

// CTS implements uart.Pins.
func (u *uartConn) CTS() gpio.PinIn {
	return gpio.INVALID
}

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
