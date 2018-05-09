// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	"periph.io/x/periph/host/fs"
)

// NewSPI opens a SPI port via its devfs interface as described at
// https://www.kernel.org/doc/Documentation/spi/spidev and
// https://www.kernel.org/doc/Documentation/spi/spi-summary
//
// The resulting object is safe for concurrent use.
//
// busNumber is the bus number as exported by devfs. For example if the path is
// /dev/spidev0.1, busNumber should be 0 and chipSelect should be 1.
//
// Do not use sysfs.NewSPI() directly as the package sysfs is providing a
// https://periph.io/x/periph/conn/spi Linux-specific implementation.
//
// periph.io works on many OSes!
//
// Instead, use https://periph.io/x/periph/conn/spi/spireg#Open. This permits
// it to work on all operating systems, or devices like SPI over USB.
func NewSPI(busNumber, chipSelect int) (*SPI, error) {
	if isLinux {
		return newSPI(busNumber, chipSelect)
	}
	return nil, errors.New("sysfs-spi: not implemented on non-linux OSes")
}

// SPI is an open SPI port.
type SPI struct {
	// Immutable
	f          ioctlCloser
	busNumber  int
	chipSelect int

	sync.Mutex
	initialized bool
	bitsPerWord uint8
	halfDuplex  bool
	noCS        bool
	maxHzPort   uint32
	maxHzDev    uint32
	clk         gpio.PinOut
	mosi        gpio.PinOut
	miso        gpio.PinIn
	cs          gpio.PinOut
	spiConn     spiConn
	// Heap optimization: reduce the amount of memory allocations.
	io [4]spiIOCTransfer
}

// Close closes the handle to the SPI driver. It is not a requirement to close
// before process termination.
//
// Note that the object is not reusable afterward.
func (s *SPI) Close() error {
	s.Lock()
	defer s.Unlock()
	if err := s.f.Close(); err != nil {
		return fmt.Errorf("sysfs-spi: %v", err)
	}
	return nil
}

func (s *SPI) String() string {
	return fmt.Sprintf("SPI%d.%d", s.busNumber, s.chipSelect)
}

// LimitSpeed implements spi.ConnCloser.
func (s *SPI) LimitSpeed(maxHz int64) error {
	if maxHz < 1 || maxHz >= 1<<32 {
		return fmt.Errorf("sysfs-spi: invalid speed %d", maxHz)
	}
	s.Lock()
	defer s.Unlock()
	s.maxHzPort = uint32(maxHz)
	return nil
}

// Connect implements spi.Port.
//
// It must be called before any I/O.
func (s *SPI) Connect(maxHz int64, mode spi.Mode, bits int) (spi.Conn, error) {
	if maxHz < 0 || maxHz >= 1<<32 {
		return nil, fmt.Errorf("sysfs-spi: invalid speed %d", maxHz)
	}
	if mode&^(spi.Mode3|spi.HalfDuplex|spi.NoCS|spi.LSBFirst) != 0 {
		return nil, fmt.Errorf("sysfs-spi: invalid mode %v", mode)
	}
	if bits < 1 || bits >= 256 {
		return nil, fmt.Errorf("sysfs-spi: invalid bits %d", bits)
	}
	s.Lock()
	defer s.Unlock()
	if s.initialized {
		return nil, errors.New("sysfs-spi: Connect() can only be called exactly once")
	}
	s.initialized = true
	s.maxHzDev = uint32(maxHz)
	s.bitsPerWord = uint8(bits)
	// Only mode needs to be set via an IOCTL, others can be specified in the
	// spiIOCTransfer packet, which saves a kernel call.
	m := mode & spi.Mode3
	if mode&spi.HalfDuplex != 0 {
		m |= threeWire
		s.halfDuplex = true
	}
	if mode&spi.NoCS != 0 {
		m |= noCS
		s.noCS = true
	}
	if mode&spi.LSBFirst != 0 {
		m |= lSBFirst
	}
	// Only the first 8 bits are used. This only works because the system is
	// running in little endian.
	if err := s.setFlag(spiIOCMode, uint64(m)); err != nil {
		return nil, fmt.Errorf("sysfs-spi: setting mode %v failed: %v", mode, err)
	}
	s.spiConn.s = s
	return &s.spiConn, nil
}

func (s *SPI) duplex() conn.Duplex {
	s.Lock()
	defer s.Unlock()
	if s.halfDuplex {
		return conn.Half
	}
	return conn.Full
}

// MaxTxSize implements conn.Limits
func (s *SPI) MaxTxSize() int {
	return drvSPI.bufSize
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
	// TODO(maruel): spi.HalfDuplex.
	return s.mosi
}

