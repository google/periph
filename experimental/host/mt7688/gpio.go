// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package mt7688

import (
	"periph.io/x/periph/conn/pin"
)

// All the pins supported by the CPU.
var (
	GPIO0  *Pin // I2S_SDI
	GPIO1  *Pin // I2S_SDO
	GPIO2  *Pin // I2S_WS
	GPIO3  *Pin // I2S_CLK
	GPIO4  *Pin // I2C_SCLK
	GPIO5  *Pin // I2C_SD
	GPIO6  *Pin // SPI_CS1
	GPIO7  *Pin // SPI_CLK
	GPIO8  *Pin // SPI_MOSI
	GPIO9  *Pin // SPI_MISO
	GPIO10 *Pin // SPI_CS0
	GPIO11 *Pin // GPIO0
	GPIO12 *Pin // UART_TXD0
	GPIO13 *Pin // UART_RXD0
	GPIO14 *Pin // MDI_TP_P1
	GPIO15 *Pin // MDI_TN_P1
	GPIO16 *Pin // MDI_RP_P1
	GPIO17 *Pin // MDI_RN_P1
	GPIO18 *Pin // MDI_RP_P2
	GPIO19 *Pin // MDI_RN_P2
	GPIO20 *Pin // MDI_TP_P2
	GPIO21 *Pin // MDI_TN_P2
	GPIO22 *Pin // MDI_TP_P3
	GPIO23 *Pin // MDI_TN_P3
	GPIO24 *Pin // MDI_RP_P3
	GPIO25 *Pin // MDI_RN_P3
	GPIO26 *Pin // MDI_RP_P4
	GPIO27 *Pin // MDI_RN_P4
	GPIO28 *Pin // MDI_TP_P4
	GPIO29 *Pin // MDI_TN_P4
	GPIO30 *Pin // EPHY_LED4_N_JTRST_N (7688KN)
	GPIO31 *Pin // EPHY_LED3_N_JTCLK (7688KN)
	GPIO32 *Pin // EPHY_LED2_N_JTMS (7688KN)
	GPIO33 *Pin // EPHY_LED1_N_JTDI (7688KN)
	GPIO34 *Pin // EPHY_LED0_N_JTDO (7688KN)
	GPIO35 *Pin // WLED_N (7688KN)
	GPIO36 *Pin // PERST_N
	GPIO37 *Pin // REF_CLKO
	GPIO38 *Pin // WDT_RST_N
	GPIO39 *Pin // EPHY_LED4_N_JTRST_N (7688AN)
	GPIO40 *Pin // EPHY_LED3_N_JTCLK (7688AN)
	GPIO41 *Pin // EPHY_LED2_N_JTMS (7688AN)
	GPIO42 *Pin // EPHY_LED1_N_JTDI (7688AN)
	GPIO43 *Pin // EPHY_LED0_N_JTDO (7688AN)
	GPIO44 *Pin // WLED_N (7688AN)
	GPIO45 *Pin // UART_TXD1
	GPIO46 *Pin // UART_RXD1
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
	{"I2SSDI", "PCMDRX"}, // 0
	{"I2SSDO", "PCMDTX"},
	{"I2SW", "PCMCLK"},
	{"I2SCLK", "PCMFS"},
	{"I2C_SCLK"},
	{"I2C_SD"}, // 5
	{"SPI_CS1", "REF_CLKO"},
	{"SPI_CLK"},
	{"SPI_MOSI"},
	{"SPI_MISO"},
	{"SPI_CS0"}, // 10
	{"GPIO#11", "REF_CLKO", "PERST_N"},
	{"UART_TXD0"},
	{"UART_RXD0"},
	{"SPIS_CS", "", "PWM_CH0"},
	{"SPIS_CLK", "", "PWM_CH1"}, // 15
	{"SPIS_MISO", "", "UART_TXD2"},
	{"SPIS_MOSI", "", "UART_RXD2"},
	{"PWM_CH0", "", "eMMC_D7"},
	{"PWM_CH1", "", "eMMC_D6"},
	{"UART_TXD2", "PWM_CH2", "eMMC_D5"}, // 20
	{"UART_RXD2", "PWM_CH3", "eMMC_D4"},
	{"SD_WP"}, // TODO: add aliases mapping eMMC
	{"SD_CD"},
	{"SD_D1"},
	{"SD_D0"}, // 25
	{"SD_CLK"},
	{"SD_CMD"},
	{"SD_D3"},
	{"SD_D2"},
	{"EPHY_LED4_K", "", "JTAG_RST_N"}, // 30
	{"EPHY_LED3_K", "", "JTAG_CLK"},
	{"EPHY_LED2_K", "", "JTAG_TMS"},
	{"EPHY_LED1_K", "", "JTAG_TDI"},
	{"EPHY_LED0_K", "", "JTAG_TDO"},
	{"WLED_N"}, // 35
	{"PERST_N"},
	{"REF_CLKO"},
	{"WDT_RST_N"},
	{"EPHY_LED4_N", "", "JTAG_RST_N"},
	{"EPHY_LED3_N", "", "JTAG_CLK"}, // 40
	{"EPHY_LED2_N", "", "JTAG_TMS"},
	{"EPHY_LED1_N", "", "JTAG_TDI"},
	{"EPHY_LED0_N", "", "JTAG_TDO"},
	{"WLED_N"},
	{"UART_TXD1", "PWM_CH0"}, // 45
	{"UART_RXD1", "PWM_CH1"},
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
