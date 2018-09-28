// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mt7688

import (
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/host/sysfs"
)

// function specifies the active functionality of a pin. The alternative
// function is GPIO pin dependent.
type function uint8

// Each pin can have one of 5 functions.
const (
	in  function = 0
	out function = 1
	// TODO: work out how to change pin functions
	alt0 function = 2
	alt1 function = 3
	alt2 function = 4
)

// cpuPins are all the pins as supported by the CPU. There is no guarantee that
// they are actually connected to anything on the board.
var cpuPins = map[string]*Pin{
	// TODO: discover default pull
	"I2S_SDI":                      {number: 0, name: "GPIO0", defaultPull: gpio.Float},
	"I2S_SDO":                      {number: 1, name: "GPIO1", defaultPull: gpio.Float},
	"I2S_WS":                       {number: 2, name: "GPIO2", defaultPull: gpio.Float},
	"I2S_CLK":                      {number: 3, name: "GPIO3", defaultPull: gpio.Float},
	"I2C_SCLK":                     {number: 4, name: "GPIO4", defaultPull: gpio.Float},
	"I2C_SD":                       {number: 5, name: "GPIO5", defaultPull: gpio.Float},
	"SPI_CS1":                      {number: 6, name: "GPIO6", defaultPull: gpio.Float},
	"SPI_CLK":                      {number: 7, name: "GPIO7", defaultPull: gpio.Float},
	"SPI_MOSI":                     {number: 8, name: "GPIO8", defaultPull: gpio.Float},
	"SPI_MISO":                     {number: 9, name: "GPIO9", defaultPull: gpio.Float},
	"SPI_CS0":                      {number: 10, name: "GPIO10", defaultPull: gpio.Float},
	"GPIO0":                        {number: 11, name: "GPIO11", defaultPull: gpio.Float},
	"UART_TXD0":                    {number: 12, name: "GPIO12", defaultPull: gpio.Float},
	"UART_RXD0":                    {number: 13, name: "GPIO13", defaultPull: gpio.Float},
	"MDI_TP_P1":                    {number: 14, name: "GPIO14", defaultPull: gpio.Float},
	"MDI_TN_P1":                    {number: 15, name: "GPIO15", defaultPull: gpio.Float},
	"MDI_RP_P1":                    {number: 16, name: "GPIO16", defaultPull: gpio.Float},
	"MDI_RN_P1":                    {number: 17, name: "GPIO17", defaultPull: gpio.Float},
	"MDI_RP_P2":                    {number: 18, name: "GPIO18", defaultPull: gpio.Float},
	"MDI_RN_P2":                    {number: 19, name: "GPIO19", defaultPull: gpio.Float},
	"MDI_TP_P2":                    {number: 20, name: "GPIO20", defaultPull: gpio.Float},
	"MDI_TN_P2":                    {number: 21, name: "GPIO21", defaultPull: gpio.Float},
	"MDI_TP_P3":                    {number: 22, name: "GPIO22", defaultPull: gpio.Float},
	"MDI_TN_P3":                    {number: 23, name: "GPIO23", defaultPull: gpio.Float},
	"MDI_RP_P3":                    {number: 24, name: "GPIO24", defaultPull: gpio.Float},
	"MDI_RN_P3":                    {number: 25, name: "GPIO25", defaultPull: gpio.Float},
	"MDI_RP_P4":                    {number: 26, name: "GPIO26", defaultPull: gpio.Float},
	"MDI_RN_P4":                    {number: 27, name: "GPIO27", defaultPull: gpio.Float},
	"MDI_TP_P4":                    {number: 28, name: "GPIO28", defaultPull: gpio.Float},
	"MDI_TN_P4":                    {number: 29, name: "GPIO29", defaultPull: gpio.Float},
	"EPHY_LED4_N_JTRST_N (7688KN)": {number: 30, name: "GPIO30", defaultPull: gpio.Float},
	"EPHY_LED3_N_JTCLK (7688KN)":   {number: 31, name: "GPIO31", defaultPull: gpio.Float},
	"EPHY_LED2_N_JTMS (7688KN)":    {number: 32, name: "GPIO32", defaultPull: gpio.Float},
	"EPHY_LED1_N_JTDI (7688KN)":    {number: 33, name: "GPIO33", defaultPull: gpio.Float},
	"EPHY_LED0_N_JTDO (7688KN)":    {number: 34, name: "GPIO34", defaultPull: gpio.Float},
	"WLED_N (7688KN)":              {number: 35, name: "GPIO35", defaultPull: gpio.Float},
	"PERST_N":                      {number: 36, name: "GPIO36", defaultPull: gpio.Float},
	"REF_CLKO":                     {number: 37, name: "GPIO37", defaultPull: gpio.Float},
	"WDT_RST_N":                    {number: 38, name: "GPIO38", defaultPull: gpio.Float},
	"EPHY_LED4_N_JTRST_N (7688AN)": {number: 39, name: "GPIO39", defaultPull: gpio.Float},
	"EPHY_LED3_N_JTCLK (7688AN)":   {number: 40, name: "GPIO40", defaultPull: gpio.Float},
	"EPHY_LED2_N_JTMS (7688AN)":    {number: 41, name: "GPIO41", defaultPull: gpio.Float},
	"EPHY_LED1_N_JTDI (7688AN)":    {number: 42, name: "GPIO42", defaultPull: gpio.Float},
	"EPHY_LED0_N_JTDO (7688AN)":    {number: 43, name: "GPIO43", defaultPull: gpio.Float},
	"WLED_N (7688AN)":              {number: 44, name: "GPIO44", defaultPull: gpio.Float},
	"UART_TXD1":                    {number: 45, name: "GPIO45", defaultPull: gpio.Float},
	"UART_RXD1":                    {number: 46, name: "GPIO46", defaultPull: gpio.Float},
}

