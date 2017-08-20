// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cci

import (
	"testing"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/devices/lepton/internal"
)

func TestStatusBit(t *testing.T) {
	v := ^StatusBit(0)
	if s := v.String(); s != "Busy|BootNormal|Booted|0xff" {
		t.Fatal(s)
	}
}

func TestNew_WaitIdle_fail(t *testing.T) {
	bus := i2ctest.Playback{DontPanic: true}
	if d, err := New(&bus); d != nil || err == nil {
		t.Fatal("WaitIdle() should have returned an error")
	}
}

func TestNew_WaitIdle(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// WaitIdle loop once.
			{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x00}},
			// WaitIdle return not booted.
			{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x01}},
			// WaitIdle return booted.
			{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		},
	}
	c, err := New(&bus)
	if err != nil {
		t.Fatal(err)
	}
	if s := c.String(); s != "playback(42)" {
		t.Fatal(s)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestInit(t *testing.T) {
	// set(agcEnable, internal.Disabled)
	ops := setOps([]byte{0x0, 0x4, 0x1, 0x1}, []byte{0, 0, 0, 0})
	// set(sysTelemetryEnable, internal.Enabled)
	ops = append(ops, setOps([]byte{0x0, 0x4, 0x2, 0x19}, []byte{0, 1, 0, 0})...)
	// set(sysTelemetryLocation, internal.Header)
	ops = append(ops, setOps([]byte{0x0, 0x4, 0x2, 0x1d}, []byte{0, 0, 0, 0})...)
	bus, d := getDev(ops)
	if err := d.Init(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestInit_fail(t *testing.T) {
	data := [][]i2ctest.IO{
		// set(agcEnable, internal.Disabled)
		setOps([]byte{0x0, 0x4, 0x1, 0x1}, []byte{0, 0, 0, 0}),
		// set(sysTelemetryEnable, internal.Enabled)
		setOps([]byte{0x0, 0x4, 0x2, 0x19}, []byte{0, 1, 0, 0}),
		// set(sysTelemetryLocation, internal.Header)
		setOps([]byte{0x0, 0x4, 0x2, 0x1d}, []byte{0, 0, 0, 0}),
	}
	var ops []i2ctest.IO
	{
		bus, d := getDev(ops)
		bus.DontPanic = true
		if d.Init() == nil {
			t.Fatal("failed")
		}
		if err := bus.Close(); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < len(data)-1; i++ {
		ops = append(ops, data[i]...)
		bus, d := getDev(ops)
		bus.DontPanic = true
		if d.Init() == nil {
			t.Fatal("failed")
		}
		if err := bus.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestWaitIdle(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
	}
	bus := i2ctest.Playback{Ops: ops}
	d := Dev{c: conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}}
	if _, err := d.WaitIdle(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestHalt(t *testing.T) {
	bus, d := getDev(runOps([]byte{0x0, 0x4, 0x48, 0x2}))
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetStatus(t *testing.T) {
	bus, d := getDev(getOps([]byte{0x0, 0x4, 0x2, 0x4}, []byte{0, 0, 0, 0, 0, 0, 0, 0}))
	if _, err := d.GetStatus(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetStatus_fail(t *testing.T) {
	if _, err := getDevFail().GetStatus(); err == nil {
		t.Fatal("failed")
	}
}

func TestGetSerial(t *testing.T) {
	bus, d := getDev(getOps([]byte{0x0, 0x4, 0x2, 0x8}, []byte{0, 0, 0, 0, 0, 0, 0, 0}))
	if _, err := d.GetSerial(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetSerial_fail(t *testing.T) {
	if _, err := getDevFail().GetSerial(); err == nil {
		t.Fatal("failed")
	}
}

func TestGetUptime(t *testing.T) {
	bus, d := getDev(getOps([]byte{0x0, 0x4, 0x2, 0xc}, []byte{0, 0, 0, 0}))
	if _, err := d.GetUptime(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetUptime_fail(t *testing.T) {
	if _, err := getDevFail().GetUptime(); err == nil {
		t.Fatal("failed")
	}
}

func TestGetTemp(t *testing.T) {
	bus, d := getDev(getOps([]byte{0x0, 0x4, 0x2, 0x14}, []byte{0, 0}))
	if _, err := d.GetTemp(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTemp_fail(t *testing.T) {
	if _, err := getDevFail().GetTemp(); err == nil {
		t.Fatal("failed")
	}
}

func TestGetTempHousing(t *testing.T) {
	bus, d := getDev(getOps([]byte{0x0, 0x4, 0x2, 0x10}, []byte{0, 0}))
	if _, err := d.GetTempHousing(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetTempHousing_fail(t *testing.T) {
	if _, err := getDevFail().GetTempHousing(); err == nil {
		t.Fatal("failed")
	}
}

func TestGetFFCModeControl(t *testing.T) {
	bus, d := getDev(getOps([]byte{0x0, 0x4, 0x2, 0x3c}, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
	if _, err := d.GetFFCModeControl(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetFFCModeControl_fail(t *testing.T) {
	if _, err := getDevFail().GetFFCModeControl(); err == nil {
		t.Fatal("failed")
	}
}

func TestGetShutterPos(t *testing.T) {
	bus, d := getDev(getOps([]byte{0x0, 0x4, 0x2, 0x38}, []byte{0, 0, 0, 0}))
	if _, err := d.GetShutterPos(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestGetShutterPos_fail(t *testing.T) {
	if _, err := getDevFail().GetShutterPos(); err == nil {
		t.Fatal("failed")
	}
}

func TestRunFFC(t *testing.T) {
	bus, d := getDev(runOps([]byte{0x0, 0x4, 0x2, 0x42}))
	if err := d.RunFFC(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

//

func TestConn_get(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, 0x4}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x4}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regData0
		{Addr: 42, W: []byte{0x00, 0x08}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	var v internal.Status
	if err := c.get(sysStatus, &v); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}

	// Test error paths.
	for len(ops) != 0 {
		ops = ops[:len(ops)-1]
		bus := i2ctest.Playback{Ops: ops, DontPanic: true}
		c = conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
		var v internal.Status
		if c.get(sysStatus, &v) == nil {
			t.Fatal("should have failed")
		}
		if err := bus.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConn_get_large(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x4, 0x0}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x4}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regData0
		{Addr: 42, W: []byte{0xf8, 0}, R: make([]byte, 2048)},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	v := make([]byte, 2048)
	if err := c.get(sysStatus, v); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestConn_get_fail_waitidle(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, 0x4}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x4}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x01, 0x00}},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	var v internal.Status
	if c.get(sysStatus, &v) == nil {
		t.Fatal("waitIdle failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestConn_get_fail(t *testing.T) {
	bus := i2ctest.Playback{}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	if c.get(sysStatus, nil) == nil {
		t.Fatal("nil value")
	}
	if c.get(sysStatus, 1) == nil {
		t.Fatal("not a pointer")
	}
	v := []byte{0}
	if c.get(sysStatus, &v) == nil {
		t.Fatal("odd length")
	}
	v = make([]byte, 2048+2)
	if c.get(sysStatus, &v) == nil {
		t.Fatal("overflow")
	}
}

func TestConn_set(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regData0
		{Addr: 42, W: []byte{0x0, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, 0x4}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x5}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	var v internal.Status
	if err := c.set(sysStatus, &v); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}

	// Test error paths.
	for len(ops) != 0 {
		ops = ops[:len(ops)-1]
		bus := i2ctest.Playback{Ops: ops, DontPanic: true}
		c = conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
		var v internal.Status
		if c.set(sysStatus, &v) == nil {
			t.Fatal("should have failed")
		}
		if err := bus.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConn_set_large(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regData0
		{Addr: 42, W: append([]byte{0xf8, 0}, make([]byte, 2048)...)},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x4, 0x0}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x5}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	v := make([]byte, 2048)
	if err := c.set(sysStatus, v); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestConn_set_fail_waitidle(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regData0
		{Addr: 42, W: []byte{0x0, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, 0x4}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x5}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x0f, 0x00}},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	var v internal.Status
	if c.set(sysStatus, &v) == nil {
		t.Fatal("waitIdle failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestConn_set_fail(t *testing.T) {
	bus := i2ctest.Playback{}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	if c.set(sysStatus, nil) == nil {
		t.Fatal("nil value")
	}
	v := []byte{0}
	if c.set(sysStatus, &v) == nil {
		t.Fatal("odd length")
	}
	v = make([]byte, 2048+2)
	if c.set(sysStatus, &v) == nil {
		t.Fatal("overflow")
	}
}

func TestConn_run(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, 0x0}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x42}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	if err := c.run(sysFCCRunNormalization); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}

	// Test error paths.
	for len(ops) != 0 {
		ops = ops[:len(ops)-1]
		bus := i2ctest.Playback{Ops: ops, DontPanic: true}
		c = conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
		if c.run(sysFCCRunNormalization) == nil {
			t.Fatal("should have failed")
		}
		if err := bus.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConn_run_fail_waitidle(t *testing.T) {
	ops := []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, 0x0}},
		// regCommandID
		{Addr: 42, W: []byte{0x0, 0x4, 0x2, 0x42}},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x0f, 0x00}},
	}
	bus := i2ctest.Playback{Ops: ops}
	c := conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: &bus, Addr: 0x2A}, Order: internal.Big16}}
	if c.run(sysFCCRunNormalization) == nil {
		t.Fatal("waitIdle failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

//

func TestStrings(t *testing.T) {
	if s := SystemReady.String(); s != "SystemReady" {
		t.Fatal(s)
	}
	if s := CameraStatus(30).String(); s != "CameraStatus(30)" {
		t.Fatal(s)
	}

	if s := agcEnable.String(); s != "agcEnable" {
		t.Fatal(s)
	}
	if s := command(0).String(); s != "command(0)" {
		t.Fatal(s)
	}

	if s := FFCShutterModeManual.String(); s != "FFCShutterModeManual" {
		t.Fatal(s)
	}
	if s := FFCShutterMode(30).String(); s != "FFCShutterMode(30)" {
		t.Fatal(s)
	}

	if s := FFCNever.String(); s != "FFCNever" {
		t.Fatal(s)
	}
	if s := FFCState(30).String(); s != "FFCState(30)" {
		t.Fatal(s)
	}

	if s := ShutterPosIdle.String(); s != "ShutterPosIdle" {
		t.Fatal(s)
	}
	if s := ShutterPos(30).String(); s != "ShutterPos(30)" {
		t.Fatal(s)
	}
	if s := ShutterPosUnknown.String(); s != "ShutterPosUnknown" {
		t.Fatal(s)
	}

	if s := ShutterTempLockoutStateInactive.String(); s != "ShutterTempLockoutStateInactive" {
		t.Fatal(s)
	}
	if s := ShutterTempLockoutState(30).String(); s != "ShutterTempLockoutState(30)" {
		t.Fatal(s)
	}
}

//

func getDev(ops []i2ctest.IO) (*i2ctest.Playback, *Dev) {
	bus := &i2ctest.Playback{Ops: ops}
	d := &Dev{c: conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: bus, Addr: 0x2A}, Order: internal.Big16}}}
	return bus, d
}

func getDevFail() *Dev {
	bus := &i2ctest.Playback{DontPanic: true}
	d := &Dev{c: conn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: bus, Addr: 0x2A}, Order: internal.Big16}}}
	return d
}

func getOps(cmd, data []byte) []i2ctest.IO {
	return []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, byte(len(data) / 2)}},
		// regCommandID
		{Addr: 42, W: cmd},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regData0
		{Addr: 42, W: []byte{0x00, 0x08}, R: data},
	}
}

func setOps(cmd, data []byte) []i2ctest.IO {
	return []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regData0
		{Addr: 42, W: append([]byte{0, 8}, data...)},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, byte(len(data) / 2)}},
		// regCommandID
		{Addr: 42, W: cmd},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
	}
}

func runOps(c []byte) []i2ctest.IO {
	return []i2ctest.IO{
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
		// regDataLength
		{Addr: 42, W: []byte{0x0, 0x6, 0x0, 0x0}},
		// regCommandID
		{Addr: 42, W: c},
		// waitIdle
		{Addr: 42, W: []byte{0x00, 0x02}, R: []byte{0x00, 0x06}},
	}
}
