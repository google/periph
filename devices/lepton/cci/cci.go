// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// "stringer" can be installed with "go get golang.org/x/tools/cmd/stringer"
//go:generate stringer -output=strings_gen.go -type=CameraStatus,command,FFCShutterMode,FFCState,ShutterPos,ShutterTempLockoutState

// Package cci declares the Camera Command Interface to interact with a FLIR
// Lepton over I²C.
//
// This protocol controls and queries the camera but is not used to read the
// images.
//
// Datasheet
//
// http://www.flir.com/uploadedFiles/OEM/Products/LWIR-Cameras/Lepton/FLIR-Lepton-Software-Interface-Description-Document.pdf
//
// Found via http://www.flir.com/cores/display/?id=51878
package cci

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/devices/lepton/internal"
)

// StatusBit is the status as returned by the FLIR Lepton.
type StatusBit uint16

// Status bitmask.
const (
	StatusBusy       StatusBit = 0x1
	StatusBootNormal StatusBit = 0x2
	StatusBooted     StatusBit = 0x4
	StatusErrorMask  StatusBit = 0xFF00
)

func (s StatusBit) String() string {
	var o []string
	if s&StatusBusy != 0 {
		o = append(o, "Busy")
	}
	if s&StatusBootNormal != 0 {
		o = append(o, "BootNormal")
	}
	if s&StatusBooted != 0 {
		o = append(o, "Booted")
	}
	if v := s & StatusErrorMask; v != 0 {
		o = append(o, "0x"+strconv.FormatUint(uint64(v)>>8, 16))
	}
	return strings.Join(o, "|")
}

// CameraStatus returns the status of the FLIR Lepton's camera.
type CameraStatus uint32

// Valid values for CameraStatus.
const (
	SystemReady              CameraStatus = 0
	SystemInitializing       CameraStatus = 1
	SystemInLowPowerMode     CameraStatus = 2
	SystemGoingIntoStandby   CameraStatus = 3
	SystemFlatFieldInProcess CameraStatus = 4
)

// Status returns the camera status as returned by the camera.
type Status struct {
	CameraStatus CameraStatus
	CommandCount uint16
}

// ShutterTempLockoutState is used in FFCMode.
type ShutterTempLockoutState uint32

// Valid values for ShutterTempLockoutState.
const (
	ShutterTempLockoutStateInactive ShutterTempLockoutState = 0
	ShutterTempLockoutStateHigh     ShutterTempLockoutState = 1
	ShutterTempLockoutStateLow      ShutterTempLockoutState = 2
)

// FFCShutterMode is used in FFCMode.
type FFCShutterMode uint32

// Valid values for FFCShutterMode.
const (
	FFCShutterModeManual   FFCShutterMode = 0
	FFCShutterModeAuto     FFCShutterMode = 1
	FFCShutterModeExternal FFCShutterMode = 2
)

// ShutterPos returns the shutter position, which is used to calibrate the
// camera.
type ShutterPos uint32

// Valid values for ShutterPos.
const (
	ShutterPosUnknown ShutterPos = 0xFFFFFFFF // -1
	ShutterPosIdle    ShutterPos = 0
	ShutterPosOpen    ShutterPos = 1
	ShutterPosClosed  ShutterPos = 2
	ShutterPosBrakeOn ShutterPos = 3
)

// FFCState describes the Flat-Field Correction state.
type FFCState uint8

const (
	// FFCNever means no FFC was requested.
	FFCNever FFCState = 0
	// FFCInProgress means a FFC is in progress. It lasts 23 frames (at 27fps) so it lasts less than a second.
	FFCInProgress FFCState = 1
	// FFCComplete means FFC was completed successfully.
	FFCComplete FFCState = 2
)

