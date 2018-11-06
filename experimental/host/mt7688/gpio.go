// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mt7688

import (
	"periph.io/x/periph/conn/pin"
)

// All the pins supported by the CPU.
var (
	GPIO0  *Pin // I2S_DIN
	GPIO1  *Pin // I2S_DOUT
	GPIO2  *Pin // I2S_WS
	GPIO3  *Pin // I2S_SCK
	GPIO4  *Pin // I2C_SCL
	GPIO5  *Pin // I2C_SDA
	GPIO6  *Pin // SPI_CS1, CLK0
	GPIO7  *Pin // SPI_CLK
	GPIO8  *Pin // SPI_MOSI
	GPIO9  *Pin // SPI_MISO
	GPIO10 *Pin // SPI_CS0
	GPIO11 *Pin // CLK0
	GPIO12 *Pin // UART0_TX
	GPIO13 *Pin // UART0_RX
	GPIO14 *Pin // PWM0
	GPIO15 *Pin // PWM1
	GPIO16 *Pin // UART2_TX
	GPIO17 *Pin // UART2_RX
	GPIO18 *Pin // PWM0
	GPIO19 *Pin // PWM1
	GPIO20 *Pin // UART2_TX, PWM2
	GPIO21 *Pin // UART2_RX, PWM3
	GPIO22 *Pin
	GPIO23 *Pin
	GPIO24 *Pin
	GPIO25 *Pin
	GPIO26 *Pin
	GPIO27 *Pin
	GPIO28 *Pin
	GPIO29 *Pin
	GPIO30 *Pin // JTAG_TRST
	GPIO31 *Pin // JTAG_TCK
	GPIO32 *Pin // JTAG_TMS
	GPIO33 *Pin // JTAG_TDI
	GPIO34 *Pin // JTAG_TDO
	GPIO35 *Pin
	GPIO36 *Pin
	GPIO37 *Pin // CLKO
	GPIO38 *Pin
	GPIO39 *Pin // JTAG_TRST
	GPIO40 *Pin // JTAG_TCK
	GPIO41 *Pin // JTAG_TMS
	GPIO42 *Pin // JTAG_TDI
	GPIO43 *Pin // JTAG_TDO
	GPIO44 *Pin
	GPIO45 *Pin // UART1_TX, PWM0
	GPIO46 *Pin // UART1_RX, PWM1
)

// mappingMT7688 describes the mapping of the MT7688 processor gpios to their
// alternate functions.
//
// It omits the in & out functions which are available on all gpio.
//
// The mapping is a combination of the naming from datasheet pages 25-31 and
// the GPIO Pin Function Mapping on page 108:
// https://labs.mediatek.com/fileMedia/download/9ef51e98-49b1-489a-b27e-391bac9f7bf3
var mapping = [][3]pin.Func{
	{"I2S_DIN"}, // 0
	{"I2S_DOUT"},
	{"I2S_WS"},
	{"I2S_SCK"},
	{"I2C_SCL"},
	{"I2C_SDA"}, // 5
	{"SPI_CS1", "", "CLK0"},
	{"SPI_CLK"},
	{"SPI_MOSI"},
	{"SPI_MISO"},
	{"SPI_CS0"}, // 10
	{"", "CLK0"},
	{"UART0_TX"},
	{"UART0_RX"},
	{"", "", "PWM0"},
	{"", "", "PWM1"}, // 15
	{"", "", "UART2_TX"},
	{"", "", "UART2_RX"},
	{"PWM0"},
	{"PWM1"},
	{"UART2_TX", "PWM2"}, // 20
	{"UART2_RX", "PWM3"},
	{""},
	{""},
	{""},
	{""}, // 25
	{""},
	{""},
	{""},
	{""},
	{"", "", "JTAG_TRST"}, // 30
	{"", "", "JTAG_TCK"},
	{"", "", "JTAG_TMS"},
	{"", "", "JTAG_TDI"},
	{"", "", "JTAG_TDO"},
	{""}, // 35
	{""},
	{"CLK0"},
	{""},
	{"", "", "JTAG_TRST"},
	{"", "", "JTAG_TCK"}, // 40
	{"", "", "JTAG_TMS"},
	{"", "", "JTAG_TDI"},
	{"", "", "JTAG_TDO"},
	{""},
	{"UART1_TX", "PWM0"}, // 45
	{"UART1_RX", "PWM1"},
}

