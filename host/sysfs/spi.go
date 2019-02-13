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
	"periph.io/x/periph/conn/physic"
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
// It is recommended to use https://periph.io/x/periph/conn/spi/spireg#Open
// instead of using NewSPI() directly as the package sysfs is providing a
// Linux-specific implementation. periph.io works on many OSes! This permits
// it to work on all operating systems, or devices like SPI over USB.
func NewSPI(busNumber, chipSelect int) (*SPI, error) {
	if isLinux {
		return newSPI(busNumber, chipSelect)
	}
	return nil, errors.New("sysfs-spi: not implemented on non-linux OSes")
}

// SPI is an open SPI port.
type SPI struct {
	conn spiConn
}

// Close closes the handle to the SPI driver. It is not a requirement to close
// before process termination.
//
// Note that the object is not reusable afterward.
func (s *SPI) Close() error {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	if err := s.conn.f.Close(); err != nil {
		return fmt.Errorf("sysfs-spi: %v", err)
	}
	s.conn.f = nil
	return nil
}

func (s *SPI) String() string {
	return s.conn.String()
}

// LimitSpeed implements spi.ConnCloser.
func (s *SPI) LimitSpeed(f physic.Frequency) error {
	if f > physic.GigaHertz {
		return fmt.Errorf("sysfs-spi: invalid speed %s; maximum supported clock is 1GHz", f)
	}
	if f < 100*physic.Hertz {
		return fmt.Errorf("sysfs-spi: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	s.conn.freqPort = f
	return nil
}

// Connect implements spi.Port.
//
// It must be called before any I/O.
func (s *SPI) Connect(f physic.Frequency, mode spi.Mode, bits int) (spi.Conn, error) {
	if f > physic.GigaHertz {
		return nil, fmt.Errorf("sysfs-spi: invalid speed %s; maximum supported clock is 1GHz", f)
	}
	if f < 100*physic.Hertz {
		return nil, fmt.Errorf("sysfs-spi: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.MegaHertz?", f)
	}
	if mode&^(spi.Mode3|spi.HalfDuplex|spi.NoCS|spi.LSBFirst) != 0 {
		return nil, fmt.Errorf("sysfs-spi: invalid mode %v", mode)
	}
	if bits < 1 || bits >= 256 {
		return nil, fmt.Errorf("sysfs-spi: invalid bits %d", bits)
	}
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()
	if s.conn.connected {
		return nil, errors.New("sysfs-spi: Connect() can only be called exactly once")
	}
	s.conn.connected = true
	s.conn.freqConn = f
	s.conn.bitsPerWord = uint8(bits)
	// Only mode needs to be set via an IOCTL, others can be specified in the
	// spiIOCTransfer packet, which saves a kernel call.
	m := mode & spi.Mode3
	s.conn.muPins.Lock()
	{
		if mode&spi.HalfDuplex != 0 {
			m |= threeWire
			s.conn.halfDuplex = true
			// In case initPins() had been called before Connect().
			s.conn.mosi = gpio.INVALID
		}
		if mode&spi.NoCS != 0 {
			m |= noCS
			s.conn.noCS = true
			// In case initPins() had been called before Connect().
			s.conn.cs = gpio.INVALID
		}
	}
	s.conn.muPins.Unlock()
	if mode&spi.LSBFirst != 0 {
		m |= lSBFirst
	}
	// Only the first 8 bits are used. This only works because the system is
	// running in little endian.
	if err := s.conn.setFlag(spiIOCMode, uint64(m)); err != nil {
		return nil, fmt.Errorf("sysfs-spi: setting mode %v failed: %v", mode, err)
	}
	return &s.conn, nil
}

// MaxTxSize implements conn.Limits
func (s *SPI) MaxTxSize() int {
	return drvSPI.bufSize
}

// CLK implements spi.Pins.
func (s *SPI) CLK() gpio.PinOut {
	return s.conn.CLK()
}

// MISO implements spi.Pins.
func (s *SPI) MISO() gpio.PinIn {
	return s.conn.MISO()
}

// MOSI implements spi.Pins.
func (s *SPI) MOSI() gpio.PinOut {
	return s.conn.MOSI()
}

// CS implements spi.Pins.
func (s *SPI) CS() gpio.PinOut {
	return s.conn.CS()
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
	return &SPI{
		spiConn{
			name:       fmt.Sprintf("SPI%d.%d", busNumber, chipSelect),
			f:          f,
			busNumber:  busNumber,
			chipSelect: chipSelect,
		},
	}, nil
}

//

// spiConn implements spi.Conn.
type spiConn struct {
	// Immutable
	name       string
	f          ioctlCloser
	busNumber  int
	chipSelect int

	mu          sync.Mutex
	freqPort    physic.Frequency // Frequency specified at LimitSpeed()
	freqConn    physic.Frequency // Frequency specified at Connect()
	bitsPerWord uint8
	connected   bool
	halfDuplex  bool
	noCS        bool
	// Heap optimization: reduce the amount of memory allocations during
	// transactions.
	io [4]spiIOCTransfer
	p  [2]spi.Packet

	// Use a separate lock for the pins, so that they can be queried while a
	// transaction is happening.
	muPins sync.Mutex
	clk    gpio.PinOut
	mosi   gpio.PinOut
	miso   gpio.PinIn
	cs     gpio.PinOut
}

func (s *spiConn) String() string {
	return s.name
}

// Read implements io.Reader.
func (s *spiConn) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("sysfs-spi: Read() with empty buffer")
	}
	if drvSPI.bufSize != 0 && len(b) > drvSPI.bufSize {
		return 0, fmt.Errorf("sysfs-spi: maximum Read length is %d, got %d bytes", drvSPI.bufSize, len(b))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.p[0].W = nil
	s.p[0].R = b
	if err := s.txPackets(s.p[:1]); err != nil {
		return 0, fmt.Errorf("sysfs-spi: Read() failed: %v", err)
	}
	return len(b), nil
}

// Write implements io.Writer.
func (s *spiConn) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, errors.New("sysfs-spi: Write() with empty buffer")
	}
	if drvSPI.bufSize != 0 && len(b) > drvSPI.bufSize {
		return 0, fmt.Errorf("sysfs-spi: maximum Write length is %d, got %d bytes", drvSPI.bufSize, len(b))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.p[0].W = b
	s.p[0].R = nil
	if err := s.txPackets(s.p[:1]); err != nil {
		return 0, fmt.Errorf("sysfs-spi: Write() failed: %v", err)
	}
	return len(b), nil
}

