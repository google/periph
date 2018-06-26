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
	"sync"
	"unsafe"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/physic"
)

// I2CSetSpeedHook can be set by a driver to enable changing the I²C buses
// speed.
func I2CSetSpeedHook(h func(f physic.Frequency) error) error {
	if h == nil {
		return errors.New("sysfs-i2c: hook must not be nil")
	}
	drvI2C.mu.Lock()
	defer drvI2C.mu.Unlock()
	if drvI2C.setSpeed != nil {
		return errors.New("sysfs-i2c: a speed hook was already set")
	}
	drvI2C.setSpeed = h
	return nil
}

// NewI2C opens an I²C bus via its sysfs interface as described at
// https://www.kernel.org/doc/Documentation/i2c/dev-interface.
//
// busNumber is the bus number as exported by sysfs. For example if the path is
// /dev/i2c-1, busNumber should be 1.
//
// The resulting object is safe for concurent use.
//
// Do not use sysfs.NewI2C() directly as the package sysfs is providing a
// https://periph.io/x/periph/conn/i2c Linux-specific implementation.
//
// periph.io works on many OSes!
//
// Instead, use https://periph.io/x/periph/conn/i2c/i2creg#Open. This permits
// it to work on all operating systems, or devices like I²C over USB.
func NewI2C(busNumber int) (*I2C, error) {
	if isLinux {
		return newI2C(busNumber)
	}
	return nil, errors.New("sysfs-i2c: is not supported on this platform")
}

// I2C is an open I²C bus via sysfs.
//
// It can be used to communicate with multiple devices from multiple goroutines.
type I2C struct {
	f         ioctlCloser
	busNumber int

	mu  sync.Mutex // In theory the kernel probably has an internal lock but not taking any chance.
	fn  functionality
	scl gpio.PinIO
	sda gpio.PinIO
}

// Close closes the handle to the I²C driver. It is not a requirement to close
// before process termination.
func (i *I2C) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if err := i.f.Close(); err != nil {
		return fmt.Errorf("sysfs-i2c: %v", err)
	}
	return nil
}

func (i *I2C) String() string {
	return fmt.Sprintf("I2C%d", i.busNumber)
}

// Tx execute a transaction as a single operation unit.
func (i *I2C) Tx(addr uint16, w, r []byte) error {
	if addr >= 0x400 || (addr >= 0x80 && i.fn&func10BitAddr == 0) {
		return errors.New("sysfs-i2c: invalid address")
	}
	if len(w) == 0 && len(r) == 0 {
		return nil
	}

	// Convert the messages to the internal format.
	var buf [2]i2cMsg
	msgs := buf[0:0]
	if len(w) != 0 {
		msgs = buf[:1]
		buf[0].addr = addr
		buf[0].length = uint16(len(w))
		buf[0].buf = uintptr(unsafe.Pointer(&w[0]))
	}
	if len(r) != 0 {
		l := len(msgs)
		msgs = msgs[:l+1] // extend the slice by one
		buf[l].addr = addr
		buf[l].flags = flagRD
		buf[l].length = uint16(len(r))
		buf[l].buf = uintptr(unsafe.Pointer(&r[0]))
	}
	p := rdwrIoctlData{
		msgs:  uintptr(unsafe.Pointer(&msgs[0])),
		nmsgs: uint32(len(msgs)),
	}
	pp := uintptr(unsafe.Pointer(&p))
	i.mu.Lock()
	defer i.mu.Unlock()
	if err := i.f.Ioctl(ioctlRdwr, pp); err != nil {
		return fmt.Errorf("sysfs-i2c: %v", err)
	}
	return nil
}

// SetSpeed implements i2c.Bus.
func (i *I2C) SetSpeed(f physic.Frequency) error {
	if f > 100*physic.MegaHertz {
		return fmt.Errorf("sysfs-i2c: invalid speed %s; maximum supported clock is 100MHz", f)
	}
	if f < physic.KiloHertz {
		return fmt.Errorf("sysfs-i2c: invalid speed %s; minimum supported clock is 1KHz; did you forget to multiply by physic.KiloHertz?", f)
	}
	drvI2C.mu.Lock()
	defer drvI2C.mu.Unlock()
	if drvI2C.setSpeed != nil {
		return drvI2C.setSpeed(f)
	}
	return errors.New("sysfs-i2c: not supported")
}

