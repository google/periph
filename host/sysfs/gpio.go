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
	"strconv"
	"sync"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host/fs"
)

// Pins is all the pins exported by GPIO sysfs.
//
// Some CPU architectures have the pin numbers start at 0 and use consecutive
// pin numbers but this is not the case for all CPU architectures, some
// have gaps in the pin numbering.
//
// This global variable is initialized once at driver initialization and isn't
// mutated afterward. Do not modify it.
var Pins map[int]*Pin

// Pin represents one GPIO pin as found by sysfs.
type Pin struct {
	number int
	name   string
	root   string // Something like /sys/class/gpio/gpio%d/

	mu         sync.Mutex
	err        error     // If open() failed
	direction  direction // Cache of the last known direction
	edge       gpio.Edge // Cache of the last edge used.
	fDirection fileIO    // handle to /sys/class/gpio/gpio*/direction; never closed
	fEdge      fileIO    // handle to /sys/class/gpio/gpio*/edge; never closed
	fValue     fileIO    // handle to /sys/class/gpio/gpio*/value; never closed
	event      fs.Event  // Initialized once
	buf        [4]byte   // scratch buffer for Function(), Read() and Out()
}

func (p *Pin) String() string {
	return p.name
}

// Name implements pins.Pin.
func (p *Pin) Name() string {
	return p.name
}

// Number implements pins.Pin.
func (p *Pin) Number() int {
	return p.number
}

// Function implements pins.Pin.
func (p *Pin) Function() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	// TODO(maruel): There's an internal bug which causes p.direction to be
	// invalid (!?) Need to figure it out ASAP.
	if err := p.open(); err != nil {
		return "ERR"
	}
	if _, err := seekRead(p.fDirection, p.buf[:]); err != nil {
		return "ERR"
	}
	if p.buf[0] == 'i' && p.buf[1] == 'n' {
		p.direction = dIn
	} else if p.buf[0] == 'o' && p.buf[1] == 'u' && p.buf[2] == 't' {
		p.direction = dOut
	}
	if p.direction == dIn {
		return "In/" + p.Read().String()
	} else if p.direction == dOut {
		return "Out/" + p.Read().String()
	}
	return "ERR"
}

// Halt implements conn.Resource.
//
// It stops edge detection if enabled.
func (p *Pin) Halt() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.haltEdge()
}

// In setups a pin as an input.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	if pull != gpio.PullNoChange && pull != gpio.Float {
		return p.wrap(errors.New("doesn't support pull-up/pull-down"))
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.direction != dIn {
		if err := p.open(); err != nil {
			return p.wrap(err)
		}
		if err := seekWrite(p.fDirection, bIn); err != nil {
			return p.wrap(err)
		}
		p.direction = dIn
	}
	// Always push none to help accumulated flush edges. This is not fool proof
	// but it seems to help.
	if p.fEdge != nil {
		if err := seekWrite(p.fEdge, bNone); err != nil {
			return p.wrap(err)
		}
	}
	// Assume that when the pin was switched, the driver doesn't recall if edge
	// triggering was enabled.
	if edge != gpio.NoEdge {
		if p.fEdge == nil {
			var err error
			p.fEdge, err = fileIOOpen(p.root+"edge", os.O_RDWR)
			if err != nil {
				return p.wrap(err)
			}
			if err = p.event.MakeEvent(p.fValue.Fd()); err != nil {
				_ = p.fEdge.Close()
				p.fEdge = nil
				return p.wrap(err)
			}
		}
		if p.edge != edge {
			var b []byte
			switch edge {
			case gpio.RisingEdge:
				b = bRising
			case gpio.FallingEdge:
				b = bFalling
			case gpio.BothEdges:
				b = bBoth
			}
			if err := seekWrite(p.fEdge, b); err != nil {
				return p.wrap(err)
			}
		}
	}
	p.edge = edge
	// This helps to remove accumulated edges but this is not 100% sufficient.
	// Most of the time the interrupts are handled promptly enough that this loop
	// flushes the accumulated interrupt.
	// Sometimes the kernel may have accumulated interrupts that haven't been
	// processed for a long time, it can easily be >300µs even on a quite idle
	// CPU. In this case, the loop below is not sufficient, since the interrupt
	// will happen afterward "out of the blue".
	if edge != gpio.NoEdge {
		p.WaitForEdge(0)
	}
	return nil
}

