// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package cap1188

import (
	"flag"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"periph.io/x/periph/conn/i2c/i2ctest"
)

func TestNewI2C(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// chip ID
			{Addr: 40, W: []byte{0xfd}, R: []byte{0x50}},
			// clear interrupt
			{Addr: 40, W: []byte{0x0}, R: []byte{0x0}},
			{Addr: 40, W: []byte{0x0, 0x0}, R: nil},
			// enable all inputs
			{Addr: 40, W: []byte{0x21, 0xff}, R: nil},
			// enable interrupts
			{Addr: 40, W: []byte{0x27, 0xff}, R: nil},
			// enable/disable repeats
			{Addr: 40, W: []byte{0x28, 0xff}, R: nil},
			// multitouch
			{Addr: 40, W: []byte{0x2a, 0x4}, R: nil},
			// sampling
			{Addr: 40, W: []byte{0x24, 0x8}, R: nil},
			// sensitivity
			{Addr: 40, W: []byte{0x1f, 0x50}, R: nil},
			// don't retrigger on hold
			{Addr: 40, W: []byte{0x28, 0x0}, R: nil},
			// config
			{Addr: 40, W: []byte{0x20, 0x30}, R: nil},
			// config 2
			{Addr: 40, W: []byte{0x44, 0x61}, R: nil},
		},
	}
	d, err := NewI2C(&bus, &Opts{Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	if s := d.String(); s != "cap1188{playback(40)}" {
		t.Fatal(s)
	}
	if err := d.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func setupPlaybackIO() []i2ctest.IO {
	return []i2ctest.IO{
		// chip ID
		{Addr: 40, W: []byte{0xfd}, R: []byte{0x50}},
		// clear interrupt
		{Addr: 40, W: []byte{0x0}, R: []byte{0x0}},
		{Addr: 40, W: []byte{0x0, 0x0}, R: nil},
		// enable all inputs
		{Addr: 40, W: []byte{0x21, 0xff}, R: nil},
		// enable interrupts
		{Addr: 40, W: []byte{0x27, 0xff}, R: nil},
		// enable/disable repeats
		{Addr: 40, W: []byte{0x28, 0xff}, R: nil},
		// multitouch
		{Addr: 40, W: []byte{0x2a, 0x4}, R: nil},
		// sampling
		{Addr: 40, W: []byte{0x24, 0x8}, R: nil},
		// sensitivity
		{Addr: 40, W: []byte{0x1f, 0x50}, R: nil},
		// linked leds
		{Addr: 40, W: []byte{0x72, 0xff}, R: nil},
		// don't retrigger on hold
		{Addr: 40, W: []byte{0x28, 0x0}, R: nil},
		// config
		{Addr: 40, W: []byte{0x20, 0x30}, R: nil},
		// config 2
		{Addr: 40, W: []byte{0x44, 0x61}, R: nil},
	}
}

func init() {
	sleep = func(time.Duration) {}
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
}
