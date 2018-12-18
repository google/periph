// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mcp9808

import (
	"encoding/binary"
	"reflect"
	"testing"
	"time"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/physic"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts Opts
		tx   []i2ctest.IO
		want error
	}{
		{
			name: "defaults",
			opts: DefaultOpts,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{resolutionConfig, 0x03}, R: []byte{}},
				{Addr: 0x18, W: []byte{configuration, 0x00, 0x00}, R: []byte{}},
			},
			want: nil,
		},
		{
			name: "bad address",
			opts: Opts{Addr: 0x40},
			want: errAddressOutOfRange,
		},
		{
			name: "io error",
			opts: Opts{Addr: 0x18},
			want: errWritingResolution,
		},
	}

	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		_, err := New(&bus, &tt.opts)
		if err != tt.want {
			t.Errorf("New(%s) expected %v but got %v ", tt.name, tt.want, err)
		}
	}
}

func TestSense(t *testing.T) {
	tests := []struct {
		name     string
		want     physic.Temperature
		tx       []i2ctest.IO
		shutdown bool
		err      error
	}{
		{
			name: "0C",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x00, 0x00}},
			},
			want: physic.ZeroCelsius,
			err:  nil,
		},
		{
			name: "10C",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x00, 0xa0}},
			},
			want: physic.ZeroCelsius + 10*physic.Kelvin,
			err:  nil,
		},
		{
			name: "-10C",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x10, 0xa0}},
			},
			want: physic.ZeroCelsius - 10*physic.Kelvin,
			err:  nil,
		},
		{
			name: "io error",
			tx:   []i2ctest.IO{},
			err:  errReadTemperature,
		},
		{
			name:     "enable error",
			shutdown: true,
			want:     physic.ZeroCelsius + 10*physic.Kelvin,
			err:      errWritingConfiguration,
		},
	}

	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
			shutdown: tt.shutdown,
		}
		e := &physic.Env{}
		err := mcp9808.Sense(e)
		if err == nil && tt.want != e.Temperature {
			t.Errorf("%s Sense() expected %v but got %v ", tt.name, tt.want, e.Temperature)
		}
		if err != tt.err {
			t.Errorf("%s Sense() expected %v but got %v ", tt.name, tt.err, err)
		}
	}
}

func TestSenseContinuous(t *testing.T) {
	tests := []struct {
		name     string
		want     physic.Temperature
		res      resolution
		interval time.Duration
		tx       []i2ctest.IO
		Halt     bool
		err      error
	}{
		{
			name:     "errTooShortInterval Max",
			res:      Maximum,
			interval: 249 * time.Millisecond,
			err:      errTooShortInterval,
		},
		{
			name:     "errTooShortInterval High",
			res:      High,
			interval: 129 * time.Millisecond,
			err:      errTooShortInterval,
		},
		{
			name:     "errTooShortInterval Med",
			res:      Medium,
			interval: 64 * time.Millisecond,
			err:      errTooShortInterval,
		},
		{
			name:     "errTooShortInterval Low",
			res:      Low,
			interval: 29 * time.Millisecond,
			err:      errTooShortInterval,
		},
		{
			name: "Halt",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x00, 0xa0}},
				{Addr: 0x18, W: []byte{configuration, 0x01, 0x00}, R: nil},
			},
			want:     physic.ZeroCelsius + 10*physic.Celsius,
			res:      Low,
			interval: 30 * time.Millisecond,
			Halt:     true,
			err:      nil,
		},
	}
	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian,
			},
			res:  tt.res,
			stop: make(chan struct{}, 1),
		}

		env, err := mcp9808.SenseContinuous(tt.interval)

		if tt.Halt {
			e := <-env
			err := mcp9808.Halt()
			if err != tt.err {
				t.Errorf("SenseContinuous() %s wanted err: %v, but got: %v", tt.name, tt.err, err)
			}
			if err == nil && e.Temperature != tt.want {
				t.Errorf("SenseContinuous() %s wanted %v, but got: %v", tt.name, tt.want, e.Temperature)
			}
		}

		if err != tt.err {
			t.Errorf("SenseContinuous() %s wanted err: %v, but got: %v", tt.name, tt.err, err)
		}
		mcp9808.Halt()
	}

}

