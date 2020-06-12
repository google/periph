// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
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
	number         int
	name           string
	root           string         // Something like /sys/class/gpio/gpio%d/
	edgeChan       chan time.Time // Used for edge detection
	cancelWaitChan chan struct{}  // Used to unblock WaitForEdge
	wg             sync.WaitGroup

	mu         sync.Mutex
	err        error     // If open() failed
	direction  direction // Cache of the last known direction
	edge       gpio.Edge // Cache of the last edge used.
	fDirection fileIO    // handle to /sys/class/gpio/gpio*/direction; never closed
	fEdge      fileIO    // handle to /sys/class/gpio/gpio*/edge; never closed
	fValue     fileIO    // handle to /sys/class/gpio/gpio*/value; never closed
	buf        [4]byte   // scratch buffer for Func(), Read() and Out()

	muCancel         sync.Mutex // Used in haltEdge() and WaitForEdge()
	cancelListenEdge func()     // Cancels a pending ListenEdges()
	until            time.Time  // Edges before this timestamp are ignored
}

// String implements conn.Resource.
func (p *Pin) String() string {
	return p.name
}

// Halt implements conn.Resource.
//
// It stops edge detection if enabled.
func (p *Pin) Halt() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.haltEdge()
}

// Name implements pin.Pin.
func (p *Pin) Name() string {
	return p.name
}

// Number implements pin.Pin.
func (p *Pin) Number() int {
	return p.number
}

// Function implements pin.Pin.
func (p *Pin) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *Pin) Func() pin.Func {
	p.mu.Lock()
	defer p.mu.Unlock()
	// TODO(maruel): There's an internal bug which causes p.direction to be
	// invalid (!?) Need to figure it out ASAP.
	if err := p.open(); err != nil {
		return pin.FuncNone
	}
	if _, err := seekRead(p.fDirection, p.buf[:]); err != nil {
		return pin.FuncNone
	}
	if p.buf[0] == 'i' && p.buf[1] == 'n' {
		p.direction = dIn
	} else if p.buf[0] == 'o' && p.buf[1] == 'u' && p.buf[2] == 't' {
		p.direction = dOut
	}
	if p.direction == dIn {
		if p.Read() {
			return gpio.IN_HIGH
		}
		return gpio.IN_LOW
	} else if p.direction == dOut {
		if p.Read() {
			return gpio.OUT_HIGH
		}
		return gpio.OUT_LOW
	}
	return pin.FuncNone
}

// SupportedFuncs implements pin.PinFunc.
func (p *Pin) SupportedFuncs() []pin.Func {
	return []pin.Func{gpio.IN, gpio.OUT}
}

// SetFunc implements pin.PinFunc.
func (p *Pin) SetFunc(f pin.Func) error {
	switch f {
	case gpio.IN:
		return p.In(gpio.PullNoChange, gpio.NoEdge)
	case gpio.OUT_HIGH:
		return p.Out(gpio.High)
	case gpio.OUT, gpio.OUT_LOW:
		return p.Out(gpio.Low)
	default:
		return p.wrap(errors.New("unsupported function"))
	}
}

// In implements gpio.PinIn.
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
	// Assume that when the pin was switched, the driver doesn't recall if edge
	// triggering was enabled.
	if edge != gpio.NoEdge {
		if p.fEdge == nil {
			var err error
			p.fEdge, err = fileIOOpen(p.root+"edge", os.O_RDWR)
			if err != nil {
				return p.wrap(err)
			}
		}

		// Cancel any previous event loop.
		p.muCancel.Lock()
		p.cancelWait()
		// Create a new context for the new event loop.
		ctx, cancel := context.WithCancel(context.Background())
		p.cancelListenEdge = cancel

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
			p.edge = edge
		}

		fd := p.fValue.Fd()
		p.wg.Add(1)
		go func() {
			p.until = time.Now()
			p.muCancel.Unlock()
			events.listen(ctx, fd, p.edgeChan)
			p.wg.Done()
		}()
		return nil
	}
	return p.haltEdge()
}

// Read implements gpio.PinIn.
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

