// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"fmt"
	"testing"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/videocore"
)

func ExamplePinsRead0To31() {
	// Print out the state of 32 GPIOs with a single read that reads all these
	// pins all at once.
	bits := PinsRead0To31()
	fmt.Printf("bits: %#x\n", bits)
	suffixes := []string{"   ", "\n"}
	for i := uint(0); i < 32; i++ {
		fmt.Printf("GPIO%-2d: %d%s", i, (bits>>i)&1, suffixes[(i%4)/3])
	}
	// Output:
	// bits: 0x80011010
	// GPIO0 : 0   GPIO1 : 0   GPIO2 : 0   GPIO3 : 0
	// GPIO4 : 1   GPIO5 : 0   GPIO6 : 0   GPIO7 : 0
	// GPIO8 : 0   GPIO9 : 0   GPIO10: 0   GPIO11: 0
	// GPIO12: 1   GPIO13: 0   GPIO14: 0   GPIO15: 0
	// GPIO16: 1   GPIO17: 0   GPIO18: 0   GPIO19: 0
	// GPIO20: 0   GPIO21: 0   GPIO22: 0   GPIO23: 0
	// GPIO24: 0   GPIO25: 0   GPIO26: 0   GPIO27: 0
	// GPIO28: 0   GPIO29: 0   GPIO30: 0   GPIO31: 1
}

func ExamplePinsRead32To46() {
	// Print out the state of 15 GPIOs with a single read that reads all these
	// pins all at once.
	bits := PinsRead32To46()
	fmt.Printf("bits: %#x\n", bits)
	suffixes := []string{"   ", "\n"}
	for i := uint(0); i < (47 - 32); i++ {
		fmt.Printf("GPIO%d: %d%s", i+32, (bits>>i)&1, suffixes[(i%4)/3])
	}
	// Output:
	// bits: 0x4101
	// GPIO32: 1   GPIO33: 0   GPIO34: 0   GPIO35: 0
	// GPIO36: 0   GPIO37: 0   GPIO38: 0   GPIO39: 0
	// GPIO40: 1   GPIO41: 0   GPIO42: 0   GPIO43: 0
	// GPIO44: 0   GPIO45: 0   GPIO46: 1
}

func ExamplePinsClear0To31() {
	// Simultaneously clears GPIO4 and GPIO16 to gpio.Low.
	PinsClear0To31(1<<16 | 1<<4)
}

func ExamplePinsSet0To31() {
	// Simultaneously sets GPIO4 and GPIO16 to gpio.High.
	PinsClear0To31(1<<16 | 1<<4)
}

func TestPresent(t *testing.T) {
	// It may return true or false, depending on hardware but it shouldn't crash.
	Present()
}

func TestPin(t *testing.T) {
	// Using Pin without the driver being initialized doesn't crash.
	defer resetGPIOMemory()
	gpioMemory = nil
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
	// When unitialized, Pin.function() returs alt5.
	if s := p.Function(); s != "UART1_RTS" {
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

	gpioMemory = &gpioMap{}
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
	defer func() {
		clockMemory = nil
		pwmMemory = nil
		dmaMemory = nil
	}()
	defer resetGPIOMemory()
	gpioMemory = nil

	p := Pin{name: "C1", number: 4, defaultPull: gpio.PullDown}
	if err := p.PWM(gpio.DutyHalf, 500*time.Nanosecond); err == nil || err.Error() != "bcm283x-gpio (C1): subsystem not initialized" {
		t.Fatal(err)
	}

	gpioMemory = &gpioMap{}
	if err := p.PWM(gpio.DutyHalf, 500*time.Nanosecond); err == nil || err.Error() != "bcm283x-gpio (C1): bcm283x-dma not initialized; try again as root?" {
		t.Fatal(err)
	}

	clockMemory = &clockMap{}
	pwmMemory = &pwmMap{}
	if err := p.PWM(gpio.DutyHalf, 9*time.Microsecond); err == nil || err.Error() != "bcm283x-gpio (C1): period must be at least 10µs" {
		t.Fatal(err)
	}
	// TODO(maruel): Fix test.
	dmaMemory = &dmaMap{}
	if err := p.PWM(gpio.DutyHalf, 10*time.Microsecond); err == nil || err.Error() != "bcm283x-gpio (C1): can't write to clock divisor CPU register" {
		t.Fatal(err)
	}
}

func TestDriver(t *testing.T) {
	d := driverGPIO{}
	if s := d.String(); s != "bcm283x-gpio" {
		t.Fatal(s)
	}
	if s := d.Prerequisites(); s != nil {
		t.Fatal(s)
	}
	// It will fail to initialize on non-bcm.
	_, _ = d.Init()
}

func TestSetSpeed(t *testing.T) {
	if setSpeed(1000) == nil {
		t.Fatal("cannot change live")
	}
}

func init() {
	dmaBufAllocator = func(s int) (*videocore.Mem, error) {
		buf := make([]byte, s)
		return &videocore.Mem{View: &pmem.View{Slice: buf}}, nil
	}
	// gpioMemory is initialized so the examples are more interesting.
	resetGPIOMemory()
}

// resetGPIOMemory resets so GPIO4, GPIO12, GPIO16, GPIO31, GPIO32, GPIO40 and
// GPIO46 are set.
func resetGPIOMemory() {
	gpioMemory = &gpioMap{
		level: [2]uint32{0x80011010, 0x4101},
	}
}