func TestPrecision(t *testing.T) {
	tests := []struct {
		name string
		want physic.Temperature
		res  resolution
	}{
		{
			name: "Maximum",
			want: 62500 * physic.MicroKelvin,
			res:  Maximum,
		},
		{
			name: "High",
			want: 125 * physic.MilliKelvin,
			res:  High,
		},
		{
			name: "Medium",
			want: 250 * physic.MilliKelvin,
			res:  Medium,
		},
		{
			name: "Low",
			want: 500 * physic.MilliKelvin,
			res:  Low,
		},
	}

	for _, tt := range tests {
		d := &Dev{res: tt.res}
		e := &physic.Env{}
		d.Precision(e)
		if e.Temperature != tt.want {
			t.Errorf("Precision(%s) wanted %v but got %v", tt.name, tt.want, e.Temperature)
		}
	}
}

func TestSenseTemp(t *testing.T) {
	tests := []struct {
		name string
		want physic.Temperature
		tx   []i2ctest.IO
		err  error
	}{
		{
			name: "0C",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x00, 0x00}},
			},
			want: physic.ZeroCelsius,
			err:  nil,
		},
		{
			name: "10C",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x00, 0xa0}},
			},
			want: physic.ZeroCelsius + 10*physic.Kelvin,
			err:  nil,
		},
		{
			name: "-10C",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x10, 0xa0}},
			},
			want: physic.ZeroCelsius - 10*physic.Kelvin,
			err:  nil,
		},
		{
			name: "io error",
			tx:   []i2ctest.IO{},
			err:  errReadTemperature,
		},
	}

	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
		}
		got, err := mcp9808.SenseTemp()
		if tt.want != got {
			t.Errorf("%s SenseTemp() expected %v but got %v ", tt.name, tt.want, got)
		}
		if err != tt.err {
			t.Errorf("%s SenseTemp() expected %v but got %v ", tt.name, tt.err, err)
		}
	}
}