// WaitForEdge implements gpio.PinIn.
func (p *Pin) WaitForEdge(timeout time.Duration) bool {
	// Run "mostly" lockless.
	p.muCancel.Lock()
	until := p.until
	p.muCancel.Unlock()

	// Edge detection is not enabled.
	if until.IsZero() {
		return false
	}

	if timeout == -1 {
		// No timeout.
		for {
			select {
			case t := <-p.edgeChan:
				if until.Before(t) {
					return true
				}
				// Else loop again.
			case <-p.cancelWaitChan:
				// Edge waiting was canceled.
				return false
			}
		}
	}

	if timeout == 0 {
		// Do a quick peek, ignoring the rest. Empties p.edgeChan if needed.
		for {
			select {
			case t := <-p.edgeChan:
				if until.Before(t) {
					return true
				}
			default:
				return false
			}
		}
	}

	// Normal timeout.
	c := time.After(timeout)
	for {
		select {
		case t := <-p.edgeChan:
			if until.Before(t) {
				return true
			}
		case <-c:
			return false
		case <-p.cancelWaitChan:
			// Edge waiting was canceled.
			return false
		}
	}
}

// Pull implements gpio.PinIn.
//
// It returns gpio.PullNoChange since gpio sysfs has no support for input pull
// resistor.
func (p *Pin) Pull() gpio.Pull {
	return gpio.PullNoChange
}

// DefaultPull implements gpio.PinIn.
//
// It returns gpio.PullNoChange since gpio sysfs has no support for input pull
// resistor.
func (p *Pin) DefaultPull() gpio.Pull {
	return gpio.PullNoChange
}

// Out implements gpio.PinOut.
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

// PWM implements gpio.PinOut.
//
// This is not supported on sysfs.
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

	// Try to open the pin if it was there. It's possible it had been exported
	// already.
	if p.fValue, p.err = fileIOOpen(p.root+"value", os.O_RDWR); p.err == nil {
		// Fast track.
		goto direction
	} else if !os.IsNotExist(p.err) {
		// It exists but not accessible, not worth doing the remainder.
		p.err = fmt.Errorf("need more access, try as root or setup udev rules: %v", p.err)
		return p.err
	}

	if _, p.err = drvGPIO.exportHandle.Write([]byte(strconv.Itoa(p.number))); p.err != nil && !isErrBusy(p.err) {
		if os.IsPermission(p.err) {
			p.err = fmt.Errorf("need more access, try as root or setup udev rules: %v", p.err)
		}
		return p.err
	}

	// There's a race condition where the file may be created but udev is still
	// running the Raspbian udev rule to make it readable to the current user.
	// It's simpler to just loop a little as if /export is accessible, it doesn't
	// make sense that gpioN/value doesn't become accessible eventually.
	for start := time.Now(); time.Since(start) < 5*time.Second; {
		// The virtual file creation is synchronous when writing to /export; albeit
		// udev rule execution is asynchronous, so file mode change via udev rules
		// takes some time to propagate.
		if p.fValue, p.err = fileIOOpen(p.root+"value", os.O_RDWR); p.err == nil || !os.IsPermission(p.err) {
			// Either success or a failure that is not a permission error.
			break
		}
	}
	if p.err != nil {
		return p.err
	}

direction:
	if p.fDirection, p.err = fileIOOpen(p.root+"direction", os.O_RDWR); p.err != nil {
		_ = p.fValue.Close()
		p.fValue = nil
	}
	return p.err
}

// haltEdge disables edge detection and unblocks any pending WaitforEdge().
//
// Must be called with mu held.
func (p *Pin) haltEdge() error {
	if p.edge != gpio.NoEdge {
		if err := seekWrite(p.fEdge, bNone); err != nil {
			return p.wrap(err)
		}
		p.edge = gpio.NoEdge
		p.muCancel.Lock()
		p.cancelWait()
		p.muCancel.Unlock()
	}
	return nil
}

// cancelWait unblocks any pending WaitForEdge(), and stop edge listening.
//
// Must be called with p.muCancel held.
func (p *Pin) cancelWait() {
	p.cancelListenEdge()
	for {
		select {
		case p.cancelWaitChan <- struct{}{}:
		case <-p.edgeChan:
		default:
			goto end
		}
	}
end:
	p.wg.Wait()
	p.cancelListenEdge = func() {}
	p.until = time.Time{}
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
			number:           i,
			name:             fmt.Sprintf("GPIO%d", i),
			root:             fmt.Sprintf("/sys/class/gpio/gpio%d/", i),
			edgeChan:         make(chan time.Time),
			cancelListenEdge: func() {},
			cancelWaitChan:   make(chan struct{}),
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

var _ conn.Resource = &Pin{}
var _ gpio.PinIn = &Pin{}
var _ gpio.PinOut = &Pin{}
var _ gpio.PinIO = &Pin{}
var _ pin.PinFunc = &Pin{}
