// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ina219

import (
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/physic"
)

func TestNew(t *testing.T) {
	stringErr := errors.New("use err.Error() error")

	type fields struct {
		currentLSB physic.ElectricCurrent
		powerLSB   physic.Power
	}

	var tests = []struct {
		name      string
		opts      Config
		want      fields
		tx        []i2ctest.IO
		err       error
		errString string
	}{
		{name: "defaults",
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
			},
			want: fields{
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   976560 * physic.NanoWatt,
			},
		},
		{name: "badAddressOption",
			opts: Config{Address: 0x60},
			err:  errAddressOutOfRange,
		},
		{name: "badSenseResistorOption",
			opts: Config{SenseResistor: -1},
			err:  errSenseResistorValueInvalid,
		},
		{name: "badMaxCurrentOption",
			opts: Config{MaxCurrent: -1},
			err:  errMaxCurrentInvalid,
		},
		{name: "setAddress",
			opts: Config{Address: 0x41},
			tx: []i2ctest.IO{
				{Addr: 0x41, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x41, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
			},
			want: fields{
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   976560 * physic.NanoWatt,
			},
			err: nil,
		},
		{name: "setMaxCurrent",
			opts: Config{MaxCurrent: 1000 * physic.MilliAmpere},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x68, 0xdc}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
			},
			want: fields{
				currentLSB: 15258 * physic.NanoAmpere,
				powerLSB:   305160 * physic.NanoWatt,
			},
			err: nil,
		},
		{name: "setSenseResistor",
			opts: Config{SenseResistor: 10 * physic.MilliOhm},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x47, 0xae}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
			},
			want: fields{
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   976560 * physic.NanoWatt,
			},
			err: nil,
		},
		{name: "txError",
			tx: []i2ctest.IO{{Addr: 0x40, W: []byte{}, R: []byte{}}},
			want: fields{
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   976560 * physic.NanoWatt,
			},
			err:       stringErr,
			errString: "unexpected write",
		},
		{name: "errWritingToConfigRegister",
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister}, R: []byte{}},
			},
			want: fields{
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   976560 * physic.NanoWatt,
			},
			err: errWritingToConfigRegister,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			bus := &i2ctest.Playback{
				Ops:       test.tx,
				DontPanic: true,
			}

			ina, err := New(bus, test.opts)

			if test.err != nil {
				if err != test.err {
					if test.err == stringErr {
						if !strings.Contains(err.Error(), test.errString) {
							t.Errorf("%v wanted err: %v, but got: %v", test.name, test.errString, err)
						}
					} else {
						t.Errorf("%v wanted err: %v, but got: %v", test.name, test.err, err)
					}
				}
			}

			if test.err == nil {
				if ina == nil {
					t.Errorf("%v wanted no err but got: %v", test.name, err)
					return
				}

				var got = fields{
					currentLSB: ina.currentLSB,
					powerLSB:   ina.powerLSB,
				}
				if got != test.want {
					t.Errorf("%v wanted: %v, but got: %v", test.name, test.want, got)
				}

			}

		})
	}

}

