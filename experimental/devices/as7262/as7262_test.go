// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package as7262

import (
	"reflect"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
)

func TestDev_Sense(t *testing.T) {

	type timefunc func(*Dev) func(time.Duration) <-chan time.Time

	defer func() {
		waitForSensor = time.After
	}()

	haltit := func(dev *Dev) func(time.Duration) <-chan time.Time {
		return func(d time.Duration) <-chan time.Time {
			t := make(chan time.Time, 1)
			dev.Halt()
			return t
		}
	}
	dontwait := func(dev *Dev) func(time.Duration) <-chan time.Time {
		return func(d time.Duration) <-chan time.Time {
			t := make(chan time.Time, 1)
			t <- time.Now()
			dev.timeout = time.Millisecond * 1
			return t
		}
	}

	intPin := &gpiotest.Pin{N: "GPIO1", Num: 1, Fn: "NotRealPin", EdgesChan: make(chan gpio.Level, 1)}

	tests := []struct {
		name     string
		tx       []i2ctest.IO
		opts     Opts
		waiter   timefunc
		want     Spectrum
		sendEdge bool
		wantErr  error
	}{
		{
			name:    "validRead",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: nil,
			want:    validSpectrum,
			tx:      sensorTestCaseValidRead,
		},
		{
			name:    "errHalted",
			opts:    DefaultOpts,
			waiter:  haltit,
			wantErr: errHalted,
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead,
		},
		{
			name: "interuptValid",
			opts: Opts{
				InterruptPin: intPin,
			},
			waiter:   dontwait,
			sendEdge: true,
			wantErr:  nil,
			want:     validSpectrum,
			tx:       sensorTestCaseInteruptValidRead,
		},
		{
			name: "interuptTimeout",
			opts: Opts{
				InterruptPin: intPin,
			},
			waiter:   dontwait,
			sendEdge: false,
			wantErr:  errPinTimeout,
			want:     Spectrum{},
			tx:       sensorTestCaseInteruptValidRead,
		},
		{
			name: "haltBeforeforEdge",
			opts: Opts{
				InterruptPin: intPin,
			},
			waiter:   dontwait,
			sendEdge: false,
			wantErr:  errHalted,
			want:     Spectrum{},
			tx:       sensorTestCaseInteruptValidRead,
		},
		{
			name: "ioErrorWritingIntergrationReg",
			opts: Opts{
				InterruptPin: intPin,
			},
			waiter:   dontwait,
			sendEdge: false,
			wantErr:  &IOError{"reading status register", nil},
			want:     Spectrum{},
			tx:       nil,
		},
		{
			name:    "ioErrorWritingLedControlReg",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: &IOError{"reading status register", nil},
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead[:4],
		},
		{
			name:    "ioErrorWritingControlReg",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: &IOError{"reading status register", nil},
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead[:8],
		},
		{
			name:    "ioErrorPollDataReady",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: &IOError{"reading status register", nil},
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead[:12],
		},
		{
			name:    "ioErrorWritingLedControlReg2",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: &IOError{"reading status register", nil},
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead[:16],
		},
		{
			name:    "ioErrorReadingRawBase",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: &IOError{"reading status register", nil},
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead[:32],
		},
		{
			name:    "ioErrorReadingCalBase",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: &IOError{"reading status register", nil},
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead[:82],
		},
		{
			name:    "ioErrorReadingTemperature",
			opts:    DefaultOpts,
			waiter:  dontwait,
			wantErr: &IOError{"reading status register", nil},
			want:    Spectrum{},
			tx:      sensorTestCaseValidRead[:164],
		},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		d, _ := New(bus, &tt.opts)

		waitForSensor = tt.waiter(d)

		if d.interrupt != nil && tt.sendEdge {
			intPin.EdgesChan <- gpio.High
		}
		if tt.name == "haltBeforeforEdge" {
			// Time must be less than senseTime.
			time.AfterFunc(time.Millisecond, func() {
				d.Halt()
			})
		}

		got, err := d.Sense(physic.MilliAmpere*100, time.Millisecond*3)

		if _, ok := tt.wantErr.(*IOError); ok {
			if _, ok := err.(*IOError); !ok {
				t.Errorf("expected IOError but %T", err)
			}
			if err.(*IOError).Op != tt.wantErr.(*IOError).Op {
				t.Errorf("expected %s, but got %s", tt.wantErr.(*IOError).Op, err.(*IOError).Op)
			}
		} else if err != tt.wantErr {
			t.Errorf("expected error: %v but got: %v", tt.wantErr, got)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Dev.Sense() = %v, want %v", got, tt.want)
		}
	}
}

