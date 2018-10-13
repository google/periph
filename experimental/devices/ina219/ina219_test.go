// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ina219

import (
	"errors"
	"strings"
	"testing"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/physic"
)

func TestAddress(t *testing.T) {
	var tests = []struct {
		name string
		arg  uint8
		dev  *Ina219
		want uint16
		err  error
	}{
		{"TestSenseResistor(0x01)",
			0x01,
			&Ina219{},
			0x40,
			errAddressOutOfRange,
		},
		{"TestSenseResistor(0x01)",
			0x50,
			&Ina219{},
			0x40,
			errAddressOutOfRange,
		},
		{"TestSenseResistor(0x40)",
			0x40,
			&Ina219{},
			0x40,
			nil,
		},
		{"TestSenseResistor(0x40)",
			0x43,
			&Ina219{c: &i2c.Dev{Bus: &i2ctest.Playback{}, Addr: 0x40}, addr: 0x40},
			0x43,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ina := test.dev
			optfn := Address(test.arg)
			err := optfn(ina)
			if err != test.err {
				t.Errorf("test:%s, wanted err: %v, but got: %v", test.name, test.err, err)
			}
			if test.err == nil {
				if ina.addr != test.want {
					t.Errorf("test:%s, wanted value: %x, but got: %x", test.name, test.want, ina.addr)
				}
			}
		})
	}
}

func TestMaxCurrent(t *testing.T) {
	var tests = []struct {
		name string
		arg  physic.ElectricCurrent
		want physic.ElectricCurrent
		err  error
	}{
		{"TestMaxCurrent(1A)", 10 * physic.Ampere, 10 * physic.Ampere, nil},
		{"TestMaxCurrent(32A)", 32 * physic.Ampere, 32 * physic.Ampere, nil},
		{"TestMaxCurrent(2A)", 2 * physic.Ampere, 2 * physic.Ampere, nil},
		{"TestMaxCurrent(0)", 0 * physic.NanoAmpere, DefaultMaxCurrent, errMaxCurrentInvalid},
		{"TestMaxCurrent(-10A)", -150 * physic.NanoAmpere, DefaultMaxCurrent, errMaxCurrentInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ina := &Ina219{
				maxCurrent: DefaultMaxCurrent,
			}
			optfn := MaxCurrent(test.arg)
			err := optfn(ina)
			if err != test.err {
				t.Errorf("test:%s, wanted err: %v, but got: %v", test.name, test.err, err)
			}
			if test.err == nil {
				if ina.maxCurrent != test.want {
					t.Errorf("test:%s, wanted value: %s, but got: %s", test.name, test.want, ina.sense)
				}
			}
		})
	}
}

