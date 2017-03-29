// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpio

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func ExamplePinIn() {
	//p := gpioreg.ByNumber(6)
	var p PinIn
	if err := p.In(PullDown, RisingEdge); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is %s\n", p, p.Read())
	for p.WaitForEdge(-1) {
		fmt.Printf("%s went %s\n", p, High)
	}
}

func ExamplePinOut() {
	//p := gpioreg.ByNumber(6)
	var p PinOut
	if err := p.Out(High); err != nil {
		log.Fatal(err)
	}
}

func TestStrings(t *testing.T) {
	if Low.String() != "Low" || High.String() != "High" {
		t.Fail()
	}
	if Float.String() != "Float" || Pull(100).String() != "Pull(100)" {
		t.Fail()
	}
	if NoEdge.String() != "NoEdge" || Edge(100).String() != "Edge(100)" {
		t.Fail()
	}
}

func TestInvalid(t *testing.T) {
	if INVALID.String() != "INVALID" || INVALID.Name() != "INVALID" || INVALID.Number() != -1 || INVALID.Function() != "" {
		t.Fail()
	}
	if INVALID.In(Float, NoEdge) != errInvalidPin || INVALID.Read() != Low || INVALID.WaitForEdge(time.Minute) || INVALID.Pull() != PullNoChange {
		t.Fail()
	}
	if INVALID.Out(Low) != errInvalidPin || INVALID.PWM(0) != errInvalidPin {
		t.Fail()
	}
}