func (p *Pin) Read() gpio.Level {
	// There's no lock here.
	if p.fValue == nil {
		return gpio.Low
	}
	if _, err := seekRead(p.fValue, p.buf[:]); err != nil {
		// Error.
		return gpio.Low
	}
	if p.buf[0] == '0' {
		return gpio.Low
	}
	if p.buf[0] == '1' {
		return gpio.High
	}
	// Error.
	return gpio.Low
}

// WaitForEdge does edge detection, returns once one is detected and implements
// gpio.PinIn.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	// Run lockless, as the normal use is to call in a busy loop.
	var ms int
	if timeout == -1 {
		ms = -1
	} else {
		ms = int(timeout / time.Millisecond)
	}
	start := time.Now()
	for {
		if nr, err := p.event.Wait(ms); err != nil {
			return false
		} else if nr == 1 {
			// TODO(maruel): According to pigpio, the correct way to consume the
			// interrupt is to call Seek().
			return true
		}
		// A signal occurred.
		if timeout != -1 {
			ms = int((timeout - time.Since(start)) / time.Millisecond)
		}
		if ms <= 0 {
			return false
		}
	}
}

// Pull returns gpio.PullNoChange since gpio sysfs has no support for input
// pull resistor.
func (p *Pin) Pull() gpio.Pull {
	return gpio.PullNoChange
}

// DefaultPull returns gpio.PullNoChange since gpio sysfs has no support for
// input pull resistor.
func (p *Pin) DefaultPull() gpio.Pull {
	return gpio.PullNoChange
}

// Out sets a pin as output; implements gpio.PinOut.
func (p *Pin) Out(l gpio.Level) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.direction != dOut {
		if err := p.open(); err != nil {
			return p.wrap(err)
		}
		if err := p.haltEdge(); err != nil {
			return err
		}
		// "To ensure glitch free operation, values "low" and "high" may be written
		// to configure the GPIO as an output with that initial value."
		var d []byte
		if l == gpio.Low {
			d = bLow
		} else {
			d = bHigh
		}
		if err := seekWrite(p.fDirection, d); err != nil {
			return p.wrap(err)
		}
		p.direction = dOut
		return nil
	}
	if l == gpio.Low {
		p.buf[0] = '0'
	} else {
		p.buf[0] = '1'
	}
	if err := seekWrite(p.fValue, p.buf[:1]); err != nil {
		return p.wrap(err)
	}
	return nil
}

// PWM implements gpio.PinOut but cannot work on sysfs.
func (p *Pin) PWM(gpio.Duty, physic.Frequency) error {
	return p.wrap(errors.New("pwm is not supported via sysfs"))
}

//

// open opens the gpio sysfs handle to /value and /direction.
//
// lock must be held.
func (p *Pin) open() error {
	if p.fDirection != nil || p.err != nil {
		return p.err
	}

	if drvGPIO.exportHandle == nil {
		return errors.New("sysfs gpio is not initialized")
	}
	var err error
	_, err = drvGPIO.exportHandle.Write([]byte(strconv.Itoa(p.number)))
	if err != nil && !isErrBusy(err) {
		p.err = err
		if os.IsPermission(p.err) {
			return fmt.Errorf("need more access, try as root or setup udev rules: %v", p.err)
		}
		return p.err
	}

	// There's a race condition where the file may be created but udev is still
	// running the Raspbian udev rule to make it readable to the current user.
	// It's simpler to just loop a little as if /export is accessible, it doesn't
	// make sense that gpioN/value doesn't become accessible eventually.
	timeout := 5 * time.Second
	for start := time.Now(); time.Since(start) < timeout; {
		p.fValue, err = fileIOOpen(p.root+"value", os.O_RDWR)
		// The virtual file creation is synchronous when writing to /export for
		// udev rule execution is asynchronous.
		if err == nil {
			break
		}
		if !os.IsPermission(err) {
			p.err = err
			break
		}
	}
	if p.err != nil {
		return p.err
	}
	p.fDirection, err = fileIOOpen(p.root+"direction", os.O_RDWR)
	if err != nil {
		p.err = err
		_ = p.fValue.Close()
		p.fValue = nil
	}
	return p.err
}

