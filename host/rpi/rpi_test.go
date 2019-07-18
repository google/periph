// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package rpi

import (
	"reflect"
	"testing"
)

func TestParseRevision(t *testing.T) {
	data := []struct {
		v uint32
		r revisionCode
	}{
		// https://www.raspberrypi.org/documentation/hardware/raspberrypi/revision-codes/README.md
		// Old style
		{0x2, newFormat | memory256MB | egoman | bcm2835 | board1B},
		{0x3, newFormat | memory256MB | egoman | bcm2835 | board1B},
		{0x4, newFormat | memory256MB | sonyUK | bcm2835 | board1B | 2},
		{0x5, newFormat | memory256MB | bcm2835 | board1B | 2},
		{0x6, newFormat | memory256MB | egoman | bcm2835 | board1B | 2},
		{0x7, newFormat | memory256MB | egoman | bcm2835 | board1A | 2},
		{0x8, newFormat | memory256MB | sonyUK | bcm2835 | board1A | 2},
		{0x9, newFormat | memory256MB | bcm2835 | board1A | 2},
		{0xd, newFormat | memory512MB | egoman | bcm2835 | board1B | 2},
		{0xe, newFormat | memory512MB | sonyUK | bcm2835 | board1B | 2},
		{0xf, newFormat | memory512MB | egoman | bcm2835 | board1B | 2},
		{0x10, newFormat | memory512MB | sonyUK | bcm2835 | board1BPlus | 2},
		{0x11, newFormat | memory512MB | sonyUK | bcm2835 | boardCM1},
		{0x12, newFormat | memory256MB | sonyUK | bcm2835 | board1APlus | 1},
		{0x13, newFormat | memory512MB | embest | bcm2835 | board1BPlus | 2},
		{0x14, newFormat | memory512MB | embest | bcm2835 | boardCM1},
		{0x15, newFormat | memory256MB | embest | bcm2835 | board1APlus | 1},
		// Test warranty bit
		{0x1000015, warrantyVoid | newFormat | memory256MB | embest | bcm2835 | board1APlus | 1},
		// New style
		{0x900021, newFormat | memory512MB | sonyUK | bcm2835 | board1APlus | 1},
		{0x900032, newFormat | memory512MB | sonyUK | bcm2835 | board1BPlus | 2},
		{0x900092, newFormat | memory512MB | sonyUK | bcm2835 | boardZero | 2},
		{0x900093, newFormat | memory512MB | sonyUK | bcm2835 | boardZero | 3},
		{0x9000c1, newFormat | memory512MB | sonyUK | bcm2835 | boardZeroW | 1},
		{0x9020e0, newFormat | memory512MB | sonyUK | bcm2837 | board3APlus},
		{0x920092, newFormat | memory512MB | embest | bcm2835 | boardZero | 2},
		{0x920093, newFormat | memory512MB | embest | bcm2835 | boardZero | 3},
		{0x900061, newFormat | memory512MB | sonyUK | bcm2835 | boardCM1 | 1},
		{0xa01040, newFormat | memory1GB | sonyUK | bcm2836 | board2B},
		{0xa01041, newFormat | memory1GB | sonyUK | bcm2836 | board2B | 1},
		{0xa02082, newFormat | memory1GB | sonyUK | bcm2837 | board3B | 2},
		{0xa020a0, newFormat | memory1GB | sonyUK | bcm2837 | boardCM3},
		{0xa020d3, newFormat | memory1GB | sonyUK | bcm2837 | board3BPlus | 3},
		{0xa21041, newFormat | memory1GB | embest | bcm2836 | board2B | 1},
		{0xa22042, newFormat | memory1GB | embest | bcm2837 | board2B | 2},
		{0xa22082, newFormat | memory1GB | embest | bcm2837 | board3B | 2},
		{0xa220a0, newFormat | memory1GB | embest | bcm2837 | boardCM3},
		{0xa32082, newFormat | memory1GB | sonyJapan | bcm2837 | board3B | 2},
		{0xa52082, newFormat | memory1GB | stadium | bcm2837 | board3B | 2},
		{0xa22083, newFormat | memory1GB | embest | bcm2837 | board3B | 3},
		{0xa02100, newFormat | memory1GB | sonyUK | bcm2837 | boardCM3Plus},
		{0xa03111, newFormat | memory1GB | sonyUK | bcm2711 | board4B | 1},
		{0xb03111, newFormat | memory2GB | sonyUK | bcm2711 | board4B | 1},
		{0xc03111, newFormat | memory4GB | sonyUK | bcm2711 | board4B | 1},
	}
	for i, line := range data {
		r, err := parseRevision(line.v)
		if err != nil {
			t.Fatalf("#%d: unexpected failure: %v", i, err)
		}
		if line.r != r {
			t.Fatalf("#%d: unexpected: %#x != %#x", i, line.r, r)
		}
	}
}