func TestSense(t *testing.T) {
	stringErr := errors.New("use err.Error() error")
	type fields struct {
		currentLSB physic.ElectricCurrent
		powerLSB   physic.Power
	}

	var tests = []struct {
		name      string
		args      Config
		want      PowerMonitor
		tx        []i2ctest.IO
		err       error
		errString string
	}{
		{
			name: "errReadShunt",
			err:  errReadShunt,
			args: Config{},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{}},
			},
		},
		{
			name: "errReadBus",
			err:  errReadBus,
			args: Config{},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{}},
			},
		},
		{
			name: "errReadCurrent",
			err:  errReadCurrent,
			args: Config{},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{}},
			},
		},
		{
			name: "errReadPower",
			err:  errReadPower,
			args: Config{},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{powerRegister}, R: []byte{}},
			},
		},
		{
			name: "readZero",
			err:  nil,
			args: Config{},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{powerRegister}, R: []byte{0x00, 0x00}},
			},
			want: PowerMonitor{Shunt: 0, Voltage: 0, Current: 0, Power: 0},
		},
		{
			name: "busVoltageOverflow",
			err:  errRegisterOverflow,
			args: Config{},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{calibrationRegister, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{configRegister, 0x1f, 0xff}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 0x01}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bus := &i2ctest.Playback{
				Ops:       test.tx,
				DontPanic: true,
			}
			ina, err := New(bus, Config{})
			if err != nil {
				t.Fatalf("set setup failure %v", err)
			}
			if ina == nil {
				t.Fatalf("device init failed")
			}
			got, err := ina.Sense()
			// fmt.Println(got, err)
			if test.err != nil {
				if err != test.err {
					if test.err == stringErr {
						if !strings.Contains(err.Error(), test.errString) {
							t.Errorf("%v wanted err: %v, but got: %v", test.name, test.errString, err)
						}
					} else {
						t.Errorf("%v wanted err: %v, but got: %v", test.name, test.err, err)
					}
				}
			}

			if test.err == nil {
				if err != nil {
					t.Errorf("%v wanted no err but got: %v", test.name, err)
					return
				}

				if got != test.want {
					t.Errorf("%v wanted: %v, but got: %v", test.name, test.want, got)
				}

			}
		})
	}
}

func TestCalibrate(t *testing.T) {
	stringErr := errors.New("use err.Error() error")

	type fields struct {
		sense      physic.ElectricResistance
		maxCurrent physic.ElectricCurrent
		currentLSB physic.ElectricCurrent
		powerLSB   physic.Power
		caibrated  bool
	}
	tests := []struct {
		name      string
		tx        []i2ctest.IO
		args      fields
		want      fields
		err       error
		errString string
	}{
		{
			name: "errBadSense",
			err:  errSenseResistorValueInvalid,
		},
		{
			name: "errBadMaxCurrent",
			args: fields{
				sense: physic.MilliOhm,
			},
			err: errMaxCurrentInvalid,
		},
		{
			name: "errIO",
			args: fields{
				sense:      physic.MilliOhm,
				maxCurrent: physic.Ampere,
			},
			err:       stringErr,
			errString: "unexpected Tx",
		},
		{
			name: "default",
			args: fields{
				sense:      100 * physic.MilliOhm,
				maxCurrent: 3200 * physic.MilliAmpere,
			},
			want: fields{
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   976560 * physic.NanoWatt,
			},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
			},
			err: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			bus := i2ctest.Playback{
				Ops:       test.tx,
				DontPanic: true,
			}

			ina := &Dev{
				// maxCurrent: test.args.maxCurrent,
				// sense:      test.args.sense,
				m: &mmr.Dev8{
					Conn:  &i2c.Dev{Bus: &bus, Addr: 0x40},
					Order: binary.BigEndian},
			}

			err := ina.Calibrate(test.args.sense, test.args.maxCurrent)
			if test.err != nil {
				if err != test.err {
					if test.err == stringErr {
						if !strings.Contains(err.Error(), test.errString) {
							t.Errorf("%v wanted err: %v, but got: %v", test.name, test.errString, err)
						}
					} else {
						t.Errorf("%v wanted err: %v, but got: %v", test.name, test.err, err)
					}
				}
			}
			if test.err == nil {
				if err != nil {
					t.Errorf("%v wanted no err but got: %v", test.name, err)
				}
				got := fields{
					// sense:      ina.sense,
					// maxCurrent: ina.maxCurrent,
					currentLSB: ina.currentLSB,
					powerLSB:   ina.powerLSB,
					// caibrated:  ina.caibrated,
				}
				if got != test.want {
					t.Errorf("%v wanted: %v, but got: %v", test.name, test.want, got)
				}
			}
		})
	}
}

func TestPowerStringer(t *testing.T) {
	var p = PowerMonitor{
		Shunt:   1,
		Voltage: 1,
		Current: 1,
		Power:   1,
	}
	want := "Bus: 1nV, Current: 1nA, Power: 1nW, Shunt: 1nV"
	got := p.String()
	if want != got {
		t.Errorf("wanted %s\n, but got: %s", want, got)
	}
}