// initPins initializes and configures pins as required.
func initPins() {
	for name := range mappingMT7688 {
		cpuPin := cpuPins[name]

		// Initialize sysfs access right away.
		cpuPin.sysfsPin = sysfs.Pins[cpuPin.number]
	}
}

type Pin struct {
	// Immutable.
	number      int
	name        string
	defaultPull gpio.Pull // Default pull at system boot, as per datasheet.

	// Immutable after driver initialization.
	sysfsPin *sysfs.Pin // Set to the corresponding sysfs.Pin, if any.
}

// String implements conn.Resource.
func (p Pin) String() string {
	return p.name
}

// Halt implements conn.Resource.
func (*Pin) Halt() error {
	return nil
}

// Name implements pin.Pin.
func (p *Pin) Name() string {
	return p.name
}

// Number implements pin.Pin.
//
// This is the GPIO number, not the pin number on a header.
func (p *Pin) Number() int {
	return p.number
}

// Function implements pin.Pin.
func (p *Pin) Function() string {
	return string(p.Func())
}

// Func implements pin.PinFunc.
func (p *Pin) Func() pin.Func {
	if drvGPIO.gpioMemory == nil {
		if p.sysfsPin == nil {
			return pin.Func("ERR")
		}
		return p.sysfsPin.Func()
	}
	switch f := p.function(); f {
	case in:
		// TODO: implement FastRead
		//if p.FastRead() {
		//	return gpio.IN_HIGH
		//}
		return gpio.IN_LOW
	case out:
		//if p.FastRead() {
		//	return gpio.OUT_HIGH
		//}
		return gpio.OUT_LOW
	case alt0:
		if s := mappingMT7688[p.name][0]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT0")
	case alt1:
		if s := mappingMT7688[p.name][1]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT1")
	case alt2:
		if s := mappingMT7688[p.name][2]; len(s) != 0 {
			return s
		}
		return pin.Func("ALT2")
	default:
		return pin.Func("ERR")
	}
}

// function returns the current GPIO pin function.
func (p *Pin) function() function {
	// TODO: implement function
	return out
}

var _ conn.Resource = &Pin{}
var _ pin.Pin = &Pin{}

// TODO: implement required interfaces
//var _ gpio.PinIO = &Pin{}
//var _ gpio.PinIn = &Pin{}
//var _ gpio.PinOut = &Pin{}