func TestSenseResistor(t *testing.T) {
	var tests = []struct {
		name string
		arg  physic.ElectricResistance
		want physic.ElectricResistance
		err  error
	}{
		{"TestSenseResistor(10mOhms)", 10 * physic.MilliOhm, 10 * physic.MilliOhm, nil},
		{"TestSenseResistor(100mOhms)", 100 * physic.MilliOhm, 100 * physic.MilliOhm, nil},
		{"TestSenseResistor(150mOhms)", 150 * physic.MilliOhm, 150 * physic.MilliOhm, nil},
		{"TestSenseResistor(0mOhms)", 0 * physic.MilliOhm, DefaultSenseResistor, errSenseResistorValueInvalid},
		{"TestSenseResistor(-150mOhms)", -150 * physic.MilliOhm, DefaultSenseResistor, errSenseResistorValueInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ina := &Ina219{
				sense: DefaultSenseResistor,
			}
			optfn := SenseResistor(test.arg)
			err := optfn(ina)
			if err != test.err {
				t.Errorf("test:%s, wanted err: %v, but got: %v", test.name, test.err, err)
			}
			if test.err == nil {
				if ina.sense != test.want {
					t.Errorf("test:%s, wanted value: %s, but got: %s", test.name, test.want, ina.sense)
				}
			}
		})
	}
}
func TestNew(t *testing.T) {
	stringErr := errors.New("use err.Error() error")

	type fields struct {
		addr       uint16
		caibrated  bool
		sense      physic.ElectricResistance
		maxCurrent physic.ElectricCurrent
		currentLSB physic.ElectricCurrent
		powerLSB   physic.Power
	}

	var tests = []struct {
		name      string
		opts      []Option
		want      fields
		tx        []i2ctest.IO
		err       error
		errString string
	}{
		{name: "defaults",
			tx: []i2ctest.IO{{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}}},
			want: fields{
				addr:       0x40,
				caibrated:  true,
				sense:      100 * physic.MilliOhm,
				maxCurrent: 3200 * physic.MilliAmpere,
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   2441 * physic.NanoWatt,
			},
		},
		{name: "badAddressOption",
			opts: []Option{Address(0x60)},
			err:  errAddressOutOfRange,
		},
		{name: "badSenseResistorOption",
			opts: []Option{SenseResistor(-1)},
			err:  errSenseResistorValueInvalid,
		},
		{name: "badMaxCurrentOption",
			opts: []Option{MaxCurrent(-1)},
			err:  errMaxCurrentInvalid,
		},
		{name: "setAddress",
			opts: []Option{Address(0x41)},
			tx:   []i2ctest.IO{{Addr: 0x41, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}}},
			want: fields{
				addr:       0x41,
				caibrated:  true,
				sense:      100 * physic.MilliOhm,
				maxCurrent: 3200 * physic.MilliAmpere,
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   2441 * physic.NanoWatt,
			},
			err: nil,
		},
		{name: "setMaxCurrent",
			opts: []Option{MaxCurrent(1000 * physic.MilliAmpere)},
			tx:   []i2ctest.IO{{Addr: 0x40, W: []byte{0x05, 0x68, 0xdc}, R: []byte{}}},
			want: fields{
				addr:       0x40,
				caibrated:  true,
				sense:      100 * physic.MilliOhm,
				maxCurrent: 1000 * physic.MilliAmpere,
				currentLSB: 15258 * physic.NanoAmpere,
				powerLSB:   762 * physic.NanoWatt,
			},
			err: nil,
		},
		{name: "setSenseResistor",
			opts: []Option{SenseResistor(10 * physic.MilliOhm)},
			tx:   []i2ctest.IO{{Addr: 0x40, W: []byte{0x05, 0x47, 0xae}, R: []byte{}}},
			want: fields{
				addr:       0x40,
				caibrated:  true,
				sense:      10 * physic.MilliOhm,
				maxCurrent: 3200 * physic.MilliAmpere,
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   2441 * physic.NanoWatt,
			},
			err: nil,
		},
		{name: "txError",
			tx: []i2ctest.IO{{Addr: 0x40, W: []byte{}, R: []byte{}}},
			want: fields{
				addr:       0x40,
				caibrated:  true,
				sense:      100 * physic.MilliOhm,
				maxCurrent: 3200 * physic.MilliAmpere,
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   2441 * physic.NanoWatt,
			},
			err:       stringErr,
			errString: "unexpected write",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			bus := &i2ctest.Playback{
				Ops:       test.tx,
				DontPanic: true,
			}

			ina, err := New(bus, test.opts...)

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
					addr:       ina.addr,
					caibrated:  ina.caibrated,
					sense:      ina.sense,
					maxCurrent: ina.maxCurrent,
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
		addr       uint16
		caibrated  bool
		currentLSB physic.ElectricCurrent
		powerLSB   physic.Power
	}

	var tests = []struct {
		name      string
		args      fields
		want      PowerMonitor
		tx        []i2ctest.IO
		err       error
		errString string
	}{
		{
			name: "errReadShunt",
			err:  errReadShunt,
			args: fields{caibrated: true},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{}},
			},
		},
		{
			name: "errReadBus",
			err:  errReadBus,
			args: fields{caibrated: true},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{}},
			},
		},
		{
			name: "errReadCurrent",
			err:  errReadCurrent,
			args: fields{caibrated: true},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{}},
			},
		},
		{
			name: "errReadPower",
			err:  errReadPower,
			args: fields{caibrated: true},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{powerRegister}, R: []byte{}},
			},
		},
		{
			name: "readZero",
			err:  nil,
			args: fields{caibrated: true},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{0x00, 0x00}},
				{Addr: 0x40, W: []byte{powerRegister}, R: []byte{0x00, 0x00}},
			},
			want: PowerMonitor{Shunt: 0, Voltage: 0, Current: 0, Power: 0},
		},
		{
			name: "notCalibrated",
			err:  nil,
			args: fields{caibrated: false},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 1}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 1 << 3}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{0x00, 0x01}},
				{Addr: 0x40, W: []byte{powerRegister}, R: []byte{0x00, 0x01}},
			},
			want: PowerMonitor{
				Shunt:   10 * physic.MicroVolt,
				Voltage: 4 * physic.MilliVolt,
				Current: 0,
				Power:   0,
			},
		},
		{
			name: "Calibrated",
			err:  nil,
			args: fields{caibrated: true},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
				{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 1}},
				{Addr: 0x40, W: []byte{busVoltageRegister}, R: []byte{0x00, 1 << 3}},
				{Addr: 0x40, W: []byte{currentRegister}, R: []byte{0x00, 0x01}},
				{Addr: 0x40, W: []byte{powerRegister}, R: []byte{0x00, 0x01}},
			},
			want: PowerMonitor{
				Shunt:   10 * physic.MicroVolt,
				Voltage: 4 * physic.MilliVolt,
				Current: 48828 * physic.NanoAmpere,
				Power:   2441 * physic.NanoWatt,
			},
		},
		{
			name: "busVoltageOverflow",
			err:  errRegisterOverflow,
			args: fields{caibrated: true},
			tx: []i2ctest.IO{
				{Addr: 0x40, W: []byte{0x05, 0x20, 0xc4}, R: []byte{}},
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
			ina, err := New(bus)
			if err != nil {
				t.Fatalf("set setup failure %v", err)
			}
			if ina == nil {
				t.Fatalf("device init failed")
			}

			ina.caibrated = test.args.caibrated
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
				sense:      100 * physic.MilliOhm,
				maxCurrent: 3200 * physic.MilliAmpere,
				currentLSB: 48828 * physic.NanoAmpere,
				powerLSB:   2441 * physic.NanoWatt,
				caibrated:  true,
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

			ina := &Ina219{
				maxCurrent: test.args.maxCurrent,
				sense:      test.args.sense,
				c: &i2c.Dev{
					Bus:  &bus,
					Addr: 0x40,
				},
			}

			err := ina.Calibrate()
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
					sense:      ina.sense,
					maxCurrent: ina.maxCurrent,
					currentLSB: ina.currentLSB,
					powerLSB:   ina.powerLSB,
					caibrated:  ina.caibrated,
				}
				if got != test.want {
					t.Errorf("%v wanted: %v, but got: %v", test.name, test.want, got)
				}
			}
		})
	}
}