func Test_calcSenseTime(t *testing.T) {
	var tests = []struct {
		name  string
		t     time.Duration
		want1 uint8
		want2 time.Duration
	}{
		{"0", 2800 * time.Microsecond, 1, 2800 * time.Microsecond},
		{"2.8ms", 0, 1, 2800 * time.Microsecond},
		{"3ms", 3 * time.Millisecond, 1, 2800 * time.Microsecond},
		{"500ms", 500 * time.Millisecond, 178, 498400 * time.Microsecond},
		{"1hour", 1 * time.Hour, 255, 714 * time.Millisecond},
		{"-1hour", -1 * time.Hour, 1, 2800 * time.Microsecond},
	}

	for _, test := range tests {
		got1, got2 := calcSenseTime(test.t)
		if got1 != test.want1 {
			t.Errorf("calcSenseTime() expected %v but got %v", test.want1, got1)
		}
		if got2 != test.want2 {
			t.Errorf("calcSenseTime() expected %v but got %v", test.want2, got2)
		}
	}
}

func Test_calcLed(t *testing.T) {
	tests := []struct {
		name  string
		drive physic.ElectricCurrent
		want  uint8
	}{
		{"Off", 0 * physic.Ampere, 0x00},
		{"12.5", 12500 * physic.MicroAmpere, 0x08},
		{"25", 25 * physic.MilliAmpere, 0x18},
		{"50", 50 * physic.MilliAmpere, 0x28},
		{"100", 100 * physic.MilliAmpere, 0x38},
		{"10", 10 * physic.MilliAmpere, 0x00},
		{"20", 20 * physic.MilliAmpere, 0x08},
		{"30", 30 * physic.MilliAmpere, 0x18},
		{"40", 40 * physic.MilliAmpere, 0x18},
		{"60", 60 * physic.MilliAmpere, 0x28},
		{"110", 110 * physic.MilliAmpere, 0x38},
		{"-1", -1 * physic.MilliAmpere, 0x00},
	}
	for _, tt := range tests {
		if got, _ := calcLed(tt.drive); got != tt.want {
			t.Errorf("calcLed() = %v, want %v", got, tt.want)
		}
	}
}

func TestDev_pollStatus(t *testing.T) {
	tests := []struct {
		name    string
		tx      []i2ctest.IO
		dir     direction
		timeout time.Duration
		halt    time.Duration
		wantErr error
	}{
		{
			name: "errStatusIO",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{}},
			},
			dir:     reading,
			timeout: time.Millisecond * 1,
			wantErr: &IOError{"reading status register", nil},
		},
		{
			name: "errStatusDeadline",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
			},
			dir:     reading,
			timeout: time.Millisecond * 1,
			wantErr: errStatusDeadline,
		},
		{
			name: "errHalted",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
			},
			dir:     reading,
			timeout: time.Millisecond * 1000,
			halt:    time.Nanosecond,
			wantErr: errHalted,
		},
		{
			name: "ReadReady",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
			},
			dir:     reading,
			timeout: time.Millisecond * 1000,
			wantErr: nil,
		},
		{
			name: "WriteReady",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
			},
			dir:     writing,
			timeout: time.Millisecond * 1000,
			wantErr: nil,
		},
		{
			name: "CleanBuffer",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
			},
			dir:     clearBuffer,
			timeout: time.Millisecond * 1000,
			wantErr: nil,
		},
		{
			name: "ClearBuffer",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
			},
			dir:     clearBuffer,
			timeout: time.Millisecond * 1000,
			wantErr: nil,
		},
		{
			name: "errClearingBuffer",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{}},
			},
			dir:     clearBuffer,
			timeout: time.Millisecond * 1000,
			wantErr: &IOError{"clearing buffer", nil},
		},
		{
			name: "retry1",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
			},
			dir:     reading,
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
		{
			name: "retrywithHalt",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
			},
			dir:     reading,
			timeout: time.Millisecond * 100,
			halt:    time.Millisecond * 3,
			wantErr: errHalted,
		},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		d := &Dev{
			c:       &i2c.Dev{Bus: bus, Addr: 0x49},
			done:    make(chan struct{}, 1),
			timeout: tt.timeout,
		}
		defer d.Halt()
		if tt.halt > time.Nanosecond {
			go func() {
				time.Sleep(tt.halt)
				d.Halt()
			}()
		} else if tt.halt != 0 {
			d.Halt()
			d.Halt()
		}

		got := d.pollStatus(tt.dir)
		// t.Errorf("expected error: %v but got: %v", tt.wantErr, got)

		if _, ok := tt.wantErr.(*IOError); ok {
			if _, ok := got.(*IOError); !ok {
				t.Errorf("expected IOError but %T", got)
			}
			if got.(*IOError).Op != tt.wantErr.(*IOError).Op {
				t.Errorf("expected %s, but got %s", tt.wantErr.(*IOError).Op, got.(*IOError).Op)
			}
		} else if got != tt.wantErr {
			t.Errorf("expected error: %v but got: %v", tt.wantErr, got)
		}
	}
}

