// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains pin mapping information that is specific to the Allwinner
// A20 model.

package allwinner

import (
	"strings"

	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/host/sysfs"
)

// mappingA20 describes the mapping of the A20 processor gpios to their
// alternate functions.
//
// It omits the in & out functions which are available on all gpio.
//
// The mapping comes from the datasheet page 241:
// http://dl.linux-sunxi.org/A20/A20%20User%20Manual%202013-03-22.pdf
var mappingA20 = map[string][5]pin.Func{
	"PA0":  {"ERXD3", "SPI1_CS0", "UART2_RTS", "GRXD3"},
	"PA1":  {"ERXD2", "SPI1_CLK", "UART2_CTS", "GRXD2"},
	"PA2":  {"ERXD1", "SPI1_MOSI", "UART2_TX", "GRXD1"},
	"PA3":  {"ERXD0", "SPI1_MISO", "UART2_RX", "GRXD0"},
	"PA4":  {"ETXD3", "SPI1_CS1", "", "GTXD3"},
	"PA5":  {"ETXD2", "SPI3_CS0", "", "GTXD2"},
	"PA6":  {"ETXD1", "SPI3_CLK", "", "GTXD1"},
	"PA7":  {"ETXD0", "SPI3_MOSI", "", "GTXD0"},
	"PA8":  {"ERXCK", "SPI3_MISO", "", "CRXCK"},
	"PA9":  {"ERXERR", "SPI3_CS1", "", "GNULL", "I2S1_MCLK"},
	"PA10": {"ERXDV", "", "UART1_TX", "GRXCTL"},
	"PA11": {"EMDC", "", "UART1_RX", "GMDC"},
	"PA12": {"EMDIO", "UART6_TX", "UART1_RTS", "GMDIO"},
	"PA13": {"ETXEN", "UART6_RX", "UART1_CTS", "GTXCTL"},
	"PA14": {"ETXCK", "UART7_TX", "UART1_DTR", "GNULL", "I2S1_SCK"},
	"PA15": {"ECRS", "UART7_RX", "UART1_DSR", "GTXCK", "I2S1_WS"},
	"PA16": {"ECOL", "CAN_TX", "UART1_DCD", "GCLKIN", "I2S1_DOUT"},
	"PA17": {"ETXERR", "CAN_RX", "UART1_RI", "GNULL", "I2S1_DIN"},
	"PB0":  {"I2C0_SCL"},
	"PB1":  {"I2C0_SDA"},
	"PB2":  {"PWM0"},
	"PB3":  {"IR0_TX", "", "SPDIF_MCLK", "", "STANBYWFI"},
	"PB4":  {"IR0_RX"},
	"PB5":  {"I2S0_MCLK", "AC97_MCLK"},
	"PB6":  {"I2S0_SCK", "AC97_SCK"},
	"PB7":  {"I2S0_WS", "AC97_SYNC"},
	"PB8":  {"I2S0_DOUT0", "AC97_DOUT"},
	"PB9":  {"I2S0_DOUT1"},
	"PB10": {"I2S0_DOUT2"},
	"PB11": {"I2S0_DOUT3"},
	"PB12": {"I2S0_DIN", "AC97_DI", "SPDIF_DI"},
	"PB13": {"SPI2_CS1", "", "SPDIF_DO"},
	"PB14": {"SPI2_CS0", "JTAG0_TMS"},
	"PB15": {"SPI2_CLK", "JTAG0_TCK"},
	"PB16": {"SPI2_MOSI", "JTAG0_TDO"},
	"PB17": {"SPI2_MISO", "JTAG0_TDI"},
	"PB18": {"I2C1_SCL"},
	"PB19": {"I2C1_SDA"},
	"PB20": {"I2C2_SCL"},
	"PB21": {"I2C2_SDA"},
	"PB22": {"UART0_TX", "IR1_TX"},
	"PB23": {"UART0_RX", "IR1_RX"},
	"PC0":  {"NWE#", "SPI0_MOSI"},
	"PC1":  {"NALE", "SPI0_MISO"},
	"PC2":  {"NCLE", "SPI0_CLK"},
	"PC3":  {"NCE1"},
	"PC4":  {"NCE0"},
	"PC5":  {"NRE#"},
	"PC6":  {"NRB0", "SDC2_CMD"},
	"PC7":  {"NRB1", "SDC2_CLK"},
	"PC8":  {"NDQ0", "SDC2_D0"},
	"PC9":  {"NDQ1", "SDC2_D1"},
	"PC10": {"NDQ2", "SDC2_D2"},
	"PC11": {"NDQ3", "SDC2_D3"},
	"PC12": {"NDQ4"},
	"PC13": {"NDQ5"},
	"PC14": {"NDQ6"},
	"PC15": {"NDQ7"},
	"PC16": {"NWP"},
	"PC17": {"NCE2"},
	"PC18": {"NCE3"},
	"PC19": {"NCE4", "SPI2_CS0", "", "", "PC_EINT12"},
	"PC20": {"NCE5", "SPI2_CLK", "", "", "PC_EINT13"},
	"PC21": {"NCE6", "SPI2_MOSI", "", "", "PC_EINT14"},
	"PC22": {"NCE7", "SPI2_MISO", "", "", "PC_EINT15"},
	"PC23": {"", "SPI2_CS0"},
	"PC24": {"NDQS"},
	"PD0":  {"LCD0_D0", "LVDS0_VP0"},
	"PD1":  {"LCD0_D1", "LVDS0_VN0"},
	"PD2":  {"LCD0_D2", "LVDS0_VP1"},
	"PD3":  {"LCD0_D3", "LVDS0_VN1"},
	"PD4":  {"LCD0_D4", "LVDS0_VP2"},
	"PD5":  {"LCD0_D5", "LVDS0_VN2"},
	"PD6":  {"LCD0_D6", "LVDS0_VPC"},
	"PD7":  {"LCD0_D7", "LVDS0_VNC"},
	"PD8":  {"LCD0_D8", "LVDS0_VP3"},
	"PD9":  {"LCD0_D9", "LVDS0_VN3"},
	"PD10": {"LCD0_D10", "LVDS1_VP0"},
	"PD11": {"LCD0_D11", "LVDS1_VN0"},
	"PD12": {"LCD0_D12", "LVDS1_VP1"},
	"PD13": {"LCD0_D13", "LVDS1_VN1"},
	"PD14": {"LCD0_D14", "LVDS1_VP2"},
	"PD15": {"LCD0_D15", "LVDS1_VN2"},
	"PD16": {"LCD0_D16", "LVDS1_VPC"},
	"PD17": {"LCD0_D17", "LVDS1_VNC"},
	"PD18": {"LCD0_D18", "LVDS1_VP3"},
	"PD19": {"LCD0_D19", "LVDS1_VN3"},
	"PD20": {"LCD0_D20", "CSI1_MCLK"},
	"PD21": {"LCD0_D21", "SMC_VPPEN"},
	"PD22": {"LCD0_D22", "SMC_VPPPP"},
	"PD23": {"LCD0_D23", "SMC_DET"},
	"PD24": {"LCD0_CLK", "SMC_VCCEN"},
	"PD25": {"LCD0_DE", "SMC_RST"},
	"PD26": {"LCD0_HSYNC", "SMC_SLK"},
	"PD27": {"LCD0_VSYNC", "SMC_SDA"},
	"PE0":  {"TS0_CLK", "CSI0_PCLK"},
	"PE1":  {"TS0_ERR", "CSI0_MCLK"},
	"PE2":  {"TS0_SYNC", "CSI0_HSYNC"},
	"PE3":  {"TS0_DLVD", "CSI0_VSYNC"},
	"PE4":  {"TS0_D0", "CSI0_D0"},
	"PE5":  {"TS0_D1", "CSI0_D1"},
	"PE6":  {"TS0_D2", "CSI0_D2"},
	"PE7":  {"TS0_D3", "CSI0_D3"},
	"PE8":  {"TS0_D4", "CSI0_D4"},
	"PE9":  {"TS0_D5", "CSI0_D5"},
	"PE10": {"TS0_D6", "CSI0_D6"},
	"PE11": {"TS0_D7", "CSI0_D7"},
	"PF0":  {"SDC0_D1", "", "JTAG1_TMS"},
	"PF1":  {"SDC0_D0", "", "JTAG1_TDI"},
	"PF2":  {"SDC0_CLK", "", "UART0_TX"},
	"PF3":  {"SDC0_CMD", "", "JTAG1_TDO"},
	"PF4":  {"SDC0_D3", "", "UART0_RX"},
	"PF5":  {"SDC0_D2", "", "JTAG1_TCK"},
	"PG0":  {"TS1_CLK", "CSI1_PCLK", "SDC1_CMD"},
	"PG1":  {"TS1_ERR", "CSI1_MCLK", "SDC1_CLK"},
	"PG2":  {"TS1_SYNC", "CSI1_HSYNC", "SDC1_D0"},
	"PG3":  {"TS1_DVLD", "CSI1_VSYNC", "SDC1_D1"},
	"PG4":  {"TS1_D0", "CSI1_D0", "SDC1_D2", "CSI0_D8"},
	"PG5":  {"TS1_D1", "CSI1_D1", "SDC1_D3", "CSI0_D9"},
	"PG6":  {"TS1_D2", "CSI1_D2", "UART3_TX", "CSI0_D10"},
	"PG7":  {"TS1_D3", "CSI1_D3", "UART3_RX", "CSI0_D11"},
	"PG8":  {"TS1_D4", "CSI1_D4", "UART3_RTS", "CSI0_D12"},
	"PG9":  {"TS1_D5", "CSI1_D4", "UART3_CTS", "CSI0_D13"},
	"PG10": {"TS1_D6", "CSI1_D6", "UART4_TX", "CSI0_D14"},
	"PG11": {"TS1_D7", "CSI1_D7", "UART4_RX", "CSI0_D15"},
	"PH0":  {"LCD1_D0", "", "UART3_TX", "", "PH_EINT0"},
	"PH1":  {"LCD1_D1", "", "UART3_RX", "", "PH_EINT1"},
	"PH2":  {"LCD1_D2", "", "UART3_RTS", "", "PH_EINT2"},
	"PH3":  {"LCD1_D3", "", "UART3_CTS", "", "PH_EINT3"},
	"PH4":  {"LCD1_D4", "", "UART4_TX", "", "PH_EINT4"},
	"PH5":  {"LCD1_D5", "", "UART4_RX", "", "PH_EINT5"},
	"PH6":  {"LCD1_D6", "", "UART5_TX", "MS_BS", "PH_EINT6"},
	"PH7":  {"LCD1_D7", "", "UART5_RX", "MS_CLK", "PH_EINT7"},
	"PH8":  {"LCD1_D8", "ERXD3", "KP_IN0", "MS_D0", "PH_EINT8"},
	"PH9":  {"LCD1_D9", "ERXD2", "KP_IN1", "MS_D1", "PH_EINT9"},
	"PH10": {"LCD1_D10", "ERXD1", "KP_IN2", "MS_D2", "PH_EINT10"},
	"PH11": {"LCD1_D11", "ERXD0", "KP_IN3", "MS_D3", "PH_EINT11"},
	"PH12": {"LCD1_D12", "", "PS2_SCK1", "", "PH_EINT12"},
	"PH13": {"LCD1_D13", "", "PS2_SDA1", "SMC_RST", "PH_EINT13"},
	"PH14": {"LCD1_D14", "ETXD3", "KP_IN4", "SMC_VPPEN", "PH_EINT14"},
	"PH15": {"LCD1_D15", "ETXD2", "KP_IN5", "SMC_VPPPP", "PH_EINT15"},
	"PH16": {"LCD1_D16", "ETXD1", "KP_IN6", "SMC_DET", "PH_EINT16"},
	"PH17": {"LCD1_D17", "ETXD0", "KP_IN7", "SMC_VCCEN", "PH_EINT17"},
	"PH18": {"LCD1_D18", "ERXCK", "KP_OUT0", "SMC_SLK", "PH_EINT18"},
	"PH19": {"LCD1_D19", "ERXERR", "KP_OUT1", "SMC_SDA", "PH_EINT19"},
	"PH20": {"LCD1_D20", "ERXDV", "CAN_TX", "", "PH_EINT20"},
	"PH21": {"LCD1_D21", "EMDC", "CAN_RX", "", "PH_EINT21"},
	"PH22": {"LCD1_D22", "EMDIO", "KP_OUT2", "SDC1_CMD", ""},
	"PH23": {"LCD1_D23", "ETXEN", "KP_OUT3", "SDC1_CLK", ""},
	"PH24": {"LCD1_CLK", "ETXCK", "KP_OUT4", "SDC1_D0", ""},
	"PH25": {"LCD1_DE", "ECRS", "KP_OUT5", "SDC1_D1", ""},
	"PH26": {"LCD1_HSYNC", "ECOL", "KP_OUT6", "SDC1_D2", ""},
	"PH27": {"LCD1_VSYNC", "ETXERR", "KP_OUT7", "SDC1_D3", ""},
	"PI0":  {"", "I2C3_SCL"},
	"PI1":  {"", "I2C3_SDA"},
	"PI2":  {"", "I2C4_SCL"},
	"PI3":  {"PWM1", "I2C4_SDA"},
	"PI4":  {"SDC3_CMD"},
	"PI5":  {"SDC3_CLK"},
	"PI6":  {"SDC3_D0"},
	"PI7":  {"SDC3_D1"},
	"PI8":  {"SDC3_D2"},
	"PI9":  {"SDC3_D3"},
	"PI10": {"SPI0_CS0", "UART5_TX", "", "PI_EINT22"},
	"PI11": {"SPI0_CLK", "UART5_RX", "", "PI_EINT23"},
	"PI12": {"SPI0_MOSI", "UART6_TX", "CLK_OUT_A", "PI_EINT24"},
	"PI13": {"SPI0_MISO", "UART6_RX", "CLK_OUT_B", "PI_EINT25"},
	"PI14": {"SPI0_CS0", "PS2_SCK1", "TCLKIN0", "PI_EINT26"},
	"PI15": {"SPI1_CS1", "PS2_SDA1", "TCLKIN1", "PI_EINT27"},
	"PI16": {"SPI1_CS0", "UART2_RTS", "", "PI_EINT28"},
	"PI17": {"SPI1_CLK", "UART2_CTS", "", "PI_EINT29"},
	"PI18": {"SPI1_MOSI", "UART2_TX", "", "PI_EINT30"},
	"PI19": {"SPI1_MISO", "UART2_RX", "", "PI_EINT31"},
	"PI20": {"PS2_SCK0", "UART7_TX", "HSCL"},
	"PI21": {"PS2_SDA0", "UART7_RX", "HSDA"},
}

// mapA20Pins uses mappingA20 to actually set the altFunc fields of all gpio
// and mark them as available.
//
// It is called by the generic allwinner processor code if an A20 is detected.
func mapA20Pins() error {
	for name, altFuncs := range mappingA20 {
		pin := cpupins[name]
		pin.altFunc = altFuncs
		pin.available = true
		if strings.Contains(string(altFuncs[4]), "_EINT") ||
			strings.Contains(string(altFuncs[3]), "_EINT") {
			pin.supportEdge = true
		}

		// Initializes the sysfs corresponding pin right away.
		pin.sysfsPin = sysfs.Pins[pin.Number()]
	}
	return nil
}