func init() {
	GPIO0 = &cpuPins[0]
	GPIO1 = &cpuPins[1]
	GPIO2 = &cpuPins[2]
	GPIO3 = &cpuPins[3]
	GPIO4 = &cpuPins[4]
	GPIO5 = &cpuPins[5]
	GPIO6 = &cpuPins[6]
	GPIO7 = &cpuPins[7]
	GPIO8 = &cpuPins[8]
	GPIO9 = &cpuPins[9]
	GPIO10 = &cpuPins[10]
	GPIO11 = &cpuPins[11]
	GPIO12 = &cpuPins[12]
	GPIO13 = &cpuPins[13]
	GPIO14 = &cpuPins[14]
	GPIO15 = &cpuPins[15]
	GPIO16 = &cpuPins[16]
	GPIO17 = &cpuPins[17]
	GPIO18 = &cpuPins[18]
	GPIO19 = &cpuPins[19]
	GPIO20 = &cpuPins[20]
	GPIO21 = &cpuPins[21]
	GPIO22 = &cpuPins[22]
	GPIO23 = &cpuPins[23]
	GPIO24 = &cpuPins[24]
	GPIO25 = &cpuPins[25]
	GPIO26 = &cpuPins[26]
	GPIO27 = &cpuPins[27]
	GPIO28 = &cpuPins[28]
	GPIO29 = &cpuPins[29]
	GPIO30 = &cpuPins[30]
	GPIO31 = &cpuPins[31]
	GPIO32 = &cpuPins[32]
	GPIO33 = &cpuPins[33]
	GPIO34 = &cpuPins[34]
	GPIO35 = &cpuPins[35]
	GPIO36 = &cpuPins[36]
	GPIO37 = &cpuPins[37]
	GPIO38 = &cpuPins[38]
	GPIO39 = &cpuPins[39]
	GPIO40 = &cpuPins[40]
	GPIO41 = &cpuPins[41]
	GPIO42 = &cpuPins[42]
	GPIO43 = &cpuPins[43]
	GPIO44 = &cpuPins[44]
	GPIO45 = &cpuPins[45]
	GPIO46 = &cpuPins[46]
}

// Mapping as
// https://labs.mediatek.com/fileMedia/download/9ef51e98-49b1-489a-b27e-391bac9f7bf3
// pages 109-110.
type gpioMap struct {
	// 0x00    RW    Direction control register (GPIO0-31)
	// 0x04    RW    Direction control register (GPIO32-63)
	// 0x08    RW    Direction control register (GPIO64-95)
	directionControl [3]uint32 // GPIO_CTRL_0~GPIO_CTRL_2
	// 0x10    RW    Polarity control register (GPIO0-31)
	// 0x14    RW    Polarity control register (GPIO32-63)
	// 0x18    RW    Polarity control register (GPIO64-95)
	polarityControl [3]uint32 // GPIO_POL_0~GPIO_POL_2
	// 0x20    RW    Data register (GPIO0-31)
	// 0x24    RW    Data register (GPIO32-63)
	// 0x28    RW    Data register (GPIO64-95)
	data [3]uint32 // GPIO_DATA_0~GPIO_DATA_2
	// 0x30    WO    Data set register (GPIO0-31)
	// 0x34    WO    Data set register (GPIO32-63)
	// 0x38    WO    Data set register (GPIO64-95)
	dataSet [3]uint32 // GPIO_DSET_0~GPIO_DSET_2
	// 0x40    WO    Data clear register (GPIO0-31)
	// 0x44    WO    Data clear register (GPIO32-63)
	// 0x48    WO    Data clear register (GPIO64-95)
	dataClear [3]uint32 // GPIO_DCLR_0~GPIO_DCLR_2
	// 0x50    RW    Rising edge interrupt enable register (GPIO0-31)
	// 0x54    RW    Rising edge interrupt enable register (GPIO32-63)
	// 0x58    RW    Rising edge interrupt enable register (GPIO64-95)
	risingEdgeIrqEnable [3]uint32 // GINT_REDGE_0~GINT_REDGE_2
	// 0x60    RW    Falling edge interrupt enable register (GPIO0-31)
	// 0x64    RW    Falling edge interrupt enable register (GPIO32-63)
	// 0x68    RW    Falling edge interrupt enable register (GPIO64-95)
	fallingEdgeIrqEnable [3]uint32 // GINT_FEDGE_0~GINT_FEDGE_2
	// 0x70    RW    High level interrupt enable register (GPIO0-31)
	// 0x74    RW    High level interrupt enable register (GPIO32-63)
	// 0x78    RW    High level interrupt enable register (GPIO64-95)
	highLevelIrqEnable [3]uint32 // GINT_HLVL_0~GINT_HLVL_2
	// 0x80    RW    Low level interrupt enable register (GPIO0-31)
	// 0x84    RW    Low level interrupt enable register (GPIO32-63)
	// 0x88    RW    Low level interrupt enable register (GPIO64-95)
	lowLevelIrqEnable [3]uint32 // GINT_LLVL_0~GINT_LLVL_2
	// 0x90    W1C   Interrupt status register (GPIO0-31)
	// 0x94    W1C   Interrupt status register (GPIO32-63)
	// 0x98    W1C   Interrupt status register (GPIO64-95)
	irqStatus [3]uint32 // GINT_STAT_0~GINT_STAT_2
	// 0xA0    W1C   Edge status register (GPIO0-31)
	// 0xA4    W1C   Edge status register (GPIO32-63)
	// 0xA8    W1C   Edge status register (GPIO64-95)
	edgeStatus [3]uint32 // GINT_EDGE_0~GINT_EDGE_2
}

// TODO: via docs: Unless specified explicitly, all the GPIO pins are in input mode after reset.