// FFCMode describes the various self-calibration settings and state.
type FFCMode struct {
	FFCShutterMode          FFCShutterMode          // Default: FFCShutterModeExternal
	ShutterTempLockoutState ShutterTempLockoutState // Default: ShutterTempLockoutStateInactive
	ElapsedTimeSinceLastFFC time.Duration           // Uptime
	DesiredFFCPeriod        time.Duration           // Default: 300s
	DesiredFFCTempDelta     physic.Temperature      // Default: 3K
	ImminentDelay           uint16                  // Default: 52
	VideoFreezeDuringFFC    bool                    // Default: true
	FFCDesired              bool                    // Default: false
	ExplicitCommandToOpen   bool                    // Default: false
}

// New returns a driver for the FLIR Lepton CCI protocol.
func New(i i2c.Bus) (*Dev, error) {
	d := &Dev{
		c: cciConn{r: mmr.Dev16{Conn: &i2c.Dev{Bus: i, Addr: 0x2A}, Order: internal.Big16}},
	}
	// Wait for the device to be booted.
	for {
		if status, err := d.c.waitIdle(); err != nil {
			return nil, err
		} else if status == StatusBootNormal|StatusBooted {
			return d, nil
		}
		//log.Printf("lepton not yet booted: 0x%02x", status)
		// Polling rocks.
		sleep(5 * time.Millisecond)
	}
}

// Dev is the Lepton specific Command and Control Interface (CCI).
//
//
// Dev can safely accessed concurrently via multiple goroutines.
//
// This interface is accessed via I²C and provides access to view and modify
// the internal state.
//
// Maximum I²C speed is 1Mhz.
type Dev struct {
	c      cciConn
	serial uint64
}

func (d *Dev) String() string {
	return d.c.String()
}

// Init initializes the FLIR Lepton in raw 14 bits mode, enables telemetry as
// header.
func (d *Dev) Init() error {
	if err := d.c.set(agcEnable, internal.Disabled); err != nil {
		return err
	}
	// Setup telemetry to always be as the header. There's no reason to make this
	// configurable by the user.
	if err := d.c.set(sysTelemetryEnable, internal.Enabled); err != nil {
		return err
	}
	if err := d.c.set(sysTelemetryLocation, internal.Header); err != nil {
		return err
	}

	/*
		// Verification code in case the I²C do not work properly.
		f := internal.Enabled
		if err := d.c.get(agcEnable, &f); err != nil {
			return err
		} else if f != internal.Disabled {
			return fmt.Errorf("lepton-cci: internal verification for AGC failed %v", f)
		}
		if err := d.c.get(sysTelemetryEnable, &f); err != nil {
			return err
		} else if f != internal.Enabled {
			return fmt.Errorf("lepton-cci: internal verification for telemetry flag failed %v", f)
		}
		hdr := internal.Footer
		if err := d.c.get(sysTelemetryLocation, &hdr); err != nil {
			return err
		} else if hdr != internal.Header {
			return fmt.Errorf("lepton-cci: internal verification for telemetry position failed %s", hdr)
		}
	*/
	return nil
}

// WaitIdle waits for camera to be ready.
//
// It loops forever and returns the StatusBit.
func (d *Dev) WaitIdle() (StatusBit, error) {
	return d.c.waitIdle()
}

// Halt stops the camera.
func (d *Dev) Halt() error {
	// TODO(maruel): Doc says it won't restart. Yo.
	return d.c.run(oemPowerDown)
}

// GetStatus return the status of the camera as known by the camera itself.
func (d *Dev) GetStatus() (*Status, error) {
	var v internal.Status
	if err := d.c.get(sysStatus, &v); err != nil {
		return nil, err
	}
	return &Status{
		CameraStatus: CameraStatus(v.CameraStatus),
		CommandCount: v.CommandCount,
	}, nil
}

// GetSerial returns the FLIR Lepton serial number.
func (d *Dev) GetSerial() (uint64, error) {
	if d.serial == 0 {
		out := uint64(0)
		if err := d.c.get(sysSerialNumber, &out); err != nil {
			return out, err
		}
		d.serial = out
	}
	return d.serial, nil
}

