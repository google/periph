// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpio

import (
	"fmt"
	"testing"
	"time"

	"periph.io/x/periph/conn"
)

func TestStrings(t *testing.T) {
	data := []struct {
		t fmt.Stringer
		s string
	}{
		{Low, "Low"},
		{High, "High"},
		{PullNoChange, "PullNoChange"},
		{Float, "Float"},
		{PullDown, "PullDown"},
		{PullUp, "PullUp"},
		{Pull(100), "Pull(100)"},
		{NoEdge, "NoEdge"},
		{Edge(100), "Edge(100)"},
	}
	for i, l := range data {
		if a := l.t.String(); a != l.s {
			t.Fatalf("#%d: %s != %s", i, l.s, a)
		}
	}
}

func TestDuty_String(t *testing.T) {
	data := []struct {
		d        Duty
		expected string
	}{
		{0, "0%"},
		{1, "0%"},
		{DutyMax / 200, "0%"},
		{DutyMax/100 - 1, "1%"},
		{DutyMax / 100, "1%"},
		{DutyMax, "100%"},
		{DutyMax - 1, "100%"},
		{DutyHalf, "50%"},
		{DutyHalf + 1, "50%"},
		{DutyHalf - 1, "50%"},
		{DutyHalf + DutyMax/100, "51%"},
		{DutyHalf - DutyMax/100, "49%"},
	}
	for i, line := range data {
		if actual := line.d.String(); actual != line.expected {
			t.Fatalf("line %d: Duty(%d).String() == %q, expected %q", i, line.d, actual, line.expected)
		}
	}
}

func TestDuty_Valid(t *testing.T) {
	if !Duty(0).Valid() {
		t.Fatal("0 is valid")
	}
	if !DutyHalf.Valid() {
		t.Fatal("half is valid")
	}
	if !DutyMax.Valid() {
		t.Fatal("half is valid")
	}
	if Duty(-1).Valid() {
		t.Fatal("-1 is not valid")
	}
	if (DutyMax + 1).Valid() {
		t.Fatal("-1 is not valid")
	}
}

func TestParseDuty(t *testing.T) {
	data := []struct {
		input  string
		d      Duty
		hasErr bool
	}{
		{"", 0, true},
		{"0", 0, false},
		{"0%", 0, false},
		{"1", 1, false},
		{"1%", 655, false},
		{"100%", DutyMax, false},
		{"65535", DutyMax, false},
		{"65536", 0, true},
		{"101%", 0, true},
		{"-1", 0, true},
		{"-1%", 0, true},
	}
	for i, line := range data {
		if d, err := ParseDuty(line.input); d != line.d || (err != nil) != line.hasErr {
			t.Fatalf("line %d: Parse(%q) == %d, %q, expected %d, %t", i, line.input, d, err, line.d, line.hasErr)
		}
	}
}

func TestInvalid(t *testing.T) {
	if INVALID.String() != "INVALID" || INVALID.Name() != "INVALID" || INVALID.Number() != -1 || INVALID.Function() != "" {
		t.Fail()
	}
	if INVALID.In(Float, NoEdge) != errInvalidPin || INVALID.Read() != Low || INVALID.WaitForEdge(time.Minute) || INVALID.Pull() != PullNoChange {
		t.Fail()
	}
	if INVALID.Out(Low) != errInvalidPin {
		t.Fail()
	}
	if err := INVALID.(conn.Resource).Halt(); err != nil {
		t.Fatal(err)
	}
}