func TestSenseWithAlerts(t *testing.T) {
	tests := []struct {
		name     string
		critical physic.Temperature
		upper    physic.Temperature
		lower    physic.Temperature
		temp     physic.Temperature
		alerts   []Alert
		tx       []i2ctest.IO
		err      error
	}{
		{
			name:     "read no alert",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			temp:     physic.ZeroCelsius + 1*physic.Kelvin,
			alerts:   nil,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x00, 0x10}},
			},
			err: nil,
		},
		{
			name:     "get lower Alert",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			temp:     physic.ZeroCelsius - 1*physic.Kelvin,
			alerts: []Alert{
				{"lower", physic.ZeroCelsius},
			},
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x30, 0x10}},
				{Addr: 0x18, W: []byte{lowerAlert}, R: []byte{0x00, 0x00}},
			},
			err: nil,
		},
		{
			name:     "get upper Alert",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			temp:     physic.ZeroCelsius + 7*physic.Kelvin,
			alerts: []Alert{
				{"upper", physic.ZeroCelsius + 5*physic.Kelvin},
			},
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x40, 0x70}},
				{Addr: 0x18, W: []byte{upperAlert}, R: []byte{0x00, 0x50}},
			},
			err: nil,
		},
		{
			name:     "get critical Alert",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			temp:     physic.ZeroCelsius + 15*physic.Kelvin,
			alerts: []Alert{
				{"critical", physic.ZeroCelsius + 10*physic.Kelvin},
				{"upper", physic.ZeroCelsius + 5*physic.Kelvin},
			},
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0xc0, 0xf0}},
				{Addr: 0x18, W: []byte{critAlert}, R: []byte{0x00, 0xa0}},
				{Addr: 0x18, W: []byte{upperAlert}, R: []byte{0x00, 0x50}},
			},
			err: nil,
		},
		{
			name:     "set critical error",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			err:      errWritingCritAlert,
		},
		{
			name:     "set upper error",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
			},
			err: errWritingUpperAlert,
		},
		{
			name:     "set lower error",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
			},
			err: errWritingLowerAlert,
		},
		{
			name:     "invalid alert config",
			critical: physic.ZeroCelsius + 1*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			tx:       []i2ctest.IO{},
			err:      errAlertInvalid,
		},
		{
			name:     "temperature io error",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
			},
			err: errReadTemperature,
		},
		{
			name:     "read critical io error",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			temp:     physic.ZeroCelsius + 15*physic.Kelvin,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0xc0, 0xf0}},
			},
			err: errReadCriticalAlert,
		},
		{
			name:     "read upper io error",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			temp:     physic.ZeroCelsius + 15*physic.Kelvin,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0xc0, 0xf0}},
				{Addr: 0x18, W: []byte{critAlert}, R: []byte{0x00, 0xa0}},
			},
			err: errReadUpperAlert,
		},
		{
			name:     "read lower io error",
			critical: physic.ZeroCelsius + 10*physic.Kelvin,
			upper:    physic.ZeroCelsius + 5*physic.Kelvin,
			lower:    physic.ZeroCelsius,
			temp:     physic.ZeroCelsius - 1*physic.Kelvin,
			alerts:   nil,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0xa0}, R: nil},
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x50}, R: nil},
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
				{Addr: 0x18, W: []byte{temperature}, R: []byte{0x30, 0x10}},
			},
			err: errReadLowerAlert,
		},
	}

	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
		}
		temp, alerts, err := mcp9808.SenseWithAlerts(tt.lower, tt.upper, tt.critical)
		if err != tt.err {
			t.Errorf("SenseWithAlerts(%s) wanted err: %v but got: %v", tt.name, tt.err, err)
		}
		if temp != tt.temp {
			t.Errorf("SenseWithAlerts(%s) wanted temp: %s but got: %s", tt.name, tt.temp, temp)
		}
		if !reflect.DeepEqual(alerts, tt.alerts) {
			t.Errorf("SenseWithAlerts(%s) expected alerts %+v but got %+v ", tt.name, tt.alerts, alerts)
		}
	}
}

func TestDevHalt(t *testing.T) {
	tests := []struct {
		name     string
		tx       []i2ctest.IO
		shutdown bool
		want     bool
		err      error
	}{
		{
			name: "shutdown",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{configuration, 0x01, 0x00}, R: nil},
			},
			shutdown: false,
			want:     true,
			err:      nil,
		},
		{
			name: "io error",
			tx:   []i2ctest.IO{
				// {Addr: 0x18, W: []byte{configuration, 0x01, 0x00}, R: nil},
			},
			shutdown: false,
			want:     false,
			err:      errWritingConfiguration,
		},
	}
	for _, tt := range tests {

		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
			shutdown: tt.shutdown,
		}

		if err := mcp9808.Halt(); err != tt.err {
			t.Errorf("Halt(%s) wanted error: %v but got: %v", tt.name, tt.err, err)
		}
		if mcp9808.shutdown != tt.want {
			t.Errorf("Halt(%s) expected dev to be: %t but is: %t", tt.name, tt.want, mcp9808.shutdown)
		}
	}
}

func TestDevString(t *testing.T) {
	mcp9808 := &Dev{}
	want := "MCP9808"
	if want != mcp9808.String() {
		t.Errorf("mpc9808.String() expected %s but got %s", want, mcp9808.String())
	}
}

func TestDev_enable(t *testing.T) {
	tests := []struct {
		name     string
		tx       []i2ctest.IO
		shutdown bool
		want     bool
		err      error
	}{
		{
			name: "shutdown",
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{configuration, 0x00, 0x00}, R: nil},
			},
			shutdown: true,
			want:     false,
			err:      nil,
		},
		{
			name: "io error",
			tx:   []i2ctest.IO{
				// {Addr: 0x18, W: []byte{configuration, 0x01, 0x00}, R: nil},
			},
			shutdown: true,
			want:     true,
			err:      errWritingConfiguration,
		},
	}
	for _, tt := range tests {

		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}
		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
			shutdown: tt.shutdown,
		}

		if err := mcp9808.enable(); err != tt.err {
			t.Errorf("enable(%s) wanted error: %v but got: %v", tt.name, tt.err, err)
		}
		if mcp9808.shutdown != tt.want {
			t.Errorf("enable(%s) expected dev to be: %t but is: %t", tt.name, tt.want, mcp9808.shutdown)
		}
	}
}

