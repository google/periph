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
	"strconv"
	"sync"
	"time"

	"github.com/google/periph"
	"github.com/google/periph/conn/gpio"
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
	direction  direction // Cache of the last known direction
	edge       gpio.Edge //
	fDirection *os.File  // handle to /sys/class/gpio/gpio*/direction; never closed
	fEdge      *os.File  // handle to /sys/class/gpio/gpio*/edge; never closed
	fValue     *os.File  // handle to /sys/class/gpio/gpio*/value; never closed
	epollFd    int       // Never closed
	event      event     // Initialized once
}

func (p *Pin) Name() string {
	return p.name
}

func (p *Pin) String() string {
	return fmt.Sprintf("%s(%d)", p.name, p.number)
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
		return err.Error()
	}
	var buf [4]byte
	if err := seekRead(p.fDirection, buf[:]); err != nil {
		return err.Error()
	}
	if buf[0] == 'i' && buf[1] == 'n' {
		p.direction = dIn
	} else if buf[0] == 'o' && buf[1] == 'u' && buf[2] == 't' {
		p.direction = dOut
	}
	if p.direction == dIn {
		return "In/" + p.Read().String()
	} else if p.direction == dOut {
		return "Out/" + p.Read().String()
	}
	return "N/A"
}

// In setups a pin as an input.
func (p *Pin) In(pull gpio.Pull, edge gpio.Edge) error {
	if pull != gpio.PullNoChange && pull != gpio.Float {
		return errors.New("sysfs-gpio: pull is not supported")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	changed := false
	if p.direction != dIn {
		if err := p.open(); err != nil {
			return err
		}
		if err := seekWrite(p.fDirection, bIn); err != nil {
			return err
		}
		p.direction = dIn
		changed = true
	}
	// Assume that when the pin was switched, the driver doesn't recall if edge
	// triggering was enabled.
	if changed || edge != p.edge {
		if edge != gpio.None {
			var err error
			if p.fEdge == nil {
				p.fEdge, err = os.OpenFile(p.root+"edge", os.O_RDWR|os.O_APPEND, 0600)
				if err != nil {
					return err
				}
			}
			if p.epollFd == 0 {
				if p.epollFd, err = p.event.makeEvent(p.fValue); err != nil {
					return err
				}
			}
			var b []byte
			switch edge {
			case gpio.Rising:
				b = bRising
			case gpio.Falling:
				b = bFalling
			case gpio.Both:
				b = bBoth
			}
			if err := seekWrite(p.fEdge, b); err != nil {
				return err
			}
		} else {
			if p.fEdge != nil {
				if err := seekWrite(p.fEdge, bNone); err != nil {
					return err
				}
			}
		}
		p.edge = edge
	}
	if p.edge != gpio.None {
		// Flush accumulated events if any.
		// BUG(maruel): Wake up any WaitForEdge() waiting for an edge.
		for {
			// Only loop if nr == -1, which means that a signal was received.
			nr, err := p.event.wait(p.epollFd, 0)
			if nr == 0 || err != nil {
				break
			}
		}
	}
	return nil
}

func (p *Pin) Read() gpio.Level {
	// There's no lock here.
	var buf [4]byte
	if err := seekRead(p.fValue, buf[:]); err != nil {
		// Error.
		//fmt.Printf("%s: %v", p, err)
		return gpio.Low
	}
	if buf[0] == '0' {
		return gpio.Low
	}
	if buf[0] == '1' {
		return gpio.High
	}
	// Error.
	return gpio.Low
}

// WaitForEdge does edge detection, returns once one is detected and implements
// gpio.PinIn.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	// Run lockless, as the normal use is to call in a busy loop.
	if p.epollFd == 0 {
		return false
	}
	var ms int
	if timeout == -1 {
		ms = -1
	} else {
		ms = int(timeout / time.Millisecond)
	}
	start := time.Now()
	for {
		nr, err := p.event.wait(p.epollFd, ms)
		if err != nil {
			return false
		}
		if nr == 1 {
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

// Out sets a pin as output; implements gpio.PinOut.
func (p *Pin) Out(l gpio.Level) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.direction != dOut {
		if err := p.open(); err != nil {
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
			return err
		}
		p.direction = dOut
		return nil
	}
	var d [1]byte
	if l == gpio.Low {
		d[0] = '0'
	} else {
		d[0] = '1'
	}
	return seekWrite(p.fValue, d[:])
}

// PWM implements gpio.PinOut.
func (p *Pin) PWM(duty int) error {
	return errors.New("sysfs-gpio: pwm is not supported")
}

//

// open opens the gpio sysfs handle to /value and /direction.
//
// lock must be held.
func (p *Pin) open() error {
	if exportHandle == nil {
		return errors.New("sysfs gpio is not initialized")
	}
	if p.fDirection != nil {
		return nil
	}
	_, err := exportHandle.Write([]byte(strconv.Itoa(p.number)))
	if err != nil && !isErrBusy(err) {
		if os.IsPermission(err) {
			return fmt.Errorf("need more access, try as root or setup udev rules: %v", err)
		}
		return err
	}
	// There's a race condition where the file may be created but udev is still
	// running the Raspbian udev rule to make it readable to the current user.
	// It's simpler to just loop a little as if /export is accessible, it doesn't
	// make sense that gpioN/value doesn't become accessible eventually.
	timeout := 5 * time.Second
	for start := time.Now(); time.Since(start) < timeout; {
		p.fValue, err = os.OpenFile(p.root+"value", os.O_RDWR, 0600)
		// The virtual file creation is synchronous when writing to /export for
		// udev rule execution is asynchronous.
		if err == nil || !os.IsPermission(err) {
			break
		}
	}
	p.fDirection, err = os.OpenFile(p.root+"direction", os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		p.fValue.Close()
		p.fValue = nil
	}
	return err
}

//

var exportHandle io.Writer // handle to /sys/class/gpio/export

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

func readInt(path string) (int, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return 0, errors.New("invalid value")
	}
	return strconv.Atoi(string(raw[:len(raw)-1]))
}

func seekRead(f *os.File, b []byte) error {
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	_, err := f.Read(b)
	return err
}

func seekWrite(f *os.File, b []byte) error {
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	_, err := f.Write(b)
	return err
}

// driverGPIO implements periph.Driver.
type driverGPIO struct {
}

func (d *driverGPIO) String() string {
	return "sysfs-gpio"
}

func (d *driverGPIO) Type() periph.Type {
	// It intentionally load later than processors.
	return periph.Pins
}

func (d *driverGPIO) Prerequisites() []string {
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
	exportHandle, err = os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0600)
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
		// Try to register real pin, but it may already be registered by the processor
		// driver. In that case register an alias instead.
		if gpio.Register(p) != nil {
			realPin := gpio.ByNumber(i)
			alias := &gpio.PinAlias{N: p.name, PinIO: realPin}
			gpio.RegisterAlias(alias)
		}
		// We cannot use gpio.MapFunction() since there is no API to determine this.
	}
	return nil
}

func init() {
	if isLinux {
		periph.MustRegister(&driverGPIO{})
	}
}

var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ gpio.PinIO = &Pin{}
