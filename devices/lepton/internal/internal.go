// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package internal

import (
	"encoding/binary"
	"time"

	"periph.io/x/periph/devices"
)

// Flag is used in FFCMode.
type Flag uint32

// Valid values for Flag.
const (
	Disabled Flag = 0
	Enabled  Flag = 1
)

// DurationMS is duration in millisecond.
//
// It is an implementation detail of the protocol.
type DurationMS uint32

// ToD converts a millisecond based timing to time.Duration.
func (d DurationMS) ToD() time.Duration {
	return time.Duration(d) * time.Millisecond
}

// CentiK is temperature in 0.01°K
//
// It is an implementation detail of the protocol.
type CentiK uint16

// ToC converts a Kelvin measurement to Celsius.
func (c CentiK) ToC() devices.Celsius {
	v := (int(c) - 27315) * 10
	return devices.Celsius(v)
}

// Status returns the camera status as returned by the camera.
type Status struct {
	CameraStatus uint32
	CommandCount uint16
	Reserved     uint16
}

// FFCMode describes the various self-calibration settings and state.
type FFCMode struct {
	FFCShutterMode          uint32     // Default: FFCShutterModeExternal
	ShutterTempLockoutState uint32     // Default: ShutterTempLockoutStateInactive
	VideoFreezeDuringFFC    Flag       // Default: Enabled
	FFCDesired              Flag       // Default: Disabled
	ElapsedTimeSinceLastFFC DurationMS // Uptime in ms.
	DesiredFFCPeriod        DurationMS // Default: 300000
	ExplicitCommandToOpen   Flag       // Default: Disabled
	DesiredFFCTempDelta     uint16     // Default: 300
	ImminentDelay           uint16     // Default: 52

	// These are documented at page 51 but not listed in the structure.
	// ClosePeriodInFrames uint16 // Default: 4
	// OpenPeriodInFrames  uint16 // Default: 1
}

// TelemetryLocation is used with SysTelemetryLocation.
type TelemetryLocation uint32

// Valid values for TelemetryLocation.
const (
	Header TelemetryLocation = 0
	Footer TelemetryLocation = 1
)

//

type table [256]uint16

const ccittFalse = 0x1021

var ccittFalseTable table

func init() {
	makeReversedTable(ccittFalse, &ccittFalseTable)
}

func makeReversedTable(poly uint16, t *table) {
	width := uint16(16)
	for i := uint16(0); i < 256; i++ {
		crc := i << (width - 8)
		for j := 0; j < 8; j++ {
			if crc&(1<<(width-1)) != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
		t[i] = crc
	}
}

func updateReversed(crc uint16, t *table, p []byte) uint16 {
	for _, v := range p {
		crc = t[byte(crc>>8)^v] ^ (crc << 8)
	}
	return crc
}

// CRC16 calculates the reversed CCITT CRC16 checksum.
func CRC16(d []byte) uint16 {
	return updateReversed(0, &ccittFalseTable, d)
}

//

// Big16 translates big endian 16bits words but everything larger is in little
// endian.
//
// It implements binary.ByteOrder.
var Big16 big16

type big16 struct{}

func (big16) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[1]) | uint16(b[0])<<8
}

func (big16) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func (big16) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[1]) | uint32(b[0])<<8 | uint32(b[3])<<16 | uint32(b[2])<<24
}

func (big16) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[1] = byte(v)
	b[0] = byte(v >> 8)
	b[3] = byte(v >> 16)
	b[2] = byte(v >> 24)
}

func (big16) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[1]) | uint64(b[0])<<8 | uint64(b[3])<<16 | uint64(b[2])<<24 |
		uint64(b[5])<<32 | uint64(b[4])<<40 | uint64(b[7])<<48 | uint64(b[6])<<56
}

func (big16) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[1] = byte(v)
	b[0] = byte(v >> 8)
	b[3] = byte(v >> 16)
	b[2] = byte(v >> 24)
	b[5] = byte(v >> 32)
	b[4] = byte(v >> 40)
	b[7] = byte(v >> 48)
	b[6] = byte(v >> 56)
}

func (big16) String() string {
	return "big16"
}

var _ binary.ByteOrder = Big16
