// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains pin mapping information that is specific to the Allwinner
// A64 model.

package allwinner

import (
	"strings"

	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/host/sysfs"
)

// A64 specific pins.
var (
	X32KFOUT *pin.BasicPin // Clock output of 32Khz crystal
	KEY_ADC  *pin.BasicPin // 6 bits resolution ADC for key application; can work up to 250Hz conversion rate; reference voltage is 2.0V
	EAROUTP  *pin.BasicPin // Earpiece amplifier negative differential output
	EAROUTN  *pin.BasicPin // Earpiece amplifier positive differential output
)

//

func init() {
	X32KFOUT = &pin.BasicPin{N: "X32KFOUT"}
	// BUG(maruel): These need to be converted to an analog.PinIO implementation
	// once analog support is implemented.
	KEY_ADC = &pin.BasicPin{N: "KEY_ADC"}
	EAROUTP = &pin.BasicPin{N: "EAROUTP"}
	EAROUTN = &pin.BasicPin{N: "EAROUTN"}
}

// mappingA64 describes the mapping of the A64 processor gpios to their
// alternate functions.
//
// It omits the in & out functions which are available on all gpio.
//
// The mapping comes from the datasheet page 23:
// http://files.pine64.org/doc/datasheet/pine64/A64_Datasheet_V1.1.pdf
//
// - The datasheet uses TWI instead of I2C but it is renamed here for
//   consistency.
// - AIF is an audio interface, i.e. to connect to S/PDIF.
// - RGMII means Reduced gigabit media-independent interface.
// - SDC means SDCard?
// - NAND connects to a NAND flash controller.
// - CSI and CCI are for video capture.
var mappingA64 = map[string][5]pin.Func{
	"PB0":  {"UART2_TX", "", "JTAG0_TMS", "", "PB_EINT0"},
	"PB1":  {"UART2_RX", "", "JTAG0_TCK", "SIM_PWREN", "PB_EINT1"},
	"PB2":  {"UART2_RTS", "", "JTAG0_TDO", "SIM_VPPEN", "PB_EINT2"},
	"PB3":  {"UART2_CTS", "I2S0_MCLK", "JTAG0_TDI", "SIM_VPPPP", "PB_EINT3"},
	"PB4":  {"AIF2_SYNC", "I2S0_WS", "", "SIM_CLK", "PB_EINT4"},
	"PB5":  {"AIF2_BCLK", "I2S0_SCK", "", "SIM_DATA", "PB_EINT5"},
	"PB6":  {"AIF2_DOUT", "I2S0_DOUT", "", "SIM_RST", "PB_EINT6"},
	"PB7":  {"AIF2_DIN", "I2S0_DIN", "", "SIM_DET", "PB_EINT7"},
	"PB8":  {"", "", "UART0_TX", "", "PB_EINT8"},
	"PB9":  {"", "", "UART0_RX", "", "PB_EINT9"},
	"PC0":  {"NAND_WE", "", "SPI0_MOSI"},
	"PC1":  {"NAND_ALE", "SDC2_DS", "SPI0_MISO"},
	"PC2":  {"NAND_CLE", "", "SPI0_CLK"},
	"PC3":  {"NAND_CE1", "", "SPI0_CS0"},
	"PC4":  {"NAND_CE0"},
	"PC5":  {"NAND_RE", "SDC2_CLK"},
	"PC6":  {"NAND_RB0", "SDC2_CMD"},
	"PC7":  {"NAND_RB1"},
	"PC8":  {"NAND_DQ0", "SDC2_D0"},
	"PC9":  {"NAND_DQ1", "SDC2_D1"},
	"PC10": {"NAND_DQ2", "SDC2_D2"},
	"PC11": {"NAND_DQ3", "SDC2_D3"},
	"PC12": {"NAND_DQ4", "SDC2_D4"},
	"PC13": {"NAND_DQ5", "SDC2_D5"},
	"PC14": {"NAND_DQ6", "SDC2_D6"},
	"PC15": {"NAND_DQ7", "SDC2_D7"},
	"PC16": {"NAND_DQS", "SDC2_RST"},
	"PD0":  {"LCD_D2", "UART3_TX", "SPI1_CS0", "CCIR_CLK"},
	"PD1":  {"LCD_D3", "UART3_RX", "SPI1_CLK", "CCIR_DE"},
	"PD2":  {"LCD_D4", "UART4_TX", "SPI1_MOSI", "CCIR_HSYNC"},
	"PD3":  {"LCD_D5", "UART4_RX", "SPI1_MISO", "CCIR_VSYNC"},
	"PD4":  {"LCD_D6", "UART4_RTS", "", "CCIR_D0"},
	"PD5":  {"LCD_D7", "UART4_CTS", "", "CCIR_D1"},
	"PD6":  {"LCD_D10", "", "", "CCIR_D2"},
	"PD7":  {"LCD_D11", "", "", "CCIR_D3"},
	"PD8":  {"LCD_D12", "", "RGMII_RXD3", "CCIR_D4"},
	"PD9":  {"LCD_D13", "", "RGMII_RXD2", "CCIR_D5"},
	"PD10": {"LCD_D14", "", "RGMII_RXD1"},
	"PD11": {"LCD_D15", "", "RGMII_RXD0"},
	"PD12": {"LCD_D18", "LVDS_VP0", "RGMII_RXCK"},
	"PD13": {"LCD_D19", "LVDS_VN0", "RGMII_RXCT"},
	"PD14": {"LCD_D20", "LVDS_VP1", "RGMII_RXER"},
	"PD15": {"LCD_D21", "LVDS_VN1", "RGMII_TXD3", "CCIR_D6"},
	"PD16": {"LCD_D22", "LVDS_VP2", "RGMII_TXD2", "CCIR_D7"},
	"PD17": {"LCD_D23", "LVDS_VN2", "RGMII_TXD1"},
	"PD18": {"LCD_CLK", "LVDS_VPC", "RGMII_TXD0"},
	"PD19": {"LCD_DE", "LVDS_VNC", "RGMII_TXCK"},
	"PD20": {"LCD_HSYNC", "LVDS_VP3", "RGMII_TXCT"},
	"PD21": {"LCD_VSYNC", "LVDS_VN3", "RGMII_CLKI"},
	"PD22": {"PWM0", "", "MDC"},
	"PD23": {"", "", "MDIO"},
	"PD24": {""},
	"PE0":  {"CSI_PCLK", "", "TS_CLK"},
	"PE1":  {"CSI_MCLK", "", "TS_ERR"},
	"PE2":  {"CSI_HSYNC", "", "TS_SYNC"},
	"PE3":  {"CSI_VSYNC", "", "TS_DVLD"},
	"PE4":  {"CSI_D0", "", "TS_D0"},
	"PE5":  {"CSI_D1", "", "TS_D1"},
	"PE6":  {"CSI_D2", "", "TS_D2"},
	"PE7":  {"CSI_D3", "", "TS_D3"},
	"PE8":  {"CSI_D4", "", "TS_D4"},
	"PE9":  {"CSI_D5", "", "TS_D5"},
	"PE10": {"CSI_D6", "", "TS_D6"},
	"PE11": {"CSI_D7", "", "TS_D7"},
	"PE12": {"CSI_SCK"},
	"PE13": {"CSI_SDA"},
	"PE14": {"PLL_LOCK_DBG", "I2C2_SCL"},
	"PE15": {"", "I2C2_SDA"},
	"PE16": {""},
	"PE17": {""},
	"PF0":  {"SDC0_D1", "JTAG1_TMS"},
	"PF1":  {"SDC0_D0", "JTAG1_TDI"},
	"PF2":  {"SDC0_CLK", "UART0_TX"},
	"PF3":  {"SDC0_CMD", "JTAG1_TDO"},
	"PF4":  {"SDC0_D3", "UART0_RX"},
	"PF5":  {"SDC0_D2", "JTAG1_TCK"},
	"PF6":  {""},
	"PG0":  {"SDC1_CLK", "", "", "", "PG_EINT0"},
	"PG1":  {"SDC1_CMD", "", "", "", "PG_EINT1"},
	"PG2":  {"SDC1_D0", "", "", "", "PG_EINT2"},
	"PG3":  {"SDC1_D1", "", "", "", "PG_EINT3"},
	"PG4":  {"SDC1_D2", "", "", "", "PG_EINT4"},
	"PG5":  {"SDC1_D3", "", "", "", "PG_EINT5"},
	"PG6":  {"UART1_TX", "", "", "", "PG_EINT6"},
	"PG7":  {"UART1_RX", "", "", "", "PG_EINT7"},
	"PG8":  {"UART1_RTS", "", "", "", "PG_EINT8"},
	"PG9":  {"UART1_CTS", "", "", "", "PG_EINT9"},
	"PG10": {"AIF3_SYNC", "I2S1_WS", "", "", "PG_EINT10"},
	"PG11": {"AIF3_BCLK", "I2S1_SCK", "", "", "PG_EINT11"},
	"PG12": {"AIF3_DOUT", "I2S1_DOUT", "", "", "PG_EINT12"},
	"PG13": {"AIF3_DIN", "I2S1_DIN", "", "", "PG_EINT13"},
	"PH0":  {"I2C0_SCL", "", "", "", "PH_EINT0"},
	"PH1":  {"I2C0_SDA", "", "", "", "PH_EINT1"},
	"PH2":  {"I2C1_SCL", "", "", "", "PH_EINT2"},
	"PH3":  {"I2C1_SDA", "", "", "", "PH_EINT3"},
	"PH4":  {"UART3_TX", "", "", "", "PH_EINT4"},
	"PH5":  {"UART3_RX", "", "", "", "PH_EINT5"},
	"PH6":  {"UART3_RTS", "", "", "", "PH_EINT6"},
	"PH7":  {"UART3_CTS", "", "", "", "PH_EINT7"},
	"PH8":  {"OWA_OUT", "", "", "", "PH_EINT8"},
	"PH9":  {"", "", "", "", "PH_EINT9"},
	"PH10": {"MIC_CLK", "", "", "", "PH_EINT10"},
	"PH11": {"MIC_DATA", "", "", "", "PH_EINT11"},
}

// mapA64Pins uses mappingA64 to actually set the altFunc fields of all gpio
// and mark them as available.
//
// It is called by the generic allwinner processor code if an A64 is detected.
func mapA64Pins() error {
	for name, altFuncs := range mappingA64 {
		pin := cpupins[name]
		pin.altFunc = altFuncs
		pin.available = true
		if strings.Contains(string(altFuncs[4]), "_EINT") {
			pin.supportEdge = true
		}

		// Initializes the sysfs corresponding pin right away.
		pin.sysfsPin = sysfs.Pins[pin.Number()]
	}
	return nil
}