// SCL implements i2c.Pins.
func (i *I2C) SCL() gpio.PinIO {
	i.initPins()
	return i.scl
}

// SDA implements i2c.Pins.
func (i *I2C) SDA() gpio.PinIO {
	i.initPins()
	return i.sda
}

// Private details.

func newI2C(busNumber int) (*I2C, error) {
	// Use the devfs path for now instead of sysfs path.
	f, err := ioctlOpen(fmt.Sprintf("/dev/i2c-%d", busNumber), os.O_RDWR)
	if err != nil {
		// Try to be helpful here. There are generally two cases:
		// - /dev/i2c-X doesn't exist. In this case, /boot/config.txt has to be
		//   edited to enable I²C then the device must be rebooted.
		// - permission denied. In this case, the user has to be added to plugdev.
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("sysfs-i2c: bus #%d is not configured: %v", busNumber, err)
		}
		// TODO(maruel): This is a debianism.
		return nil, fmt.Errorf("sysfs-i2c: are you member of group 'plugdev'? %v", err)
	}
	i := &I2C{f: f, busNumber: busNumber}

	// TODO(maruel): Changing the speed is currently doing this for all devices.
	// https://github.com/raspberrypi/linux/issues/215
	// Need to access /sys/module/i2c_bcm2708/parameters/baudrate

	// Query to know if 10 bits addresses are supported.
	if err = i.f.Ioctl(ioctlFuncs, uintptr(unsafe.Pointer(&i.fn))); err != nil {
		return nil, fmt.Errorf("sysfs-i2c: %v", err)
	}
	return i, nil
}

func (i *I2C) initPins() {
	i.mu.Lock()
	if i.scl == nil {
		if i.scl = gpioreg.ByName(fmt.Sprintf("I2C%d_SCL", i.busNumber)); i.scl == nil {
			i.scl = gpio.INVALID
		}
		if i.sda = gpioreg.ByName(fmt.Sprintf("I2C%d_SDA", i.busNumber)); i.sda == nil {
			i.sda = gpio.INVALID
		}
	}
	i.mu.Unlock()
}

// i2cdev driver IOCTL control codes.
//
// Constants and structure definition can be found at
// /usr/include/linux/i2c-dev.h and /usr/include/linux/i2c.h.
const (
	ioctlRetries = 0x701 // TODO(maruel): Expose this
	ioctlTimeout = 0x702 // TODO(maruel): Expose this; in units of 10ms
	ioctlSlave   = 0x703
	ioctlTenBits = 0x704 // TODO(maruel): Expose this but the header says it's broken (!?)
	ioctlFuncs   = 0x705
	ioctlRdwr    = 0x707
)

// flags
const (
	flagTEN        = 0x0010 // this is a ten bit chip address
	flagRD         = 0x0001 // read data, from slave to master
	flagSTOP       = 0x8000 // if funcProtocolMangling
	flagNOSTART    = 0x4000 // if I2C_FUNC_NOSTART
	flagRevDirAddr = 0x2000 // if funcProtocolMangling
	flagIgnoreNAK  = 0x1000 // if funcProtocolMangling
	flagNoRDACK    = 0x0800 // if funcProtocolMangling
	flagRecvLen    = 0x0400 // length will be first received byte

)

type functionality uint64

const (
	funcI2C                 = 0x00000001
	func10BitAddr           = 0x00000002
	funcProtocolMangling    = 0x00000004 // I2C_M_IGNORE_NAK etc.
	funcSMBusPEC            = 0x00000008
	funcNOSTART             = 0x00000010 // I2C_M_NOSTART
	funcSMBusBlockProcCall  = 0x00008000 // SMBus 2.0
	funcSMBusQuick          = 0x00010000
	funcSMBusReadByte       = 0x00020000
	funcSMBusWriteByte      = 0x00040000
	funcSMBusReadByteData   = 0x00080000
	funcSMBusWriteByteData  = 0x00100000
	funcSMBusReadWordData   = 0x00200000
	funcSMBusWriteWordData  = 0x00400000
	funcSMBusProcCall       = 0x00800000
	funcSMBusReadBlockData  = 0x01000000
	funcSMBusWriteBlockData = 0x02000000
	funcSMBusReadI2CBlock   = 0x04000000 // I2C-like block xfer
	funcSMBusWriteI2CBlock  = 0x08000000 // w/ 1-byte reg. addr.
)

