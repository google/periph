// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"sync"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiotest"
	"periph.io/x/periph/conn/physic"
)

func TestAssumption(t *testing.T) {
	f := gpiotest.Pin{}
	if f.In(gpio.PullNoChange, gpio.BothEdges) == nil {
		t.Fatal("Using gpiotest.Pin in no edge support mode")
	}
	if PollEdge(&f, 20*physic.Hertz) == nil {
		t.Fatal("expected error")
	}
}

func TestPollEdge_Short(t *testing.T) {
	p := PollEdge(&gpiotest.Pin{}, physic.Hertz)
	if err := p.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		t.Fatal(err)
	}
	if err := p.Halt(); err != nil {
		t.Fatal(err)
	}
	// timeout triggers.
	if p.WaitForEdge(time.Nanosecond) {
		t.Fatal("unexpected edge")
	}
}

func TestPollEdge_Halt(t *testing.T) {
	f := pinWait{wait: make(chan struct{})}
	p := PollEdge(&f, physic.Hertz)
	go func() {
		// Make sure the pin was read at least once, which means the code below is
		// inside WaitForEdge().
		<-f.wait
		if err := p.Halt(); err != nil {
			t.Error(err)
		}
	}()
	// p.die triggers.
	if p.WaitForEdge(-1) {
		t.Fatal("unexpected edge")
	}
}

func TestPollEdge_RisingEdge(t *testing.T) {
	f := pinLevels{levels: []gpio.Level{gpio.High, gpio.Low, gpio.High}}
	p := PollEdge(&f, physic.KiloHertz)
	if err := p.In(gpio.PullNoChange, gpio.RisingEdge); err != nil {
		t.Fatal(err)
	}
	if !p.WaitForEdge(-1) {
		t.Fatal("expected edge")
	}
	if len(f.levels) != 0 {
		t.Fatalf("unconsumed levels: %v", f.levels)
	}
}

func TestPollEdge_FallingEdge(t *testing.T) {
	f := pinLevels{levels: []gpio.Level{gpio.Low, gpio.High, gpio.Low}}
	p := PollEdge(&f, physic.KiloHertz)
	if err := p.In(gpio.PullNoChange, gpio.FallingEdge); err != nil {
		t.Fatal(err)
	}
	if !p.WaitForEdge(-1) {
		t.Fatal("expected edge")
	}
	if len(f.levels) != 0 {
		t.Fatalf("unconsumed levels: %v", f.levels)
	}
}

func TestPollEdge_BothEdges(t *testing.T) {
	f := pinLevels{levels: []gpio.Level{gpio.High, gpio.Low}}
	p := PollEdge(&f, physic.KiloHertz)
	if err := p.In(gpio.PullNoChange, gpio.BothEdges); err != nil {
		t.Fatal(err)
	}
	if !p.WaitForEdge(-1) {
		t.Fatal("expected edge")
	}
	if len(f.levels) != 0 {
		t.Fatal("unconsumed level")
	}
}

func TestPollEdge_RealPin(t *testing.T) {
	f := gpiotest.Pin{}
	p := PollEdge(&f, physic.Hertz)
	r, ok := p.(gpio.RealPin)
	if !ok {
		t.Fatal("expected gpio.RealPin")
	}
	a, ok := r.Real().(*gpiotest.Pin)
	if !ok {
		t.Fatal("expected gpiotest.Pin")
	}
	if a != &f {
		t.Fatal("expected actual pin")
	}
}

func TestPollEdge_RealPin_Deep(t *testing.T) {
	f := gpiotest.Pin{}
	p := PollEdge(PollEdge(&f, physic.Hertz), physic.Hertz)
	r, ok := p.(gpio.RealPin)
	if !ok {
		t.Fatal("expected gpio.RealPin")
	}
	a, ok := r.Real().(*gpiotest.Pin)
	if !ok {
		t.Fatal("expected gpiotest.Pin")
	}
	if a != &f {
		t.Fatal("expected actual pin")
	}
}

//

type pinLevels struct {
	gpiotest.Pin
	mu     sync.Mutex
	levels []gpio.Level
}

func (p *pinLevels) Read() gpio.Level {
	p.mu.Lock()
	defer p.mu.Unlock()
	l := p.levels[0]
	p.levels = p.levels[1:]
	return l
}

type pinWait struct {
	gpiotest.Pin
	wait chan struct{}
	once sync.Once
}

func (p *pinWait) Read() gpio.Level {
	p.once.Do(func() {
		p.wait <- struct{}{}
	})
	return true
}
