// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/videocore"
)

func TestPresent(t *testing.T) {
	// It may return true or false, depending on hardware but it shouldn't crash.
	Present()
}

func TestPin(t *testing.T) {
	defer reset()
	drvGPIO.gpioMemory = nil
	// Using Pin without the driver being initialized doesn't crash.
	p := Pin{name: "Foo", number: 42, defaultPull: gpio.PullDown}

	if s := p.String(); s != "Foo" {
		t.Fatal(s)
	}
	if s := p.Name(); s != "Foo" {
		t.Fatal(s)
	}
	if n := p.Number(); n != 42 {
		t.Fatal(n)
	}
	if d := p.DefaultPull(); d != gpio.PullDown {
		t.Fatal(d)
	}
	if s := p.Function(); s != "ERR" {
		t.Fatal(s)
	}
	if p.In(gpio.PullNoChange, gpio.NoEdge) == nil {
		t.Fatal("not initialized")
	}
	if d := p.Read(); d != gpio.Low {
		t.Fatal(d)
	}
	if d := p.Pull(); d != gpio.PullNoChange {
		t.Fatal(d)
	}
	if p.WaitForEdge(-1) {
		t.Fatal("edge not initialized")
	}
	if p.Out(gpio.Low) == nil {
		t.Fatal("not initialized")
	}

	setMemory()
	if err := p.In(gpio.PullDown, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.PullUp, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.Float, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if s := p.Function(); s != "In/Low" {
		t.Fatal(s)
	}
	if d := p.Read(); d != gpio.Low {
		t.Fatal(d)
	}
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}
	if s := p.Function(); s != "Out/Low" {
		t.Fatal(s)
	}
	if err := p.Out(gpio.High); err != nil {
		t.Fatal(err)
	}

	p.number = 25
	p.setFunction(alt0)
	if s := p.Function(); s != "<Alt0>" {
		t.Fatal(s)
	}
	p.setFunction(alt1)
	if s := p.Function(); s != "<Alt1>" {
		t.Fatal(s)
	}
	p.setFunction(alt2)
	if s := p.Function(); s != "<Alt2>" {
		t.Fatal(s)
	}
	p.setFunction(alt3)
	if s := p.Function(); s != "<Alt3>" {
		t.Fatal(s)
	}
	p.setFunction(alt4)
	if s := p.Function(); s != "<Alt4>" {
		t.Fatal(s)
	}
	p.setFunction(alt5)
	if s := p.Function(); s != "<Alt5>" {
		t.Fatal(s)
	}

	p.number = 45
	p.setFunction(alt0)
	if s := p.Function(); s != "PWM1_OUT" {
		t.Fatal(s)
	}
	p.setFunction(alt1)
	if s := p.Function(); s != "I2C0_SCL" {
		t.Fatal(s)
	}
	p.setFunction(alt2)
	if s := p.Function(); s != "I2C1_SCL" {
		t.Fatal(s)
	}
	p.setFunction(alt4)
	if s := p.Function(); s != "SPI2_CS2" {
		t.Fatal(s)
	}

	p.number = 33
	p.setFunction(alt3)
	if s := p.Function(); s != "UART0_RXD" {
		t.Fatal(s)
	}
	p.setFunction(alt5)
	if s := p.Function(); s != "UART1_RXD" {
		t.Fatal(s)
	}
}

func TestPinPWM(t *testing.T) {
	// Necessary to zap out setRaw failing on non-working fake CPU memory map.
	oldClockRawError := clockRawError
	clockRawError = nil
	defer func() {
		clockRawError = oldClockRawError
	}()
	defer reset()
	setMemory()
	p := Pin{name: "C1", number: 4, defaultPull: gpio.PullDown}
	if err := p.PWM(gpio.DutyHalf, 2*physic.MegaHertz); err == nil || err.Error() != "bcm283x-gpio (C1): bcm283x-dma not initialized; try again as root?" {
		t.Fatal(err)
	}

	drvGPIO.gpioMemory = &gpioMap{}
	if err := p.PWM(gpio.DutyHalf, 2*physic.MegaHertz); err == nil || err.Error() != "bcm283x-gpio (C1): bcm283x-dma not initialized; try again as root?" {
		t.Fatal(err)
	}

	drvDMA.clockMemory = &clockMap{}
	drvDMA.pwmMemory = &pwmMap{}
	drvDMA.pwmBaseFreq = 25 * physic.MegaHertz
	drvDMA.pwmDMAFreq = 200 * physic.KiloHertz
	if err := p.PWM(gpio.DutyHalf, 110*physic.KiloHertz); err == nil || err.Error() != "bcm283x-gpio (C1): frequency must be at most 100kHz" {
		t.Fatal(err)
	}
	drvDMA.dmaMemory = &dmaMap{}
	if err := p.PWM(gpio.DutyHalf, 100*physic.KiloHertz); err != nil {
		t.Fatal(err)
	}
}

func TestDriver(t *testing.T) {
	defer reset()
	if s := drvGPIO.String(); s != "bcm283x-gpio" {
		t.Fatal(s)
	}
	if s := drvGPIO.Prerequisites(); s != nil {
		t.Fatal(s)
	}
	// It will fail to initialize on non-bcm.
	_, _ = drvGPIO.Init()
}

func TestSetSpeed(t *testing.T) {
	if setSpeed(1000) == nil {
		t.Fatal("cannot change live")
	}
}

func init() {
	reset()
}

func reset() {
	drvGPIO.Close()
	drvDMA.Close()
	// This is needed because the examples in example_test.go run in the same
	// process as this file, even if in a separate package. This means that for
	// the examples to pass, drvGPIO.gpioMemory must be set.
	setMemory()
}

// setMemory resets so GPIO4, GPIO12, GPIO16, GPIO31, GPIO32, GPIO40 and
// GPIO46 are set and mock the DMA buffer allocator.
func setMemory() {
	drvGPIO.gpioMemory = &gpioMap{
		level: [2]uint32{0x80011010, 0x4101},
	}
	drvDMA.dmaBufAllocator = func(s int) (*videocore.Mem, error) {
		buf := make([]byte, s)
		return &videocore.Mem{View: &pmem.View{Slice: buf}}, nil
	}
}