// Tx sends and receives data simultaneously.
//
// It is OK if both w and r point to the same underlying byte slice.
//
// spidev enforces the maximum limit of transaction size. It can be as low as
// 4096 bytes. See the platform documentation to learn how to increase the
// limit.
func (s *spiConn) Tx(w, r []byte) error {
	l := len(w)
	if l == 0 {
		if l = len(r); l == 0 {
			return errors.New("sysfs-spi: Tx() with empty buffers")
		}
	} else {
		// It's not a big deal to read halfDuplex without the lock.
		if !s.halfDuplex && len(r) != 0 && len(r) != len(w) {
			return fmt.Errorf("sysfs-spi: Tx(): when both w and r are used, they must be the same size; got %d and %d bytes", len(w), len(r))
		}
	}
	if drvSPI.bufSize != 0 && l > drvSPI.bufSize {
		return fmt.Errorf("sysfs-spi: maximum Tx length is %d, got %d bytes", drvSPI.bufSize, l)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.p[0].W = w
	s.p[0].R = r
	p := s.p[:1]
	if s.halfDuplex && len(w) != 0 && len(r) != 0 {
		// Create two packets for HalfDuplex operation: one write then one read.
		s.p[0].R = nil
		s.p[0].KeepCS = true
		s.p[1].W = nil
		s.p[1].R = r
		s.p[1].KeepCS = false
		p = s.p[:2]
	} else {
		s.p[0].KeepCS = false
	}
	if err := s.txPackets(p); err != nil {
		return fmt.Errorf("sysfs-spi: Tx() failed: %v", err)
	}
	return nil
}

// TxPackets sends and receives packets as specified by the user.
//
// spidev enforces the maximum limit of transaction size. It can be as low as
// 4096 bytes. See the platform documentation to learn how to increase the
// limit.
func (s *spiConn) TxPackets(p []spi.Packet) error {
	total := 0
	for i := range p {
		lW := len(p[i].W)
		lR := len(p[i].R)
		if lW != lR && lW != 0 && lR != 0 {
			return fmt.Errorf("sysfs-spi: when both w and r are used, they must be the same size; got %d and %d bytes", lW, lR)
		}
		l := lW
		if l == 0 {
			l = lR
		}
		total += l
	}
	if total == 0 {
		return errors.New("sysfs-spi: empty packets")
	}
	if drvSPI.bufSize != 0 && total > drvSPI.bufSize {
		return fmt.Errorf("sysfs-spi: maximum TxPackets length is %d, got %d bytes", drvSPI.bufSize, total)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.halfDuplex {
		for i := range p {
			if len(p[i].W) != 0 && len(p[i].R) != 0 {
				return errors.New("sysfs-spi: can only specify one of w or r when in half duplex")
			}
		}
	}
	if err := s.txPackets(p); err != nil {
		return fmt.Errorf("sysfs-spi: TxPackets() failed: %v", err)
	}
	return nil
}

// Duplex implements conn.Conn.
func (s *spiConn) Duplex() conn.Duplex {
	if s.halfDuplex {
		return conn.Half
	}
	return conn.Full
}

// MaxTxSize implements conn.Limits.
func (s *spiConn) MaxTxSize() int {
	return drvSPI.bufSize
}

// CLK implements spi.Pins.
func (s *spiConn) CLK() gpio.PinOut {
	s.initPins()
	return s.clk
}

// MISO implements spi.Pins.
func (s *spiConn) MISO() gpio.PinIn {
	s.initPins()
	return s.miso
}

// MOSI implements spi.Pins.
func (s *spiConn) MOSI() gpio.PinOut {
	s.initPins()
	return s.mosi
}

// CS implements spi.Pins.
func (s *spiConn) CS() gpio.PinOut {
	s.initPins()
	return s.cs
}

//

func (s *spiConn) txPackets(p []spi.Packet) error {
	// Convert the packets.
	f := s.freqPort
	if s.freqConn != 0 && (s.freqPort == 0 || s.freqConn < s.freqPort) {
		f = s.freqConn
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
		csInvert := false
		if !s.noCS {
			// Invert CS behavior when a packet has KeepCS false, except for the last
			// packet when KeepCS is true.
			last := i == len(p)-1
			csInvert = p[i].KeepCS == last
		}
		m[i].reset(p[i].W, p[i].R, f, bits, csInvert)
	}
	return s.f.Ioctl(spiIOCTx(len(m)), uintptr(unsafe.Pointer(&m[0])))
}

func (s *spiConn) setFlag(op uint, arg uint64) error {
	return s.f.Ioctl(op, uintptr(unsafe.Pointer(&arg)))
}

// GetFlag allows to read back flags set via a ioctl, i.e. setFlag. It is
// exported to allow calling it from the smoke test.
func (s *spiConn) GetFlag(op uint) (arg uint64, err error) {
	err = s.f.Ioctl(op, uintptr(unsafe.Pointer(&arg)))
	return
}

func (s *spiConn) initPins() {
	s.muPins.Lock()
	defer s.muPins.Unlock()
	if s.clk != nil {
		return
	}
	if s.clk = gpioreg.ByName(fmt.Sprintf("SPI%d_CLK", s.busNumber)); s.clk == nil {
		s.clk = gpio.INVALID
	}
	if s.miso = gpioreg.ByName(fmt.Sprintf("SPI%d_MISO", s.busNumber)); s.miso == nil {
		s.miso = gpio.INVALID
	}
	// s.mosi is set to INVALID if HalfDuplex was specified.
	if s.mosi != gpio.INVALID {
		if s.mosi = gpioreg.ByName(fmt.Sprintf("SPI%d_MOSI", s.busNumber)); s.mosi == nil {
			s.mosi = gpio.INVALID
		}
	}
	// s.cs is set to INVALID if NoCS was specified.
	if s.cs != gpio.INVALID {
		if s.cs = gpioreg.ByName(fmt.Sprintf("SPI%d_CS%d", s.busNumber, s.chipSelect)); s.cs == nil {
			s.cs = gpio.INVALID
		}
	}
}

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
const spiIOCMagic uint = 'k'

var (
	spiIOCMode        = fs.IOW(spiIOCMagic, 1, 1) // SPI_IOC_WR_MODE (8 bits)
	spiIOLSBFirst     = fs.IOW(spiIOCMagic, 2, 1) // SPI_IOC_WR_LSB_FIRST
	spiIOCBitsPerWord = fs.IOW(spiIOCMagic, 3, 1) // SPI_IOC_WR_BITS_PER_WORD
	spiIOCMaxSpeedHz  = fs.IOW(spiIOCMagic, 4, 4) // SPI_IOC_WR_MAX_SPEED_HZ
	spiIOCMode32      = fs.IOW(spiIOCMagic, 5, 4) // SPI_IOC_WR_MODE32 (32 bits)
)

// spiIOCTx(l) calculates the equivalent of SPI_IOC_MESSAGE(l) to execute a
// transaction.
func spiIOCTx(l int) uint {
	return fs.IOW(spiIOCMagic, 0, uint(l)*32)
}

// spiIOCTransfer is spi_ioc_transfer in linux/spi/spidev.h.
//
// Also documented as struct spi_transfer at
// https://www.kernel.org/doc/html/latest/driver-api/spi.html
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

func (s *spiIOCTransfer) reset(w, r []byte, f physic.Frequency, bitsPerWord uint8, csInvert bool) {
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
	s.speedHz = uint32((f + 500*physic.MilliHertz) / physic.Hertz)
	s.delayUsecs = 0
	s.bitsPerWord = bitsPerWord
	if csInvert {
		s.csChange = 1
	} else {
		s.csChange = 0
	}
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