// CS implements spi.Pins.
func (s *SPI) CS() gpio.PinOut {
	s.initPins()
	// TODO(maruel): spi.NoCS and generally fix properly.
	return s.cs
}

// Private details.

func newSPI(busNumber, chipSelect int) (*SPI, error) {
	if busNumber < 0 || busNumber >= 1<<16 {
		return nil, fmt.Errorf("sysfs-spi: invalid bus %d", busNumber)
	}
	if chipSelect < 0 || chipSelect > 255 {
		return nil, fmt.Errorf("sysfs-spi: invalid chip select %d", chipSelect)
	}
	// Use the devfs path for now.
	f, err := ioctlOpen(fmt.Sprintf("/dev/spidev%d.%d", busNumber, chipSelect), os.O_RDWR)
	if err != nil {
		return nil, fmt.Errorf("sysfs-spi: %v", err)
	}
	return &SPI{f: f, busNumber: busNumber, chipSelect: chipSelect}, nil
}

func (s *SPI) txInternal(w, r []byte) (int, error) {
	l := len(w)
	if l == 0 {
		l = len(r)
	}
	if drvSPI.bufSize != 0 && l > drvSPI.bufSize {
		return 0, fmt.Errorf("sysfs-spi: maximum Tx length is %d, got at least %d bytes", drvSPI.bufSize, l)
	}

	s.Lock()
	defer s.Unlock()
	if !s.initialized {
		return 0, errors.New("sysfs-spi: Connect wasn't called")
	}
	if len(w) != 0 && len(r) != 0 && s.halfDuplex {
		return 0, errors.New("sysfs-spi: can only specify one of w or r when in half duplex")
	}
	hz := s.maxHzPort
	if s.maxHzDev != 0 && (s.maxHzPort == 0 || s.maxHzDev < s.maxHzPort) {
		hz = s.maxHzDev
	}
	// The Ioctl() call below is seen as a memory escape, so the spiIOCTransfer
	// object cannot be on the stack.
	s.io[0].reset(w, r, hz, s.bitsPerWord)
	if err := s.f.Ioctl(spiIOCTx(1), uintptr(unsafe.Pointer(&s.io[0]))); err != nil {
		return 0, fmt.Errorf("sysfs-spi: I/O failed: %v", err)
	}
	return l, nil
}

func (s *SPI) txPackets(p []spi.Packet) error {
	total := 0
	for i := range p {
		lW := len(p[i].W)
		lR := len(p[i].R)
		if lW != lR && lW != 0 && lR != 0 {
			return fmt.Errorf("sysfs-spi: when both w and r are used, they must be the same size; got %d and %d bytes", lW, lR)
		}
		l := 0
		if lW != 0 {
			l = lW
		}
		if lR != 0 {
			l = lR
		}
		if total += l; drvSPI.bufSize != 0 && total > drvSPI.bufSize {
			return fmt.Errorf("sysfs-spi: maximum TxPackets length is %d, got at least %d bytes", drvSPI.bufSize, total)
		}
	}
	if total == 0 {
		return errors.New("sysfs-spi: empty packets")
	}

	s.Lock()
	defer s.Unlock()
	if !s.initialized {
		return errors.New("sysfs-spi: Connect wasn't called")
	}
	if s.halfDuplex {
		for i := range p {
			if len(p[i].W) != 0 && len(p[i].R) != 0 {
				return errors.New("sysfs-spi: can only specify one of w or r when in half duplex")
			}
		}
	}

	// Convert the packets.
	hz := s.maxHzPort
	if s.maxHzDev != 0 && (s.maxHzPort == 0 || s.maxHzDev < s.maxHzPort) {
		hz = s.maxHzDev
	}
	var m []spiIOCTransfer
	if len(p) > len(s.io) {
		m = make([]spiIOCTransfer, len(p))
	} else {
		m = s.io[:len(p)]
	}
	for i := range p {
		bits := p[i].BitsPerWord
		if bits == 0 {
			bits = s.bitsPerWord
		}
		m[i].reset(p[i].W, p[i].R, hz, bits)
		if !s.noCS && !p[i].KeepCS {
			m[i].csChange = 1
		}
	}
	if err := s.f.Ioctl(spiIOCTx(len(m)), uintptr(unsafe.Pointer(&m[0]))); err != nil {
		return fmt.Errorf("sysfs-spi: TxPackets(%d) packets failed: %v", len(m), err)
	}
	return nil
}

