// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build go1.8

package cap1188

import (
	"reflect"
	"testing"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2ctest"
)

func TestDev_InputStatus(t *testing.T) {
	tests := []struct {
		name string
		bus  i2c.Bus
		want [8]TouchStatus
	}{
		{name: "all off",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					// status
					{Addr: 40, W: []byte{0x3}, R: []byte{0x0}},
					// deltas
					//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					// thresholds
					//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			want: [8]TouchStatus{OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus},
		},
		{name: "all pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{0xff}},
					//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			want: [8]TouchStatus{PressedStatus, PressedStatus, PressedStatus, PressedStatus, PressedStatus, PressedStatus, PressedStatus, PressedStatus},
		},
		{name: "first pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{0x80}},
					//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			want: [8]TouchStatus{PressedStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus},
		},
		{name: "second pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{0x40}},
					//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			want: [8]TouchStatus{OffStatus, PressedStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus},
		},
		{name: "third pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{0x20}},
					//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			want: [8]TouchStatus{OffStatus, OffStatus, PressedStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus},
		},
		{name: "eighth pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{0x1}},
					//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			want: [8]TouchStatus{OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, PressedStatus},
		},
		{name: "3 pressed",
			bus: &i2ctest.Playback{
				Ops: append(setupPlaybackIO(), []i2ctest.IO{
					{Addr: 40, W: []byte{0x3}, R: []byte{0x91}},
					//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
					//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				}...),
			},
			want: [8]TouchStatus{PressedStatus, OffStatus, OffStatus, PressedStatus, OffStatus, OffStatus, OffStatus, PressedStatus},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewI2C(tt.bus, &DefaultOpts)
			if err != nil {
				t.Fatal(err)
			}
			var got [8]TouchStatus
			if err := d.InputStatus(got[:]); err != nil {
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
				{Addr: 40, W: []byte{0x3}, R: []byte{0x80}},
				//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				// repeat call to get status (still pressed)
				{Addr: 40, W: []byte{0x3}, R: []byte{0x80}},
				//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				// finall call
				{Addr: 40, W: []byte{0x3}, R: []byte{0x0}},
				//{Addr: 40, W: []byte{0x10}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
				//{Addr: 40, W: []byte{0x30}, R: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}},
			}...),
		}
		// Set the recorded response to have the retrigger option on.
		bus.Ops[10] = i2ctest.IO{Addr: 40, W: []byte{0x28, 0xff}, R: nil}
		opts := DefaultOpts
		// Following option needs to be true so we can get the held status.
		opts.RetriggerOnHold = true
		d, err := NewI2C(bus, &opts)
		if err != nil {
			t.Fatal(err)
		}
		// first check
		var got [8]TouchStatus
		if err := d.InputStatus(got[:]); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, [8]TouchStatus{PressedStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus}) {
			t.Fatalf("expected to have the first sensor touched but instead got %v", got)
		}
		// 2nd check
		if err = d.InputStatus(got[:]); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, [8]TouchStatus{HeldStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus, OffStatus}) {
			t.Fatalf("expected to have the first sensor touched but instead got %v", got)
		}
	})
}