func TestReset(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x40, W: []byte{0x00, 0x80, 0x00}, R: []byte{}},
			{Addr: 0x40, W: []byte{0x00}, R: []byte{0x39, 0x9f}},
			{Addr: 0x40, W: []byte{0x00, 0x80, 0x00}, R: []byte{}},
			{Addr: 0x40, W: []byte{0x00}, R: []byte{0x3f, 0xff}},
			{Addr: 0x40, W: []byte{0x00, 0x80, 0x00}, R: []byte{}},
			{Addr: 0x40, W: []byte{0x00}, R: nil},
			{Addr: 0x40, W: []byte{}, R: []byte{}},
		},
		DontPanic: true,
	}
	ina := &Ina219{
		maxCurrent: DefaultMaxCurrent,
		sense:      DefaultSenseResistor,
		c: &i2c.Dev{
			Bus:  &bus,
			Addr: 0x40,
		},
	}

	if err := ina.reset(); err != nil {
		t.Fatal(err)
	}

	if err := ina.reset(); err != errResetError {
		t.Fatal(err)
	}

	if err := ina.reset(); !strings.Contains(err.Error(), "unexpected read buffer length") {
		t.Fatal(err)
	}

	if err := ina.reset(); !strings.Contains(err.Error(), "unexpected write") {
		t.Fatal(err)
	}
}