func (s *SPI) setFlag(op uint, arg uint64) error {
	if err := s.f.Ioctl(op|0x40000000, uintptr(unsafe.Pointer(&arg))); err != nil {
		return err
	}
	/*
		// Verification.
		actual := uint64(0)
		// getFlag() equivalent.
		if err := s.f.Ioctl(op|0x80000000, unsafe.Pointer(&actual)); err != nil {
			return err
		}
		if actual != arg {
			return fmt.Errorf("sysfs-spi: op 0x%x: set 0x%x, read 0x%x", op, arg, actual)
		}
	*/
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

// spiConn implements spi.Conn.
type spiConn struct {
	s *SPI
}

func (s *spiConn) String() string {
	return s.s.String()
}

// Read implements io.Reader.
func (s *spiConn) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("sysfs-spi: Read() with empty buffer")
	}
	return s.s.txInternal(nil, b)
}

// Write implements io.Writer.
func (s *spiConn) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("sysfs-spi: Write() with empty buffer")
	}
	return s.s.txInternal(b, nil)
}

// Tx sends and receives data simultaneously.
//
// It is OK if both w and r point to the same underlying byte slice.
//
// spidev enforces the maximum limit of transaction size. It can be as low as
// 4096 bytes. See the platform documentation to learn how to increase the
// limit.
func (s *spiConn) Tx(w, r []byte) error {
	if len(w) == 0 {
		if len(r) == 0 {
			return errors.New("sysfs-spi: Tx with empty buffers")
		}
	} else {
		if len(r) != 0 && len(w) != len(r) {
			return errors.New("sysfs-spi: Tx with zero or non-equal length w&r slices")
		}
	}
	_, err := s.s.txInternal(w, r)
	return err
}

// TxPackets sends and receives packets as specified by the user.
//
// spidev enforces the maximum limit of transaction size. It can be as low as
// 4096 bytes. See the platform documentation to learn how to increase the
// limit.
func (s *spiConn) TxPackets(p []spi.Packet) error {
	return s.s.txPackets(p)
}

func (s *spiConn) Duplex() conn.Duplex {
	return s.s.duplex()
}

func (s *spiConn) MaxTxSize() int {
	return drvSPI.bufSize
}

func (s *spiConn) CLK() gpio.PinOut {
	return s.s.CLK()
}

func (s *spiConn) MISO() gpio.PinIn {
	return s.s.MISO()
}

func (s *spiConn) MOSI() gpio.PinOut {
	return s.s.MOSI()
}

func (s *spiConn) CS() gpio.PinOut {
	return s.s.CS()
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

func (s *spiIOCTransfer) reset(w, r []byte, speedHz uint32, bitsPerWord uint8) {
	s.tx = 0
	s.rx = 0
	s.length = 0
	// w and r must be the same length.
	if l := len(w); l != 0 {
		s.tx = uint64(uintptr(unsafe.Pointer(&w[0])))
		s.length = uint32(l)
	}
	if l := len(r); l != 0 {
		s.rx = uint64(uintptr(unsafe.Pointer(&r[0])))
		s.length = uint32(l)
	}
	s.speedHz = speedHz
	s.delayUsecs = 0
	s.bitsPerWord = bitsPerWord
	s.csChange = 0
	s.txNBits = 0
	s.rxNBits = 0
	s.pad = 0
}

//

// driverSPI implements periph.Driver.
type driverSPI struct {
	// bufSize is the maximum number of bytes allowed per I/O on the SPI port.
	bufSize int
}

func (d *driverSPI) String() string {
	return "sysfs-spi"
}

func (d *driverSPI) Prerequisites() []string {
	return nil
}

func (d *driverSPI) After() []string {
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
		return false, errors.New("no SPI port found")
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
	f, err := fs.Open("/sys/module/spidev/parameters/bufsiz", os.O_RDONLY)
	if err != nil {
		return true, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return true, err
	}
	// Update the global value.
	drvSPI.bufSize, err = strconv.Atoi(strings.TrimSpace(string(b)))
	return true, err
}

type openerSPI struct {
	bus int
	cs  int
}

func (o *openerSPI) Open() (spi.PortCloser, error) {
	return NewSPI(o.bus, o.cs)
}

func init() {
	if isLinux {
		periph.MustRegister(&drvSPI)
	}
}

var drvSPI driverSPI

var _ conn.Limits = &SPI{}
var _ conn.Limits = &spiConn{}
var _ io.Reader = &spiConn{}
var _ io.Writer = &spiConn{}
var _ spi.Conn = &spiConn{}
var _ spi.Pins = &SPI{}
var _ spi.Pins = &spiConn{}
var _ spi.Port = &SPI{}
var _ spi.PortCloser = &SPI{}
var _ fmt.Stringer = &SPI{}
var _ fmt.Stringer = &spiConn{}
