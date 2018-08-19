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
	"sync"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/host/fs"
)

// LEDs is all the leds discovered on this host via sysfs.
//
// Depending on the user context, the LEDs may be read-only or writeable.
var LEDs []*LED

// LEDByName returns a *LED for the LED name, if any.
//
// For all practical purpose, a LED is considered an output-only gpio.PinOut.
func LEDByName(name string) (*LED, error) {
	// TODO(maruel): Use a bisect or a map. For now we don't expect more than a
	// handful of LEDs so it doesn't matter.
	for _, led := range LEDs {
		if led.name == name {
			if err := led.open(); err != nil {
				return nil, err
			}
			return led, nil
		}
	}
	return nil, errors.New("sysfs-led: invalid LED name")
}

// LED represents one LED on the system.
type LED struct {
	number int
	name   string
	root   string

	mu          sync.Mutex
	fBrightness *fs.File // handle to /sys/class/gpio/gpio*/direction; never closed
}

// String implements conn.Resource.
func (l *LED) String() string {
	return fmt.Sprintf("%s(%d)", l.name, l.number)
}

// Halt implements conn.Resource.
//
// It turns the light off.
func (l *LED) Halt() error {
	return l.Out(gpio.Low)
}

// Name implements pin.Pin.
func (l *LED) Name() string {
	return l.name
}

// Number implements pin.Pin.
func (l *LED) Number() int {
	return l.number
}

// Function implements pin.Pin.
func (l *LED) Function() string {
	if l.Read() {
		return "LED/On"
	}
	return "LED/Off"
}

// Func implements pin.PinFunc.
func (l *LED) Func() pin.Func {
	if l.Read() {
		return gpio.OUT_HIGH
	}
	return gpio.OUT_LOW
}

// SupportedFuncs implements pin.PinFunc.
func (l *LED) SupportedFuncs() []pin.Func {
	return []pin.Func{gpio.OUT}
}

// SetFunc implements pin.PinFunc.
func (l *LED) SetFunc(f pin.Func) error {
	switch f {
	case gpio.OUT_HIGH:
		return l.Out(gpio.High)
	case gpio.OUT, gpio.OUT_LOW:
		return l.Out(gpio.Low)
	default:
		return errors.New("sysfs-led: unsupported function")
	}
}

// In implements gpio.PinIn.
func (l *LED) In(pull gpio.Pull, edge gpio.Edge) error {
	if pull != gpio.Float && pull != gpio.PullNoChange {
		return errors.New("sysfs-led: pull is not supported on LED")
	}
	if edge != gpio.NoEdge {
		return errors.New("sysfs-led: edge is not supported on LED")
	}
	return nil
}

// Read implements gpio.PinIn.
func (l *LED) Read() gpio.Level {
	err := l.open()
	if err != nil {
		return gpio.Low
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.fBrightness.Seek(0, 0); err != nil {
		return gpio.Low
	}
	var b [4]byte
	if _, err := l.fBrightness.Read(b[:]); err != nil {
		return gpio.Low
	}
	if b[0] != '0' {
		return gpio.High
	}
	return gpio.Low
}

// WaitForEdge implements gpio.PinIn.
func (l *LED) WaitForEdge(timeout time.Duration) bool {
	return false
}

// Pull implements gpio.PinIn.
func (l *LED) Pull() gpio.Pull {
	return gpio.PullNoChange
}

// DefaultPull implements gpio.PinIn.
func (l *LED) DefaultPull() gpio.Pull {
	return gpio.PullNoChange
}

// Out implements gpio.PinOut.
func (l *LED) Out(level gpio.Level) error {
	err := l.open()
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err = l.fBrightness.Seek(0, 0); err != nil {
		return err
	}
	if level {
		_, err = l.fBrightness.Write([]byte("255"))
	} else {
		_, err = l.fBrightness.Write([]byte("0"))
	}
	return err
}

// PWM implements gpio.PinOut.
//
// This sets the intensity level, if supported. The frequency is ignored.
func (l *LED) PWM(d gpio.Duty, f physic.Frequency) error {
	err := l.open()
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err = l.fBrightness.Seek(0, 0); err != nil {
		return err
	}
	v := (d + gpio.DutyMax/512) / (gpio.DutyMax / 256)
	_, err = l.fBrightness.Write([]byte(strconv.Itoa(int(v))))
	return err
}

//

func (l *LED) open() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	// trigger, max_brightness.
	var err error
	if l.fBrightness == nil {
		p := l.root + "brightness"
		if l.fBrightness, err = fs.Open(p, os.O_RDWR); err != nil {
			// Retry with read-only. This is the default setting.
			l.fBrightness, err = fs.Open(p, os.O_RDONLY)
		}
	}
	return err
}

// driverLED implements periph.Driver.
type driverLED struct {
}

func (d *driverLED) String() string {
	return "sysfs-led"
}

func (d *driverLED) Prerequisites() []string {
	return nil
}

func (d *driverLED) After() []string {
	return nil
}

// Init initializes LEDs sysfs handling code.
//
// Uses led sysfs as described* at
// https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-class-led
//
// * for the most minimalistic meaning of 'described'.
func (d *driverLED) Init() (bool, error) {
	items, err := filepath.Glob("/sys/class/leds/*")
	if err != nil {
		return true, err
	}
	if len(items) == 0 {
		return false, errors.New("no LED found")
	}
	// This make the LEDs in deterministic order.
	sort.Strings(items)
	for i, item := range items {
		LEDs = append(LEDs, &LED{
			number: i,
			name:   filepath.Base(item),
			root:   item + "/",
		})
	}
	return true, nil
}

func init() {
	if isLinux {
		periph.MustRegister(&drvLED)
	}
}

var drvLED driverLED

var _ conn.Resource = &LED{}
var _ gpio.PinIn = &LED{}
var _ gpio.PinOut = &LED{}
var _ gpio.PinIO = &LED{}
var _ pin.PinFunc = &LED{}