func TestWriteRegister(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x40, W: []byte{shuntVoltageRegister, 0x00, 0x00}, R: []byte{}},
			{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{}},
		},
		DontPanic: true,
	}
	ina := &Ina219{
		maxCurrent: DefaultMaxCurrent,
		sense:      DefaultSenseResistor,
		c: &i2c.Dev{
			Bus:  &bus,
			Addr: 0x40,
		},
	}
	err := ina.WriteRegister(shuntVoltageRegister, 0x0000)
	if err != nil {
		t.Errorf("%v", err)
	}

	err = ina.WriteRegister(shuntVoltageRegister, 0x0000)
	if err == nil {
		t.Errorf("wanted error but got none")
	}
}
func TestReadRegister(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{0x00, 0x00}},
			{Addr: 0x40, W: []byte{shuntVoltageRegister}, R: []byte{}},
		},
		DontPanic: true,
	}
	ina := &Ina219{
		maxCurrent: DefaultMaxCurrent,
		sense:      DefaultSenseResistor,
		c: &i2c.Dev{
			Bus:  &bus,
			Addr: 0x40,
		},
	}
	data, err := ina.ReadRegister(shuntVoltageRegister)
	if err != nil {
		t.Errorf("%v %v ", err, data)
	}
	if data != 0 {
		t.Errorf("expected zero but got %d", data)
	}

	_, err = ina.ReadRegister(shuntVoltageRegister)
	if err == nil {
		t.Errorf("wanted error but got none")
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

func TestClearbit(t *testing.T) {
	var bitTests = []struct {
		n    byte
		pos  uint8
		want byte
	}{
		{0xFF, 7, 0x7F},
		{0xFF, 6, 0xBF},
		{0xFF, 5, 0xDF},
		{0xFF, 4, 0xEF},
		{0xFF, 3, 0xF7},
		{0xFF, 2, 0xFB},
		{0xFF, 1, 0xFD},
		{0xFF, 0, 0xFE},
	}
	for _, test := range bitTests {
		got := clearBit(test.n, test.pos)
		if got != test.want {
			t.Errorf("want %02x but got %02x for %02x", test.want, got, test.n)
		}
	}
}

func TestSetbit(t *testing.T) {
	var bitTests = []struct {
		n    byte
		pos  uint8
		want byte
	}{
		{0x7F, 7, 0xFF},
		{0xBF, 6, 0xFF},
		{0xDF, 5, 0xFF},
		{0xEF, 4, 0xFF},
		{0xF7, 3, 0xFF},
		{0xFB, 2, 0xFF},
		{0xFD, 1, 0xFF},
		{0xFE, 0, 0xFF},
	}
	for _, test := range bitTests {
		got := setBit(test.n, test.pos)
		if got != test.want {

			t.Errorf("want %02x but got %02x for %02x", test.want, got, test.n)
		}
	}
}

func TestHasbit(t *testing.T) {
	var bitTests = []struct {
		n    byte
		pos  uint8
		want bool
	}{
		{0x7F, 7, false},
		{0xBF, 6, false},
		{0xDF, 5, false},
		{0xEF, 4, false},
		{0xF7, 3, false},
		{0xFB, 2, false},
		{0xFD, 1, false},
		{0xFE, 0, false},
		{0x80, 7, true},
		{0x40, 6, true},
		{0x20, 5, true},
		{0x10, 4, true},
		{0x08, 3, true},
		{0x04, 2, true},
		{0x02, 1, true},
		{0x01, 0, true},
	}
	for _, test := range bitTests {
		got := hasBit(test.n, test.pos)
		if got != test.want {
			t.Errorf("want %t but got %t for %02x", test.want, got, test.n)
		}
	}
}
