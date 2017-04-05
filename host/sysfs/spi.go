// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"periph.io/x/periph"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
)

// NewSPI opens a SPI bus via its devfs interface as described at
// https://www.kernel.org/doc/Documentation/spi/spidev and
// https://www.kernel.org/doc/Documentation/spi/spi-summary
//
// The resulting object is safe for concurrent use.
//
// busNumber is the bus number as exported by deffs. For example if the path is
// /dev/spidev0.1, busNumber should be 0 and chipSelect should be 1.
func NewSPI(busNumber, chipSelect int) (*SPI, error) {
	if isLinux {
		return newSPI(busNumber, chipSelect)
	}
	return nil, errors.New("sysfs-spi: not implemented on non-linux OSes")
}

// SPI is an open SPI bus.
type SPI struct {
	// Immutable
	frwc       io.ReadWriteCloser
	fd         uintptr
	busNumber  int
	chipSelect int

	sync.Mutex
	initialized bool
	maxHzBus    uint32
	maxHzDev    uint32
	bitsPerWord uint8
	clk         gpio.PinOut
	mosi        gpio.PinOut
	miso        gpio.PinIn
	cs          gpio.PinOut
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
	return &SPI{frwc: f, fd: f.Fd(), busNumber: busNumber, chipSelect: chipSelect}, nil
}

// Close closes the handle to the SPI driver. It is not a requirement to close
// before process termination.
func (s *SPI) Close() error {
	s.Lock()
	defer s.Unlock()
	return s.frwc.Close()
}

func (s *SPI) String() string {
	return fmt.Sprintf("SPI%d.%d", s.busNumber, s.chipSelect)
}

// Speed implements spi.ConnCloser.
func (s *SPI) Speed(maxHz int64) error {
	if maxHz < 1 || maxHz >= 1<<32 {
		return fmt.Errorf("sysfs-spi: invalid speed %d", maxHz)
	}
	s.Lock()
	defer s.Unlock()
	s.maxHzBus = uint32(maxHz)
	return nil
}

// DevParams implements spi.Conn.
//
// It must be called before any I/O.
func (s *SPI) DevParams(maxHz int64, mode spi.Mode, bits int) error {
	if bits < 1 || bits >= 256 {
		return fmt.Errorf("sysfs-spi: invalid bits %d", bits)
	}
	if maxHz < 0 || maxHz >= 1<<32 {
		return fmt.Errorf("sysfs-spi: invalid speed %d", maxHz)
	}
	s.Lock()
	defer s.Unlock()
	if s.initialized {
		return errors.New("sysfs-spi: DevParams() can only be called exactly once")
	}
	s.initialized = true
	s.maxHzDev = uint32(maxHz)
	s.bitsPerWord = uint8(bits)
	// Only mode needs to be set via an IOCTL, others can be specified in the
	// spiIOCTransfer packet, which saves a kernel call.
	return s.setFlag(spiIOCMode, uint64(mode))
}

// Read implements io.Reader.
func (s *SPI) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("Read() with empty buffer")
	}
	return s.txInternal(nil, b)
}

// Write implements io.Writer.
func (s *SPI) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("Write() with empty buffer")
	}
	return s.txInternal(b, nil)
}

// Tx sends and receives data simultaneously.
//
// It is OK if both w and r point to the same underlying byte slice.
//
// Tx() implements transparent support for large I/O operations of more than
// 4096 bytes. The main problem is that if the same bus with another CS line is
// opened, it is possible that a transaction on the other device happens in
// between. So do not use multiple devices with separate CS lines on the same
// bus when doing I/O operations of more than 4096 bytes.
func (s *SPI) Tx(w, r []byte) error {
	if len(w) == 0 {
		if len(r) == 0 {
			return errors.New("Tx with empty buffers")
		}
	} else {
		if len(r) != 0 && len(w) != len(r) {
			return errors.New("Tx with zero or non-equal length w&r slices")
		}
	}
	_, err := s.txInternal(w, r)
	return err
}