func (f functionality) String() string {
	var out []string
	if f&funcI2C != 0 {
		out = append(out, "I2C")
	}
	if f&func10BitAddr != 0 {
		out = append(out, "10BIT_ADDR")
	}
	if f&funcProtocolMangling != 0 {
		out = append(out, "PROTOCOL_MANGLING")
	}
	if f&funcSMBusPEC != 0 {
		out = append(out, "SMBUS_PEC")
	}
	if f&funcNOSTART != 0 {
		out = append(out, "NOSTART")
	}
	if f&funcSMBusBlockProcCall != 0 {
		out = append(out, "SMBUS_BLOCK_PROC_CALL")
	}
	if f&funcSMBusQuick != 0 {
		out = append(out, "SMBUS_QUICK")
	}
	if f&funcSMBusReadByte != 0 {
		out = append(out, "SMBUS_READ_BYTE")
	}
	if f&funcSMBusWriteByte != 0 {
		out = append(out, "SMBUS_WRITE_BYTE")
	}
	if f&funcSMBusReadByteData != 0 {
		out = append(out, "SMBUS_READ_BYTE_DATA")
	}
	if f&funcSMBusWriteByteData != 0 {
		out = append(out, "SMBUS_WRITE_BYTE_DATA")
	}
	if f&funcSMBusReadWordData != 0 {
		out = append(out, "SMBUS_READ_WORD_DATA")
	}
	if f&funcSMBusWriteWordData != 0 {
		out = append(out, "SMBUS_WRITE_WORD_DATA")
	}
	if f&funcSMBusProcCall != 0 {
		out = append(out, "SMBUS_PROC_CALL")
	}
	if f&funcSMBusReadBlockData != 0 {
		out = append(out, "SMBUS_READ_BLOCK_DATA")
	}
	if f&funcSMBusWriteBlockData != 0 {
		out = append(out, "SMBUS_WRITE_BLOCK_DATA")
	}
	if f&funcSMBusReadI2CBlock != 0 {
		out = append(out, "SMBUS_READ_I2C_BLOCK")
	}
	if f&funcSMBusWriteI2CBlock != 0 {
		out = append(out, "SMBUS_WRITE_I2C_BLOCK")
	}
	return strings.Join(out, "|")
}

type rdwrIoctlData struct {
	msgs  uintptr // Pointer to i2cMsg
	nmsgs uint32
}

type i2cMsg struct {
	addr   uint16 // Address to communicate with
	flags  uint16 // 1 for read, see i2c.h for more details
	length uint16
	buf    uintptr
}

//

// driverI2C implements periph.Driver.
type driverI2C struct {
	mu       sync.Mutex
	buses    []string
	setSpeed func(f physic.Frequency) error
}

func (d *driverI2C) String() string {
	return "sysfs-i2c"
}

func (d *driverI2C) Prerequisites() []string {
	return nil
}

func (d *driverI2C) After() []string {
	return nil
}

func (d *driverI2C) Init() (bool, error) {
	// Do not use "/sys/bus/i2c/devices/i2c-" as Raspbian's provided udev rules
	// only modify the ACL of /dev/i2c-* but not the ones in /sys/bus/...
	prefix := "/dev/i2c-"
	items, err := filepath.Glob(prefix + "*")
	if err != nil {
		return true, err
	}
	if len(items) == 0 {
		return false, errors.New("no I²C bus found")
	}
	// Make sure they are registered in order.
	sort.Strings(items)
	for _, item := range items {
		bus, err := strconv.Atoi(item[len(prefix):])
		if err != nil {
			continue
		}
		name := fmt.Sprintf("/dev/i2c-%d", bus)
		d.buses = append(d.buses, name)
		aliases := []string{fmt.Sprintf("I2C%d", bus)}
		if err := i2creg.Register(name, aliases, bus, openerI2C(bus).Open); err != nil {
			return true, err
		}
	}
	return true, nil
}

type openerI2C int

func (o openerI2C) Open() (i2c.BusCloser, error) {
	b, err := NewI2C(int(o))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func init() {
	if isLinux {
		periph.MustRegister(&drvI2C)
	}
}

var drvI2C driverI2C

var _ i2c.Bus = &I2C{}
var _ i2c.BusCloser = &I2C{}
var _ fmt.Stringer = &I2C{}