// GetUptime returns the uptime. Rolls over after 1193 hours.
func (d *Dev) GetUptime() (time.Duration, error) {
	var v internal.DurationMS
	if err := d.c.get(sysUptime, &v); err != nil {
		return 0, err
	}
	return v.Duration(), nil
}

// GetTemp returns the temperature inside the camera.
func (d *Dev) GetTemp() (physic.Temperature, error) {
	var v internal.CentiK
	if err := d.c.get(sysTemperature, &v); err != nil {
		return 0, err
	}
	return v.Temperature(), nil
}

// GetTempHousing returns the temperature of the camera housing.
func (d *Dev) GetTempHousing() (physic.Temperature, error) {
	var v internal.CentiK
	if err := d.c.get(sysHousingTemperature, &v); err != nil {
		return 0, err
	}
	return v.Temperature(), nil
}

// GetFFCModeControl returns the internal state with regards to calibration.
func (d *Dev) GetFFCModeControl() (*FFCMode, error) {
	v := internal.FFCMode{}
	if err := d.c.get(sysFFCMode, &v); err != nil {
		return nil, err
	}
	return &FFCMode{
		FFCShutterMode:          FFCShutterMode(v.FFCShutterMode),
		ShutterTempLockoutState: ShutterTempLockoutState(v.ShutterTempLockoutState),
		ElapsedTimeSinceLastFFC: v.ElapsedTimeSinceLastFFC.Duration(),
		DesiredFFCPeriod:        v.DesiredFFCPeriod.Duration(),
		DesiredFFCTempDelta:     v.DesiredFFCTempDelta.Temperature(),
		ImminentDelay:           v.ImminentDelay,
		VideoFreezeDuringFFC:    v.VideoFreezeDuringFFC == internal.Enabled,
		FFCDesired:              v.FFCDesired == internal.Enabled,
		ExplicitCommandToOpen:   v.ExplicitCommandToOpen == internal.Enabled,
	}, nil
}

// GetShutterPos returns the position of the shutter if present.
func (d *Dev) GetShutterPos() (ShutterPos, error) {
	out := ShutterPosUnknown
	err := d.c.get(sysShutterPosition, &out)
	return out, err
}

// RunFFC forces a Flat-Field Correction to be done by the camera for
// recalibration. It takes 23 frames and the camera runs at 27fps so it lasts
// less than a second.
func (d *Dev) RunFFC() error {
	return d.c.run(sysFCCRunNormalization)
}

//

// cciConn is the low level connection.
//
// It implements the low level protocol to run the GET, SET and RUN commands
// via memory mapped registers.
type cciConn struct {
	mu sync.Mutex
	r  mmr.Dev16
}

func (c *cciConn) String() string {
	return fmt.Sprintf("%s", &c.r)
}

// waitIdle waits for the busy bit to clear.
func (c *cciConn) waitIdle() (StatusBit, error) {
	// Do not take the lock.
	for {
		if s, err := c.r.ReadUint16(regStatus); err != nil || StatusBit(s)&StatusBusy == 0 {
			return StatusBit(s), err
		}
		sleep(5 * time.Millisecond)
	}
}