func TestParseRevisionErr(t *testing.T) {
	data := []uint32{0, 1, 0xa, 0xb, 0xc, 0x16}
	for i, v := range data {
		if _, err := parseRevision(v); err == nil {
			t.Fatalf("#%d: unexpected success for %#x", i, v)
		}
	}
}

func TestFeaturesInit(t *testing.T) {
	data := []struct {
		v uint32
		f features
	}{
		{0x2, features{hdrP1P26: true, hdrAudio: true}},                             // board1B
		{0x3, features{hdrP1P26: true, hdrAudio: true}},                             // board1B
		{0x4, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1B
		{0x5, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1B
		{0x6, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1B
		{0x7, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1A
		{0x8, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1A
		{0x9, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1A
		{0xd, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1B
		{0xe, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1B
		{0xf, features{hdrP1P26: true, hdrP5: true, hdrAudio: true, hdrHDMI: true}}, // board1B
		{0x10, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}},             // board1BPlus
		{0x11, features{}}, // boardCM1
		{0x12, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}}, // board1APlus
		{0x13, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}}, // board1BPlus
		{0x14, features{}}, // boardCM1
		{0x15, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}},                        // board1APlus
		{0x900092, features{hdrP1P40: true, hdrHDMI: true}},                                    // boardZero
		{0x900093, features{hdrP1P40: true, hdrHDMI: true}},                                    // boardZero
		{0x9000c1, features{hdrP1P40: true, hdrHDMI: true}},                                    // boardZeroW
		{0x920093, features{hdrP1P40: true, hdrHDMI: true}},                                    // boardZero
		{0xa01040, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}},                    // board2B
		{0xa01041, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}},                    // board2B
		{0xa02082, features{hdrP1P40: true, hdrAudio: true, audioLeft41: true, hdrHDMI: true}}, // board3B
		{0xa020a0, features{hdrSODIMM: true}},                                                  // boardCM3
		{0xa020d3, features{hdrP1P40: true, hdrAudio: true, audioLeft41: true, hdrHDMI: true}}, // board3BPlus
		{0xa21041, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}},                    // board2B
		{0xa22042, features{hdrP1P40: true, hdrAudio: true, hdrHDMI: true}},                    // board2B
		{0xa22082, features{hdrP1P40: true, hdrAudio: true, audioLeft41: true, hdrHDMI: true}}, // board3B
		{0xa32082, features{hdrP1P40: true, hdrAudio: true, audioLeft41: true, hdrHDMI: true}}, // board3B
	}
	for i, line := range data {
		f := features{}
		if err := f.init(line.v); err != nil {
			t.Fatalf("#%d: unexpected failure for %#x: %v", i, line.v, err)
		}
		if line.f != f {
			t.Fatalf("#%d: unexpected for %#x:\nexpected: %#v\nactual:   %#v", i, line.v, line.f, f)
		}
	}
}

func TestFeaturesInitErr(t *testing.T) {
	data := []uint32{
		0x0,
		0x1,
		0x16,
		0x900021, // board1APlus
		0x900032, // board1BPlus
		0x9020e0, // board3APlus
		0x920092, // boardZero
		0x900061, // boardCM1
		0xa220a0, // boardCM3
		0xa52082, // board3B
		0xa22083, // board3B
		0xa02100, // boardCM3Plus
		0xa03111, // board4B
		0xb03111, // board4B
		0xc03111, // board4B
	}
	for i, v := range data {
		f := features{}
		if err := f.init(v); err == nil {
			t.Fatalf("#%d: unexpected success for %#x", i, v)
		}
	}
}

func TestDriver(t *testing.T) {
	if v := drv.String(); v != "rpi" {
		t.Fatal(v)
	}
	if v := drv.Prerequisites(); v != nil {
		t.Fatal(v)
	}
	if v := drv.After(); reflect.DeepEqual(v, []string{"bcm2835-gpio"}) {
		t.Fatal(v)
	}
}