// Duplex implements spi.Conn.
func (s *SPI) Duplex() conn.Duplex {
	// If half-duplex SPI is ever supported, change this code.
	return conn.Full
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

func (s *SPI) txInternal(w, r []byte) (int, error) {
	// TODO(maruel): The driver supports a series of half-duplex transfer, which
	// is needed in 3-wire SPI mode and when using 9 bits command but 8 bits data.
	s.Lock()
	defer s.Unlock()
	if !s.initialized {
		return 0, errors.New("sysfs-spi: DevParams wasn't called")
	}
	// Most spidev drivers limit each buffer to one page size (4096 bytes) so
	// chunk the I/O into multiple pages to increase compatibility, keeping CS
	// low in between (only the clock is paused).
	p := spiIOCTransfer{
		speedHz:     s.maxHzBus,
		bitsPerWord: s.bitsPerWord,
	}
	if s.maxHzDev != 0 && (s.maxHzBus == 0 || s.maxHzDev < s.maxHzBus) {
		p.speedHz = s.maxHzDev
	}

	n := 0
	for len(w) != 0 && len(r) != 0 {
		p.csChange = 1
		if l := len(w); l != 0 {
			if l > 4096 {
				// Limit kernel calls to one page and keep CS low.
				l = 4096
				p.csChange = 0
			}
			p.tx = uint64(uintptr(unsafe.Pointer(&w[0])))
			p.length = uint32(l)
			w = w[l:]
		}
		if l := len(r); l != 0 {
			if l > 4096 {
				// Limit kernel calls to one page and keep CS low.
				l = 4096
				p.csChange = 0
			}
			p.rx = uint64(uintptr(unsafe.Pointer(&r[0])))
			p.length = uint32(l)
			r = r[l:]
		}
		if err := s.ioctl(spiIOCTx(1), unsafe.Pointer(&p)); err != nil {
			return n, err
		}
		n += int(p.length)
	}
	return n, nil
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
	if err := ioctl(s.fd, op, uintptr(arg)); err != nil {
		return fmt.Errorf("sysfs-spi: ioctl: %v", err)
	}
	return nil
}

func (s *SPI) initPins() {
	s.Lock()
	isInitialized := s.clk != nil
	s.Unlock()

	if !isInitialized {
		clk := gpioreg.ByName(fmt.Sprintf("SPI%d_CLK", s.busNumber))
		if clk == nil {
			clk = gpio.INVALID
		}
		miso := gpioreg.ByName(fmt.Sprintf("SPI%d_MISO", s.busNumber))
		if miso == nil {
			miso = gpio.INVALID
		}
		mosi := gpioreg.ByName(fmt.Sprintf("SPI%d_MOSI", s.busNumber))
		if mosi == nil {
			mosi = gpio.INVALID
		}
		cs := gpioreg.ByName(fmt.Sprintf("SPI%d_CS%d", s.busNumber, s.chipSelect))
		if cs == nil {
			cs = gpio.INVALID
		}

		s.Lock()
		s.clk = clk
		s.miso = miso
		s.mosi = mosi
		s.cs = cs
		s.Unlock()
	}
}

//

const (
	cSHigh    spi.Mode = 0x4  // CS active high instead of default low (not recommended)
	lSBFirst  spi.Mode = 0x8  // Use little endian encoding for each word
	threeWire spi.Mode = 0x10 // half-duplex; MOSI and MISO are shared
	loop      spi.Mode = 0x20 // loopback mode
	noCS      spi.Mode = 0x40 // do not assert CS
	ready     spi.Mode = 0x80 // slave pulls low to pause
	// The driver optionally support dual and quad data lines.
)

// spidev driver IOCTL control codes.
//
// Constants and structure definition can be found at
// /usr/include/linux/spi/spidev.h.
const (
	spiIOCMode        = 0x16B01 // SPI_IOC_WR_MODE (8 bits)
	spiIOLSBFirst     = 0x16B02 // SPI_IOC_WR_LSB_FIRST
	spiIOCBitsPerWord = 0x16B03 // SPI_IOC_WR_BITS_PER_WORD
	spiIOCMaxSpeedHz  = 0x46B04 // SPI_IOC_WR_MAX_SPEED_HZ
	spiIOCMode32      = 0x46B05 // SPI_IOC_WR_MODE32 (32 bits)
)

// spiIOCTx(l) calculates the equivalent of SPI_IOC_MESSAGE(l) to execute a
// transaction.
//
// The IOCTL for TX was deduced from this C code:
//
//   #include "linux/spi/spidev.h"
//   #include "sys/ioctl.h"
//   #include <stdio.h>
//   int main() {
//     for (int i = 1; i < 10; i++) {
//       printf("len(%d) = 0x%08X\n", i, SPI_IOC_MESSAGE(i));
//     }
//     return 0;
//   }
//
//   $ gcc a.cc && ./a.out
//   len(1) = 0x40206B00
//   len(2) = 0x40406B00
//   len(3) = 0x40606B00
//   len(4) = 0x40806B00
//   len(5) = 0x40A06B00
//   len(6) = 0x40C06B00
//   len(7) = 0x40E06B00
//   len(8) = 0x41006B00
//   len(9) = 0x41206B00
func spiIOCTx(l int) uint {
	op := uint(0x40006B00)
	return op | uint(0x200000)*uint(l)
}

// spiIOCTransfer is spi_ioc_transfer in linux/spi/spidev.h.
type spiIOCTransfer struct {
	tx          uint64 // Pointer to byte slice
	rx          uint64 // Pointer to byte slice
	length      uint32 // buffer length of tx and rx in bytes
	speedHz     uint32 // temporarily override the speed
	delayUsecs  uint16 // µs to sleep before selecting the device before the next transfer
	bitsPerWord uint8  // temporarily override the number of bytes per word
	csChange    uint8  // true to deassert CS before next transfer
	txNBits     uint8
	rxNBits     uint8
	pad         uint16
}

//

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
		name := fmt.Sprintf("/dev/spidev%d.%d", bus, cs)
		aliases := []string{fmt.Sprintf("SPI%d.%d", bus, cs)}
		n := bus
		if cs != 0 {
			n = -1
		}
		if err := spireg.Register(name, aliases, n, (&openerSPI{bus, cs}).Open); err != nil {
			return true, err
		}
	}
	return true, nil
}

type openerSPI struct {
	bus int
	cs  int
}

func (o *openerSPI) Open() (spi.ConnCloser, error) {
	return NewSPI(o.bus, o.cs)
}

func init() {
	if isLinux {
		periph.MustRegister(&driverSPI{})
	}
}

var _ spi.Conn = &SPI{}
var _ io.Reader = &SPI{}
var _ io.Writer = &SPI{}
