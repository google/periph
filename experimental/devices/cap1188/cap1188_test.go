// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package cap1188 controls a Microchip cap1188 device over I²C.

package cap1188_test

import (
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/experimental/devices/cap1188"
)

func Example() {
	// Open the I²C bus to which the cap1188 is connected.
	i2cBus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer i2cBus.Close()

	// We will configure the cap1188 by setting some options, we can start by the defaults.
	opts := cap1188.DefaultOpts()

	// We need to set an alert ping that will let us know when a touch event
	// occurs. The alert pin is the pin connected to the IRQ/interrupt pin.
	alertPin := gpioreg.ByName("GPIO25")
	if alertPin == nil {
		log.Fatal("invalid alert GPIO pin number")
	}
	// We set the alert pin to monitor for interrupts
	if err := alertPin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		log.Fatalf("Can't monitor the alert pin")
	}

	// Optionally but highly recommended, we can also set a reset pin to
	// start/leave things in a clean state.
	resetPin := gpioreg.ByName("GPIO21")
	if resetPin == nil {
		log.Fatal("invalid reset GPIO pin number")
	}
	opts.AlertPin = alertPin
	opts.ResetPin = resetPin

	// open the device so we can detect touch events
	dev, err := cap1188.NewI2C(i2cBus, opts)
	if err != nil {
		log.Fatalf("couldn't open cap1188 - %s", err)
	}
	time.Sleep(200 * time.Millisecond)

	fmt.Println("Monitoring for touch events")
	maxTouches := 42 // stop the program after 42 touches
	for maxTouches > 0 {
		if alertPin.WaitForEdge(-1) {
			maxTouches--
			statuses, err := dev.InputStatus()
			if err != nil {
				fmt.Printf("Error reading inputs: %s\n", err)
				continue
			}
			// print the status of each sensor
			for i, st := range statuses {
				fmt.Printf("#%d: %s\t", i, st)
			}
			fmt.Println()
			// we need to clear the interrupt so it can be triggered again
			if err := dev.ClearInterrupt(); err != nil {
				fmt.Println(err, "while clearing the interrupt")
			}
		}
	}

	fmt.Print("\n")
}

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
			// linked leds
			{Addr: 40, W: []byte{0x72, 0xff}, R: nil},
			// don't retrigger on hold
			{Addr: 40, W: []byte{0x28, 0x0}, R: nil},
			// config
			{Addr: 40, W: []byte{0x20, 0x30}, R: nil},
			// config 2
			{Addr: 40, W: []byte{0x44, 0x61}, R: nil},
		},
	}
	cap1188.Debug = true
	d, err := cap1188.NewI2C(&bus, nil)
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

func TestDev_InputStatus(t *testing.T) {
	tests := []struct {
		name string
		bus  i2c.Bus
		opts *cap1188.Opts
		want []cap1188.TouchStatus
	}{
		{name: "all off",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					// status
					{Addr: 40, W: []byte{0x3}, R: []byte{0x0}},
					// deltas
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					// thresholds
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus},
		},
		{name: "all pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{0xff}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.PressedStatus, cap1188.PressedStatus, cap1188.PressedStatus, cap1188.PressedStatus, cap1188.PressedStatus, cap1188.PressedStatus, cap1188.PressedStatus, cap1188.PressedStatus},
		},
		{name: "first pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 7}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus},
		},
		{name: "second pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 6}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus},
		},
		{name: "third pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 5}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus},
		},
		{name: "forth pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 4}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus},
		},
		{name: "fifth pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 3}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus},
		},
		{name: "sixth pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 2}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus},
		},
		{name: "seventh pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 1}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus, cap1188.OffStatus},
		},
		{name: "eighth pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{1 << 0}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus},
		},
		{name: "3 pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{(1 << 7) ^ (1 << 4) ^ (1 << 0)}},
					{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			opts: cap1188.DefaultOpts(),
			want: []cap1188.TouchStatus{cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.PressedStatus},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := cap1188.NewI2C(tt.bus, tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			got, err := d.InputStatus()
			if err != nil {
				t.Fatalf("Dev.InputStatus() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dev.InputStatus() = %v, want %v", got, tt.want)
			}
		})
	}
	// test hold
	t.Run("held touch sensors", func(t *testing.T) {
		bus := &i2ctest.Playback{
			Ops: append(setupPlaybackIO(), []i2ctest.IO{
				{Addr: 40, W: []byte{0x3}, R: []byte{1 << 7}},
				{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				// repeat call to get status (still pressed)
				{Addr: 40, W: []byte{0x3}, R: []byte{1 << 7}},
				{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				// finall call
				{Addr: 40, W: []byte{0x3}, R: []byte{0x0}},
				{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
			}...),
		}
		// set the recorded response to have the retrigger option on
		bus.Ops[10] = i2ctest.IO{Addr: 40, W: []byte{0x28, 0xff}, R: nil}
		opts := cap1188.DefaultOpts()
		// following option needs to be true so we can get the held status
		opts.RetriggerOnHold = true
		d, err := cap1188.NewI2C(bus, opts)
		if err != nil {
			t.Fatal(err)
		}
		// first check
		got, err := d.InputStatus()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, []cap1188.TouchStatus{cap1188.PressedStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus}) {
			t.Fatalf("expected to have the first sensor touched but instead got %v", got)
		}
		// 2nd check
		got, err = d.InputStatus()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, []cap1188.TouchStatus{cap1188.HeldStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus, cap1188.OffStatus}) {
			t.Fatalf("expected to have the first sensor touched but instead got %v", got)
		}
	})
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
