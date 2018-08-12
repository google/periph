// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"testing"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/videocore"
)

func TestPresent(t *testing.T) {
	// It may return true or false, depending on hardware but it shouldn't crash.
	Present()
}

func TestPins(t *testing.T) {
	defer reset()
	if v := PinsRead0To31(); v != 0x80011010 {
		t.Fatal(v)
	}
	PinsClear0To31(1)
	PinsSet0To31(1)
	if v := PinsRead32To46(); v != 0x4101 {
		t.Fatal(v)
	}
	PinsClear32To46(1)
	PinsSet32To46(1)
	PinsSetup0To27(0, true, true)
	PinsSetup28To45(0, true, true)
	drvDMA.gpioPadMemory = nil
	PinsSetup0To27(0, true, true)
	PinsSetup28To45(0, true, true)
}

func TestPin_NoMem(t *testing.T) {
	defer reset()
	drvGPIO.gpioMemory = nil
	drvDMA.gpioPadMemory = nil
	// Using Pin without the driver being initialized doesn't crash.
	p := Pin{name: "Foo", number: 42, defaultPull: gpio.PullDown}

	// conn.Resource
	if s := p.String(); s != "Foo" {
		t.Fatal(s)
	}

	// pin.Pin
	if s := p.Name(); s != "Foo" {
		t.Fatal(s)
	}
	if n := p.Number(); n != 42 {
		t.Fatal(n)
	}
	if s := p.Function(); s != "ERR" {
		t.Fatal(s)
	}

	// gpio.PinIn
	if p.In(gpio.PullNoChange, gpio.NoEdge) == nil {
		t.Fatal("not initialized")
	}
	if d := p.Read(); d != gpio.Low {
		t.Fatal(d)
	}
	if d := p.Pull(); d != gpio.PullNoChange {
		t.Fatal(d)
	}
	if d := p.DefaultPull(); d != gpio.PullDown {
		t.Fatal(d)
	}
	if p.WaitForEdge(-1) {
		t.Fatal("edge not initialized")
	}

	// gpio.PinOut
	if p.Out(gpio.Low) == nil {
		t.Fatal("not initialized")
	}

	if v := p.Drive(); v != 0 {
		t.Fatal(v)
	}
	if !p.SlewLimit() {
		t.Fatal("oops")
	}
	if !p.Hysteresis() {
		t.Fatal("oops")
	}
}

func TestPin(t *testing.T) {
	p := Pin{name: "Foo", number: 42, defaultPull: gpio.PullDown}
	// pin.Pin
	if s := p.Function(); s != "In/Low" {
		t.Fatal(s)
	}

	// gpio.PinIn
	if err := p.In(gpio.PullDown, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.PullUp, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if err := p.In(gpio.Float, gpio.NoEdge); err != nil {
		t.Fatal(err)
	}
	if d := p.Read(); d != gpio.Low {
		t.Fatal(d)
	}

	// gpio.PinOut
	if err := p.Out(gpio.Low); err != nil {
		t.Fatal(err)
	}
	if s := p.Function(); s != "Out/Low" {
		t.Fatal(s)
	}
	if err := p.Out(gpio.High); err != nil {
		t.Fatal(err)
	}

	// Above 27.
	for i := 0; i < 8; i++ {
		drvDMA.gpioPadMemory.pads1 = pad(i)
		if v := p.Drive(); v != physic.ElectricCurrent(2*(i+1))*physic.MilliAmpere {
			t.Fatal(v)
		}
	}
	drvDMA.gpioPadMemory.pads1 = 1
	if v := p.Drive(); v != 4*physic.MilliAmpere {
		t.Fatal(v)
	}
	if !p.SlewLimit() {
		t.Fatal("oops")
	}
	if p.Hysteresis() {
		t.Fatal("oops")
	}
}

func TestPin_SetFunc_25(t *testing.T) {
	p := Pin{name: "Foo", number: 25, defaultPull: gpio.PullDown}
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

	// Below 28.
	if v := p.Drive(); v != 2*physic.MilliAmpere {
		t.Fatal(v)
	}
	if !p.SlewLimit() {
		t.Fatal("oops")
	}
	if p.Hysteresis() {
		t.Fatal("oops")
	}
}

func TestPin_SetFunc_33(t *testing.T) {
	p := Pin{name: "Foo", number: 33, defaultPull: gpio.PullDown}
	p.setFunction(alt3)
	if s := p.Function(); s != "UART0_RX" {
		t.Fatal(s)
	}
	p.setFunction(alt5)
	if s := p.Function(); s != "UART1_RX" {
		t.Fatal(s)
	}
}

func TestPin_SetFunc_45(t *testing.T) {
	p := Pin{name: "Foo", number: 45, defaultPull: gpio.PullDown}
	p.setFunction(alt0)
	if s := p.Function(); s != "PWM1" {
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
}

func TestPinPWM(t *testing.T) {
	defer reset()
	// Necessary to zap out setRaw failing on non-working fake CPU memory map.
	oldErrClockRegister := errClockRegister
	errClockRegister = nil
	defer func() {
		errClockRegister = oldErrClockRegister
	}()
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

func TestPinStreamIn(t *testing.T) {
	defer reset()
	p := Pin{name: "C1", number: 4, defaultPull: gpio.PullDown}
	if err := p.StreamIn(gpio.PullDown, nil); err.Error() != "bcm283x: other Stream than BitStream are not implemented yet" {
		t.Fatal(err)
	}
	if err := p.StreamIn(gpio.PullDown, &gpiostream.BitStream{}); err.Error() != "bcm283x: MSBF BitStream is not implemented yet" {
		t.Fatal(err)
	}
	if err := p.StreamIn(gpio.PullDown, &gpiostream.BitStream{LSBF: true}); err.Error() != "bcm283x: can't read to empty BitStream" {
		t.Fatal(err)
	}
	if err := p.StreamIn(gpio.PullDown, &gpiostream.BitStream{Bits: make([]byte, 1), Freq: physic.KiloHertz, LSBF: true}); err.Error() != "bcm283x-gpio (C1): frequency is too high(1kHz)" {
		t.Fatal(err)
	}
	drvGPIO.gpioMemory = nil
	if err := p.StreamIn(gpio.PullDown, &gpiostream.BitStream{Bits: make([]byte, 1), Freq: physic.KiloHertz, LSBF: true}); err.Error() != "bcm283x-gpio (C1): subsystem gpiomem not initialized" {
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
	drvDMA.gpioPadMemory = &gpioPadMap{}
}