func TestDev_setResolution(t *testing.T) {
	succeeds := []struct {
		name string
		res  resolution
		tx   []i2ctest.IO
		err  error
	}{
		{
			name: "Low",
			res:  Low,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{resolutionConfig, 0x00}, R: nil},
			},
			err: nil,
		}, {
			name: "Medium",
			res:  Medium,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{resolutionConfig, 0x01}, R: nil},
			},
			err: nil,
		}, {
			name: "High",
			res:  High,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{resolutionConfig, 0x02}, R: nil},
			},
			err: nil,
		}, {
			name: "Maximum",
			res:  Maximum,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{resolutionConfig, 0x03}, R: nil},
			},
			err: nil,
		},
	}

	fails := []struct {
		name string
		res  resolution
		tx   []i2ctest.IO
		err  error
	}{
		{
			name: "Low",
			res:  Low,
			err:  errWritingResolution,
		}, {
			name: "Medium",
			res:  Medium,
			err:  errWritingResolution,
		}, {
			name: "High",
			res:  High,
			err:  errWritingResolution,
		}, {
			name: "Maximum",
			res:  Maximum,
			err:  errWritingResolution,
		}, {
			name: "Unknown",
			res:  resolution(5),
			err:  errInvalidResolution,
		},
	}

	for _, tt := range succeeds {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
		}
		if err := mcp9808.setResolution(tt.res); err != tt.err {
			t.Errorf("setResolution(%s) expected %v but got %v", tt.name, tt.err, err)
		}

	}

	for _, tt := range fails {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
		}
		if err := mcp9808.setResolution(tt.res); err != tt.err {
			t.Errorf("setResolution(%s) expected %v but got %v", tt.name, tt.err, err)
		}

	}

}

func TestDev_setCriticalAlert(t *testing.T) {
	tests := []struct {
		name  string
		alert physic.Temperature
		crit  physic.Temperature
		tx    []i2ctest.IO
		err   error
	}{
		{
			name:  "0C",
			alert: physic.ZeroCelsius,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{critAlert, 0x00, 0x00}, R: nil},
			},
			err: nil,
		},
		{
			name:  "126C",
			alert: physic.ZeroCelsius + 126*physic.Kelvin,
			tx:    []i2ctest.IO{},
			err:   errAlertOutOfRange,
		},
		{
			name:  "io error",
			alert: physic.ZeroCelsius,
			tx:    []i2ctest.IO{},
			err:   errWritingCritAlert,
		},
		{
			name:  "nop",
			alert: physic.ZeroCelsius,
			crit:  physic.ZeroCelsius,
			tx:    []i2ctest.IO{},
			err:   nil,
		},
	}

	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
			critical: tt.crit,
		}
		if err := mcp9808.setCriticalAlert(tt.alert); err != tt.err {
			t.Errorf("setCriticalAlert(%s) expected %v but got %v", tt.name, tt.err, err)
		}
	}
}

func TestDev_setUpperAlert(t *testing.T) {
	tests := []struct {
		name  string
		alert physic.Temperature
		tx    []i2ctest.IO
		upper physic.Temperature
		err   error
	}{
		{
			name:  "0C",
			alert: physic.ZeroCelsius,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{upperAlert, 0x00, 0x00}, R: nil},
			},
			err: nil,
		},
		{
			name:  "126C",
			alert: physic.ZeroCelsius + 126*physic.Kelvin,
			tx:    []i2ctest.IO{},
			err:   errAlertOutOfRange,
		},
		{
			name:  "io error",
			alert: physic.ZeroCelsius,
			tx:    []i2ctest.IO{},
			err:   errWritingUpperAlert,
		},
		{
			name:  "nop",
			alert: physic.ZeroCelsius,
			upper: physic.ZeroCelsius,
			tx:    []i2ctest.IO{},
			err:   nil,
		},
	}

	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
			upper: tt.upper,
		}
		if err := mcp9808.setUpperAlert(tt.alert); err != tt.err {
			t.Errorf("setUpperAlert(%s) expected %v but got %v", tt.name, tt.err, err)
		}
	}
}