// get returns an attribute by querying the device.
func (c *cciConn) get(cmd command, data interface{}) error {
	if data == nil {
		return errors.New("lepton-cci: get() argument must not be nil")
	}
	if t := reflect.TypeOf(data); t.Kind() != reflect.Ptr && t.Kind() != reflect.Slice {
		return fmt.Errorf("lepton-cci: get() argument must be a pointer or a slice, got %T", data)
	}
	size := binary.Size(data)
	if size&1 != 0 {
		return errors.New("lepton-cci: get() argument must be 16 bits aligned")
	}
	nbWords := size / 2
	if nbWords > 1024 {
		return errors.New("cci: buffer too large")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.waitIdle(); err != nil {
		return err
	}
	if err := c.r.WriteUint16(regDataLength, uint16(nbWords)); err != nil {
		return err
	}
	if err := c.r.WriteUint16(regCommandID, uint16(cmd)); err != nil {
		return err
	}
	s, err := c.waitIdle()
	if err != nil {
		return err
	}
	if s&0xff00 != 0 {
		return fmt.Errorf("cci: error 0x%x", byte(s>>8))
	}
	if nbWords <= 16 {
		err = c.r.ReadStruct(regData0, data)
	} else {
		err = c.r.ReadStruct(regDataBuffer0, data)
	}
	if err != nil {
		return err
	}
	/*
		// Verify CRC:
		if crc, err := c.r.ReadUint16(regDataCRC); err != nil {
			return err
		} else if expected := internal.CRC16(data); expected != crc {
			return fmt.Errorf("invalid crc; expected 0x%04X; got 0x%04X", expected, crc)
		}
	*/
	//log.Printf("get(%s) = %v", cmd, data)
	return nil
}

// set returns an attribute on the device.
func (c *cciConn) set(cmd command, data interface{}) error {
	if data == nil {
		return errors.New("lepton-cci: set() argument must not be nil")
	}
	size := binary.Size(data)
	if size&1 != 0 {
		return errors.New("lepton-cci: set() argument must be 16 bits aligned")
	}
	nbWords := size / 2
	if nbWords > 1024 {
		return errors.New("lepton-cci: buffer too large")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.waitIdle(); err != nil {
		return err
	}
	var err error
	if nbWords <= 16 {
		err = c.r.WriteStruct(regData0, data)
	} else {
		err = c.r.WriteStruct(regDataBuffer0, data)
	}
	if err != nil {
		return err
	}
	if err := c.r.WriteUint16(regDataLength, uint16(nbWords)); err != nil {
		return err
	}
	if err := c.r.WriteUint16(regCommandID, uint16(cmd)|1); err != nil {
		return err
	}
	s, err := c.waitIdle()
	if err != nil {
		return err
	}
	if s&0xff00 != 0 {
		return fmt.Errorf("cci: error 0x%x", s>>8)
	}
	return nil
}

// run runs a command on the device that doesn't need any argument.
func (c *cciConn) run(cmd command) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.waitIdle(); err != nil {
		return err
	}
	if err := c.r.WriteUint16(regDataLength, 0); err != nil {
		return err
	}
	if err := c.r.WriteUint16(regCommandID, uint16(cmd)|2); err != nil {
		return err
	}
	s, err := c.waitIdle()
	if err != nil {
		return err
	}
	if s&0xff00 != 0 {
		return fmt.Errorf("cci: error 0x%x", s>>8)
	}
	return nil
}

//

// All the available registers.
const (
	regPower       uint16 = 0
	regStatus      uint16 = 2
	regCommandID   uint16 = 4
	regDataLength  uint16 = 6
	regData0       uint16 = 8
	regData1       uint16 = 10
	regData2       uint16 = 12
	regData3       uint16 = 14
	regData4       uint16 = 16
	regData5       uint16 = 18
	regData6       uint16 = 20
	regData7       uint16 = 22
	regData8       uint16 = 24
	regData9       uint16 = 26
	regData10      uint16 = 28
	regData11      uint16 = 30
	regData12      uint16 = 32
	regData13      uint16 = 34
	regData14      uint16 = 36
	regData15      uint16 = 38
	regDataCRC     uint16 = 40
	regDataBuffer0 uint16 = 0xF800
	regDataBuffer1 uint16 = 0xFC00
)

// command is a command supported by the FLIR Lepton over its CCI interface.
type command uint16

// All the available commands.
//
// See page 17 for more details.
//
// Number of words and supported action.
const (
	agcEnable                 command = 0x0100 // 2   GET/SET
	agcRoiSelect              command = 0x0108 // 4   GET/SET
	agcHistogramStats         command = 0x010C // 4   GET
	agcHeqDampFactor          command = 0x0124 // 1   GET/SET
	agcHeqClipLimitHigh       command = 0x012C // 1   GET/SET
	agcHeqClipLimitLow        command = 0x0130 // 1   GET/SET
	agcHeqEmptyCounts         command = 0x013C // 1   GET/SET
	agcHeqOutputScaleFactor   command = 0x0144 // 2   GET/SET
	agcCalculationEnable      command = 0x0148 // 2   GET/SET
	oemPowerDown              command = 0x4800 // 0   RUN
	oemPartNumber             command = 0x481C // 16  GET
	oemSoftwareRevision       command = 0x4820 // 4   GET
	oemVideoOutputEnable      command = 0x4824 // 2   GET/SET
	oemVideoOutputFormat      command = 0x4828 // 2   GET/SET
	oemVideoOutputSource      command = 0x482C // 2   GET/SET
	oemCustomerPartNumber     command = 0x4838 // 16  GET
	oemVideoOutputConst       command = 0x483C // 1   GET/SET
	oemCameraReboot           command = 0x4840 // 0   RUN
	oemFCCNormalizationTarget command = 0x4844 // 1   GET/SET/RUN
	oemStatus                 command = 0x4848 // 2   GET
	oemFrameMeanIntensity     command = 0x484C // 1   GET
	oemGPIOModeSelect         command = 0x4854 // 2   GET/SET
	oemGPIOVSyncPhaseDelay    command = 0x4858 // 2   GET/SET
	oemUserDefaults           command = 0x485C // 2   GET/RUN
	oemRestoreUserDefaults    command = 0x4860 // 0   RUN
	oemShutterProfile         command = 0x4064 // 2   GET/SET
	oemThermalShutdownEnable  command = 0x4868 // 2   GET/SET
	oemBadPixel               command = 0x486C // 2   GET/SET
	oemTemporalFilter         command = 0x4870 // 2   GET/SET
	oemColumnNoiseFilter      command = 0x4874 // 2   GET/SET
	oemPixelNoiseFilter       command = 0x4878 // 2   GET/SET
	sysPing                   command = 0x0200 // 0   RUN
	sysStatus                 command = 0x0204 // 4   GET
	sysSerialNumber           command = 0x0208 // 4   GET
	sysUptime                 command = 0x020C // 2   GET
	sysHousingTemperature     command = 0x0210 // 1   GET
	sysTemperature            command = 0x0214 // 1   GET
	sysTelemetryEnable        command = 0x0218 // 2   GET/SET
	sysTelemetryLocation      command = 0x021C // 2   GET/SET
	sysExecuteFrameAverage    command = 0x0220 // 0   RUN     Undocumented but listed in SDK
	sysFlatFieldFrames        command = 0x0224 // 2   GET/SET It's an enum, max is 128
	sysCustomSerialNumber     command = 0x0228 // 16  GET     It's a string
	sysRoiSceneStats          command = 0x022C // 4   GET
	sysRoiSceneSelect         command = 0x0230 // 4   GET/SET
	sysThermalShutdownCount   command = 0x0234 // 1   GET     Number of times it exceeded 80C
	sysShutterPosition        command = 0x0238 // 2   GET/SET
	sysFFCMode                command = 0x023C // 17  GET/SET Manual control; doc says 20 words but it's 17 in practice.
	sysFCCRunNormalization    command = 0x0240 // 0   RUN
	sysFCCStatus              command = 0x0244 // 2   GET
	vidColorLookupSelect      command = 0x0304 // 2   GET/SET
	vidColorLookupTransfer    command = 0x0308 // 512 GET/SET
	vidFocusCalculationEnable command = 0x030C // 2   GET/SET
	vidFocusRoiSelect         command = 0x0310 // 4   GET/SET
	vidFocusMetricThreshold   command = 0x0314 // 2   GET/SET
	vidFocusMetricGet         command = 0x0318 // 2   GET
	vidVideoFreezeEnable      command = 0x0324 // 2   GET/SET
)

// TODO(maruel): Enable RadXXX commands.

var sleep = time.Sleep

var _ conn.Resource = &Dev{}
var _ fmt.Stringer = &Dev{}
var _ fmt.Stringer = &cciConn{}
