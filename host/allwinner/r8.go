// Copyright 2016 Thorsten von Eicken. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file contains pin mapping information that is specific to the Allwinner R8 model.

package allwinner

import (
	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/conn/pins"
)

// pins that are in the R8 datasheet but not generic (?)
var (
	FEL      pins.Pin   = &pins.BasicPin{Name: "FEL"}
	MIC_IN   gpio.PinIO = &gpio.BasicPin{Name: "MIX_IN"}
	MIC_GND  pins.Pin   = &pins.BasicPin{Name: "MIC_GND"}
	HP_LEFT  gpio.PinIO = &gpio.BasicPin{Name: "HP_LEFT"}
	HP_RIGHT gpio.PinIO = &gpio.BasicPin{Name: "HP_RIGHT"}
	HP_COM   pins.Pin   = &pins.BasicPin{Name: "HP_COM"}

	X1 gpio.PinIO = &gpio.BasicPin{Name: "X1"} // touch screen
	X2 gpio.PinIO = &gpio.BasicPin{Name: "X2"} // touch screen
	Y1 gpio.PinIO = &gpio.BasicPin{Name: "Y1"} // touch screen
	Y2 gpio.PinIO = &gpio.BasicPin{Name: "Y2"} // touch screen
)

// mappingR8 for Allwinner R8.
// Datasheet, page 18:
// https://github.com/NextThingCo/CHIP-Hardware/raw/master/CHIP%5Bv1_0%5D/CHIPv1_0-BOM-Datasheets/Allwinner%20R8%20Datasheet%20V1.2.pdf
//
// - The datasheet uses TWI instead of I2C but this is renamed here for consistency.
var mappingR8 = map[string][5]string{
	"PB0":  {"I2C0_SCL"},
	"PB1":  {"I2C0_SDA"},
	"PB2":  {"PWM", "", "", "", "EINT16"},
	"PB3":  {"IR_TX", "", "", "", "EINT17"},
	"PB4":  {"IR_RX", "", "", "", "EINT18"},
	"PB10": {"SPI2_CS1"},
	"PB15": {"I2C1_SCL"},
	"PB16": {"I2C1_SDA"},
	"PB17": {"I2C2_SCL"},
	"PB18": {"I2C2_SDA"},
	"PC0":  {"NAND_WE", "SPI0_MOSI"},
	"PC1":  {"NAND_ALE", "SPI0_MISO"},
	"PC2":  {"NAND_CLE", "SPI0_CLK"},
	"PC3":  {"NAND_CE1", "SPI0_CS0"},
	"PC4":  {"NAND_CE0"},
	"PC5":  {"NAND_RE"},
	"PC6":  {"NAND_RB0", "SDC2_CMD"},
	"PC7":  {"NAND_RB1", "SDC2_CLK"},
	"PC8":  {"NAND_DQ0", "SDC2_D0"},
	"PC9":  {"NAND_DQ1", "SDC2_D1"},
	"PC10": {"NAND_DQ2", "SDC2_D2"},
	"PC11": {"NAND_DQ3", "SDC2_D3"},
	"PC12": {"NAND_DQ4", "SDC2_D4"},
	"PC13": {"NAND_DQ5", "SDC2_D5"},
	"PC14": {"NAND_DQ6", "SDC2_D6"},
	"PC15": {"NAND_DQ7", "SDC2_D7"},
	"PD2":  {"LCD_D2", "UART2_TX"},
	"PD3":  {"LCD_D3", "UART2_RX"},
	"PD4":  {"LCD_D4", "UART2_CTX"},
	"PD5":  {"LCD_D5", "UART2_RTS"},
	"PD6":  {"LCD_D6", "ECRS"},
	"PD7":  {"LCD_D7", "ECOL"},
	"PD10": {"LCD_D10", "ERXD0"},
	"PD11": {"LCD_D11", "ERXD1"},
	"PD12": {"LCD_D12", "ERXD2"},
	"PD13": {"LCD_D13", "ERXD3"},
	"PD14": {"LCD_D14", "ERXCK"},
	"PD15": {"LCD_D15", "ERXERR"},
	"PD18": {"LCD_D18", "ERXDV"},
	"PD19": {"LCD_D19", "ETXD0"},
	"PD20": {"LCD_D20", "ETXD1"},
	"PD21": {"LCD_D21", "ETXD2"},
	"PD22": {"LCD_D22", "ETXD3"},
	"PD23": {"LCD_D23", "ETXEN"},
	"PD24": {"LCD_CLK", "ETXCK"},
	"PD25": {"LCD_DE", "ETXERR"},
	"PD26": {"LCD_HSYNC", "EMDC"},
	"PD27": {"LCD_VSYNC", "EMDIO"},
	"PE0":  {"TS_CLK", "CSI_PCLK", "SPI2_CS0", "", "EINT14"},
	"PE1":  {"TS_ERR", "CSI_MCLK", "SPI2_CLK", "", "EINT15"},
	"PE2":  {"TS_SYNC", "CSI_HSYNC", "SPI2_MOSI"},
	"PE3":  {"TS_DVLD", "CSI_VSYNC", "SPI2_MISO"},
	"PE4":  {"TS_D0", "CSI_D0", "SDC2_D0"},
	"PE5":  {"TS_D1", "CSI_D1", "SDC2_D1"},
	"PE6":  {"TS_D2", "CSI_D2", "SDC2_D2"},
	"PE7":  {"TS_D3", "CSI_D3", "SDC2_D3"},
	"PE8":  {"TS_D4", "CSI_D4", "SDC2_CMD"},
	"PE9":  {"TS_D5", "CSI_D5", "SDC2_CLK"},
	"PE10": {"TS_D6", "CSI_D6", "UART1_TX"},
	"PE11": {"TS_D7", "CSI_D7", "UART1_RX"},
	"PF0":  {"SDC0_D1", "", "JTAG_MS1"},
	"PF1":  {"SDC0_D0", "", "JTAG_DI1"},
	"PF2":  {"SDC0_CLK", "", "UART0_TX"},
	"PF3":  {"SDC0_CMD", "", "JTAG_DO1"},
	"PF4":  {"SDC0_D3", "", "UART0_RX"},
	"PF5":  {"SDC0_D2", "", "JTAG_CK1"},
	"PG0":  {"GPS_CLK", "", "", "", "EINT0"},
	"PG1":  {"GPS_SIGN", "", "", "", "EINT1"},
	"PG2":  {"GPS_MAG", "", "", "", "EINT2"},
	"PG3":  {"", "", "UART1_TX", "", "EINT3"},
	"PG4":  {"", "", "UART1_RX", "", "EINT4"},
	"PG9":  {"SPI1_CS0", "UART3_TX", "", "", "PG_EINT9"},
	"PG10": {"SPI1_CLK", "UART3_RX", "", "", "PG_EINT10"},
	"PG11": {"SPI1_MOSI", "UART3_CTS", "", "", "PG_EINT11"},
	"PG12": {"SPI1_MISO", "UART3_RTS", "", "", "PG_EINT12"},
}

func mapR8Pins() {
	// set the altFunc fields of all pins that are on the R8
	for name, altFuncs := range mappingR8 {
		if pin := pinByName(name); pin != nil {
			pin.altFunc = altFuncs
		}
	}
}