func TestDev_setLowerAlert(t *testing.T) {
	tests := []struct {
		name  string
		alert physic.Temperature
		tx    []i2ctest.IO
		lower physic.Temperature
		err   error
	}{
		{
			name:  "0C",
			alert: physic.ZeroCelsius,
			tx: []i2ctest.IO{
				{Addr: 0x18, W: []byte{lowerAlert, 0x00, 0x00}, R: nil},
			},
			err: nil,
		},
		{
			name:  "126C",
			alert: physic.ZeroCelsius + 126*physic.Kelvin,
			tx:    []i2ctest.IO{},
			err:   errAlertOutOfRange,
		},
		{
			name:  "io error",
			alert: physic.ZeroCelsius,
			tx:    []i2ctest.IO{},
			err:   errWritingLowerAlert,
		},
		{
			name:  "nop",
			alert: physic.ZeroCelsius,
			lower: physic.ZeroCelsius,
			tx:    []i2ctest.IO{},
			err:   nil,
		},
	}

	for _, tt := range tests {
		bus := i2ctest.Playback{
			Ops:       tt.tx,
			DontPanic: true,
		}

		mcp9808 := &Dev{
			m: mmr.Dev8{
				Conn:  &i2c.Dev{Bus: &bus, Addr: 0x18},
				Order: binary.BigEndian},
			lower: tt.lower,
		}
		if err := mcp9808.setLowerAlert(tt.alert); err != tt.err {
			t.Errorf("setLowerAlert(%s) expected %v but got %v", tt.name, tt.err, err)
		}
	}
}

func Test_alertBitsToTemperature(t *testing.T) {
	tests := []struct {
		name string
		bits uint16
		want physic.Temperature
	}{
		{"0°C", 0x0000, physic.ZeroCelsius},
		{"0.25°C", 0x0004, physic.ZeroCelsius + 250*physic.MilliKelvin},
		{"-0.25°C", 0x1004, physic.ZeroCelsius - 250*physic.MilliKelvin},
	}

	for _, tt := range tests {
		if got := alertBitsToTemperature(tt.bits); got != tt.want {
			t.Errorf("alertBitsToTemperature(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func Test_temperatureToAlertBits(t *testing.T) {
	succeeds := []struct {
		name  string
		alert physic.Temperature
		want  uint16
	}{
		{"0°C", physic.ZeroCelsius, 0x0000},
		{"0.25°C", physic.ZeroCelsius + 250*physic.MilliKelvin, 0x0004},
		{"-0.25°C", physic.ZeroCelsius - 250*physic.MilliKelvin, 0x1004},
		{"124.75°C", physic.ZeroCelsius + 124750*physic.MilliKelvin, 0x07cc},
		{"-39.75°C", physic.ZeroCelsius - 39750*physic.MilliKelvin, 0x127c},
	}

	fails := []struct {
		name  string
		alert physic.Temperature
		want  error
	}{
		{"126°C", physic.ZeroCelsius + 126*physic.Kelvin, errAlertOutOfRange},
		{"-41°C", physic.ZeroCelsius - 41*physic.Kelvin, errAlertOutOfRange},
	}

	for _, tt := range succeeds {
		got, err := alertTemperatureToBits(tt.alert)
		if got != tt.want && err == nil {
			t.Errorf("alertBitsToTemperature(%s) = %x, want %x", tt.name, got, tt.want)
		}
		if err != nil {
			t.Errorf("alertBitsToTemperature(%s) got unexpected error: %v", tt.name, err)
		}
	}

	for _, tt := range fails {
		if _, err := alertTemperatureToBits(tt.alert); err == nil {
			t.Errorf("alertBitsToTemperature(%s) expected error %v", tt.name, tt.want)
		}
	}

}
