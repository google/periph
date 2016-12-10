// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/google/periph"
	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/spi"
)

// NewSPI opens a SPI bus via its devfs interface as described at
// https://www.kernel.org/doc/Documentation/spi/spidev and
// https://www.kernel.org/doc/Documentation/spi/spi-summary
//
// busNumber is the bus number as exported by deffs. For example if the path is
// /dev/spidev0.1, busNumber should be 0 and chipSelect should be 1.
//
// Default configuration is Mode3 and 8 bits.
func NewSPI(busNumber, chipSelect int) (*SPI, error) {
	if isLinux {
		return newSPI(busNumber, chipSelect)
	}
	return nil, errors.New("sysfs-spi: not implemented on non-linux OSes")
}

// SPI is an open SPI bus.
type SPI struct {
	f          *os.File
	busNumber  int
	chipSelect int
	clk        gpio.PinOut
	mosi       gpio.PinOut
	miso       gpio.PinIn
	cs         gpio.PinOut
}

func newSPI(busNumber, chipSelect int) (*SPI, error) {
	if busNumber < 0 || busNumber >= 1<<16 {
		return nil, fmt.Errorf("sysfs-spi: invalid bus %d", busNumber)
	}
	if chipSelect < 0 || chipSelect > 255 {
		return nil, fmt.Errorf("sysfs-spi: invalid chip select %d", chipSelect)
	}
	// Use the devfs path for now.
	f, err := os.OpenFile(fmt.Sprintf("/dev/spidev%d.%d", busNumber, chipSelect), os.O_RDWR, os.ModeExclusive)
	if err != nil {
		return nil, err
	}
	s := &SPI{f: f, busNumber: busNumber, chipSelect: chipSelect}
	if err := s.Configure(spi.Mode3, 8); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the handle to the SPI driver. It is not a requirement to close
// before process termination.
func (s *SPI) Close() error {
	err := s.f.Close()
	s.f = nil
	return err
}

func (s *SPI) String() string {
	return fmt.Sprintf("SPI%d.%d", s.busNumber, s.chipSelect)
}

// Speed implements spi.Conn.
func (s *SPI) Speed(hz int64) error {
	if hz < 1000 {
		return fmt.Errorf("sysfs-spi: invalid speed %d", hz)
	}
	return s.setFlag(spiIOCMaxSpeedHz, uint64(hz))
}

// Configure implements spi.Conn.
func (s *SPI) Configure(mode spi.Mode, bits int) error {
	if bits < 1 || bits > 256 {
		return fmt.Errorf("sysfs-spi: invalid bits %d", bits)
	}
	if err := s.setFlag(spiIOCMode, uint64(mode)); err != nil {
		return err
	}
	return s.setFlag(spiIOCBitsPerWord, uint64(bits))
}

// Write implements spi.Conn.
func (s *SPI) Write(b []byte) (int, error) {
	return s.f.Write(b)
}

// Tx sends and receives data simultaneously.
func (s *SPI) Tx(w, r []byte) error {
	if len(w) == 0 || len(w) != len(r) {
		return errors.New("Tx with zero or non-equal length w&r slices")
	}
	p := spiIOCTransfer{
		tx:          uint64(uintptr(unsafe.Pointer(&w[0]))),
		rx:          uint64(uintptr(unsafe.Pointer(&r[0]))),
		length:      uint32(len(w)),
		bitsPerWord: 8,
	}
	return s.ioctl(spiIOCTx|0x40000000, unsafe.Pointer(&p))
}

// CLK implements spi.Pins.
func (s *SPI) CLK() gpio.PinOut {
	s.initPins()
	return s.clk
}

// MISO implements spi.Pins.
func (s *SPI) MISO() gpio.PinIn {
	s.initPins()
	return s.miso
}

// MOSI implements spi.Pins.
func (s *SPI) MOSI() gpio.PinOut {
	s.initPins()
	return s.mosi
}

// CS implements spi.Pins.
func (s *SPI) CS() gpio.PinOut {
	s.initPins()
	return s.cs
}

// Private details.

const (
	cSHigh    spi.Mode = 0x4
	lSBFirst  spi.Mode = 0x8
	threeWire spi.Mode = 0x10
	loop      spi.Mode = 0x20
	noCS      spi.Mode = 0x40
)

// spidev driver IOCTL control codes.
//
// Constants and structure definition can be found at
// /usr/include/linux/spi/spidev.h.
const (
	spiIOCMode        = 0x16B01
	spiIOCBitsPerWord = 0x16B03
	spiIOCMaxSpeedHz  = 0x46B04
	spiIOCTx          = 0x206B00
)

type spiIOCTransfer struct {
	tx          uint64 // Pointer to byte slice
	rx          uint64 // Pointer to byte slice
	length      uint32
	speedHz     uint32
	delayUsecs  uint16
	bitsPerWord uint8
	csChange    uint8
	txNBits     uint8
	rxNBits     uint8
	pad         uint16
}

func (s *SPI) setFlag(op uint, arg uint64) error {
	if err := s.ioctl(op|0x40000000, unsafe.Pointer(&arg)); err != nil {
		return err
	}
	actual := uint64(0)
	// getFlag() equivalent.
	if err := s.ioctl(op|0x80000000, unsafe.Pointer(&actual)); err != nil {
		return err
	}
	if actual != arg {
		return fmt.Errorf("sysfs-spi: op 0x%x: set 0x%x, read 0x%x", op, arg, actual)
	}
	return nil
}

func (s *SPI) ioctl(op uint, arg unsafe.Pointer) error {
	if err := ioctl(s.f.Fd(), op, uintptr(arg)); err != nil {
		return fmt.Errorf("sysfs-spi: ioctl: %v", err)
	}
	return nil
}

func (s *SPI) initPins() {
	if s.clk == nil {
		if s.clk = gpio.ByName(fmt.Sprintf("SPI%d_CLK", s.busNumber)); s.clk == nil {
			s.clk = gpio.INVALID
		}
		if s.miso = gpio.ByName(fmt.Sprintf("SPI%d_MISO", s.busNumber)); s.miso == nil {
			s.miso = gpio.INVALID
		}
		if s.mosi = gpio.ByName(fmt.Sprintf("SPI%d_MOSI", s.busNumber)); s.mosi == nil {
			s.mosi = gpio.INVALID
		}
		if s.cs = gpio.ByName(fmt.Sprintf("SPI%d_CS%d", s.busNumber, s.chipSelect)); s.cs == nil {
			s.cs = gpio.INVALID
		}
	}
}

// driverSPI implements periph.Driver.
type driverSPI struct {
}

func (d *driverSPI) String() string {
	return "sysfs-spi"
}

func (d *driverSPI) Prerequisites() []string {
	return nil
}

func (d *driverSPI) Init() (bool, error) {
	// This driver is only registered on linux, so there is no legitimate time to
	// skip it.

	// Do not use "/sys/bus/spi/devices/spi" as Raspbian's provided udev rules
	// only modify the ACL of /dev/spidev* but not the ones in /sys/bus/...
	prefix := "/dev/spidev"
	items, err := filepath.Glob(prefix + "*")
	if err != nil {
		return true, err
	}
	if len(items) == 0 {
		return false, errors.New("no SPI bus found")
	}
	sort.Strings(items)
	for _, item := range items {
		parts := strings.Split(item[len(prefix):], ".")
		if len(parts) != 2 {
			continue
		}
		bus, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		cs, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		if err := spi.Register(fmt.Sprintf("SPI%d.%d", bus, cs), bus, cs, func() (spi.ConnCloser, error) {
			return NewSPI(bus, cs)
		}); err != nil {
			return true, err
		}
	}
	return true, nil
}

func init() {
	if isLinux {
		periph.MustRegister(&driverSPI{})
	}
}

var _ spi.Conn = &SPI{}