func TestDev_writeVirtualRegister(t *testing.T) {
	tests := []struct {
		name    string
		tx      []i2ctest.IO
		timeout time.Duration
		halt    time.Duration
		wantErr error
	}{
		{
			name: "errHalted",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x02}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x02}},
			},
			timeout: time.Millisecond * 100,
			halt:    time.Nanosecond,
			wantErr: errHalted,
		},
		{
			name: "errSettingVirtualRegister",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"setting virtual register", nil},
		},
		{
			name: "errStatusIO",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x84}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"reading status register", nil},
		},
		{
			name: "errWritingVirtualRegister",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x84}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"writing virtual register", nil},
		},
		{
			name: "writeOk",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x84}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0xFF}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		d := &Dev{
			c:       &i2c.Dev{Bus: bus, Addr: 0x49},
			done:    make(chan struct{}, 1),
			timeout: tt.timeout,
		}
		defer d.Halt()
		if tt.halt > time.Nanosecond {
			go func() {
				time.Sleep(tt.halt)
				d.Halt()
			}()
		} else if tt.halt != 0 {
			d.Halt()
		}

		got := d.writeVirtualRegister(0x04, 0xFF)
		if _, ok := tt.wantErr.(*IOError); ok {
			if _, ok := got.(*IOError); !ok {
				t.Errorf("expected IOError but %T", got)
			}
			if got.(*IOError).Op != tt.wantErr.(*IOError).Op {
				t.Errorf("expected %s, but got %s", tt.wantErr.(*IOError).Op, got.(*IOError).Op)
			}
		} else if got != tt.wantErr {
			t.Errorf("expected error: %v but got: %v", tt.wantErr, got)
		}
	}
}

func TestDev_readVirtualRegister(t *testing.T) {
	tests := []struct {
		name    string
		tx      []i2ctest.IO
		timeout time.Duration
		data    []byte
		halt    time.Duration
		wantErr error
	}{
		{
			name:    "nodata",
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
		{
			name:    "errHalted",
			tx:      []i2ctest.IO{},
			data:    []byte{0x00},
			halt:    time.Nanosecond,
			timeout: time.Millisecond * 100,
			wantErr: errHalted,
		},
		{
			name: "errClearingBuffer",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				// {Addr: 0x49, W: []byte{statusReg}, R: []byte{0x02}},
			},
			data:    []byte{0x00},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"clearing buffer", nil},
		},
		{
			name: "errSettingVirtualRegister",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg}, R: []byte{}},
			},
			data:    []byte{0x00},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"setting virtual register", nil},
		},
		{
			name: "errStatusIO",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x04}, R: []byte{}},
			},
			data:    []byte{0x00},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"reading status register", nil},
		},
		{
			name: "errReadingVirtualRegister",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x04}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{}},
			},
			data:    []byte{0x00},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"reading virtual register", nil},
		},
		{
			name: "readSingleByteOk",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x04}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x00}},
			},
			data:    []byte{0x00},
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
		{
			name: "readTwoBytesOk",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x04}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x05}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x00}},
			},
			data:    []byte{0x00, 0x00},
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		d := &Dev{
			c:       &i2c.Dev{Bus: bus, Addr: 0x49},
			done:    make(chan struct{}, 1),
			timeout: tt.timeout,
		}
		// defer d.Halt()
		if tt.halt > time.Nanosecond {
			go func() {
				time.Sleep(tt.halt)
				d.Halt()
			}()
		} else if tt.halt != 0 {
			d.Halt()
		}

		got := d.readVirtualRegister(0x04, tt.data)
		if _, ok := tt.wantErr.(*IOError); ok {
			if _, ok := got.(*IOError); !ok {
				t.Errorf("expected IOError but %T", got)
			}
			if got.(*IOError).Op != tt.wantErr.(*IOError).Op {
				t.Errorf("expected %s, but got %s", tt.wantErr.(*IOError).Op, got.(*IOError).Op)
			}
		} else if got != tt.wantErr {
			t.Errorf("expected error: %v but got: %v", tt.wantErr, got)
		}
	}
}