// haltEdge stops any on-going edge detection.
func (p *Pin) haltEdge() error {
	if p.edge != gpio.NoEdge {
		if err := seekWrite(p.fEdge, bNone); err != nil {
			return p.wrap(err)
		}
		p.edge = gpio.NoEdge
		// This is still important to remove an accumulated edge.
		p.WaitForEdge(0)
	}
	return nil
}

func (p *Pin) wrap(err error) error {
	return fmt.Errorf("sysfs-gpio (%s): %v", p, err)
}

//

type direction int

const (
	dUnknown direction = 0
	dIn      direction = 1
	dOut     direction = 2
)

var (
	bIn      = []byte("in")
	bLow     = []byte("low")
	bHigh    = []byte("high")
	bNone    = []byte("none")
	bRising  = []byte("rising")
	bFalling = []byte("falling")
	bBoth    = []byte("both")
)

// readInt reads a pseudo-file (sysfs) that is known to contain an integer and
// returns the parsed number.
func readInt(path string) (int, error) {
	f, err := fileIOOpen(path, os.O_RDONLY)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var b [24]byte
	n, err := f.Read(b[:])
	if err != nil {
		return 0, err
	}
	raw := b[:n]
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return 0, errors.New("invalid value")
	}
	return strconv.Atoi(string(raw[:len(raw)-1]))
}

// driverGPIO implements periph.Driver.
type driverGPIO struct {
	exportHandle io.Writer // handle to /sys/class/gpio/export
}

func (d *driverGPIO) String() string {
	return "sysfs-gpio"
}

func (d *driverGPIO) Prerequisites() []string {
	return nil
}

func (d *driverGPIO) After() []string {
	return nil
}

// Init initializes GPIO sysfs handling code.
//
// Uses gpio sysfs as described at
// https://www.kernel.org/doc/Documentation/gpio/sysfs.txt
//
// GPIO sysfs is often the only way to do edge triggered interrupts. Doing this
// requires cooperation from a driver in the kernel.
//
// The main drawback of GPIO sysfs is that it doesn't expose internal pull
// resistor and it is much slower than using memory mapped hardware registers.
func (d *driverGPIO) Init() (bool, error) {
	items, err := filepath.Glob("/sys/class/gpio/gpiochip*")
	if err != nil {
		return true, err
	}
	if len(items) == 0 {
		return false, errors.New("no GPIO pin found")
	}

	// There are hosts that use non-continuous pin numbering so use a map instead
	// of an array.
	Pins = map[int]*Pin{}
	for _, item := range items {
		if err := d.parseGPIOChip(item + "/"); err != nil {
			return true, err
		}
	}
	drvGPIO.exportHandle, err = fileIOOpen("/sys/class/gpio/export", os.O_WRONLY)
	if os.IsPermission(err) {
		return true, fmt.Errorf("need more access, try as root or setup udev rules: %v", err)
	}
	return true, err
}

func (d *driverGPIO) parseGPIOChip(path string) error {
	base, err := readInt(path + "base")
	if err != nil {
		return err
	}
	number, err := readInt(path + "ngpio")
	if err != nil {
		return err
	}
	// TODO(maruel): The chip driver may lie and lists GPIO pins that cannot be
	// exported. The only way to know about it is to export it before opening.
	for i := base; i < base+number; i++ {
		if _, ok := Pins[i]; ok {
			return fmt.Errorf("found two pins with number %d", i)
		}
		p := &Pin{
			number: i,
			name:   fmt.Sprintf("GPIO%d", i),
			root:   fmt.Sprintf("/sys/class/gpio/gpio%d/", i),
		}
		Pins[i] = p
		if err := gpioreg.Register(p); err != nil {
			return err
		}
		// If there is a CPU memory mapped gpio pin with the same number, the
		// driver has to unregister this pin and map its own after.
		if err := gpioreg.RegisterAlias(strconv.Itoa(i), p.name); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	if isLinux {
		periph.MustRegister(&drvGPIO)
	}
}

var drvGPIO driverGPIO

var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ gpio.PinIO = &Pin{}
var _ fmt.Stringer = &Pin{}