func TestDev_pollDataReady(t *testing.T) {
	tests := []struct {
		name    string
		tx      []i2ctest.IO
		timeout time.Duration
		halt    time.Duration
		wantErr error
	}{
		{
			name:    "errHalted",
			tx:      []i2ctest.IO{},
			halt:    time.Nanosecond,
			timeout: time.Millisecond * 100,
			wantErr: errHalted,
		},
		{
			name: "errStatusIO",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"reading status register", nil},
		},
		{
			name: "errSettingVirtualRegister",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"setting virtual register", nil},
		},
		{
			name: "errStatusIO2",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"reading status register", nil},
		},
		{
			name: "errReadingVirtualRegister",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"reading virtual register", nil},
		},
		{
			name: "errStatusDeadline",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x00}},
			},
			timeout: 1 * time.Nanosecond,
			wantErr: errStatusDeadline,
		},
		{
			name: "ok",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x03}},
			},
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
		{
			name: "retryOk",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x03}},
			},
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
		{
			name: "errHalted2",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, controlReg}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x01}},
				{Addr: 0x49, W: []byte{readReg}, R: []byte{0x03}},
			},
			halt:    time.Millisecond * 2,
			timeout: time.Millisecond * 100,
			wantErr: errHalted,
		},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		d := &Dev{
			c:       &i2c.Dev{Bus: bus, Addr: 0x49},
			done:    make(chan struct{}, 1),
			timeout: tt.timeout,
		}
		defer d.Halt()
		if tt.halt > time.Nanosecond {
			go func() {
				time.Sleep(tt.halt)
				d.Halt()
			}()
		} else if tt.halt != 0 {
			d.Halt()
		}

		got := d.pollDataReady()
		if _, ok := tt.wantErr.(*IOError); ok {
			if _, ok := got.(*IOError); !ok {
				t.Errorf("expected IOError but %T", got)
			}
			if got.(*IOError).Op != tt.wantErr.(*IOError).Op {
				t.Errorf("expected %s, but got %s", tt.wantErr.(*IOError).Op, got.(*IOError).Op)
			}
		} else if got != tt.wantErr {
			t.Errorf("expected error: %v but got: %v", tt.wantErr, got)
		}
	}
}

func TestNew(t *testing.T) {

	tests := []struct {
		name    string
		opts    Opts
		want1   Gain
		want2   gpio.PinIn
		wantErr bool
	}{
		{name: "defautls",
			opts:    DefaultOpts,
			want1:   G1x,
			want2:   nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {

		bus := &i2ctest.Playback{DontPanic: true}
		d, err := New(bus, &tt.opts)
		if err != nil != tt.wantErr {
			t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if tt.want1 != d.gain {
			t.Errorf("New() wanted %v but got %v", tt.want1, d.gain)
		}
		if tt.want2 != d.interrupt {
			t.Errorf("New() wanted %v but got %v", tt.want2, d.interrupt)
		}
	}
}

func TestDev_Gain(t *testing.T) {
	tests := []struct {
		name    string
		tx      []i2ctest.IO
		timeout time.Duration
		gain    Gain
		wantErr error
	}{
		{
			name: "errStatusIO",
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{}},
				// {Addr: 0x49, W: []byte{writeReg}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: &IOError{"reading status register", nil},
		},
		{
			name: "ok",
			gain: G16x,
			tx: []i2ctest.IO{
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x84}, R: []byte{}},
				{Addr: 0x49, W: []byte{statusReg}, R: []byte{0x00}},
				{Addr: 0x49, W: []byte{writeReg, 0x20}, R: []byte{}},
			},
			timeout: time.Millisecond * 100,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		bus := &i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		d, _ := New(bus, &DefaultOpts)

		got := d.Gain(tt.gain)

		if _, ok := tt.wantErr.(*IOError); ok {
			if _, ok := got.(*IOError); !ok {
				t.Errorf("expected IOError but %T", got)
			}
			if got.(*IOError).Op != tt.wantErr.(*IOError).Op {
				t.Errorf("expected %s, but got %s", tt.wantErr.(*IOError).Op, got.(*IOError).Op)
			}
		} else if got != tt.wantErr {
			t.Errorf("expected error: %v but got: %v", tt.wantErr, got)
		}
	}
}

func TestDev_String(t *testing.T) {
	want := "AMS AS7262 6 channel visible spectrum sensor"
	d := &Dev{}
	if d.String() != want {
		t.Errorf("expected %s but got %s", want, d.String())
	}
}

func TestIOError_Error(t *testing.T) {
	tests := []struct {
		name string
		op   string
		err  error
		want string
	}{
		{"nil", "doing nothing", nil, "ioerror while doing nothing"},
		{"errTimeoutPin", "", errPinTimeout, "ioerror while : timeout waiting for interrupt signal on pin"},
	}
	for _, tt := range tests {
		e := &IOError{tt.op, tt.err}
		got := e.Error()
		if tt.want != got {
			t.Errorf("expected %s but got %s", tt.want, got)
		}
	}
}
